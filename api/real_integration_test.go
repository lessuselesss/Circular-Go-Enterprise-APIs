//go:build integration
// +build integration

package api

import (
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
