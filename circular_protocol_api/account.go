// Package circular_protocol_api provides functionalities to interact with the Circular Protocol API.
// It allows for account management, transaction signing, certificate submission, and other related operations.
package circular_protocol_api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	btcec "github.com/btcsuite/btcd/btcec/v2" // aliased v2 to btcec
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
)

// Note: Duplicate import block removed for clarity. Ensure this is handled if it was intentional.

const (
	libVersion   = "1.0.1"
	networkURL   = "https://circularlabs.io/network/getNAG?network="
	defaultChain = "0x8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2"
	defaultNAG   = "https://nag.circularlabs.io/NAG.php?cep="
)

// CEPAccount represents a Circular Enterprise Protocol (CEP) account.
// It holds all necessary information to interact with the CEP network,
// including account details, network configuration, and transaction data.
type CEPAccount struct {
	Address     string                 // Address is the unique identifier of the CEP account.
	PublicKey   string                 // PublicKey associated with the account, used for verifying signatures.
	Info        string                 // Info stores general information or metadata related to the account.
	CodeVersion string                 // CodeVersion indicates the version of the library or client.
	LastError   string                 // LastError stores the last error message encountered by account operations.
	NAGURL      string                 // NAGURL is the Network Access Gateway URL used for API requests.
	NetworkNode string                 // NetworkNode specifies the particular network node to connect to.
	Blockchain  string                 // Blockchain identifier for the target blockchain.
	LatestTxID  string                 // LatestTxID stores the ID of the most recent transaction.
	Nonce       int                    // Nonce is the transaction sequence number to prevent replay attacks.
	Data        map[string]interface{} // Data is a flexible map to store any additional account-specific data.
	IntervalSec int                    // IntervalSec defines the polling interval in seconds for operations like GetTransactionOutcome.
}

// NewCEPAccount creates and initializes a new CEPAccount with default values.
// It sets the CodeVersion to the current library version, NAGURL to the default NAG,
// Blockchain to the default chain, initializes an empty Data map, and sets a default IntervalSec.
func NewCEPAccount() *CEPAccount {
	return &CEPAccount{
		CodeVersion: libVersion, // Initialize with the current library version.
		NAGURL:      defaultNAG,
		Blockchain:  defaultChain,
		Data:        make(map[string]interface{}),
		IntervalSec: 2,
	}
}

// Open initializes a CEPAccount with the given address.
// It validates that the address is not empty.
//
// Parameters:
//
//	address: The CEP account address to open.
//
// Returns:
//
//	An error if the address is empty, otherwise nil.
func (a *CEPAccount) Open(address string) error {
	// Ensure the provided address is not an empty string.
	if address == "" {
		return errors.New("invalid address format")
	}
	a.Address = address
	return nil
}

// UpdateAccount fetches the latest nonce for the account from the CEP network.
// It constructs a request to the NAG (Network Access Gateway) to get the wallet nonce.
// The Nonce field of the CEPAccount is updated upon successful retrieval.
//
// Returns:
//
//	A boolean indicating success (true) or failure (false).
//	An error if the account is not open, or if any network or parsing error occurs.
func (a *CEPAccount) UpdateAccount() (bool, error) {
	// Check if the account address has been set.
	if a.Address == "" {
		return false, errors.New("account is not open")
	}

	// Prepare the data payload for the nonce request.
	data := map[string]string{
		"Blockchain": hexFix(a.Blockchain), // Ensure blockchain is hex-encoded with "0x" prefix.
		"Address":    hexFix(a.Address),    // Ensure address is hex-encoded with "0x" prefix.
		"Version":    a.CodeVersion,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		// Error during JSON marshaling.
		return false, err
	}

	// Construct the NAG URL for getting wallet nonce.
	nagAPIURL := a.NAGURL + "Circular_GetWalletNonce_" + a.NetworkNode
	resp, err := http.Post(nagAPIURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		// Network error during the POST request.
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// HTTP error if status is not OK.
		return false, fmt.Errorf("HTTP error! status: %d", resp.StatusCode)
	}

	var jsonResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
		// Error decoding the JSON response.
		return false, err
	}

	// Check the result and extract the nonce.
	if result, ok := jsonResponse["Result"].(float64); ok && result == 200 {
		if response, ok := jsonResponse["Response"].(map[string]interface{}); ok {
			if nonce, ok := response["Nonce"].(float64); ok {
				a.Nonce = int(nonce) + 1 // Update account nonce (incremented by 1 as per original logic).
				return true, nil
			}
		}
	}

	// If the response format is invalid or Nonce is missing.
	return false, errors.New("invalid response format or missing Nonce field")
}

// SetNetwork configures the network for the CEP account by fetching the NAG URL
// associated with the provided network identifier.
//
// Parameters:
//
//	network: The network identifier (e.g., "mainnet", "testnet").
//
// Returns:
//
//	An error if the network request fails, the response status is not OK,
//	or if the response format is invalid. Otherwise, nil.
func (a *CEPAccount) SetNetwork(network string) error {
	// Fetch network configuration from the predefined network URL.
	resp, err := http.Get(networkURL + network)
	if err != nil {
		// Network error during GET request.
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// HTTP error if status is not OK.
		return fmt.Errorf("HTTP error! status: %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		// Error decoding JSON response.
		return err
	}

	// Check for successful status and extract the NAG URL.
	if status, ok := data["status"].(string); ok && status == "success" {
		if url, ok := data["url"].(string); ok {
			a.NAGURL = url // Update the NAGURL for the account.
			return nil
		}
	}

	// If the response indicates failure, return the error message from the response.
	if message, ok := data["message"].(string); ok {
		return errors.New(message)
	}

	// Default error if the URL is not found or another issue occurs.
	return errors.New("failed to get URL")
}

// SetBlockchain sets the blockchain identifier for the CEP account.
//
// Parameters:
//
//	chain: The blockchain identifier string (e.g., a specific chain hash).
func (a *CEPAccount) SetBlockchain(chain string) {
	a.Blockchain = chain
}

// Close resets the CEPAccount fields to their default states.
// This includes clearing address, public key, info, error messages,
// and resetting network configurations, nonce, and data.
func (a *CEPAccount) Close() {
	a.Address = ""
	a.PublicKey = ""
	a.Info = ""
	a.LastError = ""
	a.NAGURL = defaultNAG
	a.NetworkNode = ""
	a.Blockchain = defaultChain
	a.LatestTxID = ""
	a.Nonce = 0
	a.Data = make(map[string]interface{}) // Re-initialize Data to an empty map.
	a.IntervalSec = 2                     // Reset IntervalSec to its default value.
}

// SignData signs the given data string using the provided private key.
// It first decodes the hexadecimal private key, then hashes the input data using SHA256.
// The hash is then signed using the ECDSA private key.
// The resulting signature is returned as a hexadecimal string.
//
// Parameters:
//
//	data: The string data to be signed.
//	privateKey: The hexadecimal string representation of the private key.
//
// Returns:
//
//	A hexadecimal string representation of the signature, or an empty string and an error
//	if the account is not open, or if any error occurs during decoding, hashing, or signing.
func (a *CEPAccount) SignData(data, privateKey string) (string, error) {
	// Ensure the account is open.
	if a.Address == "" {
		return "", errors.New("account is not open")
	}

	// Decode the private key from hex to bytes.
	privKeyBytes, err := hex.DecodeString(hexFix(privateKey)) // hexFix ensures "0x" prefix if needed.
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}

	// Create a private key object from the bytes.
	// S256 is the Bitcoin secp256k1 curve.
	privKey, _ := btcec.PrivKeyFromBytes(privKeyBytes) // Second return value is public key, ignored here. S256() is no longer needed with v2.

	// Hash the data to be signed.
	hasher := sha256.New()
	hasher.Write([]byte(data))
	msgHash := hasher.Sum(nil)

	// Sign the hash.
	// Note: The original code used privKey.Sign which produces an ECDSA signature.
	// For Schnorr signatures, schnorr.Sign(privKey, msgHash) would be used.
	// Using ecdsa.Sign as per v2 library structure.
	signature := ecdsa.Sign(privKey, msgHash)
	if signature == nil {
		return "", errors.New("failed to sign data: signature is nil")
	}

	// Encode the signature to a hex string.
	return hex.EncodeToString(signature.Serialize()), nil
}

// GetTransactionbyID retrieves a transaction by its ID from the CEP network.
// It queries within a specified block range (start and end parameters, though their usage in the API call isn't fully clear from the context).
//
// Parameters:
//
//	txID: The ID of the transaction to retrieve.
//	start: The starting block number for the search range (usage unclear in current NAG call).
//	end: The ending block number for the search range (usage unclear in current NAG call).
//
// Returns:
//
//	A map[string]interface{} containing the JSON response from the NAG, or nil and an error
//	if any network or parsing error occurs, or if the network response status is not OK.
func (a *CEPAccount) GetTransactionbyID(txID string, start, end int) (map[string]interface{}, error) {
	// Prepare the data payload for the request.
	data := map[string]string{
		"Blockchain": hexFix(a.Blockchain), // Ensure blockchain is hex-encoded.
		"ID":         hexFix(txID),         // Ensure transaction ID is hex-encoded.
		"Start":      strconv.Itoa(start),  // Convert start block number to string.
		"End":        strconv.Itoa(end),    // Convert end block number to string.
		"Version":    a.CodeVersion,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Construct the NAG URL for getting a transaction by ID.
	nagAPIURL := a.NAGURL + "Circular_GetTransactionbyID_" + a.NetworkNode
	resp, err := http.Post(nagAPIURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("network response was not ok")
	}

	var jsonResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
		return nil, err
	}

	return jsonResponse, nil
}

// SubmitCertificate prepares and submits a certificate to the CEP network.
// It involves creating a payload, generating a transaction ID, signing the ID,
// and then submitting all necessary data to the NAG.
//
// Parameters:
//
//	pdata: The data string to be included in the certificate.
//	privateKey: The hexadecimal string representation of the private key for signing.
//
// Returns:
//
//	A map[string]interface{} containing the JSON response from the NAG, or nil and an error
//	if the account is not open, or if any error occurs during JSON marshaling, signing,
//	network requests, or response parsing.
func (a *CEPAccount) SubmitCertificate(pdata, privateKey string) (map[string]interface{}, error) {
	// Ensure the account is open.
	if a.Address == "" {
		return nil, errors.New("account is not open")
	}

	// Create the payload object for the certificate.
	payloadObject := map[string]string{
		"Action": "CP_CERTIFICATE",   // Define the action type.
		"Data":   stringToHex(pdata), // Convert certificate data to hex.
	}
	jsonStr, err := json.Marshal(payloadObject)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload object: %w", err)
	}
	payload := stringToHex(string(jsonStr)) // Convert the JSON string of the payload object to hex.

	// Generate a timestamp for the transaction.
	timestamp := getFormattedTimestamp() // Assumes getFormattedTimestamp() is defined elsewhere.

	// Construct the string to be hashed for the transaction ID.
	// This string includes blockchain, addresses, payload, nonce, and timestamp.
	strToHash := hexFix(a.Blockchain) + hexFix(a.Address) + hexFix(a.Address) + payload + strconv.Itoa(a.Nonce) + timestamp
	hasher := sha256.New()
	hasher.Write([]byte(strToHash))
	id := hex.EncodeToString(hasher.Sum(nil)) // The transaction ID.

	// Sign the generated transaction ID.
	signature, err := a.SignData(id, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction ID: %w", err)
	}

	// Prepare the final data for submitting the transaction.
	data := map[string]string{
		"ID":         id,
		"From":       hexFix(a.Address),
		"To":         hexFix(a.Address), // Certificate is sent from and to the same address.
		"Timestamp":  timestamp,
		"Payload":    payload,
		"Nonce":      strconv.Itoa(a.Nonce),
		"Signature":  signature,
		"Blockchain": hexFix(a.Blockchain),
		"Type":       "C_TYPE_CERTIFICATE", // Define the transaction type.
		"Version":    a.CodeVersion,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal final transaction data: %w", err)
	}

	// Construct the NAG URL for adding a transaction.
	nagAPIURL := a.NAGURL + "Circular_AddTransaction_" + a.NetworkNode
	resp, err := http.Post(nagAPIURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("network response was not ok")
	}

	var jsonResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
		return nil, err
	}

	// Upon successful submission, the Nonce might need to be incremented here or after confirming the transaction outcome.
	// The current UpdateAccount logic fetches and increments nonce, so care should be taken to avoid conflicts.
	// a.Nonce++ // Example: Increment nonce locally, or rely on UpdateAccount.

	return jsonResponse, nil
}

// GetTransactionOutcome polls the CEP network for the outcome of a transaction specified by its ID.
// It repeatedly calls GetTransactionbyID until the transaction status is no longer "Pending"
// or a timeout occurs.
//
// Parameters:
//
//	txID: The ID of the transaction to get the outcome for.
//	timeoutSec: The timeout duration in seconds.
//
// Returns:
//
//	A map[string]interface{} containing the transaction response details once it's no longer pending,
//	or nil and an error if a timeout occurs, or if any network or parsing error happens during polling.
func (a *CEPAccount) GetTransactionOutcome(txID string, timeoutSec int) (map[string]interface{}, error) {
	timeout := time.After(time.Duration(timeoutSec) * time.Second)       // Channel that signals after timeoutSec.
	ticker := time.NewTicker(time.Duration(a.IntervalSec) * time.Second) // Channel that ticks at IntervalSec.
	defer ticker.Stop()                                                  // Ensure the ticker is stopped when the function exits.

	for {
		select {
		case <-timeout:
			// Timeout occurred before the transaction outcome was determined.
			return nil, errors.New("timeout exceeded")
		case <-ticker.C:
			// Time to poll for the transaction outcome.
			// The block range 0, 10 is used here as in the original code,
			// its specific meaning in the context of GetTransactionbyID for outcome polling should be verified.
			data, err := a.GetTransactionbyID(txID, 0, 10)
			if err != nil {
				// Error during polling.
				return nil, err
			}

			// Check if the transaction outcome is available.
			if result, ok := data["Result"].(float64); ok && result == 200 {
				if response, ok := data["Response"].(map[string]interface{}); ok {
					if status, ok := response["Status"].(string); ok && status != "Pending" {
						// Transaction is no longer pending, return its details.
						return response, nil
					}
					// If status is still "Pending", continue polling.
				}
			}
			// If response format is unexpected or result is not 200, continue polling or handle error.
			// The current logic will continue polling. Consider adding error handling for non-200 results if they indicate a final failure.
		}
	}
}

// hexFix ensures that a hexadecimal string has a "0x" prefix.
// This is a helper function used internally.
// It is not part of the public API of CEPAccount but is used by its methods.
// If the input string already has "0x", it's returned as is.
// Otherwise, "0x" is prepended.
func hexFix(hexStr string) string {
	if strings.HasPrefix(hexStr, "0x") {
		return hexStr
	}
	return "0x" + hexStr
}

// stringToHex converts a string to its hexadecimal representation.
// This is a helper function used internally.
func stringToHex(s string) string {
	return hex.EncodeToString([]byte(s))
}

// getFormattedTimestamp returns the current UTC timestamp as a string.
// The format is "YYYY-MM-DD HH:MM:SS.mmm UTC".
// This is a helper function used internally.
func getFormattedTimestamp() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05.000") + " UTC"
}

// // SubmitResponse is the expected response structure after submitting a certificate.
// type SubmitResponse struct {
// 	Result   int    `json:"result"`
// 	Message  string `json:"message,omitempty"`
// 	Response struct {
// 		TxID string `json:"TxID,omitempty"`
// 	} `json:"response,omitempty"`
// }

// // TransactionOutcomeResponse is the expected response for a transaction outcome query.
// type TransactionOutcomeResponse struct {
// 	Result  int    `json:"result"`
// 	Message string `json:"message,omitempty"`
// 	BlockID string `json:"BlockID,omitempty"` // Or any other fields NAG returns
// 	Status  string `json:"status,omitempty"`  // e.g., "confirmed", "pending", "failed"
// }

// // TransactionResponse is the expected response when fetching a specific transaction.
// type TransactionResponse struct {
// 	Result      int    `json:"result"`
// 	Message     string `json:"message,omitempty"`
// 	Transaction struct {
// 		TxID      string `json:"txID"`
// 		Data      string `json:"data"` // Example field
// 		Timestamp string `json:"timestamp"`
// 		// Add other relevant transaction fields
// 	} `json:"transaction,omitempty"`
// }

// // NetworkConfig holds URLs for a specific network.
// type NetworkConfig struct {
// 	NAGURL         string
// 	NetworkNodeURL string
// }

// // Define network configurations
// var networkMap = map[string]NetworkConfig{
// 	"testnet": {
// 		NAGURL:         "https://nag-testnet.circular.io/v1", // Placeholder
// 		NetworkNodeURL: "https://node-testnet.circular.io",   // Placeholder
// 	},
// 	"mainnet": {
// 		NAGURL:         "https://nag-mainnet.circular.io/v1", // Placeholder
// 		NetworkNodeURL: "https://node-mainnet.circular.io",   // Placeholder
// 	},
// }

// // Account manages interactions with the Circular Protocol for a specific account.
// type Account struct {
// 	Address        string
// 	Network        string // e.g., "testnet", "mainnet"
// 	Blockchain     string // Blockchain identifier
// 	Nonce          int64  // Account nonce, should be int64 for safety
// 	NAGURL         string // Network Access Gateway URL
// 	NetworkNodeURL string // Network Node URL for things like nonce fetching
// 	Client         *http.Client
// }

// // Open initializes a new account with the given address.
// // It also initializes an HTTP client for the account.
// func (a *Account) Open(address string) {
// 	a.Address = address
// 	a.Client = &http.Client{Timeout: 30 * time.Second} // Default timeout
// 	fmt.Println("Account opened for address:", address)
// }

// // SetNetwork sets the blockchain network (e.g., "testnet", "mainnet")
// // and updates NAGURL and NetworkNodeURL accordingly.
// func (a *Account) SetNetwork(networkName string) error {
// 	config, ok := networkMap[strings.ToLower(networkName)]
// 	if !ok {
// 		return fmt.Errorf("unknown network: %s", networkName)
// 	}
// 	a.Network = networkName
// 	a.NAGURL = config.NAGURL
// 	a.NetworkNodeURL = config.NetworkNodeURL
// 	fmt.Printf("Network set to: %s (NAG: %s, Node: %s)\n", networkName, a.NAGURL, a.NetworkNodeURL)
// 	return nil
// }

// // SetBlockchain sets the blockchain identifier for the account.
// func (a *Account) SetBlockchain(chain string) {
// 	a.Blockchain = chain
// 	fmt.Println("Blockchain set to:", chain)
// }

// // UpdateAccount fetches the latest Nonce for the account from the network node.
// // This is a simplified version. The actual endpoint and response structure
// // for fetching nonce need to be defined based on NAG/Node API.
// func (a *Account) UpdateAccount() error {
// 	if a.Address == "" || a.NetworkNodeURL == "" {
// 		return fmt.Errorf("address or network node URL not set")
// 	}
// 	// Example: GET <NetworkNodeURL>/accounts/<address>/nonce
// 	// The actual endpoint and response structure will depend on the Circular API.
// 	// For now, let's assume a placeholder endpoint and response.
// 	reqURL := fmt.Sprintf("%s/accounts/%s/nonce", a.NetworkNodeURL, a.Address)
// 	resp, err := a.Client.Get(reqURL)
// 	if err != nil {
// 		return fmt.Errorf("failed to fetch nonce: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return fmt.Errorf("failed to fetch nonce: received status code %d", resp.StatusCode)
// 	}

// 	var nonceResp struct {
// 		Nonce int64 `json:"nonce"`
// 	}
// 	if err := json.NewDecoder(resp.Body).Decode(&nonceResp); err != nil {
// 		return fmt.Errorf("failed to decode nonce response: %w", err)
// 	}
// 	a.Nonce = nonceResp.Nonce
// 	fmt.Println("Account Nonce updated to:", a.Nonce)
// 	return nil
// }

// // SignData signs the provided data string using the given private key hex string.
// // It returns the signature as a hex string.
// // This uses SECP256k1, common in many blockchains.
// func (a *Account) SignData(data string, privateKeyHex string) (string, error) {
// 	privKeyBytes, err := hex.DecodeString(privateKeyHex)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to decode private key: %w", err)
// 	}
// 	privKey, _ := btcec.PrivKeyFromBytes(privKeyBytes)

// 	// Hash the data before signing (standard practice)
// 	hashedData := sha256.Sum256([]byte(data))

// 	// Sign using Schnorr signatures (often used with SECP256k1)
// 	// If ECDSA is required: signature, err := ecdsa.SignASN1(rand.Reader, privKey.ToECDSA(), hashedData[:])
// 	signature, err := schnorr.Sign(privKey, hashedData[:])
// 	if err != nil {
// 		return "", fmt.Errorf("failed to sign data: %w", err)
// 	}

// 	return hex.EncodeToString(signature.Serialize()), nil
// }

// // SubmitCertificate creates a certificate from data, signs it, and submits it.
// // This is a simplified version. The actual request payload for NAG needs to be defined.
// func (a *Account) SubmitCertificate(data string, privateKeyHex string) (SubmitResponse, error) {
// 	var response SubmitResponse
// 	if a.Address == "" || a.NAGURL == "" {
// 		response.Result = -1
// 		response.Message = "account address or NAG URL not set"
// 		return response, fmt.Errorf(response.Message)
// 	}

// 	// 1. Create certificate (using the Certificate struct from certificate.go)
// 	cert := Certificate{Data: data}
// 	// In a real scenario, you might structure the data to be signed more carefully,
// 	// possibly including nonce, address, etc. For now, just signing the raw data.
// 	// jsonCert, _ := cert.GetJSONCertificate() // Or some other canonical representation

// 	// 2. Sign the data
// 	// The data to be signed might need to be a specific format including nonce, etc.
// 	// For this example, we'll just sign the raw data string.
// 	// A more robust implementation would construct a specific message to sign.
// 	signature, err := a.SignData(data, privateKeyHex) // Or sign(jsonCert, ...)
// 	if err != nil {
// 		response.Result = -1
// 		response.Message = fmt.Sprintf("failed to sign data: %v", err)
// 		return response, err
// 	}

// 	// 3. Construct payload for NAG
// 	// This is a placeholder. The actual payload structure depends on the NAG API.
// 	payload := map[string]interface{}{
// 		"address":     a.Address,
// 		"blockchain":  a.Blockchain,
// 		"nonce":       a.Nonce,
// 		"certificate": cert.Data, // Or perhaps the JSON representation
// 		"signature":   signature,
// 	}
// 	jsonPayload, err := json.Marshal(payload)
// 	if err != nil {
// 		response.Result = -1
// 		response.Message = fmt.Sprintf("failed to marshal payload: %v", err)
// 		return response, err
// 	}

// 	// 4. Submit to NAG
// 	reqURL := fmt.Sprintf("%s/submit_certificate", a.NAGURL) // Placeholder endpoint
// 	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonPayload))
// 	if err != nil {
// 		response.Result = -1
// 		response.Message = fmt.Sprintf("failed to create request: %v", err)
// 		return response, err
// 	}
// 	req.Header.Set("Content-Type", "application/json")

// 	resp, err := a.Client.Do(req)
// 	if err != nil {
// 		response.Result = -1
// 		response.Message = fmt.Sprintf("failed to submit certificate: %v", err)
// 		return response, err
// 	}
// 	defer resp.Body.Close()

// 	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
// 		// Try to read body for more info if JSON decoding fails
// 		// bodyBytes, _ := io.ReadAll(resp.Body)
// 		// return SubmitResponse{Result: resp.StatusCode, Message: "failed to decode response: " + string(bodyBytes)}, err
// 		return SubmitResponse{Result: resp.StatusCode, Message: "failed to decode response body"}, fmt.Errorf("failed to decode submit response: %w", err)
// 	}

// 	if response.Result == 200 || (resp.StatusCode >= 200 && resp.StatusCode < 300) {
// 		a.Nonce++ // Increment nonce on successful submission
// 		fmt.Println("Certificate submitted. New Nonce:", a.Nonce)
// 	} else if response.Message == "" {
// 		response.Message = fmt.Sprintf("NAG returned non-success status: %d", resp.StatusCode)
// 	}
// 	return response, nil
// }

// // GetTransactionOutcome polls for the outcome of a transaction.
// // This is a simplified version. Endpoint and polling logic might be more complex.
// func (a *Account) GetTransactionOutcome(txID string, timeoutSec int) (TransactionOutcomeResponse, error) {
// 	var response TransactionOutcomeResponse
// 	if a.NAGURL == "" {
// 		response.Result = -1
// 		response.Message = "NAG URL not set"
// 		return response, fmt.Errorf(response.Message)
// 	}

// 	startTime := time.Now()
// 	endpoint := fmt.Sprintf("%s/transactions/%s/outcome", a.NAGURL, txID) // Placeholder

// 	for {
// 		if time.Since(startTime).Seconds() > float64(timeoutSec) {
// 			response.Result = -1
// 			response.Message = "timeout waiting for transaction outcome"
// 			return response, fmt.Errorf(response.Message)
// 		}

// 		req, err := http.NewRequest("GET", endpoint, nil)
// 		if err != nil {
// 			response.Result = -1
// 			response.Message = "failed to create request"
// 			return response, err
// 		}

// 		httpResp, err := a.Client.Do(req)
// 		if err != nil {
// 			// Network error, wait and retry
// 			time.Sleep(5 * time.Second) // Wait 5 seconds before retrying
// 			continue
// 		}

// 		err = json.NewDecoder(httpResp.Body).Decode(&response)
// 		httpResp.Body.Close() // Close body promptly

// 		if err != nil {
// 			// Error decoding, maybe NAG is down or format changed
// 			response.Result = httpResp.StatusCode
// 			response.Message = "failed to decode transaction outcome response"
// 			return response, fmt.Errorf("%s: %w", response.Message, err)
// 		}

// 		// Assuming a "status" field indicates completion (e.g., "confirmed", "failed")
// 		// or result indicates finality. This logic depends on NAG API.
// 		if response.Result == 200 && (response.Status == "confirmed" || response.Status == "failed") {
// 			fmt.Printf("Transaction outcome received for %s: %s\n", txID, response.Status)
// 			return response, nil
// 		}
// 		if response.Result != 200 && response.Result != 0 && response.Result != http.StatusProcessing && response.Result != http.StatusAccepted {
// 			// If NAG indicates an error other than "still processing"
// 			return response, fmt.Errorf("NAG error fetching outcome: %s (status: %d)", response.Message, response.Result)
// 		}

// 		time.Sleep(5 * time.Second) // Polling interval
// 	}
// }

// // GetTransaction retrieves details of a specific transaction.
// // This is a simplified version. Endpoint and response structure depend on NAG API.
// func (a *Account) GetTransaction(blockID string, txID string) (TransactionResponse, error) {
// 	var response TransactionResponse
// 	if a.NAGURL == "" {
// 		response.Result = -1
// 		response.Message = "NAG URL not set"
// 		return response, fmt.Errorf(response.Message)
// 	}

// 	// Example: NAGURL/transactions/<blockID>/<txID> - This is a guess
// 	endpoint := fmt.Sprintf("%s/transactions/%s/%s", a.NAGURL, blockID, txID) // Placeholder
// 	req, err := http.NewRequest("GET", endpoint, nil)
// 	if err != nil {
// 		response.Result = -1
// 		response.Message = "failed to create request"
// 		return response, err
// 	}

// 	httpResp, err := a.Client.Do(req)
// 	if err != nil {
// 		response.Result = -1
// 		response.Message = "failed to get transaction"
// 		return response, err
// 	}
// 	defer httpResp.Body.Close()

// 	if err := json.NewDecoder(httpResp.Body).Decode(&response); err != nil {
// 		response.Result = httpResp.StatusCode
// 		response.Message = "failed to decode transaction response"
// 		return response, fmt.Errorf("%s: %w", response.Message, err)
// 	}

// 	if response.Result != 200 && response.Message == "" {
// 		response.Message = fmt.Sprintf("NAG returned non-success status: %d for GetTransaction", httpResp.StatusCode)
// 	}
// 	return response, nil
// }

// // Close resets the account fields.
// func (a *Account) Close() {
// 	a.Address = ""
// 	a.Network = ""
// 	a.Blockchain = ""
// 	a.Nonce = 0
// 	a.NAGURL = ""
// 	a.NetworkNodeURL = ""
// 	a.Client = nil // Release the client
// 	fmt.Println("Account closed and reset.")
// }
