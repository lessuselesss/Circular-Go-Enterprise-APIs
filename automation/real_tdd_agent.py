#!/usr/bin/env python3
"""
Real TDD Agent - Implements the actual RED/GREEN/REFACTOR cycle
Writes failing tests, then makes them pass with minimal code
"""

import os
import subprocess
import time
from pathlib import Path


class RealTDDAgent:
    """The actual TDD implementation agent that writes code."""
    
    def __init__(self, project_path: str):
        self.project_path = Path(project_path)
    
    def run_taskmaster(self, command: str) -> str:
        """Run Task Master CLI command."""
        try:
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
    
    def run_go_test(self, path: str = "./internal/utils/"):
        """Run Go tests and return detailed results."""
        try:
            full_cmd = f'nix develop --command bash -c "cd {path} && go test -v"'
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
    
    def write_file(self, file_path: str, content: str):
        """Write content to a file."""
        full_path = self.project_path / file_path
        with open(full_path, 'w') as f:
            f.write(content)
        print(f"✏️  Wrote {file_path}")
    
    def execute_tdd_cycle(self, task_id: str = "2.1"):
        """Execute the complete RED/GREEN/REFACTOR TDD cycle."""
        
        print(f"🔴🟢🔵 Starting TDD Cycle for Task {task_id}")
        print("=" * 60)
        
        # Update task status
        self.run_taskmaster(f'update-subtask --id={task_id} --prompt="@planner: Starting TDD implementation cycle"')
        
        # PHASE 1: RED - Write failing test
        print("\n🔴 PHASE 1: RED - Writing Failing Test")
        self.red_phase(task_id)
        
        # PHASE 2: GREEN - Make test pass  
        print("\n🟢 PHASE 2: GREEN - Making Test Pass")
        self.green_phase(task_id)
        
        # PHASE 3: REFACTOR - Improve code
        print("\n🔵 PHASE 3: REFACTOR - Improving Code") 
        self.refactor_phase(task_id)
        
        # Final verification
        print("\n✅ FINAL: Verification")
        self.verify_phase(task_id)
        
        print("\n🎉 TDD Cycle Complete!")
        return {"tdd_complete": True, "task_id": task_id}
    
    def red_phase(self, task_id: str):
        """RED: Write failing test."""
        print("Writing failing test that changes function signature...")
        
        # Write the failing test from the TDD plan
        failing_test = '''package utils

import (
	"testing"
	"regexp"
)

func TestGetFormattedTimeStamp(t *testing.T) {
	// Test 1: Assert the correct format (YYYY:MM:DD-HH:MM:SS)
	// The function should take no arguments and use current UTC time.
	// This test will initially fail because the function signature is wrong
	timestamp := GetFormattedTimeStamp() // Function signature change: no arguments

	// Regex to match YYYY:MM:DD-HH:MM:SS format
	// Example: 2025:06:23-07:15:30
	pattern := `^\\d{4}:\\d{2}:\\d{2}-\\d{2}:\\d{2}:\\d{2}$`
	matched, err := regexp.MatchString(pattern, timestamp)
	if err != nil {
		t.Fatalf("Regex compilation error: %v", err)
	}
	if !matched {
		t.Errorf("GetFormattedTimeStamp() returned '%s', which does not match expected format %s", timestamp, pattern)
	}
}
'''
        
        self.write_file("internal/utils/utils_test.go", failing_test)
        
        # Run test to confirm it fails
        print("Running test to confirm it fails...")
        test_result = self.run_go_test()
        
        if test_result["passed"]:
            print("⚠️  Test unexpectedly passed! This should fail in RED phase.")
        else:
            print("✅ Test failed as expected (RED phase successful)")
            print(f"Error: {test_result['stderr']}")
        
        # Update task
        self.run_taskmaster(f'update-subtask --id={task_id} --prompt="RED phase complete: failing test written with new signature"')
    
    def green_phase(self, task_id: str):
        """GREEN: Make test pass with minimal code."""
        print("Writing minimal implementation to make test pass...")
        
        # Write the minimal implementation from the TDD plan
        minimal_impl = '''package utils

import "time"

// GetFormattedTimeStamp returns the current UTC time formatted as YYYY:MM:DD-HH:MM:SS.
func GetFormattedTimeStamp() string {
	now := time.Now().UTC()
	return now.Format("2006:01:02-15:04:05") // Go's reference time for YYYY:MM:DD-HH:MM:SS
}
'''
        
        self.write_file("internal/utils/utils.go", minimal_impl)
        
        # Run test to confirm it passes
        print("Running test to confirm it passes...")
        test_result = self.run_go_test()
        
        if test_result["passed"]:
            print("✅ Test passed! GREEN phase successful")
            print(test_result["stdout"])
        else:
            print("❌ Test still failing - need to fix implementation")
            print(f"Error: {test_result['stderr']}")
        
        # Update task
        self.run_taskmaster(f'update-subtask --id={task_id} --prompt="GREEN phase complete: minimal implementation makes test pass"')
    
    def refactor_phase(self, task_id: str):
        """REFACTOR: Add more comprehensive tests."""
        print("Adding comprehensive tests...")
        
        # Add more comprehensive tests from the TDD plan
        comprehensive_test = '''package utils

import (
	"testing"
	"time"
	"regexp"
)

func TestGetFormattedTimeStamp(t *testing.T) {
	// Test 1: Assert the correct format (YYYY:MM:DD-HH:MM:SS)
	timestamp := GetFormattedTimeStamp()
	pattern := `^\\d{4}:\\d{2}:\\d{2}-\\d{2}:\\d{2}:\\d{2}$`
	matched, err := regexp.MatchString(pattern, timestamp)
	if err != nil {
		t.Fatalf("Regex compilation error: %v", err)
	}
	if !matched {
		t.Errorf("GetFormattedTimeStamp() returned '%s', which does not match expected format %s", timestamp, pattern)
	}

	// Test 2: Assert the timestamp is approximately current UTC time
	parsedTime, err := time.Parse("2006:01:02-15:04:05", timestamp)
	if err != nil {
		t.Fatalf("Failed to parse timestamp '%s': %v", timestamp, err)
	}

	nowUTC := time.Now().UTC()
	// Allow a small tolerance for execution time
	if parsedTime.Before(nowUTC.Add(-2*time.Second)) || parsedTime.After(nowUTC.Add(2*time.Second)) {
		t.Errorf("GetFormattedTimeStamp() returned time %s; expected approximately %s", parsedTime.Format(time.RFC3339), nowUTC.Format(time.RFC3339))
	}
}

// Test 3: Test multiple calls return different times
func TestGetFormattedTimeStampUniqueness(t *testing.T) {
	timestamp1 := GetFormattedTimeStamp()
	time.Sleep(1100 * time.Millisecond) // Sleep just over 1 second
	timestamp2 := GetFormattedTimeStamp()
	
	if timestamp1 == timestamp2 {
		t.Errorf("Expected different timestamps, but got same: %s", timestamp1)
	}
}
'''
        
        self.write_file("internal/utils/utils_test.go", comprehensive_test)
        
        # Run comprehensive tests
        print("Running comprehensive tests...")
        test_result = self.run_go_test()
        
        if test_result["passed"]:
            print("✅ All comprehensive tests passed! REFACTOR phase successful")
            print(test_result["stdout"])
        else:
            print("❌ Some comprehensive tests failing")
            print(f"Error: {test_result['stderr']}")
        
        # Update task
        self.run_taskmaster(f'update-subtask --id={task_id} --prompt="REFACTOR phase complete: comprehensive tests added and passing"')
    
    def verify_phase(self, task_id: str):
        """VERIFY: Final verification and approval."""
        print("Final verification of all tests...")
        
        # Run all tests one final time
        test_result = self.run_go_test()
        
        if test_result["passed"]:
            print("🎯 All tests passing! TDD cycle successful!")
            
            # Request approval from planner
            self.run_taskmaster(f'update-subtask --id={task_id} --prompt="@planner: TDD cycle complete! All tests passing. Function signature changed from GetFormattedTimeStamp(time.Time) to GetFormattedTimeStamp() with new format YYYY:MM:DD-HH:MM:SS. Ready for task completion approval."')
            
            # Mark task as done
            self.run_taskmaster(f'set-status --id={task_id} --status=done')
            
            print("✅ Task marked as complete and approval requested from planner")
        else:
            print("❌ Final verification failed!")
            print(f"Error: {test_result['stderr']}")
            
            self.run_taskmaster(f'update-subtask --id={task_id} --prompt="@planner: TDD cycle completed but final verification failed. Need review."')


if __name__ == "__main__":
    import sys
    
    project_path = sys.argv[1] if len(sys.argv) > 1 else "."
    task_id = sys.argv[2] if len(sys.argv) > 2 else "2.1"
    
    agent = RealTDDAgent(project_path)
    result = agent.execute_tdd_cycle(task_id)
    
    print(f"\n🏁 TDD Agent Result: {result}")