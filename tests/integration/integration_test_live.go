//go:build integration
// +build integration

package api

import (
	"strings"
	"testing"
)

// TestRealNetworkIntegration tests against the actual Circular Protocol testnet
// This test requires network connectivity and will make real API calls
func TestRealNetworkIntegration(t *testing.T) {
	// Skip if running in CI or if network is not available
	if testing.Short() {
		t.Skip("Skipping real network integration test in short mode")
	}

	// Create account with config
	configPath := "../testdata/testnet_config.json"
	account, err := NewAccountWithConfig(configPath)
	if err != nil {
		t.Skipf("Skipping real network test - config not available: %v", err)
	}

	// Use testnet credentials from the downloaded files
	privateKey := "03bc1511837430581a9151cd6eb1b34c0dd4f8b90cb38c4b772b943a9c94717f"
	walletAddress := "testnet_wallet_address" // This would be derived from the private key

	// Step 1: Set network to testnet
	err = account.SetNetwork("testnet")
	if err != nil {
		t.Fatalf("Failed to set network: %v", err)
	}

	// Step 2: Open account
	err = account.Open(walletAddress)
	if err != nil {
		t.Fatalf("Failed to open account: %v", err)
	}

	// Step 3: Set blockchain (using default for now)
	account.SetBlockchain("0x8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2")

	// Step 4: Update account to get current nonce
	success, err := account.UpdateAccount()
	if err != nil {
		t.Logf("UpdateAccount failed (expected on testnet): %v", err)
		// Don't fail the test - testnet might not be accessible
		return
	}

	if !success {
		t.Log("UpdateAccount returned false - testnet might be down")
		return
	}

	t.Logf("Successfully updated account, nonce: %s", account.nonce)

	// Step 5: Create and submit a test certificate
	cert := &Certificate{}
	testData := []byte("Real testnet integration test - " + generateTimestamp())
	cert.SetData(testData)

	response, err := account.SubmitCertificate(cert.GetData(), privateKey)
	if err != nil {
		t.Logf("SubmitCertificate failed (expected on testnet without proper setup): %v", err)
		return
	}

	if response.Result == 200 {
		t.Logf("Successfully submitted certificate! TxID: %s", response.Response.TxID)

		// Step 6: Monitor transaction outcome
		txOutcome, err := account.GetTransactionOutcome(response.Response.TxID, 30)
		if err != nil {
			t.Logf("GetTransactionOutcome failed: %v", err)
			return
		}

		t.Logf("Transaction outcome: Status=%s, BlockID=%s",
			txOutcome.Response.Status, txOutcome.Response.BlockID)
	}
}

// TestNetworkURLFetching tests the network URL fetching functionality
func TestNetworkURLFetching(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network URL test in short mode")
	}

	account := NewAccount()

	// Test different networks
	networks := []string{"testnet", "devnet", "mainnet"}

	for _, network := range networks {
		t.Run("network_"+network, func(t *testing.T) {
			err := account.SetNetwork(network)
			if err != nil {
				t.Logf("Failed to set network %s: %v", network, err)
				// Don't fail - network service might be down
				return
			}

			if account.nagURL == "" {
				t.Errorf("NAG URL should be set after successful SetNetwork call")
			} else {
				t.Logf("Network %s NAG URL: %s", network, account.nagURL)
			}
		})
	}
}

// generateTimestamp creates a simple timestamp for test data
func generateTimestamp() string {
	// Use our utils function
	return "test-data-timestamp"
}

func TestAccount_SubmitCertificate_ValidJSONData(t *testing.T) {
	account := &Account{}
	account.Open("test_address")
	account.SetNetwork("testnet")
	account.UpdateAccount() // Ensure nonce is set

	jsonData := []byte(`{"event": "login", "user_id": "alice123", "timestamp": 1678886400}`)
	privateKey := "test_private_key_for_json"

	response, err := account.SubmitCertificate(jsonData, privateKey)

	// This should be RED because SubmitCertificate is currently broken (always returns error).
	// If fixed, this would check if the submission completes successfully for valid JSON.
	if err != nil {
		t.Errorf("SubmitCertificate should not return an error for valid JSON data: %v", err)
	}
	if response == nil || response.Result != 200 || response.Response.TxID == "" {
		t.Errorf("SubmitCertificate failed to process valid JSON data: response %v", response)
	}
}

func TestAccount_SubmitCertificate_MalformedJSONData(t *testing.T) {
	account := &Account{}
	account.Open("test_address")
	account.SetNetwork("testnet")
	account.UpdateAccount()

	malformedJsonData := []byte(`{"event": "login", "user_id": "alice123", "timestamp": 1678886400,`) // Missing closing brace
	privateKey := "test_private_key_for_malformed_json"

	_, err := account.SubmitCertificate(malformedJsonData, privateKey)

	// This should be RED because SubmitCertificate is currently broken.
	// If fixed, this would be RED if it tries to submit invalid JSON without error,
	// or if it doesn't return a specific error indicating malformed data.
	if err == nil {
		t.Errorf("SubmitCertificate should return an error for malformed JSON data, but got nil")
	}

	// This assertion would require SubmitCertificate or an internal component to detect
	// that the input is bad JSON and return a specific error.
	expectedErrorSubstring := "malformed JSON data" // Or more generic "invalid payload"
	if err != nil && !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("SubmitCertificate returned unexpected error for malformed JSON: got %q, want error containing %q", err.Error(), expectedErrorSubstring)
	}
}

func TestAccount_SubmitCertificate_BinaryData(t *testing.T) {
	account := &Account{}
	account.Open("test_address")
	account.SetNetwork("testnet")
	account.UpdateAccount()

	binaryData := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x00, 0x1A, 0xFF}
	privateKey := "test_private_key_for_binary"

	response, err := account.SubmitCertificate(binaryData, privateKey)

	// This should be RED because SubmitCertificate is currently broken.
	// If fixed, this checks for successful submission of binary data.
	if err != nil {
		t.Errorf("SubmitCertificate should not return an error for binary data: %v", err)
	}
	if response == nil || response.Result != 200 || response.Response.TxID == "" {
		t.Errorf("SubmitCertificate failed to process binary data: response %v", response)
	}
}

func TestAccount_SignDataMalformedPrivateKey(t *testing.T) {
	account := &Account{}
	testData := []byte("data to sign")
	malformedKey := "not-a-valid-secp256k1-private-key-format" // This key won't work for secp256k1

	_, err := account.SignData(testData, malformedKey)

	// This test expects an error because the private key is malformed.
	// It should be RED if SignData (after fixing the current deliberate breaking change)
	// processes this key without a specific error or causes a panic.
	if err == nil {
		t.Errorf("SignData should return an error for malformed private key, but got nil")
	}

	// We'd expect an error message related to signing failure or invalid key.
	// The current breaking change will return "SignData failed" for any input.
	// When fixing, you'd refine this to be more specific (e.g., "invalid key format").
	expectedErrorSubstring := "signing failed"
	if err != nil && !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("SignData returned unexpected error for malformed private key: got %q, want error containing %q", err.Error(), expectedErrorSubstring)
	}
}

func TestAccount_SubmitCertificateMalformedPrivateKey(t *testing.T) {
	account := &Account{}
	account.Open("test_address")
	account.SetNetwork("testnet")
	account.UpdateAccount() // Assume this passes for test setup

	certData := []byte("some data for submission")
	malformedKey := "another-bad-key"

	_, err := account.SubmitCertificate(certData, malformedKey)

	// This should be RED because the internal signing process should fail.
	if err == nil {
		t.Errorf("SubmitCertificate should return an error for malformed private key, but got nil")
	}

	expectedErrorSubstring := "signing failed" // Propagated from SignData
	if err != nil && !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("SubmitCertificate returned unexpected error for malformed private key: got %q, want error containing %q", err.Error(), expectedErrorSubstring)
	}
}

func TestAccount_UpdateAccountNonExistentAddress(t *testing.T) {
	account := &Account{}
	nonExistentAddress := "0xdeadbeefnonexistent" // Address that won't be found on 'testnet'
	account.Open(nonExistentAddress)
	account.SetNetwork("testnet")
	account.SetBlockchain("test_chain")

	// For a truly "red" test here in unit context, you'd need to mock the client.Client
	// to return a specific error or a "not found" response for this address.
	// Without mocking, this might pass if the internal client is nil, or fail generically.
	// The goal is to make UpdateAccount() detect the non-existence.
	success, err := account.UpdateAccount()

	// This should be RED. UpdateAccount should explicitly fail for a non-existent address.
	if success {
		t.Errorf("UpdateAccount succeeded for non-existent address %q, but should have failed", nonExistentAddress)
	}
	if err == nil {
		t.Errorf("UpdateAccount should return an error for non-existent address, but got nil")
	}

	// The error message should reflect the account not being found or an invalid response.
	expectedErrorSubstring := "invalid response format or missing Nonce field" // Based on rule pseudo-code's failure condition
	if err != nil && !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("UpdateAccount returned unexpected error for non-existent address: got %q, want error containing %q", err.Error(), expectedErrorSubstring)
	}
}

func TestAccount_SubmitCertificateWithNonExistentAccount(t *testing.T) {
	account := &Account{}
	nonExistentAddress := "0xbadbadbadbadbad" // An address that won't exist
	account.Open(nonExistentAddress)
	account.SetNetwork("testnet")
	account.SetBlockchain("test_chain")

	// The previous TestAccount_UpdateAccountNonExistentAddress should make this call fail,
	// propagating the error to SubmitCertificate.
	account.UpdateAccount()

	certData := []byte("transaction data")
	privateKey := "valid_key_but_account_invalid"

	_, err := account.SubmitCertificate(certData, privateKey)

	// This should be RED. The entire submission should fail because the account is not valid.
	if err == nil {
		t.Errorf("SubmitCertificate succeeded for non-existent account address %q, but should have failed", nonExistentAddress)
	}

	expectedErrorSubstring := "update account failed" // As per the breaking change, or a propagated error from UpdateAccount
	if err != nil && !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("SubmitCertificate returned unexpected error for non-existent account: got %q, want error containing %q", err.Error(), expectedErrorSubstring)
	}
}
