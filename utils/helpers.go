package utils

import (
	"time"
)

// GetFormattedUTCTimestamp generates a UTC timestamp string in the format:
// YYYY-MM-DDTHH:MM:SSZ (e.g., "2023-10-27T10:30:00Z")
func GetFormattedUTCTimestamp() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

// Add other utility functions here if needed in the future, for example:
// HexEncode a byte slice to string (wrapper for hex.EncodeToString if specific logic needed)
// HexDecode a string to byte slice (wrapper for hex.DecodeString if specific error handling needed)
