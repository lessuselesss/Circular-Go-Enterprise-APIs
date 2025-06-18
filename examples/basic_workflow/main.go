package main

import (
	"fmt"
	"log"
	"strings" // Added for strings.HasPrefix

	"github.com/circular-protocol/circular-go/circular_protocol_api"
	"github.com/circular-protocol/circular-go/utils"
)

func main() {
	fmt.Println("Circular Go Enterprise API - Basic Workflow Example")
	fmt.Printf("Current UTC Timestamp: %s\n", utils.GetFormattedUTCTimestamp())

	// Replace with your actual address and private key
	const userAddress = "your-wallet-address-hex"
	// IMPORTANT: Never hardcode private keys in production code.
	// This is for demonstration purposes only. Use environment variables or a secure vault.
	const userPrivateKeyHex = "11842e4034999297038f59f054d1794389758469070e15999837078cec243f55" // Example valid-looking hex
	const blockchainID = "0x8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2"    // Example blockchain ID

	account := &circular_protocol_api.Account{}

	// 1. Open Account
	account.Open(userAddress)
	fmt.Printf("Account opened for address: %s\n", account.Address)

	// 2. Set Network and Blockchain
	err := account.SetNetwork("testnet") // Or "mainnet"
	if err != nil {
		log.Fatalf("Failed to set network: %v", err)
	}
	account.SetBlockchain(blockchainID)
	fmt.Printf("Network: %s, NAG URL: %s\n", account.Network, account.NAGURL)
	fmt.Printf("Blockchain ID: %s\n", account.Blockchain)

	// 3. Update Account (Fetch Nonce)
	fmt.Println("Attempting to update account (fetch nonce)...")
	if err := account.UpdateAccount(); err != nil {
		log.Printf("Warning: Failed to update account (fetch nonce): %v. Using default Nonce: %d\n", err, account.Nonce)
	} else {
		fmt.Printf("Account Nonce updated to: %d\n", account.Nonce)
	}

	// 4. Prepare data and submit certificate
	certificateData := fmt.Sprintf("My important data for certification at %s", utils.GetFormattedUTCTimestamp())

	fmt.Printf("Submitting certificate with data: '%s'\n", certificateData)
	submitResp, err := account.SubmitCertificate(certificateData, userPrivateKeyHex)
	if err != nil {
		log.Printf("Failed to submit certificate (likely due to placeholder NAG): %v\n", err)
		if submitResp.Response.TxID == "" {
			submitResp.Response.TxID = "mockTxID_for_example_flow_due_to_error"
			fmt.Println("Using mock TxID for example flow as submission failed.")
		}
	}

	if submitResp.Result == 200 || (submitResp.Response.TxID != "" && !strings.HasPrefix(submitResp.Response.TxID, "mockTxID_")) {
		fmt.Printf("Certificate submission potentially successful or pending. Transaction ID: %s\n", submitResp.Response.TxID)
	} else if submitResp.Response.TxID != "" && strings.HasPrefix(submitResp.Response.TxID, "mockTxID_") {
		fmt.Printf("Certificate submission was mocked. Transaction ID: %s\n", submitResp.Response.TxID)
	} else {
		fmt.Printf("Certificate submission failed. Result: %d, Message: %s\n", submitResp.Result, submitResp.Message)
		if submitResp.Response.TxID == "" {
			submitResp.Response.TxID = "mockTxID_for_failed_submission"
			fmt.Println("Using mock TxID for further example steps after submission failure.")
		}
	}

	if submitResp.Response.TxID != "" {
		txID := submitResp.Response.TxID

		// 5. Get Transaction Outcome
		fmt.Printf("Polling for transaction outcome for TxID: %s...\n", txID)
		outcomeResp, err := account.GetTransactionOutcome(txID, 15)
		if err != nil {
			log.Printf("Failed to get transaction outcome (likely due to placeholder NAG): %v\n", err)
			outcomeResp.Result = -1
			outcomeResp.Status = "mock_failure_or_timeout"
			outcomeResp.Message = "Mocked outcome due to placeholder error"
			// Ensure BlockID is also mocked if not present
			if outcomeResp.BlockID == "" {
				outcomeResp.BlockID = "mockBlockID_after_outcome_error"
			}
		}
		fmt.Printf("Transaction Outcome: Result %d, Status '%s', BlockID '%s', Message '%s'\n",
			outcomeResp.Result, outcomeResp.Status, outcomeResp.BlockID, outcomeResp.Message)

		// 6. Get Transaction Details
		if outcomeResp.BlockID != "" || (outcomeResp.Status == "confirmed" || outcomeResp.Status == "mock_success") {
			blockIDToUse := outcomeResp.BlockID
			if blockIDToUse == "" { // Should be covered by above but as a safeguard
				blockIDToUse = "mockBlockID_for_example"
				fmt.Println("Using mock BlockID for GetTransaction example.")
			}

			fmt.Printf("Fetching transaction details for BlockID %s, TxID %s...\n", blockIDToUse, txID)
			txDetailsResp, err := account.GetTransaction(blockIDToUse, txID)
			if err != nil {
				log.Printf("Failed to get transaction details (likely due to placeholder NAG): %v\n", err)
				txDetailsResp.Result = -1
				txDetailsResp.Message = "Mocked transaction details due to error"
				txDetailsResp.Transaction.TxID = txID // Ensure TxID is present in mock
				txDetailsResp.Transaction.Data = "Mocked transaction data"
				txDetailsResp.Transaction.Timestamp = utils.GetFormattedUTCTimestamp()
			}
			fmt.Printf("Transaction Details: Result %d, TxID '%s', Data '%s', Timestamp '%s', Message: '%s'\n",
				txDetailsResp.Result, txDetailsResp.Transaction.TxID, txDetailsResp.Transaction.Data, txDetailsResp.Transaction.Timestamp, txDetailsResp.Message)
		} else {
			fmt.Printf("Skipping GetTransaction as BlockID is missing or status is not confirmed/mock_success. Status: '%s', BlockID: '%s'\n", outcomeResp.Status, outcomeResp.BlockID)
		}
	} else {
		fmt.Println("Skipping transaction outcome and details fetching as no TxID was available from submission.")
	}

	// 7. Close Account
	account.Close()
	fmt.Println("Account closed.")
}
