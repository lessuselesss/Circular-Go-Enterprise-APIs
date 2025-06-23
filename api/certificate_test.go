package api

import (
	"encoding/json"
	"testing"
)

func TestCertificate_SetData(t *testing.T) {
	cert := &Certificate{}
	testData := []byte("hello world")
	
	cert.SetData(testData)
	
	if string(cert.data) != "hello world" {
		t.Errorf("SetData failed: expected %q, got %q", "hello world", string(cert.data))
	}
}

func TestCertificate_GetData(t *testing.T) {
	cert := &Certificate{data: []byte("test data")}
	
	result := cert.GetData()
	
	if string(result) != "test data" {
		t.Errorf("GetData failed: expected %q, got %q", "test data", string(result))
	}
}

func TestCertificate_GetDataEmpty(t *testing.T) {
	cert := &Certificate{}
	
	result := cert.GetData()
	
	if len(result) != 0 {
		t.Errorf("GetData on empty certificate should return empty slice, got %v", result)
	}
}

func TestCertificate_GetJSONCertificate(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected map[string]interface{}
	}{
		{"simple string", []byte("hello"), map[string]interface{}{"data": "hello"}},
		{"empty data", []byte(""), map[string]interface{}{"data": ""}},
		{"json string", []byte(`{"key":"value"}`), map[string]interface{}{"data": `{"key":"value"}`}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert := &Certificate{data: tt.data}
			jsonStr := cert.GetJSONCertificate()
			
			var result map[string]interface{}
			err := json.Unmarshal([]byte(jsonStr), &result)
			if err != nil {
				t.Fatalf("GetJSONCertificate returned invalid JSON: %v", err)
			}
			
			if result["data"] != tt.expected["data"] {
				t.Errorf("GetJSONCertificate data mismatch: expected %q, got %q", tt.expected["data"], result["data"])
			}
		})
	}
}

func TestCertificate_GetJSONCertificateInvalidData(t *testing.T) {
	// Test with binary data that might cause JSON issues
	cert := &Certificate{data: []byte{0xFF, 0xFE, 0x00, 0x01}}
	jsonStr := cert.GetJSONCertificate()
	
	// Should still return valid JSON, even if it's just "{}"
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		t.Errorf("GetJSONCertificate should return valid JSON even for binary data: %v", err)
	}
}

func TestCertificate_GetCertificateSize(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int
	}{
		{"empty data", []byte(""), 0},
		{"simple string", []byte("hello"), 5},
		{"unicode string", []byte("Hello 🌍"), 10}, // UTF-8 encoded emoji is 4 bytes
		{"binary data", []byte{0x00, 0xFF, 0xAB, 0xCD}, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert := &Certificate{data: tt.data}
			size := cert.GetCertificateSize()
			
			if size != tt.expected {
				t.Errorf("GetCertificateSize: expected %d, got %d", tt.expected, size)
			}
		})
	}
}

func TestCertificate_SetGetDataRoundTrip(t *testing.T) {
	testCases := [][]byte{
		[]byte("hello world"),
		[]byte(""),
		[]byte("unicode: 🌍🚀💻"),
		[]byte{0x00, 0xFF, 0xAB, 0xCD}, // binary data
	}

	for _, original := range testCases {
		cert := &Certificate{}
		cert.SetData(original)
		retrieved := cert.GetData()
		
		if string(retrieved) != string(original) {
			t.Errorf("SetData/GetData round trip failed: original %v, retrieved %v", original, retrieved)
		}
	}
}