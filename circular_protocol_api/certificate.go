// Package circular_protocol_api provides functionalities to interact with
// the Circular Protocol API. It includes types and methods for managing
// accounts, certificates, and transactions.
package circular_protocol_api

import (
	"encoding/hex"
	"encoding/json"
)

// CCertificate represents a Circular Protocol Certificate.
// It stores the certificate data, references to previous transactions/blocks,
// and the code version used to create it.
// The struct tags `json:"..."` are used to control the field names during JSON serialization,
// ensuring compatibility with systems expecting specific (e.g., camelCase) naming.
type CCertificate struct {
	Data          string `json:"data"`          // Data is the hex-encoded content of the certificate.
	PreviousTxID  string `json:"previousTxID"`  // PreviousTxID is the transaction ID of the previous certificate in a chain, if applicable.
	PreviousBlock string `json:"previousBlock"` // PreviousBlock is the block ID where the previous certificate was recorded, if applicable.
	CodeVersion   string `json:"version"`       // CodeVersion indicates the version of the client or library that generated this certificate.
}

// NewCCertificate creates and initializes a new CCertificate instance with default values.
// Specifically, it sets the CodeVersion to the current library version.
// Other fields (Data, PreviousTxID, PreviousBlock) are initialized to their zero values (empty strings).
//
// Returns:
//
//	A pointer to the newly created CCertificate.
func NewCCertificate() *CCertificate {
	return &CCertificate{
		// String fields (Data, PreviousTxID, PreviousBlock) are automatically
		// initialized to "" (empty string) in Go, which is the desired default.
		CodeVersion: libVersion, // Set the version of the library.
	}
}

// SetData encodes the provided string data into its hexadecimal representation
// and assigns it to the Data field of the CCertificate.
//
// Parameters:
//
//	data: The string data to be encoded and stored in the certificate.
func (c *CCertificate) SetData(data string) {
	c.Data = hex.EncodeToString([]byte(data))
}

// GetData decodes the hexadecimal string from the CCertificate's Data field
// back into its original string representation.
//
// Returns:
//
//	The decoded string data.
//	An error if the hex decoding fails.
func (c *CCertificate) GetData() (string, error) {
	// Uses the internal hexToString helper for decoding.
	return hexToString(c.Data)
}

// GetJSONCertificate serializes the CCertificate instance into a JSON formatted string.
// The field names in the JSON output are determined by the `json` struct tags
// (e.g., "data", "previousTxID").
//
// Returns:
//
//	A string containing the JSON representation of the certificate.
//	An error if JSON marshaling fails.
func (c *CCertificate) GetJSONCertificate() (string, error) {
	// Marshal the certificate struct 'c' into a JSON byte slice.
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		// If marshaling fails, return an empty string and the error.
		return "", err
	}
	// Convert the JSON byte slice to a string.
	return string(jsonBytes), nil
}

// GetCertificateSize calculates and returns the size in bytes of the
// JSON-serialized representation of the CCertificate.
// This method ensures that the size calculation is based on the actual
// serialized form, correctly accounting for JSON formatting and character encoding.
//
// Returns:
//
//	The size of the JSON-serialized certificate in bytes.
//	An error if JSON marshaling (which is a prerequisite for size calculation) fails.
func (c *CCertificate) GetCertificateSize() (int, error) {
	// Marshal the certificate to its JSON byte representation to accurately determine its size.
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		// If marshaling fails, the size cannot be determined, so return 0 and the error.
		return 0, err
	}
	// The length of the byte slice is the size of the JSON data.
	return len(jsonBytes), nil
}

// hexToString is a non-exported helper function that decodes a hexadecimal string
// into its original string representation. It handles optional "0x" prefixes.
//
// Parameters:
//
//	h: The hexadecimal string to decode. It may optionally start with "0x".
//
// Returns:
//
//	The decoded string.
//	An error if the hexadecimal decoding fails (e.g., invalid characters).
func hexToString(h string) (string, error) {
	// Make a copy to modify if "0x" prefix is present.
	cleanHex := h
	// Check for and remove "0x" prefix if it exists.
	if len(h) >= 2 && h[0:2] == "0x" {
		cleanHex = h[2:]
	}

	// Decode the cleaned hexadecimal string to a byte slice.
	decodedBytes, err := hex.DecodeString(cleanHex)
	if err != nil {
		// If decoding fails, return an empty string and the error.
		return "", err
	}
	// Convert the decoded byte slice to a string.
	return string(decodedBytes), nil
}
