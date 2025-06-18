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
