//go:build integration
// +build integration

package integration

import (
	"testing"

	"github.com/lessuselesss/circular-go-enterprise-apis/api"
)

func TestAccount_ChainedOperations(t *testing.T) {
	account := &api.Account{}

	// Test a sequence of operations
	err := account.Open("test_address")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	account.SetNetwork("testnet")
	account.SetBlockchain("test_chain")

	success, err := account.UpdateAccount()
	if err != nil {
		t.Errorf("UpdateAccount failed: %v", err)
	}
	if !success {
		t.Errorf("UpdateAccount should succeed")
	}

	// Verify state
	if account.GetNetwork() != "testnet" {
		t.Errorf("Network not set correctly: %q", account.GetNetwork())
	}
	if account.GetBlockchain() != "test_chain" {
		t.Errorf("Blockchain not set correctly: %q", account.GetBlockchain())
	}

	// Clean up
	account.Close()

	// Verify cleanup
	if account.GetNetwork() != "" || account.GetBlockchain() != "" {
		t.Errorf("Close did not clean up properly")
	}
}

func TestAccount_SubmitCertificate(t *testing.T) {
	account := &api.Account{}
	testData := []byte("certificate data")
	privateKey := "test_private_key_123"

	response, err := account.SubmitCertificate(testData, privateKey)
	if err != nil {
		t.Errorf("SubmitCertificate should not return error: %v", err)
	}

	if response == nil {
		t.Error("SubmitCertificate should return response")
	}

	if response.Result != 200 {
		t.Errorf("Expected Result 200, got %d", response.Result)
	}

	if response.Response.TxID == "" {
		t.Error("Response should contain TxID")
	}

	if response.Response.Timestamp == "" {
		t.Error("Response should contain Timestamp")
	}
}

func TestAccount_GetTransactionOutcome(t *testing.T) {
	account := &api.Account{}
	txID := "test_transaction_id_123"
	timeoutSec := 30

	response, err := account.GetTransactionOutcome(txID, timeoutSec)
	if err != nil {
		t.Errorf("GetTransactionOutcome should not return error: %v", err)
	}

	if response == nil {
		t.Error("GetTransactionOutcome should return response")
	}

	if response.Result != 200 {
		t.Errorf("Expected Result 200, got %d", response.Result)
	}

	if response.Response.ID != txID {
		t.Errorf("Expected transaction ID %q, got %q", txID, response.Response.ID)
	}

	if response.Response.Status == "" {
		t.Error("Response should contain Status")
	}

	if response.Response.Timestamp == "" {
		t.Error("Response should contain Timestamp")
	}
}

func TestAccount_GetTransactionByID(t *testing.T) {
	account := &api.Account{}
	txID := "test_transaction_id_456"
	start := "start_block_123"
	end := "end_block_456"

	response, err := account.GetTransactionByID(txID, start, end)
	if err != nil {
		t.Errorf("GetTransactionByID should not return error: %v", err)
	}

	if response == nil {
		t.Error("GetTransactionByID should return response")
	}

	if response.Result != 200 {
		t.Errorf("Expected Result 200, got %d", response.Result)
	}

	if response.Response.ID != txID {
		t.Errorf("Expected transaction ID %q, got %q", txID, response.Response.ID)
	}

	if response.Response.BlockID != start {
		t.Errorf("Expected BlockID %q, got %q", start, response.Response.BlockID)
	}
}

func TestAccount_GetTransactionOutcomeTimeout(t *testing.T) {
	account := &api.Account{}
	txID := "test_transaction_id_timeout"
	timeoutSec := 0 // Zero timeout

	// Should still work even with zero timeout (immediate return)
	response, err := account.GetTransactionOutcome(txID, timeoutSec)
	if err != nil {
		t.Errorf("GetTransactionOutcome should handle zero timeout: %v", err)
	}

	if response == nil {
		t.Error("GetTransactionOutcome should return response even with zero timeout")
	}
}
