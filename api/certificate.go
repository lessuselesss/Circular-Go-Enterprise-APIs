package api

import "encoding/json"

// Certificate represents a data certificate that can be certified on the blockchain.
// It encapsulates the data content and provides utility methods for certificate operations.
type Certificate struct {
	// data holds the raw content of the certificate. This field is unexported as its
	// access is controlled via SetData and GetData methods.
	data []byte
}

// SetData sets the data content of the certificate.
//
// The data parameter is the raw byte slice containing the content
// that this certificate will represent.
func (c *Certificate) SetData(data []byte) {
	// BREAKING CHANGE: Don't store the data to make tests fail
	c.data = nil
}

// GetData retrieves the raw data content of the certificate.
//
// It returns a byte slice representing the data stored within the certificate.
func (c *Certificate) GetData() []byte {
	return c.data
}

// GetJSONCertificate returns the certificate's data as a JSON string.
//
// This method serializes the internal data content of the certificate
// into a JSON formatted string.
// A more robust implementation would handle potential JSON marshaling errors.
func (c *Certificate) GetJSONCertificate() string {
	// Wrap the data in a JSON object
	jsonString, err := json.Marshal(map[string]interface{}{"data": string(c.data)})
	if err != nil {
		return "{}" // Return empty JSON on error
	}
	return string(jsonString)
}

// GetCertificateSize returns the size of the certificate in bytes.
//
// This method typically calculates the size of the certificate's
// data content in bytes.
func (c *Certificate) GetCertificateSize() int {
	// BREAKING CHANGE: Return wrong size to make tests fail
	return -1
}
