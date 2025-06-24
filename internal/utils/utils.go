package utils

import (
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

// GetFormattedTimeStamp returns the current UTC time formatted as YYYY:MM:DD-HH:MM:SS.
func GetFormattedTimeStamp() string {
	now := time.Now().UTC()
	return now.Format("2006:01:02-15:04:05.000") // Go's reference time for YYYY:MM:DD-HH:MM:SS.000 (milliseconds)
}

// PadNumber adds a leading zero to numbers less than 10.
// It takes an integer 'num' and returns a string representation of the padded number.
func PadNumber(num int) string {
	if num < 10 && num >= 0 {
		return "0" + strconv.Itoa(num)
	}
	return strconv.Itoa(num)
}

// HexFix removes the "0x" prefix from a hexadecimal string if present.
// If the input 'word' is not a string, it returns an empty string.
// Otherwise, it returns the string with the "0x" prefix removed.
// This matches the NodeJS implementation behavior.
func HexFix(word string) string {
	if word == "" {
		return ""
	}
	if strings.HasPrefix(word, "0x") || strings.HasPrefix(word, "0X") {
		return word[2:]
	}
	return word
}

// StringToHex converts a string to its hexadecimal representation.
// It correctly handles multi-byte Unicode characters by encoding the string as UTF-8
// before conversion. The resulting hexadecimal string does not include a "0x" prefix.
func StringToHex(str string) string {
	return hex.EncodeToString([]byte(str))
}

// HexToString converts a hexadecimal string back to a regular string.
// It first removes any "0x" prefix using HexFix and then decodes the hexadecimal
// string into a byte buffer, interpreting it as a UTF-8 string to correctly reconstruct
// multi-byte Unicode characters.
func HexToString(hexStr string) string {
	fixedHex := HexFix(hexStr)
	decoded, err := hex.DecodeString(fixedHex)
	if err != nil {
		return "" // Return empty string on error
	}
	return string(decoded)
}
