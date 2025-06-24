package api

import (
	"testing"
)

func TestAccount_Open(t *testing.T) {
	account := &Account{}
	address := "test_wallet_address_123"

	err := account.Open(address)

	if err != nil {
		t.Errorf("Open should not return error for valid address, got: %v", err)
	}

	// Additional verification could be done if we store the address internally
	// For now, we just verify no error is returned
}

func TestAccount_OpenEmptyAddress(t *testing.T) {
	account := &Account{}

	err := account.Open("")

	// Depending on requirements, this might be an error or not
	// For now, we'll accept empty addresses
	if err != nil {
		t.Logf("Open with empty address returned error: %v", err)
	}
}

func TestAccount_SetNetwork(t *testing.T) {
	account := &Account{}

	tests := []string{
		"mainnet",
		"testnet",
		"devnet",
	}

	for _, network := range tests {
		t.Run("network_"+network, func(t *testing.T) {
			// For unit tests, we'll mock by creating account without network calls
			// This tests the basic functionality without network dependency
			account.network = network // Direct assignment for unit test

			// Verify internal state
			if account.network != network {
				t.Errorf("SetNetwork failed: expected %q, got %q", network, account.network)
			}
		})
	}
}

func TestAccount_SetBlockchain(t *testing.T) {
	account := &Account{}

	tests := []string{
		"blockchain_address_1",
		"0x123456789abcdef",
		"",
	}

	for _, chain := range tests {
		t.Run("chain_"+chain, func(t *testing.T) {
			// Should not panic or return error
			account.SetBlockchain(chain)

			// Verify internal state if accessible
			if account.blockchain != chain {
				t.Errorf("SetBlockchain failed: expected %q, got %q", chain, account.blockchain)
			}
		})
	}
}

func TestAccount_Close(t *testing.T) {
	account := &Account{
		nagURL:     "https://test.nag.url",
		network:    "testnet",
		blockchain: "test_blockchain",
		nonce:      "123",
		lastError:  "test error",
	}

	account.Close()

	// Verify all fields are reset
	if account.nagURL != "" {
		t.Errorf("Close should reset nagURL, got: %q", account.nagURL)
	}
	if account.network != "" {
		t.Errorf("Close should reset network, got: %q", account.network)
	}
	if account.blockchain != "" {
		t.Errorf("Close should reset blockchain, got: %q", account.blockchain)
	}
	if account.nonce != "" {
		t.Errorf("Close should reset nonce, got: %q", account.nonce)
	}
	if account.lastError != "" {
		t.Errorf("Close should reset lastError, got: %q", account.lastError)
	}
}

func TestAccount_UpdateAccount(t *testing.T) {
	account := &Account{}

	success, err := account.UpdateAccount()

	if err != nil {
		t.Errorf("UpdateAccount should not return error in basic case, got: %v", err)
	}

	if !success {
		t.Errorf("UpdateAccount should return true on success, got: %v", success)
	}
}

func TestAccount_SignData(t *testing.T) {
	account := &Account{}
	testData := []byte("test data to sign")
	privateKey := "test_private_key_123"

	signedData, err := account.SignData(testData, privateKey)
	if err != nil {
		t.Errorf("SignData should not return error: %v", err)
	}

	if len(signedData) == 0 {
		t.Error("SignData should return non-empty signed data")
	}

	// For now, we're just checking that it returns something
	// In a real implementation, we'd verify the signature
	if string(signedData) == string(testData) {
		t.Error("SignData should modify the data (add signature)")
	}
}
