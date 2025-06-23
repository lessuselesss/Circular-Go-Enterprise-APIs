#!/usr/bin/env python3
"""
Session-Based TDD Automation with Git Integration
Orchestrator manages git operations, session runs until completion or failure
"""

import subprocess
import json
import time
from pathlib import Path
from typing import Dict, List, Optional
from dataclasses import dataclass
from enum import Enum

# Import our functional agents
from functional_tdd import (
    ActionType, AgentResult, TaskContext,
    tester_function, planner_function, researcher_function,
    run_taskmaster, run_go_test
)


class SessionResult(Enum):
    COMPLETED = "completed"
    BLOCKED = "blocked" 
    NO_MORE_TASKS = "no_more_tasks"
    ERROR = "error"


@dataclass
class GitContext:
    """Git state for the session."""
    base_branch: str = "main"
    current_branch: str = "main"
    task_branches: Dict[str, str] = None
    commits: List[str] = None
    
    def __post_init__(self):
        if self.task_branches is None:
            self.task_branches = {}
        if self.commits is None:
            self.commits = []


def run_git_command(command: str, project_path: Path) -> Dict:
    """Run git command and return result."""
    try:
        full_cmd = f'nix develop --command bash -c "git {command}"'
        result = subprocess.run(
            full_cmd,
            shell=True,
            cwd=project_path,
            capture_output=True,
            text=True
        )
        return {
            "success": result.returncode == 0,
            "stdout": result.stdout.strip(),
            "stderr": result.stderr.strip(),
            "exit_code": result.returncode
        }
    except Exception as e:
        return {
            "success": False,
            "stdout": "",
            "stderr": str(e),
            "exit_code": -1
        }


def orchestrator_git_function(context: TaskContext, git_context: GitContext, 
                             action: str, **kwargs) -> AgentResult:
    """
    Enhanced orchestrator that handles git operations.
    
    Git actions: create_branch, commit_phase, tag_completion, merge_task
    """
    print(f"🎯 ORCHESTRATOR: Git operation '{action}' for {context.task_id}")
    
    if action == "create_branch":
        # Create feature branch for task
        branch_name = f"task-{context.task_id}"
        
        # Switch to base branch first
        run_git_command(f"checkout {git_context.base_branch}", context.project_path)
        run_git_command("pull origin main", context.project_path)
        
        # Create and switch to task branch
        git_result = run_git_command(f"checkout -b {branch_name}", context.project_path)
        
        if git_result["success"]:
            git_context.current_branch = branch_name
            git_context.task_branches[context.task_id] = branch_name
            
            # Update Task Master
            run_taskmaster(f'update-subtask --id={context.task_id} --prompt="Created branch: {branch_name}"', context.project_path)
            
            return AgentResult(
                action=ActionType.SUCCESS,
                data={"branch": branch_name},
                message=f"Created branch {branch_name} for task {context.task_id}"
            )
        else:
            return AgentResult(
                action=ActionType.FAIL,
                data={"git_error": git_result["stderr"]},
                message=f"Failed to create branch: {git_result['stderr']}"
            )
    
    elif action == "commit_phase":
        # Commit after TDD phase (RED/GREEN/REFACTOR)
        phase = kwargs.get("phase", "implementation")
        
        # Add all changes
        run_git_command("add .", context.project_path)
        
        # Commit with descriptive message
        commit_msg = f"{phase}: task {context.task_id}\n\n🤖 Generated with TDD automation"
        git_result = run_git_command(f'commit -m "{commit_msg}"', context.project_path)
        
        if git_result["success"]:
            commit_hash = run_git_command("rev-parse HEAD", context.project_path)["stdout"][:8]
            git_context.commits.append(commit_hash)
            
            return AgentResult(
                action=ActionType.SUCCESS,
                data={"commit": commit_hash, "phase": phase},
                message=f"Committed {phase} phase: {commit_hash}"
            )
        else:
            # No changes to commit is OK
            if "nothing to commit" in git_result["stdout"]:
                return AgentResult(
                    action=ActionType.SUCCESS,
                    data={"no_changes": True},
                    message="No changes to commit"
                )
            else:
                return AgentResult(
                    action=ActionType.FAIL,
                    data={"git_error": git_result["stderr"]},
                    message=f"Commit failed: {git_result['stderr']}"
                )
    
    elif action == "tag_completion":
        # Tag successful task completion
        tag_name = f"task-{context.task_id}-complete"
        
        git_result = run_git_command(f'tag -a {tag_name} -m "Task {context.task_id} completed with TDD"', context.project_path)
        
        if git_result["success"]:
            return AgentResult(
                action=ActionType.SUCCESS,
                data={"tag": tag_name},
                message=f"Tagged completion: {tag_name}"
            )
        else:
            return AgentResult(
                action=ActionType.FAIL,
                data={"git_error": git_result["stderr"]},
                message=f"Tagging failed: {git_result['stderr']}"
            )
    
    elif action == "merge_task":
        # Merge completed task back to main
        task_branch = git_context.task_branches.get(context.task_id)
        
        if not task_branch:
            return AgentResult(
                action=ActionType.FAIL,
                data={"error": "No branch found for task"},
                message=f"No branch found for task {context.task_id}"
            )
        
        # Switch to main and merge
        run_git_command(f"checkout {git_context.base_branch}", context.project_path)
        git_result = run_git_command(f"merge {task_branch} --no-ff -m 'Merge task {context.task_id}'", context.project_path)
        
        if git_result["success"]:
            # Clean up branch
            run_git_command(f"branch -d {task_branch}", context.project_path)
            git_context.current_branch = git_context.base_branch
            
            return AgentResult(
                action=ActionType.SUCCESS,
                data={"merged": task_branch},
                message=f"Merged and cleaned up {task_branch}"
            )
        else:
            return AgentResult(
                action=ActionType.FAIL,
                data={"git_error": git_result["stderr"]},
                message=f"Merge failed: {git_result['stderr']}"
            )
    
    else:
        return AgentResult(
            action=ActionType.FAIL,
            data={"error": f"Unknown git action: {action}"},
            message=f"Unknown git action: {action}"
        )


def enhanced_tester_function(context: TaskContext, git_context: GitContext) -> AgentResult:
    """Enhanced tester that coordinates with orchestrator for git operations."""
    
    # CRITICAL: Validate all preconditions before starting
    from validators import validate_all_preconditions
    
    validation_result = validate_all_preconditions(context.task_id, context.project_path)
    
    if not validation_result.valid:
        return AgentResult(
            action=ActionType.FAIL,
            data={
                "validation_errors": validation_result.errors,
                "blocking_tasks": validation_result.blocking_tasks,
                "conflicting_branches": validation_result.conflicting_branches
            },
            message=f"Task {context.task_id} failed validation: {', '.join(validation_result.errors)}"
        )
    
    # First, ensure we have a task branch
    if context.task_id not in git_context.task_branches:
        branch_result = orchestrator_git_function(context, git_context, "create_branch")
        if branch_result.action != ActionType.SUCCESS:
            return branch_result
    
    # Execute TDD cycle with git commits per phase
    print(f"🧪 TESTER: Executing TDD cycle for {context.task_id}")
    
    try:
        # Phase 1: RED
        print("🔴 RED Phase")
        # ... TDD RED logic ...
        orchestrator_git_function(context, git_context, "commit_phase", phase="red")
        
        # Phase 2: GREEN  
        print("🟢 GREEN Phase")
        # ... TDD GREEN logic ...
        orchestrator_git_function(context, git_context, "commit_phase", phase="green")
        
        # Phase 3: REFACTOR
        print("🔵 REFACTOR Phase")
        # ... TDD REFACTOR logic ...
        orchestrator_git_function(context, git_context, "commit_phase", phase="refactor")
        
        # Tag completion
        tag_result = orchestrator_git_function(context, git_context, "tag_completion")
        
        # Mark task done in Task Master
        run_taskmaster(f'set-status --id={context.task_id} --status=done', context.project_path)
        
        return AgentResult(
            action=ActionType.SUCCESS,
            data={"tdd_complete": True, "git_managed": True},
            message=f"TDD cycle completed with git management for {context.task_id}"
        )
        
    except Exception as e:
        return AgentResult(
            action=ActionType.FAIL,
            data={"error": str(e)},
            message=f"Enhanced tester failed: {e}"
        )


def get_next_task(project_path: Path) -> Optional[str]:
    """Get next available task from Task Master."""
    try:
        result = run_taskmaster("next", project_path)
        
        # Parse task ID from output (simplified)
        # In practice, would parse structured output
        if "Next Task to Work On:" in result:
            # Extract task ID using regex or parsing
            lines = result.split('\n')
            for line in lines:
                if "ID:" in line and "-" in line:
                    # Extract task ID like "2.1" or "3.2"
                    parts = line.split()
                    for part in parts:
                        if part.replace('.', '').isdigit() and '.' in part:
                            return part
        
        return None
    except:
        return None


def run_development_session(project_path: str) -> Dict:
    """
    Run a complete development session.
    Processes tasks until completion, blockage, or no more tasks.
    """
    project_path = Path(project_path)
    git_context = GitContext()
    
    session_result = {
        "status": SessionResult.COMPLETED,
        "tasks_completed": [],
        "tasks_blocked": [],
        "git_operations": [],
        "session_start": time.time(),
        "session_end": None
    }
    
    print("🚀 Starting Development Session")
    print("=" * 60)
    
    max_tasks = 10  # Prevent runaway sessions
    task_count = 0
    
    while task_count < max_tasks:
        # Get next available task
        next_task = get_next_task(project_path)
        
        if not next_task:
            print("📋 No more tasks available")
            session_result["status"] = SessionResult.NO_MORE_TASKS
            break
        
        print(f"\n🎯 Processing Task: {next_task}")
        task_count += 1
        
        # Create task context
        context = TaskContext(
            task_id=next_task,
            project_path=project_path,
            current_phase="tdd_with_git"
        )
        
        # Execute enhanced workflow with validation
        try:
            print(f"🔒 Validating task {next_task} preconditions...")
            result = enhanced_tester_function(context, git_context)
            
            if result.action == ActionType.SUCCESS:
                print(f"✅ Task {next_task} completed successfully")
                session_result["tasks_completed"].append(next_task)
                
                # Merge task back to main
                merge_result = orchestrator_git_function(context, git_context, "merge_task")
                if merge_result.action == ActionType.SUCCESS:
                    print(f"🔀 Task {next_task} merged to main")
                    
            elif result.action == ActionType.FAIL:
                print(f"❌ Task {next_task} failed: {result.message}")
                session_result["tasks_blocked"].append({
                    "task_id": next_task,
                    "reason": result.message,
                    "data": result.data
                })
                
                # Stop session on failure - human intervention needed
                session_result["status"] = SessionResult.BLOCKED
                break
                
            else:
                print(f"⚠️  Task {next_task} needs escalation: {result.message}")
                # Could implement escalation chain here
                # For now, treat as blocked
                session_result["tasks_blocked"].append({
                    "task_id": next_task,
                    "reason": "escalation_needed",
                    "message": result.message
                })
                
        except Exception as e:
            print(f"💥 Session error on task {next_task}: {e}")
            session_result["status"] = SessionResult.ERROR
            session_result["error"] = str(e)
            break
    
    session_result["session_end"] = time.time()
    session_result["duration"] = session_result["session_end"] - session_result["session_start"]
    
    print(f"\n🏁 Development Session Complete")
    print(f"Status: {session_result['status'].value}")
    print(f"Tasks Completed: {len(session_result['tasks_completed'])}")
    print(f"Tasks Blocked: {len(session_result['tasks_blocked'])}")
    print(f"Duration: {session_result['duration']:.1f}s")
    
    return session_result


if __name__ == "__main__":
    import sys
    
    project_path = sys.argv[1] if len(sys.argv) > 1 else "."
    
    result = run_development_session(project_path)
    
    print(f"\n📊 Final Session Result:")
    print(json.dumps(result, indent=2, default=str))