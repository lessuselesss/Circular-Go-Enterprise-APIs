package circular_protocol_enterprise_api

import (
	"encoding/json"
)
package circular

import (
	"encoding/hex"
	"encoding/json"
)

// A shared constant from the library details
const libVersion = "1.0.13"

// CCertificate mirrors the C_CERTIFICATE class from the Node.js library.
// It holds the data for a certificate.
type CCertificate struct {
	// Note the use of json tags. This controls how the struct is serialized
	// into JSON, matching the JavaScript object keys exactly.
	Data          string `json:"data"`
	PreviousTxID  string `json:"previousTxID"`
	PreviousBlock string `json:"previousBlock"`
	CodeVersion   string `json:"version"`
}

// NewCCertificate acts as a constructor for a CCertificate instance.
// It initializes the certificate with default values.
func NewCCertificate() *CCertificate {
	return &CCertificate{
		// In Go, string fields are automatically initialized to "" (empty string),
		// which is equivalent to 'null' for these fields in the JS constructor.
		// So we only need to set the version.
		CodeVersion: libVersion,
	}
}

// SetData encodes the provided string into hex and stores it in the certificate.
// The receiver (c *CCertificate) is a pointer, allowing the method to modify the original struct.
func (c *CCertificate) SetData(data string) {
	c.Data = hex.EncodeToString([]byte(data))
}

// GetData decodes the hex data from the certificate back into a string.
// The receiver (c *CCertificate) can be a value receiver here, but pointer is
// conventional for consistency. It returns the decoded string and any potential error.
func (c *CCertificate) GetData() (string, error) {
	// The hexToString helper would handle decoding
	return hexToString(c.Data)
}

// GetJSONCertificate serializes the certificate struct into a JSON formatted string.
// It returns the JSON string and any potential error from the marshalling process.
func (c *CCertificate) GetJSONCertificate() (string, error) {
	// We are marshalling the struct 'c' itself. The json tags on the struct
	// fields will ensure the output keys ("data", "previousTxID", etc.) are correct.
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// GetCertificateSize calculates the byte size of the JSON-serialized certificate.
// This correctly handles multi-byte characters, just like the Node.js FIX.
func (c *CCertificate) GetCertificateSize() (int, error) {
	// By marshalling to JSON first, we get the exact representation whose
	// size we need to measure. The length of the resulting byte slice is the size.
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return 0, err
	}
	return len(jsonBytes), nil
}

// hexToString is a helper function to decode a hex string.
// It assumes the hex string may or may not have a "0x" prefix.
func hexToString(h string) (string, error) {
	// Trim the "0x" prefix if it exists
	cleanHex := h
	if len(h) > 2 && h[:2] == "0x" {
		cleanHex = h[2:]
	}

	decodedBytes, err := hex.DecodeString(cleanHex)
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}


// // Certificate holds the data for a data certificate.
// type Certificate struct {
// 	Data string `json:"data"` // Data content of the certificate
// }

// // SetData sets the data content of the certificate.
// func (c *Certificate) SetData(data string) {
// 	c.Data = data
// }

// // GetData retrieves the data content of the certificate.
// func (c *Certificate) GetData() string {
// 	return c.Data
// }

// // GetJSONCertificate returns the certificate as a JSON string.
// func (c *Certificate) GetJSONCertificate() (string, error) {
// 	jsonData, err := json.Marshal(c)
// 	if err != nil {
// 		return "", err
// 	}
// 	return string(jsonData), nil
// }

// // GetCertificateSize returns the size of the certificate data in bytes.
// func (c *Certificate) GetCertificateSize() int {
// 	return len([]byte(c.Data))
// }
