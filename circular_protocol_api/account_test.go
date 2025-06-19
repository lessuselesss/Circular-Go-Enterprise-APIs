package circular_protocol_enterprise_api

import (
	// "bytes" // Removed as unused
	"encoding/json"
	"encoding/hex" // Added for SignData test
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	// "time" // Removed as unused
)

// mockHTTPResponse represents a mock HTTP response for a specific endpoint condition.
type mockHTTPResponse struct {
	Body       string
	StatusCode int
	Err        error // For simulating network errors from the RoundTripper itself
}

// mockEndpointResponses maps an endpoint identifier (e.g., URL suffix or a custom key)
// to a list of responses to be served in sequence.
var mockResponses map[string][]*mockHTTPResponse

// mockRequestCounter tracks how many times a mock endpoint has been called.
var mockRequestCounters map[string]int

// setupMockHTTP configures the mock responses for tests.
// Call this at the beginning of each test function that makes HTTP calls.
func setupMockHTTP() {
	mockResponses = make(map[string][]*mockHTTPResponse)
	mockRequestCounters = make(map[string]int)

	originalTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = roundTripFunc(func(req *http.Request) *http.Response {
		// Try to identify the endpoint based on URL suffix
		var endpointKey string
		urlStr := req.URL.String()

		// Specific for SetNetwork (GET request)
		if strings.Contains(urlStr, "/getNAG?network=") {
			// Extract network name to use as part of the key if needed, or use a generic key
			queryParts := strings.Split(req.URL.RawQuery, "=")
			if len(queryParts) == 2 && queryParts[0] == "network" {
				endpointKey = "/getNAG?network=" + queryParts[1]
			} else {
				endpointKey = urlStr // Fallback
			}
		} else if strings.Contains(urlStr, "_GetWalletNonce_") {
			endpointKey = "_GetWalletNonce_"
		} else if strings.Contains(urlStr, "_GetTransactionbyID_") {
			endpointKey = "_GetTransactionbyID_"
		} else if strings.Contains(urlStr, "_AddTransaction_") {
			endpointKey = "_AddTransaction_"
		} else {
			// Fallback for unexpected requests
			return &http.Response{
				StatusCode: http.StatusNotImplemented,
				Body:       io.NopCloser(strings.NewReader(fmt.Sprintf("No mock configured for URL: %s", urlStr))),
				Header:     make(http.Header),
			}
		}

		responses, ok := mockResponses[endpointKey]
		if !ok || len(responses) == 0 {
			return &http.Response{
				StatusCode: http.StatusNotImplemented,
				Body:       io.NopCloser(strings.NewReader(fmt.Sprintf("No mock responses defined for key: %s (URL: %s)", endpointKey, urlStr))),
				Header:     make(http.Header),
			}
		}

		counter := mockRequestCounters[endpointKey]
		if counter >= len(responses) {
			// Reuse last response if counter exceeds available responses (useful for polling "Pending")
			counter = len(responses) - 1
		}

		respData := responses[counter]
		mockRequestCounters[endpointKey]++

		if respData.Err != nil {
			// This part is tricky as RoundTrip itself returns an error, not an *http.Response with an error.
			// For now, we'll assume respData.Err means a problem creating the response,
			// not a network error to be returned by the RoundTripper itself.
			// To simulate network errors, the RoundTripper should return an error directly.
			// This needs refinement if direct network errors from RoundTripper are to be simulated.
			// For now, we use StatusCode to signal HTTP errors.
		}

		headers := make(http.Header)
		headers.Set("Content-Type", "application/json")
		return &http.Response{
			StatusCode: respData.StatusCode,
			Body:       io.NopCloser(strings.NewReader(respData.Body)),
			Header:     headers,
		}
	})
	// Return a teardown function to restore the original transport
	// This is typically done with `defer teardown()` in the test function.
	// However, since we are setting this for multiple tests and subtests,
	// managing teardown carefully is important.
	// For simplicity here, we assume tests run sequentially or http.DefaultClient is managed per test package.
	// A better approach for parallel tests would be custom clients per test.
	// For now, we'll rely on a single test file running sequentially or careful test structuring.
	// The `defer` in each test function is the safest.
	_ = originalTransport // Keep a reference to it if needed for a package-level teardown
}

// addMockResponse adds a response for a given endpoint key.
func addMockResponse(key string, statusCode int, body string) {
	if _, ok := mockResponses[key]; !ok {
		mockResponses[key] = []*mockHTTPResponse{}
	}
	mockResponses[key] = append(mockResponses[key], &mockHTTPResponse{Body: body, StatusCode: statusCode})
}

// Helper to create a standard success JSON response for NAG
func nagSuccessResponse(dataContent string) string {
	return fmt.Sprintf(`{"Result": 200, "Response": %s}`, dataContent)
}

// Helper to create a standard error JSON response for NAG
func nagErrorResponse(resultCode int, message string) string {
	return fmt.Sprintf(`{"Result": %d, "Message": "%s"}`, resultCode, message)
}


func TestCEPAccount_Open(t *testing.T) {
	acc := NewCEPAccount()

	// Test with a valid address
	validAddress := "test_address"
	err := acc.Open(validAddress)
	if err != nil {
		t.Errorf("Open() with valid address failed: %v", err)
	}
	if acc.Address != validAddress {
		t.Errorf("Open() did not set address correctly: expected %s, got %s", validAddress, acc.Address)
	}

	// Test with an empty address
	err = acc.Open("")
	if err == nil {
		t.Errorf("Open() with empty address should have returned an error, but did not")
	}
}

func TestCEPAccount_SetBlockchain(t *testing.T) {
	acc := NewCEPAccount()
	testChain := "test_chain_id_123"
	acc.SetBlockchain(testChain)
	if acc.Blockchain != testChain {
		t.Errorf("SetBlockchain() failed: expected %s, got %s", testChain, acc.Blockchain)
	}
}

func TestCEPAccount_Close(t *testing.T) {
	acc := NewCEPAccount()

	// Setup with non-default values
	acc.Address = "some_address"
	acc.PublicKey = "some_public_key"
	acc.Info = "some_info"
	acc.NAGURL = "http://custom-nag.com"
	acc.NetworkNode = "custom_node"
	acc.Blockchain = "custom_blockchain"
	acc.LatestTxID = "tx_id_1"
	acc.Nonce = 10
	acc.Data = map[string]interface{}{"key": "value"}
	acc.IntervalSec = 5

	acc.Close()

	if acc.Address != "" { t.Errorf("Close() Address: got %s, want \"\"", acc.Address) }
	if acc.PublicKey != "" { t.Errorf("Close() PublicKey: got %s, want \"\"", acc.PublicKey) }
	if acc.Info != "" { t.Errorf("Close() Info: got %s, want \"\"", acc.Info) }
	if acc.NAGURL != defaultNAG { t.Errorf("Close() NAGURL: got %s, want %s", acc.NAGURL, defaultNAG) }
	if acc.NetworkNode != "" { t.Errorf("Close() NetworkNode: got %s, want \"\"", acc.NetworkNode) }
	if acc.Blockchain != defaultChain { t.Errorf("Close() Blockchain: got %s, want %s", acc.Blockchain, defaultChain) }
	if acc.LatestTxID != "" { t.Errorf("Close() LatestTxID: got %s, want \"\"", acc.LatestTxID) }
	if acc.Nonce != 0 { t.Errorf("Close() Nonce: got %d, want 0", acc.Nonce) }
	if len(acc.Data) != 0 { t.Errorf("Close() Data: got %v, want empty map", acc.Data) }
	if acc.IntervalSec != 2 { t.Errorf("Close() IntervalSec: got %d, want 2", acc.IntervalSec) }
	if acc.CodeVersion != libVersion { t.Errorf("Close() CodeVersion: got %s, want %s (should not change)", acc.CodeVersion, libVersion) }
	if acc.LastError != "" { t.Errorf("Close() LastError: got %s, want \"\"", acc.LastError) }
}

func TestCEPAccount_SetNetwork_WithMock(t *testing.T) {
	originalTransport := http.DefaultClient.Transport
	setupMockHTTP() // Sets up the mock transport
	defer func() { http.DefaultClient.Transport = originalTransport }()

	tests := []struct {
		name           string
		networkName    string // Used to form the endpoint key for mockResponses
		mockStatusCode int
		mockBody       string
		expectedNAGURL string
		expectError    bool
		expectedErrMsg string
	}{
		{"Success", "test_ok", http.StatusOK, `{"status": "success", "url": "http://new-mock-nag.com"}`, "http://new-mock-nag.com", false, ""},
		{"HTTPError", "test_http_err", http.StatusInternalServerError, "HTTP Error", "", true, "HTTP error! status: 500"},
		{"BadJSON", "test_bad_json", http.StatusOK, `{"status": "success", "url": "http://new-mock-nag.com"`, "", true, "unexpected EOF"},
		{"LogicError", "test_logic_err", http.StatusOK, `{"status": "failure", "message": "Specific Error"}`, "", true, "Specific Error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := NewCEPAccount()
			initialNAGURL := acc.NAGURL // Default NAG URL

			// Key for SetNetwork is specific including networkName
			mockKey := "/getNAG?network=" + tt.networkName
			addMockResponse(mockKey, tt.mockStatusCode, tt.mockBody)

			err := acc.SetNetwork(tt.networkName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				} else if tt.expectedErrMsg != "" && !strings.Contains(err.Error(), tt.expectedErrMsg) {
					t.Errorf("Error message mismatch: got '%s', expected to contain '%s'", err.Error(), tt.expectedErrMsg)
				}
				if acc.NAGURL != initialNAGURL { // NAGURL should not change on error
					t.Errorf("NAGURL changed on error: got %s, expected %s", acc.NAGURL, initialNAGURL)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if acc.NAGURL != tt.expectedNAGURL {
					t.Errorf("NAGURL mismatch: got %s, want %s", acc.NAGURL, tt.expectedNAGURL)
				}
			}
			// Clear responses for this specific key for next sub-test if any, or rely on full setupMockHTTP for next top-level test
			delete(mockResponses, mockKey)
			delete(mockRequestCounters, mockKey)
		})
	}
}


func TestCEPAccount_UpdateAccount(t *testing.T) {
	originalTransport := http.DefaultClient.Transport
	setupMockHTTP()
	defer func() { http.DefaultClient.Transport = originalTransport }()

	acc := NewCEPAccount()
	acc.Open("test_address_for_update") // Open account

	// Case 1: Success
	t.Run("Success", func(t *testing.T) {
		mockResponses["_GetWalletNonce_"] = nil // Clear previous mocks for this key
		mockRequestCounters["_GetWalletNonce_"] = 0
		addMockResponse("_GetWalletNonce_", http.StatusOK, nagSuccessResponse(`{"Nonce": 123}`))
		success, err := acc.UpdateAccount()
		if err != nil { t.Fatalf("Expected no error, got %v", err) }
		if !success { t.Fatal("Expected success true, got false") }
		if acc.Nonce != 124 { t.Errorf("Expected Nonce 124, got %d", acc.Nonce) }
		mockRequestCounters["_GetWalletNonce_"] = 0 // Reset counter for this endpoint
	})

	// Case 2: NAG Error
	t.Run("NAGError", func(t *testing.T) {
		mockResponses["_GetWalletNonce_"] = nil // Clear previous mocks for this key
		mockRequestCounters["_GetWalletNonce_"] = 0
		addMockResponse("_GetWalletNonce_", http.StatusOK, nagErrorResponse(400, "Bad Request"))
		success, err := acc.UpdateAccount()
		if err == nil { t.Fatal("Expected error, got nil") }
		if success { t.Fatal("Expected success false, got true") }
		if !strings.Contains(err.Error(), "invalid response format or missing Nonce field") && !strings.Contains(err.Error(), "Bad Request") { // Error can be generic or specific
			t.Errorf("Unexpected error message: %s", err.Error())
		}
		mockRequestCounters["_GetWalletNonce_"] = 0
	})

	// Case 3: HTTP Error
	t.Run("HTTPError", func(t *testing.T) {
		mockResponses["_GetWalletNonce_"] = nil // Clear previous mocks for this key
		mockRequestCounters["_GetWalletNonce_"] = 0
		addMockResponse("_GetWalletNonce_", http.StatusInternalServerError, "Internal Server Error")
		success, err := acc.UpdateAccount()
		if err == nil { t.Fatal("Expected error, got nil") }
		if success { t.Fatal("Expected success false, got true") }
		if !strings.Contains(err.Error(), "HTTP error! status: 500") {
			t.Errorf("Unexpected error message for HTTP error: %s", err.Error())
		}
		mockRequestCounters["_GetWalletNonce_"] = 0
	})

	// Case 4: Malformed JSON
	t.Run("MalformedJSON", func(t *testing.T) {
		mockResponses["_GetWalletNonce_"] = nil // Clear previous mocks for this key
		mockRequestCounters["_GetWalletNonce_"] = 0
		addMockResponse("_GetWalletNonce_", http.StatusOK, `{"Result": 200, "Response": {"Nonce": 123`) // Malformed
		success, err := acc.UpdateAccount()
		if err == nil { t.Fatal("Expected error from malformed JSON, got nil") }
		if success { t.Fatal("Expected success false with malformed JSON, got true") }
		mockRequestCounters["_GetWalletNonce_"] = 0
	})

	// Case 5: Account Not Open
	t.Run("AccountNotOpen", func(t *testing.T) {
		closedAcc := NewCEPAccount() // Fresh, unopened account
		success, err := closedAcc.UpdateAccount()
		if err == nil { t.Fatal("Expected error for unopened account, got nil") }
		if success { t.Fatal("Expected success false for unopened account, got true") }
		if err.Error() != "account is not open" { t.Errorf("Unexpected error for unopened account: %s", err.Error())}
	})
}

func TestCEPAccount_SignData(t *testing.T) {
	acc := NewCEPAccount()

	// Case 1: Success
	t.Run("Success", func(t *testing.T) {
		acc.Open("test_sign_address")
		// Example private key (replace with a real one if specific signature is needed for verification)
		// This is a dummy one, 32 bytes, hex encoded.
		// Corresponds to public key: 04d0d1a9a72f7a8949a8776976ac0b3c240dda76399a5903541b4495417760098509178990284597951099993925560493a8704218f909607370f9c05ba1080ff3
		privateKeyHex := "21e1a76903c9200ef047197e1f314e270a697539a490791782610500a202b586"
		signature, err := acc.SignData("some data to sign", privateKeyHex)
		if err != nil { t.Fatalf("Expected no error, got %v", err) }
		if signature == "" { t.Error("Expected a signature string, got empty") }
		// Basic check for hex string
		if _, decodeErr := hex.DecodeString(signature); decodeErr != nil {
			t.Errorf("Signature is not a valid hex string: %v", decodeErr)
		}
	})

	// Case 2: Invalid Private Key
	t.Run("InvalidPrivateKey", func(t *testing.T) {
		acc.Open("test_sign_address_bad_key")
		_, err := acc.SignData("some data", "invalid-hex-key")
		if err == nil { t.Fatal("Expected error for invalid private key, got nil") }
	})

	// Case 3: Account Not Open
	t.Run("AccountNotOpen", func(t *testing.T) {
		closedAcc := NewCEPAccount()
		_, err := closedAcc.SignData("some data", "21e1a76903c9200ef047197e1f314e270a697539a490791782610500a202b586")
		if err == nil { t.Fatal("Expected error for unopened account, got nil") }
		if err.Error() != "account is not open" { t.Errorf("Unexpected error for unopened account: %s", err.Error())}
	})
}

func TestCEPAccount_GetTransactionbyID(t *testing.T) {
	originalTransport := http.DefaultClient.Transport
	setupMockHTTP()
	defer func() { http.DefaultClient.Transport = originalTransport }()

	acc := NewCEPAccount()
	acc.Open("test_address_for_gettx")

	// Case 1: Success
	t.Run("Success", func(t *testing.T) {
		mockResponses["_GetTransactionbyID_"] = nil // Clear previous mocks for this key
		mockRequestCounters["_GetTransactionbyID_"] = 0
		// expectedTxData removed as it was unused
		mockRespBody := nagSuccessResponse(`{"tx_id": "test_txid", "data": "some_payload"}`)
		addMockResponse("_GetTransactionbyID_", http.StatusOK, mockRespBody)

		txData, err := acc.GetTransactionbyID("test_txid", 0, 10)
		if err != nil { t.Fatalf("Expected no error, got %v", err) }
		// The function returns the entire NAG response, so we expect {"Result": 200, "Response": expectedTxData}
		// Let's adjust expectedTxData or the check. The function returns jsonResponse (raw).
		var fullExpectedResponse map[string]interface{}
		json.Unmarshal([]byte(mockRespBody), &fullExpectedResponse)

		if !reflect.DeepEqual(txData, fullExpectedResponse) {
			t.Errorf("Transaction data mismatch: got %v, want %v", txData, fullExpectedResponse)
		}
		mockRequestCounters["_GetTransactionbyID_"] = 0
	})

	// Case 2: NAG Error / Not Found
	t.Run("NAGErrorNotFound", func(t *testing.T) {
		mockResponses["_GetTransactionbyID_"] = nil // Clear previous mocks for this key
		mockRequestCounters["_GetTransactionbyID_"] = 0
		addMockResponse("_GetTransactionbyID_", http.StatusOK, nagErrorResponse(404, "Transaction not found"))
		txData, err := acc.GetTransactionbyID("unknown_txid", 0, 10)
		// The current GetTransactionbyID doesn't return an error for NAG errors, it returns the JSON.
		// This might be something to refine in the main code.
		// For now, we test current behavior: err is nil, txData contains the error JSON.
		if err != nil { t.Fatalf("Expected nil error for NAG error response, got %v", err) }
		if txData == nil {t.Fatal("Expected error response from NAG, got nil map")}
		resultVal, ok := txData["Result"].(float64)
		if !ok {
			t.Errorf("Result field is not a float64 or not present in NAG error response: %v", txData)
		} else if resultVal == 200 {
			t.Errorf("Expected non-200 NAG result code in error case, got 200. Full response: %v", txData)
		}
		mockRequestCounters["_GetTransactionbyID_"] = 0
	})

	// Case 3: HTTP Error
	t.Run("HTTPError", func(t *testing.T) {
		mockResponses["_GetTransactionbyID_"] = nil // Clear previous mocks for this key
		mockRequestCounters["_GetTransactionbyID_"] = 0
		addMockResponse("_GetTransactionbyID_", http.StatusInternalServerError, "HTTP Internal Error")
		_, err := acc.GetTransactionbyID("any_txid", 0, 10)
		if err == nil { t.Fatal("Expected HTTP error, got nil") }
		if !strings.Contains(err.Error(), "network response was not ok") { // Error from main code
			t.Errorf("Unexpected error for HTTP error: %s", err.Error())
		}
		mockRequestCounters["_GetTransactionbyID_"] = 0
	})
}


func TestCEPAccount_SubmitCertificate(t *testing.T) {
	originalTransport := http.DefaultClient.Transport
	setupMockHTTP()
	defer func() { http.DefaultClient.Transport = originalTransport }()

	validPrivateKey := "21e1a76903c9200ef047197e1f314e270a697539a490791782610500a202b586"

	// Case 1: Success
	t.Run("Success", func(t *testing.T) {
		acc := NewCEPAccount()
		acc.Open("test_submit_cert_addr")
		acc.Nonce = 5 // Pre-set nonce as UpdateAccount is not called here
		mockResponses["_AddTransaction_"] = nil // Clear previous mocks for this key
		mockRequestCounters["_AddTransaction_"] = 0
		mockRespBody := nagSuccessResponse(`{"TxID": "new_tx_id_123"}`)
		addMockResponse("_AddTransaction_", http.StatusOK, mockRespBody)

		responseData, err := acc.SubmitCertificate("my certificate data", validPrivateKey)
		if err != nil { t.Fatalf("Expected no error, got %v", err) }

		var expectedFullResponse map[string]interface{}
		json.Unmarshal([]byte(mockRespBody), &expectedFullResponse)
		if !reflect.DeepEqual(responseData, expectedFullResponse) {
			t.Errorf("Response data mismatch: got %v, want %v", responseData, expectedFullResponse)
		}
		mockRequestCounters["_AddTransaction_"] = 0
	})

	// Case 2: Account Not Open
	t.Run("AccountNotOpen", func(t *testing.T) {
		closedAcc := NewCEPAccount()
		_, err := closedAcc.SubmitCertificate("some data", validPrivateKey)
		if err == nil { t.Fatal("Expected error for unopened account, got nil") }
		if err.Error() != "account is not open" { t.Errorf("Unexpected error: %s", err.Error())}
	})

	// Case 3: SignData Error (bad private key)
	t.Run("SignDataError", func(t *testing.T) {
		acc := NewCEPAccount()
		acc.Open("test_submit_badsig_addr")
		acc.Nonce = 1
		_, err := acc.SubmitCertificate("some data", "invalid-key")
		if err == nil { t.Fatal("Expected error from SignData failure, got nil") }
		if !strings.Contains(err.Error(), "failed to decode private key") {
			t.Errorf("Expected private key decode error, got: %s", err.Error())
		}
	})

	// Case 4: NAG Error
	t.Run("NAGError", func(t *testing.T) {
		acc := NewCEPAccount()
		acc.Open("test_submit_nagerr_addr")
		acc.Nonce = 2
		addMockResponse("_AddTransaction_", http.StatusOK, nagErrorResponse(501, "NAG Processing Error"))
		_, err := acc.SubmitCertificate("cert data", validPrivateKey)
		// Similar to GetTransactionByID, current SubmitCertificate returns the NAG error JSON as data, not an SDK error.
		// This should be tested as is, or SubmitCertificate should be refactored.
		// For now, assume err is nil, and responseData contains the NAG error.
		if err != nil {t.Fatalf("Expected nil error for NAG error, got %v", err)}
		// responseData, _ := acc.SubmitCertificate("cert data", validPrivateKey) // This would re-run
		// This test needs to capture the responseData from the call above.
		// Let's re-do this sub-test structure slightly for clarity on responseData
		mockRequestCounters["_AddTransaction_"] = 0 // Reset for this sub-test
	})

	t.Run("NAGError_Redux", func(t *testing.T) {
		 acc := NewCEPAccount()
		 acc.Open("test_submit_nagerr_addr_redux")
		 acc.Nonce = 3
		 mockResponses["_AddTransaction_"] = nil // Clear previous mocks for this key
		 mockRequestCounters["_AddTransaction_"] = 0
		 addMockResponse("_AddTransaction_", http.StatusOK, nagErrorResponse(501, "NAG Processing Error"))
		 responseData, err := acc.SubmitCertificate("cert data", validPrivateKey)
		 if err != nil {t.Fatalf("Expected nil error for NAG error, got %v", err)}
		 if responseData == nil {t.Fatal("Expected NAG error response, got nil map")}
		 resultVal, ok := responseData["Result"].(float64)
		 if !ok {
			t.Errorf("Result field is not a float64 or not present in NAG error response: %v", responseData)
		 } else if resultVal == 200 {
			t.Errorf("Expected non-200 NAG result code in error case, got 200. Full response: %v", responseData)
		 }
		 mockRequestCounters["_AddTransaction_"] = 0
	})


	// Case 5: HTTP Error
	t.Run("HTTPError", func(t *testing.T) {
		acc := NewCEPAccount()
		acc.Open("test_submit_httperr_addr")
		acc.Nonce = 4
		mockResponses["_AddTransaction_"] = nil // Clear previous mocks for this key
		mockRequestCounters["_AddTransaction_"] = 0
		addMockResponse("_AddTransaction_", http.StatusBadGateway, "HTTP Bad Gateway")
		_, err := acc.SubmitCertificate("cert data", validPrivateKey)
		if err == nil { t.Fatal("Expected HTTP error, got nil") }
		if !strings.Contains(err.Error(), "network response was not ok") {
			t.Errorf("Unexpected error for HTTP error: %s", err.Error())
		}
		mockRequestCounters["_AddTransaction_"] = 0
	})
}


func TestCEPAccount_GetTransactionOutcome(t *testing.T) {
	originalTransport := http.DefaultClient.Transport
	setupMockHTTP()
	defer func() { http.DefaultClient.Transport = originalTransport }()

	acc := NewCEPAccount()
	acc.Open("test_txoutcome_addr")
	originalInterval := acc.IntervalSec
	acc.IntervalSec = 20 // Speed up polling for tests (20ms)
	defer func() { acc.IntervalSec = originalInterval }()


	// Case 1: Success (Pending -> Confirmed)
	t.Run("SuccessConfirmed", func(t *testing.T) {
		mockResponses["_GetTransactionbyID_"] = nil // Clear previous mocks for this key
		mockRequestCounters["_GetTransactionbyID_"] = 0

		pendingResp := nagSuccessResponse(`{"Status": "Pending"}`)
		confirmedRespBody := `{"Status": "Confirmed", "Detail": "Transaction was successful"}`
		confirmedFullResp := nagSuccessResponse(confirmedRespBody)

		// Add two responses for _GetTransactionbyID_: first Pending, then Confirmed
		addMockResponse("_GetTransactionbyID_", http.StatusOK, pendingResp)
		addMockResponse("_GetTransactionbyID_", http.StatusOK, confirmedFullResp)

		outcome, err := acc.GetTransactionOutcome("tx_success", 100) // 100ms timeout (5 * 20ms interval)
		if err != nil { t.Fatalf("Expected no error, got %v", err) }

		var expectedOutcome map[string]interface{}
		// We expect the "Response" part of the NAG success response
		json.Unmarshal([]byte(confirmedRespBody), &expectedOutcome)

		if !reflect.DeepEqual(outcome, expectedOutcome) {
			t.Errorf("Outcome data mismatch: got %v, want %v", outcome, expectedOutcome)
		}
		mockRequestCounters["_GetTransactionbyID_"] = 0
		mockResponses["_GetTransactionbyID_"] = nil // Clear responses for this key
	})

	// Case 2: Timeout
	t.Run("Timeout", func(t *testing.T) {
		mockResponses["_GetTransactionbyID_"] = nil // Clear previous mocks
		mockRequestCounters["_GetTransactionbyID_"] = 0
		pendingResp := nagSuccessResponse(`{"Status": "Pending"}`)
		addMockResponse("_GetTransactionbyID_", http.StatusOK, pendingResp) // Always pending

		_, err := acc.GetTransactionOutcome("tx_timeout", 50) // 50ms timeout
		if err == nil { t.Fatal("Expected timeout error, got nil") }
		if !strings.Contains(err.Error(), "timeout exceeded") {
			t.Errorf("Expected timeout error, got: %s", err.Error())
		}
		mockRequestCounters["_GetTransactionbyID_"] = 0
		mockResponses["_GetTransactionbyID_"] = nil
	})

	// Case 3: GetTransactionbyID returns HTTP error during polling
	t.Run("GetTxByID_HTTPErrorInPoll", func(t *testing.T) {
		mockResponses["_GetTransactionbyID_"] = nil // Clear previous mocks
		mockRequestCounters["_GetTransactionbyID_"] = 0
		addMockResponse("_GetTransactionbyID_", http.StatusInternalServerError, "Simulated HTTP error during poll")

		_, err := acc.GetTransactionOutcome("tx_http_error_poll", 50) // 50ms timeout
		if err == nil {t.Fatal("Expected error to propagate from GetTransactionbyID's HTTP error, got nil")}
		if !strings.Contains(err.Error(), "network response was not ok") { // This is error from GetTransactionbyID
			t.Errorf("Expected propagated HTTP error, got %s", err.Error())
		}
		mockRequestCounters["_GetTransactionbyID_"] = 0
		mockResponses["_GetTransactionbyID_"] = nil
	})
}


// roundTripFunc is a helper type for mocking http.Client transport
type roundTripFunc func(req *http.Request) *http.Response

// RoundTrip wraps the function to satisfy the http.RoundTripper interface.
func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	// Check if the transport itself should return an error (e.g. network down)
	// This needs to be handled by inspecting the mockResponse struct if we add an Err field to it.
	// For now, all errors are via StatusCode.
	return f(req), nil
}
