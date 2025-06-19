// Package utils provides a collection of helper functions and utilities
// supporting the Circular Protocol Go library. This includes functions for making HTTP requests,
// data type conversions (e.g., hex to string), cryptographic operations like hashing and signing,
// and other general-purpose utilities.
package utils

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// SendRequest performs an HTTP POST request with JSON data to a specified NAG (Network Access Gateway) function.
// It handles JSON marshalling of the request data and unmarshalling of the JSON response.
// Note: This function currently returns a map[string]interface{} for both success and error cases.
// The error case is indicated by "Result": "500" within the map.
// A more idiomatic Go approach would be to return (map[string]interface{}, error).
//
// Parameters:
//
//	data: The data to be sent as the JSON body of the request. Can be any type that can be marshalled to JSON.
//	nagFunction: The specific NAG API function name to be appended to the nagURL.
//	nagURL: The base URL of the Network Access Gateway.
//
// Returns:
//
//	A map[string]interface{} representing the JSON response from the server or an error map.
//	If an error occurs during request creation, sending, or response processing,
//	the map will typically contain "Result": "500" and a "Response" field with an error message.
func SendRequest(data interface{}, nagFunction string, nagURL string) map[string]interface{} {
	url := nagURL + nagFunction // Construct the full URL.

	// Convert the request data to JSON.
	jsonData, err := json.Marshal(data)
	if err != nil {
		// Handle error during JSON marshalling.
		return map[string]interface{}{
			"Result":   "500",
			"Response": "Wrong JSON format for request data",
		}
	}

	// Create a new HTTP POST request.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		// Handle error during request creation.
		return map[string]interface{}{
			"Result":   "500",
			"Response": "Error during the creation of the HTTP request",
		}
	}
	req.Header.Set("Content-Type", "application/json") // Set JSON content type.

	// Send the request using the default HTTP client.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// Handle error during request sending.
		return map[string]interface{}{
			"Result":   "500",
			"Response": "Error during the sending of the HTTP request",
		}
	}
	defer resp.Body.Close() // Ensure the response body is closed.

	// Read the entire response body.
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		// Handle error during response body reading.
		return map[string]interface{}{
			"Result":   "500",
			"Response": "Error during the reading of the HTTP response body",
		}
	}

	// Check if the HTTP status code indicates success (200 OK).
	if resp.StatusCode != http.StatusOK {
		// Handle non-OK HTTP status.
		// The body might contain more error details from NAG, so we try to return it if possible,
		// otherwise a generic error.
		errorMsg := fmt.Sprintf("HTTP request failed with status code %d. Response body: %s", resp.StatusCode, string(bodyBytes))
		// Attempt to unmarshal into a map anyway, as NAG might still send JSON.
		var errorResponse map[string]interface{}
		if json.Unmarshal(bodyBytes, &errorResponse) == nil {
			if _, ok := errorResponse["Response"]; !ok { // If no "Response" field, put the body there.
				errorResponse["Response"] = errorMsg
			}
			if _, ok := errorResponse["Result"]; !ok { // Ensure Result indicates error.
				errorResponse["Result"] = fmt.Sprintf("%d", resp.StatusCode)
			}
			return errorResponse
		}
		// If unmarshalling fails, return a generic error map.
		return map[string]interface{}{
			"Result":   fmt.Sprintf("%d", resp.StatusCode),
			"Response": errorMsg,
		}
	}

	// Unmarshal the response body into a map.
	var response map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		// Handle error during JSON unmarshalling of the response.
		return map[string]interface{}{
			"Result":   "500",
			"Response": "Error during the decoding of the JSON response",
		}
	}

	return response // Return the successfully parsed response.
}

// PadNumber pads a given integer with a leading zero if it is less than 10.
//
// Parameters:
//
//	number: The integer to pad.
//
// Returns:
//
//	A string representation of the number, padded with a leading zero if necessary.
func PadNumber(number int) string {
	if number < 10 {
		return fmt.Sprintf("0%d", number) // Add leading zero for single-digit numbers.
	}
	return fmt.Sprintf("%d", number) // Return number as string for two or more digits.
}

// GetFormattedTimestamp generates a formatted timestamp string representing the current time.
// The format produced is "YYYY:MM:DD-HH:MM:SS".
// Note: This function uses the local time of the system, not UTC.
//
// Returns:
//
//	A string with the current time formatted as "YYYY:MM:DD-HH:MM:SS".
func GetFormattedTimestamp() string {
	t := time.Now() // Get current local time.
	// Format the timestamp.
	return fmt.Sprintf("%d:%s:%s-%s:%s:%s",
		t.Year(),
		PadNumber(int(t.Month())),
		PadNumber(t.Day()),
		PadNumber(t.Hour()),
		PadNumber(t.Minute()),
		PadNumber(t.Second()))
}

// ECDSASignature defines the structure for an ECDSA signature, containing
// the R and S integer components. This structure is typically used for
// ASN.1 DER encoding/decoding of ECDSA signatures.
type ECDSASignature struct {
	R, S *big.Int // R and S are the two integer components of an ECDSA signature.
}

// SignMessage signs a message string using a hexadecimal ECDSA private key.
// It hashes the message using chainhash.HashB (similar to Bitcoin's double SHA256)
// and then signs the hash. The resulting signature is DER encoded and returned as a hex string.
// The function also returns the R and S components of the signature.
// Note: This function, like SendRequest, returns a map for success/error cases.
//
// Parameters:
//
//	message: The string message to be signed.
//	privateKey: The ECDSA private key in hexadecimal string format.
//
// Returns:
//
//	A map[string]interface{}. On success, it contains "Signature" (hex string), "R" (*big.Int), and "S" (*big.Int).
//	On error (e.g., private key decoding, signing, ASN.1 marshalling), it contains "Result": "500"
//	and "Response" with an error message.
func SignMessage(message string, privateKey string) map[string]interface{} {
	// Decode the hexadecimal private key.
	bytesPrivateKey, err := hex.DecodeString(privateKey)
	if err != nil {
		return map[string]interface{}{
			"Result":   "500",
			"Response": "Error during the decoding of the private key: " + err.Error(),
		}
	}

	// Reconstruct the private key.
	privKey := secp256k1.PrivKeyFromBytes(bytesPrivateKey)

	// Hash the message. chainhash.HashB performs a single SHA256 in this context, not double.
	// For double SHA256, chainhash.DoubleHashB would be used. The original code used HashB.
	messageHash := chainhash.HashB([]byte(message))

	// Sign the hash.
	r, s, err := ecdsa.Sign(rand.Reader, privKey.ToECDSA(), messageHash)
	if err != nil {
		return map[string]interface{}{
			"Result":   "500",
			"Response": "Error during the signing of the message: " + err.Error(),
		}
	}

	// Marshal the signature into ASN.1 DER format.
	derSignature, err := asn1.Marshal(ECDSASignature{R: r, S: s})
	if err != nil {
		return map[string]interface{}{
			"Result":   "500",
			"Response": "Error during the ASN.1 DER encoding of the signature: " + err.Error(),
		}
	}

	// Encode the DER signature to a hex string.
	stringDERSignature := hex.EncodeToString(derSignature)
	return map[string]interface{}{ // Success case
		"Signature": stringDERSignature,
		"R":         r,
		"S":         s,
	}
}

// StringToHex converts a standard string to its hexadecimal representation.
//
// Parameters:
//
//	str: The string to convert.
//
// Returns:
//
//	The hexadecimal encoded string.
func StringToHex(str string) string {
	return hex.EncodeToString([]byte(str))
}

// HexToString converts a hexadecimal encoded string back to its original string representation.
//
// Parameters:
//
//	hexStr: The hexadecimal string to decode.
//
// Returns:
//
//	The decoded string, or an empty string and an error if decoding fails.
func HexToString(hexStr string) (string, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err // Propagate the error from hex.DecodeString.
	}
	return string(bytes), nil
}

// HexFix processes a word, expecting it to be a hexadecimal representation.
// If it's an integer, it formats it as a hexadecimal string.
// If it's a string, it removes the "0x" prefix if present.
// For other types, it returns an empty string.
//
// Parameters:
//
//	word: The input value (int or string) to be processed.
//
// Returns:
//
//	A string representing the processed hexadecimal value, or an empty string for unsupported types.
func HexFix(word interface{}) string {
	switch v := word.(type) {
	case int:
		return fmt.Sprintf("%x", v) // Format integer as hex.
	case string:
		if len(v) >= 2 && v[0:2] == "0x" { // Check for "0x" prefix.
			return v[2:] // Return string without prefix.
		}
		return v // Return original string if no "0x" prefix.
	default:
		// For unsupported types, return an empty string.
		// Consider returning an error for unexpected types in a stricter implementation.
		return ""
	}
}

// Sha256 calculates the SHA-256 hash of a given string data and returns it
// as a hexadecimal string.
//
// Parameters:
//
//	data: The string data to hash.
//
// Returns:
//
//	The SHA-256 hash encoded as a hexadecimal string.
func Sha256(data string) string {
	hash := sha256.Sum256([]byte(data)) // Calculate SHA-256 hash.
	return hex.EncodeToString(hash[:])  // Convert hash bytes to hex string.
}

// bytesToHex converts a slice of bytes to its hexadecimal string representation.
// This is an unexported helper function.
func bytesToHex(b []byte) string {
	return hex.EncodeToString(b)
}

// NAG (Network Access Gateway) function name constants.
// These constants define the specific API endpoint names used when interacting with the NAG.
const (
	TEST_CONTRACT               = "Circular_TestContract_"            // TEST_CONTRACT is used for testing contract interactions.
	CALL_CONTRACT               = "Circular_CallContract_"            // CALL_CONTRACT is used for calling smart contract functions.
	CHECK_WALLET                = "Circular_CheckWallet_"             // CHECK_WALLET is used for checking wallet status or existence.
	GET_WALLET                  = "Circular_GetWallet_"               // GET_WALLET is used for retrieving wallet details.
	GET_LATEST_TRANSACTIONS     = "Circular_GetLatestTransactions_"   // GET_LATEST_TRANSACTIONS is used for fetching the most recent transactions.
	GET_WALLET_BALANCE          = "Circular_GetWalletBalance_"        // GET_WALLET_BALANCE is used for querying a wallet's balance.
	REGISTER_WALLET             = "Circular_RegisterWallet_"          // REGISTER_WALLET is used for registering a new wallet.
	GET_DOMAIN                  = "Circular_GetDomain_"               // GET_DOMAIN is used for retrieving domain-related information.
	GET_ASSET_LIST              = "Circular_GetAssetList_"            // GET_ASSET_LIST is used for fetching a list of assets.
	GET_ASSET                   = "Circular_GetAsset_"                // GET_ASSET is used for retrieving details of a specific asset.
	GET_ASSET_SUPPLY            = "Circular_GetAssetSupply_"          // GET_ASSET_SUPPLY is used for getting the total supply of an asset.
	GET_VOUCHER                 = "Circular_GetVoucher_"              // GET_VOUCHER is used for retrieving voucher information.
	GET_BLOCK_RANGE             = "Circular_GetBlockRange_"           // GET_BLOCK_RANGE is used for fetching a range of blocks.
	GET_BLOCK                   = "Circular_GetBlock_"                // GET_BLOCK is used for retrieving a specific block.
	GET_BLOCK_COUNT             = "Circular_GetBlockCount_"           // GET_BLOCK_COUNT is used for getting the total number of blocks.
	GET_ANALYTICS               = "Circular_GetAnalytics_"            // GET_ANALYTICS is used for fetching analytics data.
	GET_BLOCKCHAINS             = "Circular_GetBlockchains_"          // GET_BLOCKCHAINS is used for getting a list of available blockchains.
	GET_PENDING_TRANSACTION     = "Circular_GetPendingTransaction_"   // GET_PENDING_TRANSACTION is used for fetching pending transactions.
	GET_TRANSACTION_BY_ID       = "Circular_GetTransactionbyID_"      // GET_TRANSACTION_BY_ID is used for retrieving a transaction by its ID.
	GET_TRANSACTION_BY_NODE     = "Circular_GetTransactionbyNode_"    // GET_TRANSACTION_BY_NODE is used for fetching transactions by a specific node.
	GET_TRANSACTIONS_BY_ADDRESS = "Circular_GetTransactionbyAddress_" // GET_TRANSACTIONS_BY_ADDRESS is used for fetching transactions for a given address.
	GET_TRANSACTION_BY_DATE     = "Circular_GetTransactionbyDate_"    // GET_TRANSACTION_BY_DATE is used for fetching transactions by date.
	SEND_TRANSACTION            = "Circular_AddTransaction_"          // SEND_TRANSACTION is used for submitting a new transaction.
	GET_WALLET_NONCE            = "Circular_GetWalletNonce_"          // GET_WALLET_NONCE is used for retrieving the nonce of a wallet.
)
