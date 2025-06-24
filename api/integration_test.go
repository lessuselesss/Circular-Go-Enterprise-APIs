//go:build integration
// +build integration

package api

import (
	"testing"
)

// TestFullWorkflowIntegration tests the complete workflow of using the Circular API
func TestFullWorkflowIntegration(t *testing.T) {
	// Step 1: Create and configure account
	account := &Account{}

	// Step 2: Open account with wallet address
	walletAddress := "0x1234567890abcdef"
	err := account.Open(walletAddress)
	if err != nil {
		t.Fatalf("Failed to open account: %v", err)
	}

	// Step 3: Configure network and blockchain (mock for integration test)
	account.network = "testnet"
	account.SetBlockchain("test_blockchain_address")

	// Step 4: Update account to get latest nonce
	success, err := account.UpdateAccount()
	if err != nil {
		t.Fatalf("Failed to update account: %v", err)
	}
	if !success {
		t.Fatal("UpdateAccount should return true")
	}

	// Step 5: Create a certificate
	cert := &Certificate{}
	testData := []byte("Integration test certificate data")
	cert.SetData(testData)

	// Verify certificate data
	if string(cert.GetData()) != string(testData) {
		t.Errorf("Certificate data mismatch")
	}

	// Verify certificate size
	if cert.GetCertificateSize() != len(testData) {
		t.Errorf("Certificate size mismatch: expected %d, got %d", len(testData), cert.GetCertificateSize())
	}

	// Step 6: Sign the certificate data
	privateKey := "test_private_key_for_integration"
	signedData, err := account.SignData(cert.GetData(), privateKey)
	if err != nil {
		t.Fatalf("Failed to sign data: %v", err)
	}

	if len(signedData) == 0 {
		t.Fatal("Signed data should not be empty")
	}

	// Step 7: Submit certificate to blockchain
	submitResp, err := account.SubmitCertificate(cert.GetData(), privateKey)
	if err != nil {
		t.Fatalf("Failed to submit certificate: %v", err)
	}

	if submitResp.Result != 200 {
		t.Errorf("Submit certificate should return 200, got %d", submitResp.Result)
	}

	if submitResp.Response.TxID == "" {
		t.Error("Submit response should contain transaction ID")
	}

	txID := submitResp.Response.TxID

	// Step 8: Monitor transaction outcome
	txOutcome, err := account.GetTransactionOutcome(txID, 30)
	if err != nil {
		t.Fatalf("Failed to get transaction outcome: %v", err)
	}

	if txOutcome.Result != 200 {
		t.Errorf("Transaction outcome should return 200, got %d", txOutcome.Result)
	}

	if txOutcome.Response.ID != txID {
		t.Errorf("Transaction ID mismatch: expected %s, got %s", txID, txOutcome.Response.ID)
	}

	if txOutcome.Response.Status != "Executed" {
		t.Errorf("Expected transaction status 'Executed', got %s", txOutcome.Response.Status)
	}

	// Step 9: Query transaction by ID
	blockID := txOutcome.Response.BlockID
	txByID, err := account.GetTransactionByID(txID, blockID, "")
	if err != nil {
		t.Fatalf("Failed to get transaction by ID: %v", err)
	}

	if txByID.Response.ID != txID {
		t.Errorf("GetTransactionByID ID mismatch: expected %s, got %s", txID, txByID.Response.ID)
	}

	// Step 10: Clean up
	account.Close()

	// Verify cleanup
	if account.network != "" || account.blockchain != "" {
		t.Error("Account should be properly closed")
	}
}

// TestCertificateWorkflow tests the certificate creation and manipulation workflow
func TestCertificateWorkflow(t *testing.T) {
	// Test data scenarios
	testCases := []struct {
		name string
		data []byte
	}{
		{"simple text", []byte("Hello, Circular Protocol!")},
		{"json data", []byte(`{"user":"alice","action":"create","timestamp":"2024-01-01T00:00:00Z"}`)},
		{"unicode data", []byte("Hello 🌍 Unicode test data 🚀")},
		{"binary data", []byte{0x00, 0xFF, 0xAB, 0xCD, 0xEF}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create certificate
			cert := &Certificate{}
			cert.SetData(tc.data)

			// Verify data round trip
			retrievedData := cert.GetData()
			if string(retrievedData) != string(tc.data) {
				t.Errorf("Data round trip failed for %s", tc.name)
			}

			// Verify size
			if cert.GetCertificateSize() != len(tc.data) {
				t.Errorf("Size mismatch for %s: expected %d, got %d", tc.name, len(tc.data), cert.GetCertificateSize())
			}

			// Verify JSON representation
			jsonCert := cert.GetJSONCertificate()
			if jsonCert == "" {
				t.Errorf("JSON certificate should not be empty for %s", tc.name)
			}

			// JSON should be valid (basic check)
			if jsonCert == "{}" && len(tc.data) > 0 {
				t.Errorf("JSON certificate should not be empty object for non-empty data in %s", tc.name)
			}
		})
	}
}

// TestAccountConfigurationWorkflow tests various account configuration scenarios
func TestAccountConfigurationWorkflow(t *testing.T) {
	networks := []string{"mainnet", "testnet", "devnet"}
	blockchains := []string{"chain1", "chain2", "0x123456"}
	walletAddresses := []string{"0xabcdef", "wallet123", "test_address"}

	for _, network := range networks {
		for _, blockchain := range blockchains {
			for _, wallet := range walletAddresses {
				t.Run(network+"_"+blockchain+"_"+wallet, func(t *testing.T) {
					account := &Account{}

					// Configure account
					err := account.Open(wallet)
					if err != nil {
						t.Fatalf("Failed to open account with wallet %s: %v", wallet, err)
					}

					account.SetNetwork(network)
					account.SetBlockchain(blockchain)

					// Verify configuration
					if account.network != network {
						t.Errorf("Network configuration failed: expected %s, got %s", network, account.network)
					}

					if account.blockchain != blockchain {
						t.Errorf("Blockchain configuration failed: expected %s, got %s", blockchain, account.blockchain)
					}

					// Test account update
					success, err := account.UpdateAccount()
					if err != nil {
						t.Errorf("UpdateAccount failed for %s/%s/%s: %v", network, blockchain, wallet, err)
					}
					if !success {
						t.Errorf("UpdateAccount should succeed for %s/%s/%s", network, blockchain, wallet)
					}

					// Clean up
					account.Close()
				})
			}
		}
	}
}

// TestErrorHandlingWorkflow tests error scenarios and edge cases
func TestErrorHandlingWorkflow(t *testing.T) {
	account := &Account{}

	// Test with minimal configuration
	err := account.Open("")
	if err != nil {
		t.Logf("Open with empty address returned error: %v", err)
	}

	// Test operations on unconfigured account
	success, err := account.UpdateAccount()
	if err != nil {
		t.Logf("UpdateAccount on minimal account returned error: %v", err)
	} else if !success {
		t.Log("UpdateAccount on minimal account returned false")
	}

	// Test signing with empty data
	signedData, err := account.SignData([]byte{}, "test_key")
	if err != nil {
		t.Logf("SignData with empty data returned error: %v", err)
	} else if len(signedData) == 0 {
		t.Log("SignData with empty data returned empty result")
	}

	// Test certificate submission with empty data
	submitResp, err := account.SubmitCertificate([]byte{}, "test_key")
	if err != nil {
		t.Logf("SubmitCertificate with empty data returned error: %v", err)
	} else if submitResp == nil {
		t.Log("SubmitCertificate with empty data returned nil response")
	}

	// Test transaction monitoring with empty ID
	txResp, err := account.GetTransactionOutcome("", 1)
	if err != nil {
		t.Logf("GetTransactionOutcome with empty ID returned error: %v", err)
	} else if txResp == nil {
		t.Log("GetTransactionOutcome with empty ID returned nil response")
	}
}
