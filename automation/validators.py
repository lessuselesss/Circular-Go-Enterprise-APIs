#!/usr/bin/env python3
"""
TDD Automation Validators
Ensures proper task dependencies and git state before execution
"""

import subprocess
import json
import re
from pathlib import Path
from typing import Dict, List, Optional, Tuple
from dataclasses import dataclass

from functional_tdd import run_taskmaster


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


@dataclass
class ValidationResult:
    """Result of validation checks."""
    valid: bool
    errors: List[str]
    warnings: List[str]
    blocking_tasks: List[str] = None
    conflicting_branches: List[str] = None
    
    def __post_init__(self):
        if self.blocking_tasks is None:
            self.blocking_tasks = []
        if self.conflicting_branches is None:
            self.conflicting_branches = []


def validate_task_dependencies(task_id: str, project_path: Path) -> ValidationResult:
    """
    Validate that all dependencies for a task are completed.
    
    Returns ValidationResult with blocking tasks if any.
    """
    errors = []
    warnings = []
    blocking_tasks = []
    
    try:
        # Get task details from Task Master
        task_output = run_taskmaster(f"show {task_id}", project_path)
        
        # Also validate dependencies using Task Master's built-in validator
        deps_output = run_taskmaster("validate-dependencies", project_path)
        
        # Get current task status
        if "✓ done" in task_output:
            warnings.append(f"Task {task_id} is already marked as done")
        elif "○ pending" not in task_output and "► in-progress" not in task_output:
            errors.append(f"Task {task_id} has invalid status")
        
        # Extract dependencies from task output
        dependencies = _extract_dependencies(task_output)
        
        if dependencies:
            print(f"🔍 Checking dependencies for task {task_id}: {dependencies}")
            
            # Check each dependency
            for dep_id in dependencies:
                dep_output = run_taskmaster(f"show {dep_id}", project_path)
                
                if "✓ done" not in dep_output:
                    blocking_tasks.append(dep_id)
                    if "○ pending" in dep_output:
                        errors.append(f"Dependency {dep_id} is still pending")
                    elif "► in-progress" in dep_output:
                        errors.append(f"Dependency {dep_id} is still in progress")
                    else:
                        errors.append(f"Dependency {dep_id} has unknown status")
        
        # Check Task Master's dependency validation
        if "invalid dependencies" in deps_output.lower() or "dependency issues" in deps_output.lower():
            errors.append("Task Master reports dependency issues")
        
        return ValidationResult(
            valid=len(errors) == 0,
            errors=errors,
            warnings=warnings,
            blocking_tasks=blocking_tasks
        )
        
    except Exception as e:
        return ValidationResult(
            valid=False,
            errors=[f"Failed to validate dependencies: {str(e)}"],
            warnings=[]
        )


def validate_git_state(task_id: str, project_path: Path) -> ValidationResult:
    """
    Validate git state is clean for task execution.
    
    Checks for:
    - No existing task branches
    - Clean working directory
    - On main/master branch
    """
    errors = []
    warnings = []
    conflicting_branches = []
    
    try:
        # Check current branch
        current_branch_result = run_git_command("branch --show-current", project_path)
        if not current_branch_result["success"]:
            errors.append("Failed to get current git branch")
            return ValidationResult(valid=False, errors=errors, warnings=warnings)
        
        current_branch = current_branch_result["stdout"]
        
        # Should be on main/master
        if current_branch not in ["main", "master"]:
            if current_branch.startswith("task-"):
                errors.append(f"Already on task branch: {current_branch}. Complete or cleanup first.")
            else:
                warnings.append(f"Not on main branch (currently on: {current_branch})")
        
        # Check for existing task branches
        branches_result = run_git_command("branch", project_path)
        if branches_result["success"]:
            branches = branches_result["stdout"].split('\n')
            
            for branch in branches:
                branch = branch.strip().replace('*', '').strip()
                if branch.startswith('task-'):
                    # Extract task ID from branch name
                    branch_task_id = branch.replace('task-', '')
                    
                    # Check if this conflicts with our task or its dependencies
                    if branch_task_id == task_id:
                        errors.append(f"Branch 'task-{task_id}' already exists")
                        conflicting_branches.append(branch)
                    elif _is_blocking_branch(branch_task_id, task_id, project_path):
                        errors.append(f"Blocking task branch exists: {branch}")
                        conflicting_branches.append(branch)
        
        # Check working directory is clean
        status_result = run_git_command("status --porcelain", project_path)
        if status_result["success"] and status_result["stdout"]:
            # Has uncommitted changes
            warnings.append("Working directory has uncommitted changes")
        
        # Check for unmerged branches that might conflict
        unmerged_result = run_git_command("branch --no-merged main", project_path)
        if unmerged_result["success"] and unmerged_result["stdout"]:
            unmerged_branches = [b.strip().replace('*', '').strip() 
                               for b in unmerged_result["stdout"].split('\n') if b.strip()]
            
            for branch in unmerged_branches:
                if branch.startswith('task-'):
                    warnings.append(f"Unmerged task branch exists: {branch}")
        
        return ValidationResult(
            valid=len(errors) == 0,
            errors=errors,
            warnings=warnings,
            conflicting_branches=conflicting_branches
        )
        
    except Exception as e:
        return ValidationResult(
            valid=False,
            errors=[f"Failed to validate git state: {str(e)}"],
            warnings=[]
        )


def validate_task_order(task_id: str, project_path: Path) -> ValidationResult:
    """
    Validate task is being executed in proper order.
    
    Ensures we're not skipping ahead or working on tasks out of sequence.
    """
    errors = []
    warnings = []
    
    try:
        # Get all tasks to check ordering
        task_list = run_taskmaster("list", project_path)
        
        # Extract task information
        task_info = _parse_task_list(task_list)
        
        # Find our task
        current_task = None
        for task in task_info:
            if task["id"] == task_id:
                current_task = task
                break
        
        if not current_task:
            errors.append(f"Task {task_id} not found in task list")
            return ValidationResult(valid=False, errors=errors, warnings=warnings)
        
        # Check if there are pending tasks that should be done first
        pending_tasks = [t for t in task_info if t["status"] == "pending"]
        
        for pending_task in pending_tasks:
            # If there's a pending task with lower ID, we might be skipping
            if _is_prerequisite_task(pending_task["id"], task_id):
                warnings.append(f"Skipping earlier task {pending_task['id']} which is still pending")
        
        # Check if task is ready based on dependencies
        if current_task["dependencies"]:
            for dep_id in current_task["dependencies"]:
                dep_task = next((t for t in task_info if t["id"] == dep_id), None)
                if dep_task and dep_task["status"] != "done":
                    errors.append(f"Cannot start {task_id}: dependency {dep_id} not complete")
        
        return ValidationResult(
            valid=len(errors) == 0,
            errors=errors,
            warnings=warnings
        )
        
    except Exception as e:
        return ValidationResult(
            valid=False,
            errors=[f"Failed to validate task order: {str(e)}"],
            warnings=[]
        )


def validate_all_preconditions(task_id: str, project_path: Path) -> ValidationResult:
    """
    Run all validation checks before allowing task execution.
    
    Returns combined validation result.
    """
    print(f"🔒 Validating preconditions for task {task_id}")
    
    errors = []
    warnings = []
    conflicting_branches = []
    
    # 1. Simple check: does task exist and is it available?
    try:
        task_output = run_taskmaster(f"show {task_id}", project_path)
        
        if "not found" in task_output.lower() or "error" in task_output.lower():
            errors.append(f"Task {task_id} not found")
        elif "✓ done" in task_output:
            errors.append(f"Task {task_id} is already completed")
        elif "blocked" in task_output.lower():
            errors.append(f"Task {task_id} is blocked")
            
    except Exception as e:
        errors.append(f"Failed to check task status: {e}")
    
    # 2. Check Task Master's dependency validation
    try:
        deps_output = run_taskmaster("validate-dependencies", project_path)
        if "no invalid dependencies found" not in deps_output.lower() and "all dependencies are valid" not in deps_output.lower():
            errors.append("Task Master reports dependency issues")
    except Exception as e:
        warnings.append(f"Could not validate dependencies: {e}")
    
    # 3. Check git state for conflicts
    try:
        # Check for existing task branches
        branches_result = run_git_command("branch", project_path)
        if branches_result["success"]:
            branches = branches_result["stdout"].split('\n')
            
            for branch in branches:
                branch = branch.strip().replace('*', '').strip()
                if branch == f"task-{task_id}":
                    errors.append(f"Branch 'task-{task_id}' already exists")
                    conflicting_branches.append(branch)
        
        # Check current branch
        current_branch_result = run_git_command("branch --show-current", project_path)
        if current_branch_result["success"]:
            current_branch = current_branch_result["stdout"]
            if current_branch.startswith("task-") and current_branch != f"task-{task_id}":
                warnings.append(f"Currently on task branch: {current_branch}")
    
    except Exception as e:
        warnings.append(f"Could not validate git state: {e}")
    
    # 4. Check if task is actually next available
    try:
        next_output = run_taskmaster("next", project_path)
        if task_id not in next_output:
            warnings.append(f"Task {task_id} may not be the next recommended task")
    except Exception as e:
        warnings.append(f"Could not check next task: {e}")
    
    combined_result = ValidationResult(
        valid=len(errors) == 0,
        errors=errors,
        warnings=warnings,
        conflicting_branches=conflicting_branches
    )
    
    # Print results
    if combined_result.valid:
        print("✅ All preconditions satisfied")
        if warnings:
            print("⚠️  Warnings:")
            for warning in warnings:
                print(f"   - {warning}")
    else:
        print("❌ Validation failed:")
        for error in errors:
            print(f"   - {error}")
        
        if conflicting_branches:
            print(f"🌿 Clean up these branches first: {', '.join(conflicting_branches)}")
    
    return combined_result


# Helper functions

def _extract_dependencies(task_output: str) -> List[str]:
    """Extract dependency IDs from task output."""
    dependencies = []
    
    # Look for dependency patterns in task output
    # Format might be "Dependencies: 1, 2.1" or similar
    lines = task_output.split('\n')
    for line in lines:
        if 'dependencies' in line.lower() or 'depends' in line.lower():
            # Extract task IDs (numbers with optional dots)
            task_ids = re.findall(r'\b\d+(?:\.\d+)*\b', line)
            dependencies.extend(task_ids)
    
    # Remove duplicates and our own task ID
    return list(set(dependencies))


def _is_blocking_branch(branch_task_id: str, current_task_id: str, project_path: Path) -> bool:
    """Check if a branch task ID blocks the current task."""
    # Simple heuristic: if branch task is a dependency or has lower ID
    try:
        current_task_info = run_taskmaster(f"show {current_task_id}", project_path)
        dependencies = _extract_dependencies(current_task_info)
        
        return branch_task_id in dependencies
    except:
        return False


def _parse_task_list(task_list_output: str) -> List[Dict]:
    """Parse task list output into structured data."""
    tasks = []
    
    # This is a simplified parser - would need to be more robust for production
    lines = task_list_output.split('\n')
    
    for line in lines:
        if '│' in line and any(status in line for status in ['✓ done', '○ pending', '► in-progress']):
            # Extract task info from table row
            parts = [part.strip() for part in line.split('│') if part.strip()]
            
            if len(parts) >= 4:
                task_id = parts[0].strip()
                status = 'done' if '✓' in parts[2] else 'pending' if '○' in parts[2] else 'in-progress'
                
                # Extract dependencies if present
                deps = []
                if len(parts) > 4:
                    dep_text = parts[4]
                    if 'None' not in dep_text:
                        deps = re.findall(r'\b\d+(?:\.\d+)*\b', dep_text)
                
                tasks.append({
                    'id': task_id,
                    'status': status,
                    'dependencies': deps
                })
    
    return tasks


def _is_prerequisite_task(potential_prereq: str, current_task: str) -> bool:
    """Check if potential_prereq should be done before current_task based on ID ordering."""
    # Simple numeric comparison for now
    try:
        prereq_num = float(potential_prereq.replace('.', ''))
        current_num = float(current_task.replace('.', ''))
        return prereq_num < current_num
    except:
        return False


if __name__ == "__main__":
    import sys
    
    if len(sys.argv) < 3:
        print("Usage: python validators.py <task_id> <project_path>")
        sys.exit(1)
    
    task_id = sys.argv[1]
    project_path = Path(sys.argv[2])
    
    result = validate_all_preconditions(task_id, project_path)
    
    if result.valid:
        print(f"\n✅ Task {task_id} is ready for execution")
        sys.exit(0)
    else:
        print(f"\n❌ Task {task_id} cannot be executed yet")
        sys.exit(1)