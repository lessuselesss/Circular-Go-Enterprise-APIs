package circular_protocol_enterprise_api

import (
	"encoding/hex"
	"strings"
	"strconv"
	"time"
)

// hexFix ensures the hex string has a "0x" prefix.
// If the string already has "0x", it's returned as is.
// If not, "0x" is prepended.
func hexFix(s string) string {
	if strings.HasPrefix(s, "0x") {
		return s
	}
	return "0x" + s
}

// stringToHex converts a string to its hexadecimal representation.
// The resulting hex string does NOT have a "0x" prefix.
func stringToHex(s string) string {
	return hex.EncodeToString([]byte(s))
}

// Helper function to remove "0x" prefix if present and decode hex.
// This is similar to hexToString in certificate.go but used internally.
func decodeHex(h string) ([]byte, error) {
	cleanHex := h
	if strings.HasPrefix(h, "0x") {
		cleanHex = h[2:]
	}
	return hex.DecodeString(cleanHex)
}

// getFormattedTimestamp returns the current time in milliseconds since epoch as a string.
func getFormattedTimestamp() string {
	return strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
}
