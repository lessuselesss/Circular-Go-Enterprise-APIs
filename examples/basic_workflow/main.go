// Package main provides a basic command-line application demonstrating a typical workflow
// for interacting with the circular_protocol_api. This example covers account initialization,
// network configuration, certificate submission, and transaction status checking.
// It is intended for illustrative purposes and highlights how to use various API functionalities.
package main

import (
	"fmt"
	"log"
	"strings" // Used for checking mock TxID prefixes.

	"github.com/circular-protocol/circular-go/circular_protocol_api"
	"github.com/circular-protocol/circular-go/utils"
)

// main function demonstrates a basic workflow using the circular_protocol_api.
// The steps include:
// 1. Opening a user account with a specified address.
// 2. Setting the desired network (e.g., "testnet") and blockchain ID.
// 3. Updating the account, which typically involves fetching the latest nonce from the network.
// 4. Preparing sample data and submitting it as a certificate to the network.
//   - This step includes handling potential errors and using a mock TxID if the
//     actual submission fails (e.g., due to placeholder Network Access Gateway - NAG - URLs
//     or network issues), to allow the example workflow to continue.
//
// 5. Polling for the transaction outcome using the TxID obtained from the submission.
//   - Similar to submission, this includes error handling and mocking for example purposes.
//
// 6. Fetching detailed information about the transaction if a BlockID is available from the outcome.
//   - Error handling and mocking are also applied here.
//
// 7. Closing the account to reset its state.
//
// Note: This example uses placeholder values for address and private key.
// In a real application, these should be managed securely and not hardcoded.
// The NAG interactions might fail if the configured NAG URLs are placeholders or unreachable,
// hence the mocking logic to illustrate the complete flow.
func main() {
	fmt.Println("Circular Go Enterprise API - Basic Workflow Example")
	// Display current timestamp for reference.
	fmt.Printf("Current Timestamp (from utils.GetFormattedTimestamp): %s\n", utils.GetFormattedTimestamp())

	// Configuration: Replace with your actual address and private key for real testing.
	// For this example, placeholders are used.
	const userAddress = "your-wallet-address-hex" // Example: "0x..."
	// IMPORTANT: Never hardcode private keys in production applications.
	// This is for demonstration purposes only. Use environment variables, a secure vault, or other secure means.
	const userPrivateKeyHex = "11842e4034999297038f59f054d1794389758469070e15999837078cec243f55" // Example: A valid hex private key string
	const blockchainID = "0x8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2"    // Example: A specific blockchain identifier

	// Initialize a new Account object from the API.
	account := &circular_protocol_api.CEPAccount{}

	// --- 1. Open Account ---
	// Initializes the account object with the user's address.
	account.Open(userAddress)
	fmt.Printf("Account opened for address: %s\n", account.Address)

	// --- 2. Set Network and Blockchain ---
	// Configures the account to use a specific network (e.g., "testnet" or "mainnet")
	// and sets the target blockchain ID.
	// The SetNetwork call might fetch network-specific configurations like the NAG URL.
	err := account.SetNetwork("testnet") // Use "testnet" for testing, or "mainnet" for production.
	if err != nil {
		// If network setup fails, it's a critical error for this workflow.
		log.Fatalf("Failed to set network: %v", err)
	}
	account.SetBlockchain(blockchainID) // Set the specific blockchain to operate on.
	// NAGURL and NetworkNode are set by SetNetwork. The specific network name ("testnet") isn't stored as an exported field.
	fmt.Printf("Network configured (NAG URL: %s, Network Node: %s)\n", account.NAGURL, account.NetworkNode)
	fmt.Printf("Blockchain ID: %s\n", account.Blockchain)

	// --- 3. Update Account (Fetch Nonce) ---
	// Updates account information, primarily fetching the current nonce from the network.
	// The nonce is crucial for sequencing transactions.
	fmt.Println("Attempting to update account (fetch nonce)...")
	success, err := account.UpdateAccount()
	if err != nil {
		// This error means the call to UpdateAccount itself failed (e.g., network issue).
		log.Printf("Error calling UpdateAccount: %v. Using current Nonce: %d\n", err, account.Nonce)
	} else if !success {
		// The call succeeded but the update operation on the backend might have failed or returned an unsuccessful status.
		log.Printf("Warning: UpdateAccount reported not successful. Using current Nonce: %d\n", account.Nonce)
	} else {
		fmt.Printf("Account Nonce updated successfully to: %d\n", account.Nonce)
	}

	// --- 4. Prepare data and submit certificate ---
	// Create some sample data for the certificate.
	certificateData := fmt.Sprintf("My important data for certification at %s", utils.GetFormattedTimestamp())

	fmt.Printf("Submitting certificate with data: '%s'\n", certificateData)
	// Submit the certificate data, signed with the user's private key.
	submitRespMap, err := account.SubmitCertificate(certificateData, userPrivateKeyHex)

	var txID string
	var submissionFailed bool = false

	if err != nil {
		// This error typically occurs if the NAG is unreachable or returns an error during the HTTP request itself.
		log.Printf("Error calling SubmitCertificate: %v\n", err)
		submissionFailed = true
	} else {
		// Parse the submitRespMap
		result, _ := submitRespMap["Result"].(float64) // NAG often returns numbers as float64
		message, _ := submitRespMap["Message"].(string)
		responseMap, responseOK := submitRespMap["Response"].(map[string]interface{})

		if responseOK {
			txID, _ = responseMap["TxID"].(string)
		}

		log.Printf("SubmitCertificate Response: Result: %v, Message: '%s', TxID: '%s'\n", result, message, txID)

		if result != 200 && result != 0 { // 0 might be used by some NAGs for pending/success without explicit 200
			log.Printf("Submission reported non-success result: %v\n", result)
			submissionFailed = true
		}
		if txID == "" && !submissionFailed { // If result was 200 but no TxID, that's also an issue.
			log.Println("Submission successful but no TxID received.")
			submissionFailed = true // Treat as failure for example flow if no TxID.
		}
	}

	// If submission failed or no TxID, use a mock TxID for example flow.
	if submissionFailed || txID == "" {
		txID = "mockTxID_after_submission_issue"
		fmt.Printf("Using mock TxID for example flow: %s\n", txID)
		// Ensure submitRespMap is not nil for mocking, if original call failed badly.
		if submitRespMap == nil {
			submitRespMap = make(map[string]interface{})
		}
		// Mock parts of the response for consistent logging if needed later, though txID is primary.
		submitRespMap["Result"] = -1.0 // Mocked result
		submitRespMap["Message"] = "Mocked due to submission issue"
		if _, ok := submitRespMap["Response"].(map[string]interface{}); !ok {
			submitRespMap["Response"] = make(map[string]interface{})
		}
		submitRespMap["Response"].(map[string]interface{})["TxID"] = txID
	}

	// Log final decision on TxID
	if strings.HasPrefix(txID, "mockTxID_") {
		fmt.Printf("Proceeding with MOCKED Transaction ID: %s\n", txID)
	} else {
		fmt.Printf("Proceeding with REAL Transaction ID: %s\n", txID)
	}


	// Proceed only if we have a TxID (real or mocked).
	if txID != "" {
		// txID is already defined and populated above.

		// --- 5. Get Transaction Outcome ---
		// Poll for the outcome of the submitted transaction using its TxID.
		var outcomeResult float64
		var outcomeStatus, outcomeBlockID, outcomeMessage string
		var outcomeFailed bool = false

		fmt.Printf("Polling for transaction outcome for TxID: %s...\n", txID)
		outcomeRespMap, err := account.GetTransactionOutcome(txID, 15) // 15 seconds timeout.
		if err != nil {
			log.Printf("Error calling GetTransactionOutcome: %v\n", err)
			outcomeFailed = true
		} else {
			outcomeResult, _ = outcomeRespMap["Result"].(float64)
			outcomeStatus, _ = outcomeRespMap["Status"].(string)
			outcomeBlockID, _ = outcomeRespMap["BlockID"].(string)
			outcomeMessage, _ = outcomeRespMap["Message"].(string)
			log.Printf("GetTransactionOutcome Response: Result: %v, Status: '%s', BlockID: '%s', Message: '%s'\n",
				outcomeResult, outcomeStatus, outcomeBlockID, outcomeMessage)
			if outcomeResult != 200 || outcomeStatus == "Pending" || outcomeStatus == "" {
				// Consider "Pending" or empty status as a case where we might want to mock for example flow.
				log.Printf("Transaction outcome not yet confirmed or failed. Result: %v, Status: '%s'\n", outcomeResult, outcomeStatus)
				outcomeFailed = true // Treat as failure for example if not clearly successful.
			}
		}

		if outcomeFailed {
			// Mock an outcome response for example continuation.
			outcomeResult = -1.0 // Indicate mock failure.
			outcomeStatus = "mock_failure_or_timeout"
			outcomeMessage = "Mocked outcome due to error or timeout during GetTransactionOutcome."
			if outcomeBlockID == "" { // Only mock BlockID if it wasn't already populated (e.g. from a partial real response)
				outcomeBlockID = "mockBlockID_after_outcome_error"
			}
			fmt.Println("Using mocked transaction outcome.")
		}
		fmt.Printf("Transaction Outcome: Result %v, Status '%s', BlockID '%s', Message '%s'\n",
			outcomeResult, outcomeStatus, outcomeBlockID, outcomeMessage)

		// --- 6. Get Transaction Details ---
		// Fetch detailed information about the transaction if a BlockID was returned
		// or if status is "confirmed" (or a mock success).
		if outcomeBlockID != "" || (outcomeStatus == "confirmed" || outcomeStatus == "mock_success") {
			blockIDToUse := outcomeBlockID
			if blockIDToUse == "" { // Should be covered by outcomeFailed logic, but as a safeguard.
				blockIDToUse = "mockBlockID_for_example_get_transaction"
				fmt.Println("Using mock BlockID for GetTransaction example as it was unexpectedly empty but status allowed proceeding.")
			}

			var detailsResult float64
			var detailsMessage, txDetailsTxID, txDetailsData, txDetailsTimestamp string
			var detailsFailed bool = false

			// Note: GetTransactionbyID does not take blockID as a direct param.
			// The blockID might be implicit in the NAG's handling of unique TxIDs, or start/end params might be used.
			// Using 0,0 for start/end as their specific usage for a single TxID is marked unclear in API.
			fmt.Printf("Fetching transaction details for TxID %s (BlockID %s was for context)...\n", txID, blockIDToUse)
			txDetailsRespMap, err := account.GetTransactionbyID(txID, 0, 0)
			if err != nil {
				log.Printf("Error calling GetTransactionbyID: %v\n", err)
				detailsFailed = true
			} else {
				detailsResult, _ = txDetailsRespMap["Result"].(float64)
				detailsMessage, _ = txDetailsRespMap["Message"].(string)
				transactionMap, txOK := txDetailsRespMap["Transaction"].(map[string]interface{})
				if txOK {
					txDetailsTxID, _ = transactionMap["TxID"].(string)
					txDetailsData, _ = transactionMap["Data"].(string)
					txDetailsTimestamp, _ = transactionMap["Timestamp"].(string)
				}
				log.Printf("GetTransaction Response: Result: %v, Message: '%s', Tx.TxID: '%s', Tx.Data: '%s', Tx.Timestamp: '%s'\n",
					detailsResult, detailsMessage, txDetailsTxID, txDetailsData, txDetailsTimestamp)
				if detailsResult != 200 {
					detailsFailed = true
				}
			}

			if detailsFailed {
				detailsResult = -1.0 // Indicate mock failure.
				detailsMessage = "Mocked transaction details due to error during GetTransaction."
				txDetailsTxID = txID // Ensure TxID is present in the mocked transaction.
				txDetailsData = "Mocked transaction data content for example."
				txDetailsTimestamp = utils.GetFormattedTimestamp() // Use current time for mock.
				fmt.Println("Using mocked transaction details.")
			}
			fmt.Printf("Transaction Details: Result %v, TxID '%s', Data '%s', Timestamp '%s', Message: '%s'\n",
				detailsResult, txDetailsTxID, txDetailsData, txDetailsTimestamp, detailsMessage)
		} else {
			// Skip if BlockID is missing or transaction not confirmed.
			fmt.Printf("Skipping GetTransaction as BlockID is missing or status is not confirmed/mock_success. Status: '%s', BlockID: '%s'\n", outcomeStatus, outcomeBlockID)
		}
	} else {
		// Skip if no TxID was available from the submission step.
		fmt.Println("Skipping transaction outcome and details fetching as no TxID was available from submission.")
	}

	// --- 7. Close Account ---
	// Resets the account object's state, clearing sensitive information.
	account.Close()
	fmt.Println("Account closed.")
}
