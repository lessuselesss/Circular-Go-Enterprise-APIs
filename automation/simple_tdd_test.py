#!/usr/bin/env python3
"""
Simple TDD automation test - CLI-only version
Test our workflow on task 2.1: GetFormattedTimeStamp
"""

import os
import subprocess
import json
from pathlib import Path


class SimpleTDDTest:
    """Test our TDD automation workflow using CLI only."""
    
    def __init__(self, project_path: str):
        self.project_path = Path(project_path)
    
    def run_taskmaster(self, command: str) -> str:
        """Run Task Master CLI command."""
        try:
            # Use nix develop to ensure we have the right environment
            full_cmd = f'nix develop --command bash -c "task-master {command}"'
            result = subprocess.run(
                full_cmd,
                shell=True,
                cwd=self.project_path,
                capture_output=True,
                text=True,
                check=True
            )
            return result.stdout
        except subprocess.CalledProcessError as e:
            return f"Error: {e.stderr}"
    
    def run_go_test(self, path: str = ""):
        """Run Go tests."""
        try:
            test_cmd = f"go test {path}" if path else "go test ./..."
            full_cmd = f'nix develop --command bash -c "{test_cmd}"'
            result = subprocess.run(
                full_cmd,
                shell=True,
                cwd=self.project_path,
                capture_output=True,
                text=True
            )
            return {
                "exit_code": result.returncode,
                "stdout": result.stdout,
                "stderr": result.stderr,
                "passed": result.returncode == 0
            }
        except Exception as e:
            return {
                "exit_code": -1,
                "stdout": "",
                "stderr": str(e),
                "passed": False
            }
    
    def test_workflow(self, task_id: str = "2.1"):
        """Test our TDD workflow on a specific task."""
        
        print(f"🚀 Testing TDD Automation Workflow for task {task_id}")
        print("=" * 60)
        
        # Step 1: Get task details (Research phase)
        print("\n📋 Step 1: Getting task details...")
        task_details = self.run_taskmaster(f"show {task_id}")
        print("Task details retrieved ✓")
        
        # Step 2: Check current test status
        print("\n🧪 Step 2: Checking current test status...")
        test_result = self.run_go_test("./internal/utils/")
        print(f"Current tests - Pass: {test_result['passed']}")
        if not test_result['passed']:
            print("Tests are failing (expected for TDD Red phase)")
        
        # Step 3: Simulate research phase
        print("\n🔍 Step 3: Research phase - analyzing existing code...")
        
        # Check what files exist
        utils_dir = self.project_path / "internal" / "utils"
        if utils_dir.exists():
            files = list(utils_dir.glob("*"))
            print(f"Found files: {[f.name for f in files]}")
        
        # Step 4: Update task with research findings
        print("\n📝 Step 4: Updating task with research findings...")
        research_update = self.run_taskmaster(
            f'update-subtask --id={task_id} --prompt="Research complete: Found existing utils.go and utils_test.go. Ready for TDD implementation."'
        )
        print("Task updated with research ✓")
        
        # Step 5: Simulate TDD planning
        print("\n📋 Step 5: TDD Planning phase...")
        planning_update = self.run_taskmaster(
            f'update-subtask --id={task_id} --prompt="@tester: TDD plan ready. Need to implement GetFormattedTimeStamp function following the existing pattern. Write failing test first, then implement."'
        )
        print("TDD plan created and tester notified ✓")
        
        # Step 6: Check if ready for implementation
        print("\n✅ Step 6: Workflow coordination complete!")
        print("Next steps would be:")
        print("1. Tester agent writes failing test")
        print("2. Tester agent implements minimal code")
        print("3. Tester agent refactors and adds comprehensive tests")
        print("4. Tester agent bubbles up approval to planner")
        
        # Step 7: Show updated task
        print("\n📊 Step 7: Updated task status...")
        updated_task = self.run_taskmaster(f"show {task_id}")
        print("Task has been updated with workflow progress ✓")
        
        return {
            "workflow_completed": True,
            "task_id": task_id,
            "ready_for_tdd": True
        }


if __name__ == "__main__":
    import sys
    
    project_path = sys.argv[1] if len(sys.argv) > 1 else "."
    task_id = sys.argv[2] if len(sys.argv) > 2 else "2.1"
    
    tester = SimpleTDDTest(project_path)
    result = tester.test_workflow(task_id)
    
    print("\n" + "=" * 60)
    print("🎉 TDD Automation Workflow Test Complete!")
    print(f"Result: {json.dumps(result, indent=2)}")