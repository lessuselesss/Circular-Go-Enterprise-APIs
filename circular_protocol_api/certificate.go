package circular_protocol_api

import (
	"encoding/json"
)

// Certificate holds the data for a data certificate.
type Certificate struct {
	Data string `json:"data"` // Data content of the certificate
}

// SetData sets the data content of the certificate.
func (c *Certificate) SetData(data string) {
	c.Data = data
}

// GetData retrieves the data content of the certificate.
func (c *Certificate) GetData() string {
	return c.Data
}

// GetJSONCertificate returns the certificate as a JSON string.
func (c *Certificate) GetJSONCertificate() (string, error) {
	jsonData, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// GetCertificateSize returns the size of the certificate data in bytes.
func (c *Certificate) GetCertificateSize() int {
	return len([]byte(c.Data))
}
