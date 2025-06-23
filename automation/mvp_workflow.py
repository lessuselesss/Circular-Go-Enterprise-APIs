#!/usr/bin/env python3
"""
MVP TDD Automation Workflow - Simple End-to-End Loop
One orchestrator coordinates: Research → Plan → TDD → Complete
"""

import os
import json
import subprocess
from pathlib import Path
from typing import Dict, Optional

from mcp_agent import MCPApp, Agent


class SimpleTDDWorkflow:
    """MVP: One simple end-to-end TDD automation loop."""
    
    def __init__(self, project_path: str):
        self.project_path = Path(project_path)
        self.app = MCPApp()
        
        # Single orchestrator agent that coordinates everything
        self.orchestrator = Agent(
            name="tdd_orchestrator",
            instruction="""You are a TDD workflow orchestrator. You coordinate a simple loop:
            
            1. Get next task from Task Master
            2. Research the task and store findings  
            3. Plan implementation with approval gates
            4. Execute TDD cycle (test → code → verify)
            5. Mark complete and move to next
            
            Use Task Master MCP tools throughout. Keep it simple and focused.""",
            server_names=["task-master-ai"] if self._has_taskmaster_mcp() else []
        )
    
    def _has_taskmaster_mcp(self) -> bool:
        """Check if Task Master MCP is available."""
        mcp_config = self.project_path / ".mcp.json"
        if mcp_config.exists():
            try:
                with open(mcp_config) as f:
                    config = json.load(f)
                    return "task-master-ai" in config.get("mcpServers", {})
            except:
                pass
        return False
    
    def _run_taskmaster(self, command: str) -> str:
        """Run Task Master CLI as fallback."""
        try:
            result = subprocess.run(
                f"task-master {command}",
                shell=True,
                cwd=self.project_path,
                capture_output=True,
                text=True,
                check=True
            )
            return result.stdout
        except subprocess.CalledProcessError as e:
            return f"Error: {e.stderr}"
    
    async def run_single_loop(self) -> Dict:
        """Execute one complete TDD automation loop."""
        
        loop_result = {
            "status": "started",
            "task_id": None,
            "steps": {
                "get_task": {"status": "pending"},
                "research": {"status": "pending"}, 
                "plan": {"status": "pending"},
                "tdd": {"status": "pending"},
                "complete": {"status": "pending"}
            },
            "outputs": {},
            "errors": []
        }
        
        try:
            # Step 1: Get next task
            loop_result["steps"]["get_task"]["status"] = "in_progress"
            next_task = await self._get_next_task()
            loop_result["task_id"] = next_task.get("id")
            loop_result["steps"]["get_task"]["status"] = "done"
            loop_result["outputs"]["task"] = next_task
            
            if not next_task.get("id"):
                loop_result["status"] = "no_tasks_available"
                return loop_result
            
            # Step 2: Research task
            loop_result["steps"]["research"]["status"] = "in_progress"
            research_result = await self._research_task(next_task)
            loop_result["steps"]["research"]["status"] = "done"
            loop_result["outputs"]["research"] = research_result
            
            # Step 3: Plan implementation
            loop_result["steps"]["plan"]["status"] = "in_progress"  
            plan_result = await self._plan_task(next_task, research_result)
            loop_result["steps"]["plan"]["status"] = "done"
            loop_result["outputs"]["plan"] = plan_result
            
            # Step 4: Execute TDD
            loop_result["steps"]["tdd"]["status"] = "in_progress"
            tdd_result = await self._execute_tdd(next_task, plan_result)
            loop_result["steps"]["tdd"]["status"] = "done"
            loop_result["outputs"]["tdd"] = tdd_result
            
            # Step 5: Mark complete
            loop_result["steps"]["complete"]["status"] = "in_progress"
            complete_result = await self._complete_task(next_task, tdd_result)
            loop_result["steps"]["complete"]["status"] = "done"
            loop_result["outputs"]["completion"] = complete_result
            
            loop_result["status"] = "completed"
            
        except Exception as e:
            loop_result["status"] = "error"
            loop_result["errors"].append(str(e))
            
        return loop_result
    
    async def _get_next_task(self) -> Dict:
        """Step 1: Get next available task from Task Master."""
        
        prompt = """Get the next available task to work on:
        
        1. Use task-master next to find the next task
        2. Use task-master show <id> to get full details
        3. Return task information including ID, title, description
        
        If no tasks available, return empty dict.
        """
        
        response = await self.orchestrator.query(prompt)
        
        # Also try CLI fallback
        cli_next = self._run_taskmaster("next")
        
        return {
            "ai_response": response,
            "cli_output": cli_next,
            "id": "extracted_from_response"  # Would need proper parsing
        }
    
    async def _research_task(self, task: Dict) -> Dict:
        """Step 2: Research task and store findings."""
        
        task_id = task.get("id", "unknown")
        
        prompt = f"""Research this task thoroughly:
        
        Task: {task}
        
        Steps:
        1. Use task-master research to analyze requirements
        2. Store findings in .taskmaster/reports/research-{task_id}.md
        3. Update task with research notes using update-subtask
        4. Tag research as complete
        
        Focus on understanding what needs to be implemented.
        """
        
        research_response = await self.orchestrator.query(prompt)
        
        return {
            "research_completed": True,
            "report_path": f".taskmaster/reports/research-{task_id}.md",
            "findings": research_response
        }
    
    async def _plan_task(self, task: Dict, research: Dict) -> Dict:
        """Step 3: Plan implementation with approval gates."""
        
        task_id = task.get("id", "unknown")
        
        prompt = f"""Plan implementation for this task:
        
        Task: {task}
        Research: {research['findings']}
        
        Steps:
        1. Use task-master expand to break into subtasks if needed
        2. Define clear acceptance criteria  
        3. Plan TDD approach (what tests to write)
        4. Update task with plan using update-task
        5. Request approval: "@tdd_agent: ready for implementation"
        
        Make plan concrete and testable.
        """
        
        plan_response = await self.orchestrator.query(prompt)
        
        return {
            "plan_created": True,
            "approach": plan_response,
            "approval_requested": True
        }
    
    async def _execute_tdd(self, task: Dict, plan: Dict) -> Dict:
        """Step 4: Coordinate TDD cycle - but only TESTER writes code."""
        
        prompt = f"""Coordinate TDD cycle for this task:
        
        Task: {task}
        Plan: {plan['approach']}
        
        IMPORTANT: You are the orchestrator - you coordinate but DON'T write code.
        Only the TESTER agent has file write permissions.
        
        Steps:
        1. Use task-master update-subtask to request tester implementation
        2. Update task: "@tester: ready for TDD implementation, see plan"
        3. Monitor task status updates from tester
        4. Verify tester marks tests as passing
        5. Get approval from tester before marking complete
        
        You coordinate - the tester implements.
        """
        
        tdd_response = await self.orchestrator.query(prompt)
        
        return {
            "tdd_requested": True,
            "tester_notified": True,
            "awaiting_implementation": True,
            "details": tdd_response
        }
    
    async def _complete_task(self, task: Dict, tdd_result: Dict) -> Dict:
        """Step 5: Mark task complete and prepare for next."""
        
        task_id = task.get("id", "unknown")
        
        prompt = f"""Complete this task:
        
        Task ID: {task_id}
        TDD Result: {tdd_result}
        
        Steps:
        1. Use task-master set-status --id={task_id} --status=done
        2. Add completion tag
        3. Update with final notes
        4. Prepare for next task
        
        Mark as officially complete.
        """
        
        completion_response = await self.orchestrator.query(prompt)
        
        return {
            "task_completed": True,
            "status_updated": True,
            "ready_for_next": True,
            "details": completion_response
        }


# Simple CLI interface for testing
if __name__ == "__main__":
    import asyncio
    import sys
    
    if len(sys.argv) < 2:
        print("Usage: python mvp_workflow.py <project_path>")
        sys.exit(1)
    
    project_path = sys.argv[1]
    workflow = SimpleTDDWorkflow(project_path)
    
    async def run():
        print("Starting MVP TDD Automation Loop...")
        result = await workflow.run_single_loop()
        print(json.dumps(result, indent=2))
    
    asyncio.run(run())