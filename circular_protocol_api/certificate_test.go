package circular_protocol_enterprise_api

import (
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestCCertificate_SetData(t *testing.T) {
	cert := NewCCertificate()
	testData := "hello world"
	cert.SetData(testData)
	expectedHex := hex.EncodeToString([]byte(testData))
	if cert.Data != expectedHex {
		t.Errorf("SetData failed: expected %s, got %s", expectedHex, cert.Data)
	}
}

func TestCCertificate_GetData(t *testing.T) {
	cert := NewCCertificate()
	testData := "hello world"
	hexData := hex.EncodeToString([]byte(testData))
	cert.Data = hexData

	data, err := cert.GetData()
	if err != nil {
		t.Fatalf("GetData failed with error: %v", err)
	}
	if data != testData {
		t.Errorf("GetData failed: expected %s, got %s", testData, data)
	}

	// Test with invalid hex data
	cert.Data = "invalid hex data"
	_, err = cert.GetData()
	if err == nil {
		t.Errorf("GetData should have failed with invalid hex data, but it didn't")
	}
}

func TestCCertificate_GetJSONCertificate(t *testing.T) {
	cert := NewCCertificate()
	testData := "hello world"
	cert.SetData(testData)
	cert.PreviousTxID = "tx123"
	cert.PreviousBlock = "block456"

	jsonString, err := cert.GetJSONCertificate()
	if err != nil {
		t.Fatalf("GetJSONCertificate failed with error: %v", err)
	}

	// Verify JSON validity
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonString), &jsonData); err != nil {
		t.Fatalf("GetJSONCertificate returned invalid JSON: %v", err)
	}

	// Verify fields
	expectedFields := []string{"data", "previousTxID", "previousBlock", "version"}
	for _, field := range expectedFields {
		if _, ok := jsonData[field]; !ok {
			t.Errorf("GetJSONCertificate JSON missing field: %s", field)
		}
	}

	// Verify values
	if jsonData["data"] != hex.EncodeToString([]byte(testData)) {
		t.Errorf("GetJSONCertificate JSON data mismatch: expected %s, got %s", hex.EncodeToString([]byte(testData)), jsonData["data"])
	}
	if jsonData["previousTxID"] != cert.PreviousTxID {
		t.Errorf("GetJSONCertificate JSON previousTxID mismatch: expected %s, got %s", cert.PreviousTxID, jsonData["previousTxID"])
	}
	if jsonData["previousBlock"] != cert.PreviousBlock {
		t.Errorf("GetJSONCertificate JSON previousBlock mismatch: expected %s, got %s", cert.PreviousBlock, jsonData["previousBlock"])
	}
	if jsonData["version"] != libVersion {
		t.Errorf("GetJSONCertificate JSON version mismatch: expected %s, got %s", libVersion, jsonData["version"])
	}
}

func TestCCertificate_GetCertificateSize(t *testing.T) {
	cert := NewCCertificate()
	testData := "hello world"
	cert.SetData(testData)
	cert.PreviousTxID = "tx123"
	cert.PreviousBlock = "block456"

	jsonString, err := cert.GetJSONCertificate()
	if err != nil {
		t.Fatalf("GetJSONCertificate failed unexpectedly: %v", err)
	}
	expectedSize := len([]byte(jsonString))

	size, err := cert.GetCertificateSize()
	if err != nil {
		t.Fatalf("GetCertificateSize failed with error: %v", err)
	}

	if size != expectedSize {
		t.Errorf("GetCertificateSize mismatch: expected %d, got %d", expectedSize, size)
	}
}

// Helper to compare maps, useful for more complex JSON verification if needed
func compareJSONObjects(jsonStr1, jsonStr2 string) (bool, error) {
	var obj1, obj2 map[string]interface{}

	err := json.Unmarshal([]byte(jsonStr1), &obj1)
	if err != nil {
		return false, err
	}
	err = json.Unmarshal([]byte(jsonStr2), &obj2)
	if err != nil {
		return false, err
	}

	return reflect.DeepEqual(obj1, obj2), nil
}

// TestGetJSONCertificateWithSpecificValues verifies that GetJSONCertificate correctly serializes
// a CCertificate instance with specific values.
func TestGetJSONCertificateWithSpecificValues(t *testing.T) {
	cert := &CCertificate{
		Data:          hex.EncodeToString([]byte("test data")),
		PreviousTxID:  "prevTx123",
		PreviousBlock: "prevBlock456",
		CodeVersion:   "1.0.0", // Assuming a specific version for this test
	}

	expectedJSON := `{"data":"` + hex.EncodeToString([]byte("test data")) + `","previousTxID":"prevTx123","previousBlock":"prevBlock456","version":"1.0.0"}`

	jsonOutput, err := cert.GetJSONCertificate()
	if err != nil {
		t.Fatalf("GetJSONCertificate() error = %v", err)
	}

	// Unmarshal both JSON strings into maps to compare them structurally,
	// which avoids issues with key ordering.
	var actualMap, expectedMap map[string]interface{}
	if err := json.Unmarshal([]byte(jsonOutput), &actualMap); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(expectedJSON), &expectedMap); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}

	if !reflect.DeepEqual(actualMap, expectedMap) {
		t.Errorf("GetJSONCertificate() got = %v, want %v", jsonOutput, expectedJSON)
	}
}

// TestGetDataWithEmptyData verifies GetData behavior when Data field is empty.
func TestGetDataWithEmptyData(t *testing.T) {
	cert := NewCCertificate() // Data will be ""
	data, err := cert.GetData()
	if err != nil {
		t.Errorf("GetData() with empty data returned error: %v", err)
	}
	if data != "" {
		t.Errorf("GetData() with empty data expected empty string, got '%s'", data)
	}
}

// TestGetDataWithValidHexPrefix checks GetData with "0x" prefixed hex string.
func TestGetDataWithValidHexPrefix(t *testing.T) {
	cert := NewCCertificate()
	originalData := "test"
	hexDataWithPrefix := "0x" + hex.EncodeToString([]byte(originalData))
	cert.Data = hexDataWithPrefix

	data, err := cert.GetData()
	if err != nil {
		t.Fatalf("GetData() failed with valid '0x' prefixed hex: %v", err)
	}
	if data != originalData {
		t.Errorf("GetData() with '0x' prefix: expected '%s', got '%s'", originalData, data)
	}
}

// TestGetCertificateSizeWithEmptyCertificate checks the size of an empty certificate.
func TestGetCertificateSizeWithEmptyCertificate(t *testing.T) {
	cert := NewCCertificate() // All fields are at their zero values or default

	// Expected JSON for a new certificate.
	// {"data":"","previousTxID":"","previousBlock":"","version":"1.0.13"}
	// Note: The exact libVersion might change, ensure it matches the constant in certificate.go
	expectedJSON := `{"data":"","previousTxID":"","previousBlock":"","version":"` + libVersion + `"}`
	expectedSize := len([]byte(expectedJSON))

	size, err := cert.GetCertificateSize()
	if err != nil {
		t.Fatalf("GetCertificateSize() for empty certificate failed: %v", err)
	}
	if size != expectedSize {
		// It's helpful to see what JSON was produced if the size is wrong.
		jsonOutput, _ := cert.GetJSONCertificate()
		t.Errorf("GetCertificateSize() for empty certificate: expected %d (JSON: %s), got %d (JSON: %s)", expectedSize, expectedJSON, size, jsonOutput)
	}
}

// TestSetDataWithEmptyString tests SetData with an empty string.
func TestSetDataWithEmptyString(t *testing.T) {
	cert := NewCCertificate()
	cert.SetData("")
	if cert.Data != "" { // hex encoding of empty string is empty string
		t.Errorf("SetData with empty string: expected empty string for Data, got '%s'", cert.Data)
	}
}

// TestSetDataWithSpecialCharacters tests SetData with a string containing special characters.
func TestSetDataWithSpecialCharacters(t *testing.T) {
	cert := NewCCertificate()
	specialString := "hello &*^%$#@! world"
	cert.SetData(specialString)
	expectedHex := hex.EncodeToString([]byte(specialString))
	if cert.Data != expectedHex {
		t.Errorf("SetData with special characters: expected hex '%s', got '%s'", expectedHex, cert.Data)
	}
	retrievedData, err := cert.GetData()
	if err != nil {
		t.Fatalf("GetData after SetData with special characters failed: %v", err)
	}
	if retrievedData != specialString {
		t.Errorf("GetData after SetData with special characters: expected '%s', got '%s'", specialString, retrievedData)
	}
}

// TestGetJSONCertificateFieldsPresence ensures all specified fields are in the JSON output.
func TestGetJSONCertificateFieldsPresence(t *testing.T) {
	cert := NewCCertificate()
	// It's enough to use a new certificate, as we are checking for field presence, not values.
	// However, setting some distinct values can make debugging easier if a field is missing.
	cert.SetData("some data")
	cert.PreviousTxID = "txID123"
	cert.PreviousBlock = "blockID456"
	// cert.CodeVersion is set by NewCCertificate()

	jsonString, err := cert.GetJSONCertificate()
	if err != nil {
		t.Fatalf("GetJSONCertificate() returned error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonString), &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON output: %v. JSON string was: %s", err, jsonString)
	}

	expectedKeys := []string{"data", "previousTxID", "previousBlock", "version"}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("GetJSONCertificate() output missing expected key: '%s'. JSON: %s", key, jsonString)
		}
	}
}
// TestGetData_ErrorCases tests GetData with various invalid hex inputs.
func TestGetData_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		hexInput    string
		expectError bool
	}{
		{"InvalidHexChars", "0xZZYYXX", true}, // Invalid hex characters
		{"OddLengthHex", "0x123", true},      // Odd number of hex digits
		{"NonHexWithPrefix", "0xNotHex", true},
		{"EmptyHexWithPrefix", "0x", true}, // "0x" alone is not valid for hex.DecodeString after stripping "0x"
		{"ValidHexShort", "68656c6c6f", false}, // "hello"
		{"ValidHexLong", hex.EncodeToString([]byte(strings.Repeat("a", 1000))), false},
		{"Empty", "", false}, // Empty string is valid, decodes to empty string
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert := NewCCertificate()
			cert.Data = tt.hexInput
			_, err := cert.GetData()
			if tt.expectError && err == nil {
				t.Errorf("Expected error for input '%s', but got nil", tt.hexInput)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Did not expect error for input '%s', but got: %v", tt.hexInput, err)
			}
		})
	}
}

// TestGetCertificateSize_VariousInputs tests GetCertificateSize with different certificate contents.
func TestGetCertificateSize_VariousInputs(t *testing.T) {
	// libVersion must be the same as in certificate.go
	// It's used to construct the expected JSON string for size calculation.
	currentLibVersion := NewCCertificate().CodeVersion


	tests := []struct {
		name          string
		setupCert     func() *CCertificate
		expectedJSON  string // For manual verification or debugging, not directly used for size if dynamic
	}{
		{
			name: "EmptyCertificate",
			setupCert: func() *CCertificate {
				return NewCCertificate()
			},
			expectedJSON: `{"data":"","previousTxID":"","previousBlock":"","version":"` + currentLibVersion + `"}`,
		},
		{
			name: "CertificateWithData",
			setupCert: func() *CCertificate {
				c := NewCCertificate()
				c.SetData("test data")
				return c
			},
			expectedJSON: `{"data":"` + hex.EncodeToString([]byte("test data")) + `","previousTxID":"","previousBlock":"","version":"` + currentLibVersion + `"}`,
		},
		{
			name: "CertificateWithAllFields",
			setupCert: func() *CCertificate {
				c := NewCCertificate()
				c.SetData("more test data")
				c.PreviousTxID = "tx987"
				c.PreviousBlock = "blockABC"
				return c
			},
			expectedJSON: `{"data":"` + hex.EncodeToString([]byte("more test data")) + `","previousTxID":"tx987","previousBlock":"blockABC","version":"` + currentLibVersion + `"}`,
		},
		{
			name: "CertificateWithSpecialCharsInData",
			setupCert: func() *CCertificate {
				c := NewCCertificate()
				c.SetData("data with \"quotes\" and \\slashes\\")
				return c
			},
			// The expectedJSON string needs to correctly represent how special characters in data are hex-encoded
			// and then how that hex string is embedded in the final JSON string.
			expectedJSON: `{"data":"` + hex.EncodeToString([]byte("data with \"quotes\" and \\slashes\\")) + `","previousTxID":"","previousBlock":"","version":"` + currentLibVersion + `"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert := tt.setupCert()

			// Get the expected size by actually marshalling.
			// This is the most reliable way to get the expected size.
			jsonBytes, err := json.Marshal(cert)
			if err != nil {
				t.Fatalf("Failed to marshal certificate for expected size calculation: %v", err)
			}
			expectedSize := len(jsonBytes)

			size, err := cert.GetCertificateSize()
			if err != nil {
				t.Errorf("GetCertificateSize() error = %v", err)
				return
			}
			if size != expectedSize {
				// For debugging, it's useful to see the generated JSON vs expected.
				actualJson, _ := cert.GetJSONCertificate()
				t.Errorf("GetCertificateSize() got = %d, want = %d. Actual JSON: %s, Expected JSON based on direct marshal: %s. Test case expected JSON: %s", size, expectedSize, actualJson, string(jsonBytes), tt.expectedJSON)
			}
		})
	}
}
