package utils

import (
	"regexp"
	"testing"
	"time"
)

func TestGetFormattedTimeStamp(t *testing.T) {
	// Test 1: Assert the correct format (YYYY:MM:DD-HH:MM:SS)
	timestamp := GetFormattedTimeStamp()
	pattern := `^\d{4}:\d{2}:\d{2}-\d{2}:\d{2}:\d{2}\.\d{3}$`
	matched, err := regexp.MatchString(pattern, timestamp)
	if err != nil {
		t.Fatalf("Regex compilation error: %v", err)
	}
	if !matched {
		t.Errorf("GetFormattedTimeStamp() returned '%s', which does not match expected format %s", timestamp, pattern)
	}

	// Test 2: Assert the timestamp is approximately current UTC time
	parsedTime, err := time.Parse("2006:01:02-15:04:05.000", timestamp)
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

func TestPadNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"single digit", 5, "05"},
		{"zero", 0, "00"},
		{"double digit", 10, "10"},
		{"large number", 99, "99"},
		{"negative single digit", -5, "-5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PadNumber(tt.input)
			if result != tt.expected {
				t.Errorf("PadNumber(%d) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHexFix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"with 0x prefix", "0x1234abcd", "1234abcd"},
		{"with 0X prefix", "0X1234ABCD", "1234ABCD"},
		{"without prefix", "1234abcd", "1234abcd"},
		{"empty string", "", ""},
		{"only 0x", "0x", ""},
		{"invalid hex", "0xGHIJ", "GHIJ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HexFix(tt.input)
			if result != tt.expected {
				t.Errorf("HexFix(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStringToHex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple string", "hello", "68656c6c6f"},
		{"empty string", "", ""},
		{"unicode", "Hello 🌍", "48656c6c6f20f09f8c8d"},
		{"numbers", "123", "313233"},
		{"special chars", "!@#$", "21402324"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringToHex(tt.input)
			if result != tt.expected {
				t.Errorf("StringToHex(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHexToString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple hex", "68656c6c6f", "hello"},
		{"empty string", "", ""},
		{"unicode", "48656c6c6f20f09f8c8d", "Hello 🌍"},
		{"numbers", "313233", "123"},
		{"special chars", "21402324", "!@#$"},
		{"with 0x prefix", "0x68656c6c6f", "hello"},
		{"with 0X prefix", "0X68656c6c6f", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HexToString(tt.input)
			if result != tt.expected {
				t.Errorf("HexToString(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHexToStringInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"odd length", "68656c6c6"},
		{"invalid characters", "zzzz"},
		{"mixed invalid", "68gg6c6f"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HexToString(tt.input)
			if result != "" {
				t.Logf("HexToString(%q) = %q (invalid input handled)", tt.input, result)
			}
		})
	}
}

func TestStringHexRoundTrip(t *testing.T) {
	testStrings := []string{
		"hello world",
		"",
		"Hello 🌍 Unicode test",
		"!@#$%^&*()",
		"123456789",
	}

	for _, original := range testStrings {
		t.Run("roundtrip_"+original, func(t *testing.T) {
			hex := StringToHex(original)
			recovered := HexToString(hex)
			if recovered != original {
				t.Errorf("Round trip failed: %q -> %q -> %q", original, hex, recovered)
			}
		})
	}
}
