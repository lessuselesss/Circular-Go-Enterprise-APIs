Implement GetFormattedTimeStamp() in internal/utils/utils.go to return the current UTC time formatted as YYYY:MM:DD-HH:MM:SS, matching the behavior of the NodeJS getFormattedTimestamp() function.

TDD Steps:

Step 1: Initial Failing Test (Red)
*   File: internal/utils/utils_test.go
*   Action: Modify the existing TestGetFormattedTimeStamp to assert the correct format and ensure it takes no arguments.
*   Expected Test Code (Failing):
    ```go
    package utils

    import (
    	"testing"
    	"time"
    	"regexp" // Import regexp for pattern matching
    )

    func TestGetFormattedTimeStamp(t *testing.T) {
    	// Test 1: Assert the correct format (YYYY:MM:DD-HH:MM:SS)
    	// The function should take no arguments and use current UTC time.
    	// This test will initially fail because the function signature is wrong
    	// and the format might not match.
    	timestamp := GetFormattedTimeStamp() // Function signature change: no arguments

    	// Regex to match YYYY:MM:DD-HH:MM:SS format
    	// Example: 2025:06:23-07:15:30
    	pattern := `^\d{4}:\d{2}:\d{2}-\d{2}:\d{2}:\d{2}$`
    	matched, err := regexp.MatchString(pattern, timestamp)
    	if err != nil {
    		t.Fatalf("Regex compilation error: %v", err)
    	}
    	if !matched {
    		t.Errorf("GetFormattedTimeStamp() returned '%s', which does not match expected format %s", timestamp, pattern)
    	}

    	// Further tests will verify the actual time value.
    }
    ```
*   Reason for Failure: GetFormattedTimeStamp function signature mismatch (expects time.Time argument, but test calls without), and potentially incorrect format if the function were to compile.

Step 2: Minimum Code to Pass (Green)
*   File: internal/utils/utils.go
*   Action: Modify GetFormattedTimeStamp to take no arguments, use time.Now().UTC(), and format it according to YYYY:MM:DD-HH:MM:SS.
*   Expected Code:
    ```go
    package utils

    import "time"

    // GetFormattedTimeStamp returns the current UTC time formatted as YYYY:MM:DD-HH:MM:SS.
    func GetFormattedTimeStamp() string {
    	now := time.Now().UTC()
    	return now.Format("2006:01:02-15:04:05") // Go's reference time for YYYY:MM:DD-HH:MM:SS
    }
    ```
*   Expected Outcome: The test from Step 1 should now pass.

Step 3: Refactor and Add More Tests (Red/Green/Refactor Cycle)
*   File: internal/utils/utils_test.go
*   Action: Add a test to verify the accuracy of the timestamp, not just the format. This will involve comparing the generated timestamp with a known UTC time (within a small tolerance).
*   Expected Test Code (Failing initially, then passing):
    ```go
    package utils

    import (
    	"testing"
    	"time"
    	"regexp"
    )

    func TestGetFormattedTimeStamp(t *testing.T) {
    	// Test 1: Assert the correct format (YYYY:MM:DD-HH:MM:SS)
    	timestamp := GetFormattedTimeStamp()
    	pattern := `^\d{4}:\d{2}:\d{2}-\d{2}:\d{2}:\d{2}$`
    	matched, err := regexp.MatchString(pattern, timestamp)
    	if err != nil {
    		t.Fatalf("Regex compilation error: %v", err)
    	}
    	if !matched {
    		t.Errorf("GetFormattedTimeStamp() returned '%s', which does not match expected format %s", timestamp, pattern)
    	}

    	// Test 2: Assert the timestamp is approximately current UTC time
    	// This test might be flaky if run exactly at a second boundary,
    	// but should generally pass.
    	parsedTime, err := time.Parse("2006:01:02-15:04:05", timestamp)
    	if err != nil {
    		t.Fatalf("Failed to parse timestamp '%s': %v", timestamp, err)
    	}

    	nowUTC := time.Now().UTC()
    	// Allow a small tolerance for execution time
    	if parsedTime.Before(nowUTC.Add(-2*time.Second)) || parsedTime.After(nowUTC.Add(2*time.Second)) {
    		t.Errorf("GetFormattedTimeStamp() returned time %s; expected approximately %s", parsedTime.Format(time.RFC3339), nowUTC.Format(time.RFC3339))
    	}

    	// Test 3: (Future consideration) Test with a mocked time to ensure deterministic output
    	// This would require dependency injection or a global variable for time.Now,
    	// which might be overkill for a simple utility function unless specifically required.
    }
    ```
*   Refactoring (if needed): Ensure the GetFormattedTimeStamp function is clean and efficient. For this simple function, the initial implementation is likely sufficient.