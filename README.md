# Circular Protocol Go Enterprise API

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Official Circular Protocol Enterprise Go API for data certification on the blockchain.

## Description

This library provides tools for creating and submitting data certificates to the Circular blockchain using the Go language. It includes functionalities for managing accounts, setting network configurations, signing data, and interacting with the Circular Network Access Gateway (NAG).

## Features

-   **Account Management:** Open, update (fetch nonce), and close accounts.
-   **Network Configuration:** Set blockchain network (e.g., "testnet", "mainnet") and chain identifier.
-   **Data Certification:** Create certificate data and submit it to the blockchain via NAG.
-   **Data Signing:** Sign arbitrary data or certificate data using private keys (SECP256k1 with Schnorr signatures).
-   **Transaction Monitoring:** Poll for transaction outcomes and retrieve transaction details from NAG.
-   **Timestamp Generation:** Generate formatted UTC timestamps.

## Installation

To install the Circular Protocol Go Enterprise API package, run the following commands:

```bash
go get github.com/circular-protocol/circular-go
```

Then, ensure your project dependencies are tidy:

```bash
go mod tidy
```

## Usage

Below is an example demonstrating the basic workflow for using the API:

```go
package main

import (
	"fmt"
	"log"
	"strings" // Added for example logic

	"github.com/circular-protocol/circular-go/circular_protocol_api"
	"github.com/circular-protocol/circular-go/utils"
)

func main() {
	fmt.Println("Circular Go Enterprise API Example")
	fmt.Printf("Current UTC Timestamp: %s\n", utils.GetFormattedUTCTimestamp())

	// Replace with your actual address and private key
	const userAddress = "your-wallet-address-hex"
	// IMPORTANT: Never hardcode private keys in production code.
	// This is for demonstration purposes only. Use environment variables or a secure vault.
	const userPrivateKeyHex = "your-private-key-hex" // Example: "11842e4034999297038f59f054d1794389758469070e15999837078cec243f55"
	const blockchainID = "0x8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2" // Example blockchain ID

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
	// In a real scenario, this queries the network for the latest nonce.
	// The current implementation has a placeholder for this HTTP call.
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
		// Mock TxID if submission fails to allow example flow to continue
		if submitResp.Response.TxID == "" {
			submitResp.Response.TxID = "mockTxID_for_example_flow_due_to_error"
			fmt.Println("Using mock TxID for example flow as submission failed.")
		}
	}

	if submitResp.Result == 200 || (submitResp.Response.TxID != "" && !strings.HasPrefix(submitResp.Response.TxID, "mockTxID_")) {
		fmt.Printf("Certificate submission potentially successful or pending. Transaction ID: %s\n", submitResp.Response.TxID)
	} else if submitResp.Response.TxID != "" && strings.HasPrefix(submitResp.Response.TxID, "mockTxID_") {
        // This case handles when mock TxID was used due to initial error.
        fmt.Printf("Certificate submission was mocked. Transaction ID: %s\n", submitResp.Response.TxID)
    } else {
		fmt.Printf("Certificate submission failed. Result: %d, Message: %s\n", submitResp.Result, submitResp.Message)
        // If submission truly failed and we don't have a TxID, we can't proceed with below steps.
        // For example's sake, we might still use a mock TxID if none was set.
        if submitResp.Response.TxID == "" {
            submitResp.Response.TxID = "mockTxID_for_failed_submission"
            fmt.Println("Using mock TxID for further example steps after submission failure.")
        }
	}

    // Proceed with outcome and details fetching only if we have a TxID (real or mocked)
    if submitResp.Response.TxID != "" {
        txID := submitResp.Response.TxID

        // 5. Get Transaction Outcome
        fmt.Printf("Polling for transaction outcome for TxID: %s...\n", txID)
        outcomeResp, err := account.GetTransactionOutcome(txID, 15) // Reduced timeout for example
        if err != nil {
            log.Printf("Failed to get transaction outcome (likely due to placeholder NAG): %v\n", err)
            // Mock an outcome if the call fails, for example purposes
            outcomeResp.Result = -1
            outcomeResp.Status = "mock_failure_or_timeout"
            outcomeResp.Message = "Mocked outcome due to placeholder error"
			outcomeResp.BlockID = "" // Ensure BlockID is empty for mock failure
        }
        fmt.Printf("Transaction Outcome: Result %d, Status '%s', BlockID '%s', Message '%s'\n",
            outcomeResp.Result, outcomeResp.Status, outcomeResp.BlockID, outcomeResp.Message)

        // 6. Get Transaction Details (if outcome was notionally successful or has a BlockID)
        // For example, we might try if BlockID is present or if status indicates some success.
        // Here, we'll proceed if a mock or real BlockID is present.
        if outcomeResp.BlockID != "" || (outcomeResp.Status == "confirmed" || outcomeResp.Status == "mock_success") { // Added mock_success for demo
            // If BlockID is empty from a failed/mocked outcome, create a mock BlockID
            blockIDToUse := outcomeResp.BlockID
            if blockIDToUse == "" {
                blockIDToUse = "mockBlockID_for_example"
                fmt.Println("Using mock BlockID for GetTransaction example.")
            }

            fmt.Printf("Fetching transaction details for BlockID %s, TxID %s...\n", blockIDToUse, txID)
            txDetailsResp, err := account.GetTransaction(blockIDToUse, txID)
            if err != nil {
                log.Printf("Failed to get transaction details (likely due to placeholder NAG): %v\n", err)
                // Mock details if call fails
                txDetailsResp.Result = -1
				txDetailsResp.Message = "Mocked transaction details due to error"
                txDetailsResp.Transaction.Data = "Mocked transaction data"
                txDetailsResp.Transaction.Timestamp = utils.GetFormattedUTCTimestamp()
				txDetailsResp.Transaction.TxID = txID
            }
            fmt.Printf("Transaction Details: Result %d, TxID '%s', Data '%s', Timestamp '%s', Message: '%s'\n",
                txDetailsResp.Result, txDetailsResp.Transaction.TxID, txDetailsResp.Transaction.Data, txDetailsResp.Transaction.Timestamp, txDetailsResp.Message)
        } else {
			fmt.Println("Skipping GetTransaction details as BlockID is not available or transaction status is not suitable.")
		}
    } else {
		fmt.Println("Skipping transaction outcome and details fetching as no TxID was available from submission.")
	}


	// 7. Close Account
	account.Close()
	fmt.Println("Account closed.")
}
```

## API Overview

The primary components of this API are:

*   **`circular_protocol_api.Account`**:
    *   Manages account state (address, network, nonce).
    *   Handles interactions with the Circular Network Access Gateway (NAG) for operations like updating account nonce, submitting certificates, and querying transactions.
    *   Provides data signing capabilities.
*   **`circular_protocol_api.Certificate`**:
    *   Represents a data certificate.
    *   Methods to set/get data, get JSON representation, and certificate size.
*   **`utils.GetFormattedUTCTimestamp()`**:
    *   A utility function to generate a UTC timestamp string in the format `YYYY-MM-DDTHH:MM:SSZ`.

**Note on Network Interactions:** The methods involving direct communication with the Circular Network Access Gateway (`UpdateAccount`, `SubmitCertificate`, `GetTransactionOutcome`, `GetTransaction`) currently use placeholder URLs and logic. For actual use, these will need to be configured with the correct NAG API endpoints and request/response structures provided by Circular Protocol. The example above includes mocked responses for these calls to illustrate the intended API flow.

## Contributing

Contributions are welcome! Please fork the repository, make your changes, and submit a pull request.

1.  Fork the Project
2.  Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3.  Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4.  Push to the Branch (`git push origin feature/AmazingFeature`)
5.  Open a Pull Request

## License

Distributed under the MIT License. See `LICENSE` file for more information.
(A `LICENSE` file with the MIT license text should exist in the repository root.)
