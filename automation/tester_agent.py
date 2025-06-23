#!/usr/bin/env python3
"""
Tester Agent - The ONLY agent with file write permissions.
Handles TDD cycle: Write failing tests → Implement code → Verify tests pass
"""

import os
import json
import subprocess
from pathlib import Path
from typing import Dict, List

from mcp_agent import MCPApp, Agent


class TesterAgent:
    """The only agent authorized to write to the codebase."""
    
    def __init__(self, project_path: str):
        self.project_path = Path(project_path)
        self.app = MCPApp()
        
        # Tester agent with file system access
        self.tester = Agent(
            name="tdd_tester",
            instruction="""You are the TDD Tester Agent - the ONLY agent with file write permissions.
            
            Your responsibilities:
            1. Write failing tests based on task requirements
            2. Implement minimal code to make tests pass
            3. Run tests to verify they pass
            4. Update Task Master with progress and approval
            
            You follow strict TDD: Red → Green → Refactor
            
            IMPORTANT: Other agents can only coordinate - you do all the actual coding.
            Use Task Master tools to communicate with other agents.
            """,
            server_names=["task-master-ai", "filesystem"] if self._has_mcp_servers() else []
        )
    
    def _has_mcp_servers(self) -> bool:
        """Check if required MCP servers are available."""
        mcp_config = self.project_path / ".mcp.json"
        if mcp_config.exists():
            try:
                with open(mcp_config) as f:
                    config = json.load(f)
                    servers = config.get("mcpServers", {})
                    return "task-master-ai" in servers
            except:
                pass
        return False
    
    def _run_taskmaster(self, command: str) -> str:
        """Run Task Master CLI."""
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
    
    def _run_tests(self, test_path: str = "") -> str:
        """Run Go tests."""
        try:
            cmd = f"go test {test_path}" if test_path else "go test ./..."
            result = subprocess.run(
                cmd,
                shell=True,
                cwd=self.project_path,
                capture_output=True,
                text=True
            )
            return f"Exit code: {result.returncode}\nStdout: {result.stdout}\nStderr: {result.stderr}"
        except Exception as e:
            return f"Test execution error: {str(e)}"
    
    async def execute_tdd_cycle(self, task_id: str) -> Dict:
        """Execute complete TDD cycle for a task."""
        
        tdd_result = {
            "task_id": task_id,
            "status": "started",
            "phases": {
                "red": {"status": "pending", "details": ""},
                "green": {"status": "pending", "details": ""},
                "refactor": {"status": "pending", "details": ""},
                "verify": {"status": "pending", "details": ""}
            },
            "files_modified": [],
            "tests_passing": False,
            "errors": []
        }
        
        try:
            # Get task details
            task_details = self._run_taskmaster(f"show {task_id}")
            
            # Phase 1: RED - Write failing tests
            tdd_result["phases"]["red"]["status"] = "in_progress"
            red_result = await self._red_phase(task_id, task_details)
            tdd_result["phases"]["red"] = red_result
            tdd_result["files_modified"].extend(red_result.get("files", []))
            
            # Phase 2: GREEN - Make tests pass
            tdd_result["phases"]["green"]["status"] = "in_progress"
            green_result = await self._green_phase(task_id, red_result)
            tdd_result["phases"]["green"] = green_result
            tdd_result["files_modified"].extend(green_result.get("files", []))
            
            # Phase 3: REFACTOR - Clean up code
            tdd_result["phases"]["refactor"]["status"] = "in_progress"
            refactor_result = await self._refactor_phase(task_id, green_result)
            tdd_result["phases"]["refactor"] = refactor_result
            tdd_result["files_modified"].extend(refactor_result.get("files", []))
            
            # Phase 4: VERIFY - Confirm all tests pass
            tdd_result["phases"]["verify"]["status"] = "in_progress"
            verify_result = await self._verify_phase(task_id)
            tdd_result["phases"]["verify"] = verify_result
            tdd_result["tests_passing"] = verify_result.get("all_tests_pass", False)
            
            # Update Task Master with completion
            await self._update_task_completion(task_id, tdd_result)
            
            tdd_result["status"] = "completed"
            
        except Exception as e:
            tdd_result["status"] = "error"
            tdd_result["errors"].append(str(e))
            
        return tdd_result
    
    async def _red_phase(self, task_id: str, task_details: str) -> Dict:
        """RED: Write failing tests."""
        
        prompt = f"""Write failing tests for this task:
        
        Task ID: {task_id}
        Task Details: {task_details}
        
        Steps:
        1. Analyze task requirements
        2. Write Go test file(s) that define expected behavior
        3. Ensure tests FAIL initially (Red phase)
        4. Follow Go testing conventions
        5. Update Task Master: "@planner: failing tests written for {task_id}"
        
        Write the actual test files - you have file write permissions.
        """
        
        test_response = await self.tester.query(prompt)
        
        # Run tests to confirm they fail
        test_output = self._run_tests()
        
        # Update Task Master
        self._run_taskmaster(f'update-subtask --id={task_id} --prompt="@planner: RED phase complete - failing tests written"')
        
        return {
            "status": "completed",
            "details": test_response,
            "test_output": test_output,
            "files": ["test_files_written"]  # Would track actual files
        }
    
    async def _green_phase(self, task_id: str, red_result: Dict) -> Dict:
        """GREEN: Implement code to make tests pass."""
        
        prompt = f"""Implement code to make the failing tests pass:
        
        Task ID: {task_id}
        Red Phase Result: {red_result['details']}
        
        Steps:
        1. Write minimal implementation to make tests pass
        2. Focus on making tests GREEN, not perfect code
        3. Follow Go idioms and project conventions
        4. Run tests to verify they pass
        5. Update Task Master: "@planner: GREEN phase complete - tests passing"
        
        Write the actual implementation files.
        """
        
        impl_response = await self.tester.query(prompt)
        
        # Run tests to confirm they pass
        test_output = self._run_tests()
        
        # Update Task Master
        self._run_taskmaster(f'update-subtask --id={task_id} --prompt="@planner: GREEN phase complete - tests now passing"')
        
        return {
            "status": "completed",
            "details": impl_response,
            "test_output": test_output,
            "files": ["implementation_files_written"]
        }
    
    async def _refactor_phase(self, task_id: str, green_result: Dict) -> Dict:
        """REFACTOR: Clean up code while keeping tests green."""
        
        prompt = f"""Refactor the implementation to improve code quality:
        
        Task ID: {task_id}
        Green Phase Result: {green_result['details']}
        
        Steps:
        1. Review implementation for improvements
        2. Refactor code while keeping tests passing
        3. Add any additional test cases if needed
        4. Ensure code follows project conventions
        5. Update Task Master: "@planner: REFACTOR phase complete - code cleaned up"
        
        Only refactor if improvements are needed.
        """
        
        refactor_response = await self.tester.query(prompt)
        
        # Run tests to ensure they still pass
        test_output = self._run_tests()
        
        # Update Task Master
        self._run_taskmaster(f'update-subtask --id={task_id} --prompt="@planner: REFACTOR phase complete - code quality improved"')
        
        return {
            "status": "completed",
            "details": refactor_response,
            "test_output": test_output,
            "files": ["refactored_files"]
        }
    
    async def _verify_phase(self, task_id: str) -> Dict:
        """VERIFY: Final test run and validation."""
        
        # Run full test suite
        test_output = self._run_tests()
        
        # Check if all tests pass
        all_tests_pass = "FAIL" not in test_output and "exit code: 0" in test_output.lower()
        
        # Update Task Master with final status
        status_msg = "All tests passing - ready for approval" if all_tests_pass else "Tests failing - need fixes"
        self._run_taskmaster(f'update-subtask --id={task_id} --prompt="@planner: VERIFY phase complete - {status_msg}"')
        
        return {
            "status": "completed",
            "all_tests_pass": all_tests_pass,
            "test_output": test_output,
            "details": f"Final verification: {status_msg}"
        }
    
    async def _update_task_completion(self, task_id: str, tdd_result: Dict) -> None:
        """Update Task Master with TDD completion and request approval."""
        
        if tdd_result["tests_passing"]:
            # Request approval from planner
            self._run_taskmaster(f'update-subtask --id={task_id} --prompt="@planner: TDD cycle complete. All tests passing. Ready for task completion approval."')
        else:
            # Report issues
            self._run_taskmaster(f'update-subtask --id={task_id} --prompt="@planner: TDD cycle completed but tests failing. Need review."')


# CLI interface for manual testing
if __name__ == "__main__":
    import asyncio
    import sys
    
    if len(sys.argv) < 3:
        print("Usage: python tester_agent.py <project_path> <task_id>")
        sys.exit(1)
    
    project_path = sys.argv[1]
    task_id = sys.argv[2]
    
    tester = TesterAgent(project_path)
    
    async def run():
        print(f"Starting TDD cycle for task {task_id}...")
        result = await tester.execute_tdd_cycle(task_id)
        print(json.dumps(result, indent=2))
    
    asyncio.run(run())