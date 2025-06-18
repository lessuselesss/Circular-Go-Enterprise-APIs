package circular_protocol_api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
)

// SubmitResponse is the expected response structure after submitting a certificate.
type SubmitResponse struct {
	Result   int    `json:"result"`
	Message  string `json:"message,omitempty"`
	Response struct {
		TxID string `json:"TxID,omitempty"`
	} `json:"response,omitempty"`
}

// TransactionOutcomeResponse is the expected response for a transaction outcome query.
type TransactionOutcomeResponse struct {
	Result  int    `json:"result"`
	Message string `json:"message,omitempty"`
	BlockID string `json:"BlockID,omitempty"` // Or any other fields NAG returns
	Status  string `json:"status,omitempty"`  // e.g., "confirmed", "pending", "failed"
}

// TransactionResponse is the expected response when fetching a specific transaction.
type TransactionResponse struct {
	Result      int    `json:"result"`
	Message     string `json:"message,omitempty"`
	Transaction struct {
		TxID      string `json:"txID"`
		Data      string `json:"data"` // Example field
		Timestamp string `json:"timestamp"`
		// Add other relevant transaction fields
	} `json:"transaction,omitempty"`
}

// NetworkConfig holds URLs for a specific network.
type NetworkConfig struct {
	NAGURL         string
	NetworkNodeURL string
}

// Define network configurations
var networkMap = map[string]NetworkConfig{
	"testnet": {
		NAGURL:         "https://nag-testnet.circular.io/v1", // Placeholder
		NetworkNodeURL: "https://node-testnet.circular.io",   // Placeholder
	},
	"mainnet": {
		NAGURL:         "https://nag-mainnet.circular.io/v1", // Placeholder
		NetworkNodeURL: "https://node-mainnet.circular.io",   // Placeholder
	},
}

// Account manages interactions with the Circular Protocol for a specific account.
type Account struct {
	Address        string
	Network        string // e.g., "testnet", "mainnet"
	Blockchain     string // Blockchain identifier
	Nonce          int64  // Account nonce, should be int64 for safety
	NAGURL         string // Network Access Gateway URL
	NetworkNodeURL string // Network Node URL for things like nonce fetching
	Client         *http.Client
}

// Open initializes a new account with the given address.
// It also initializes an HTTP client for the account.
func (a *Account) Open(address string) {
	a.Address = address
	a.Client = &http.Client{Timeout: 30 * time.Second} // Default timeout
	fmt.Println("Account opened for address:", address)
}

// SetNetwork sets the blockchain network (e.g., "testnet", "mainnet")
// and updates NAGURL and NetworkNodeURL accordingly.
func (a *Account) SetNetwork(networkName string) error {
	config, ok := networkMap[strings.ToLower(networkName)]
	if !ok {
		return fmt.Errorf("unknown network: %s", networkName)
	}
	a.Network = networkName
	a.NAGURL = config.NAGURL
	a.NetworkNodeURL = config.NetworkNodeURL
	fmt.Printf("Network set to: %s (NAG: %s, Node: %s)\n", networkName, a.NAGURL, a.NetworkNodeURL)
	return nil
}

// SetBlockchain sets the blockchain identifier for the account.
func (a *Account) SetBlockchain(chain string) {
	a.Blockchain = chain
	fmt.Println("Blockchain set to:", chain)
}

// UpdateAccount fetches the latest Nonce for the account from the network node.
// This is a simplified version. The actual endpoint and response structure
// for fetching nonce need to be defined based on NAG/Node API.
func (a *Account) UpdateAccount() error {
	if a.Address == "" || a.NetworkNodeURL == "" {
		return fmt.Errorf("address or network node URL not set")
	}
	// Example: GET <NetworkNodeURL>/accounts/<address>/nonce
	// The actual endpoint and response structure will depend on the Circular API.
	// For now, let's assume a placeholder endpoint and response.
	reqURL := fmt.Sprintf("%s/accounts/%s/nonce", a.NetworkNodeURL, a.Address)
	resp, err := a.Client.Get(reqURL)
	if err != nil {
		return fmt.Errorf("failed to fetch nonce: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch nonce: received status code %d", resp.StatusCode)
	}

	var nonceResp struct {
		Nonce int64 `json:"nonce"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&nonceResp); err != nil {
		return fmt.Errorf("failed to decode nonce response: %w", err)
	}
	a.Nonce = nonceResp.Nonce
	fmt.Println("Account Nonce updated to:", a.Nonce)
	return nil
}

// SignData signs the provided data string using the given private key hex string.
// It returns the signature as a hex string.
// This uses SECP256k1, common in many blockchains.
func (a *Account) SignData(data string, privateKeyHex string) (string, error) {
	privKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode private key: %w", err)
	}
	privKey, _ := btcec.PrivKeyFromBytes(privKeyBytes)

	// Hash the data before signing (standard practice)
	hashedData := sha256.Sum256([]byte(data))

	// Sign using Schnorr signatures (often used with SECP256k1)
	// If ECDSA is required: signature, err := ecdsa.SignASN1(rand.Reader, privKey.ToECDSA(), hashedData[:])
	signature, err := schnorr.Sign(privKey, hashedData[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign data: %w", err)
	}

	return hex.EncodeToString(signature.Serialize()), nil
}

// SubmitCertificate creates a certificate from data, signs it, and submits it.
// This is a simplified version. The actual request payload for NAG needs to be defined.
func (a *Account) SubmitCertificate(data string, privateKeyHex string) (SubmitResponse, error) {
	var response SubmitResponse
	if a.Address == "" || a.NAGURL == "" {
		response.Result = -1
		response.Message = "account address or NAG URL not set"
		return response, fmt.Errorf(response.Message)
	}

	// 1. Create certificate (using the Certificate struct from certificate.go)
	cert := Certificate{Data: data}
	// In a real scenario, you might structure the data to be signed more carefully,
	// possibly including nonce, address, etc. For now, just signing the raw data.
	// jsonCert, _ := cert.GetJSONCertificate() // Or some other canonical representation

	// 2. Sign the data
	// The data to be signed might need to be a specific format including nonce, etc.
	// For this example, we'll just sign the raw data string.
	// A more robust implementation would construct a specific message to sign.
	signature, err := a.SignData(data, privateKeyHex) // Or sign(jsonCert, ...)
	if err != nil {
		response.Result = -1
		response.Message = fmt.Sprintf("failed to sign data: %v", err)
		return response, err
	}

	// 3. Construct payload for NAG
	// This is a placeholder. The actual payload structure depends on the NAG API.
	payload := map[string]interface{}{
		"address":     a.Address,
		"blockchain":  a.Blockchain,
		"nonce":       a.Nonce,
		"certificate": cert.Data, // Or perhaps the JSON representation
		"signature":   signature,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		response.Result = -1
		response.Message = fmt.Sprintf("failed to marshal payload: %v", err)
		return response, err
	}

	// 4. Submit to NAG
	reqURL := fmt.Sprintf("%s/submit_certificate", a.NAGURL) // Placeholder endpoint
	req, err := http.NewRequest("POST", reqURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		response.Result = -1
		response.Message = fmt.Sprintf("failed to create request: %v", err)
		return response, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.Client.Do(req)
	if err != nil {
		response.Result = -1
		response.Message = fmt.Sprintf("failed to submit certificate: %v", err)
		return response, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		// Try to read body for more info if JSON decoding fails
		// bodyBytes, _ := io.ReadAll(resp.Body)
		// return SubmitResponse{Result: resp.StatusCode, Message: "failed to decode response: " + string(bodyBytes)}, err
		return SubmitResponse{Result: resp.StatusCode, Message: "failed to decode response body"}, fmt.Errorf("failed to decode submit response: %w", err)
	}

	if response.Result == 200 || (resp.StatusCode >= 200 && resp.StatusCode < 300) {
		a.Nonce++ // Increment nonce on successful submission
		fmt.Println("Certificate submitted. New Nonce:", a.Nonce)
	} else if response.Message == "" {
		response.Message = fmt.Sprintf("NAG returned non-success status: %d", resp.StatusCode)
	}
	return response, nil
}

// GetTransactionOutcome polls for the outcome of a transaction.
// This is a simplified version. Endpoint and polling logic might be more complex.
func (a *Account) GetTransactionOutcome(txID string, timeoutSec int) (TransactionOutcomeResponse, error) {
	var response TransactionOutcomeResponse
	if a.NAGURL == "" {
		response.Result = -1
		response.Message = "NAG URL not set"
		return response, fmt.Errorf(response.Message)
	}

	startTime := time.Now()
	endpoint := fmt.Sprintf("%s/transactions/%s/outcome", a.NAGURL, txID) // Placeholder

	for {
		if time.Since(startTime).Seconds() > float64(timeoutSec) {
			response.Result = -1
			response.Message = "timeout waiting for transaction outcome"
			return response, fmt.Errorf(response.Message)
		}

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			response.Result = -1
			response.Message = "failed to create request"
			return response, err
		}

		httpResp, err := a.Client.Do(req)
		if err != nil {
			// Network error, wait and retry
			time.Sleep(5 * time.Second) // Wait 5 seconds before retrying
			continue
		}

		err = json.NewDecoder(httpResp.Body).Decode(&response)
		httpResp.Body.Close() // Close body promptly

		if err != nil {
			// Error decoding, maybe NAG is down or format changed
			response.Result = httpResp.StatusCode
			response.Message = "failed to decode transaction outcome response"
			return response, fmt.Errorf("%s: %w", response.Message, err)
		}

		// Assuming a "status" field indicates completion (e.g., "confirmed", "failed")
		// or result indicates finality. This logic depends on NAG API.
		if response.Result == 200 && (response.Status == "confirmed" || response.Status == "failed") {
			fmt.Printf("Transaction outcome received for %s: %s\n", txID, response.Status)
			return response, nil
		}
		if response.Result != 200 && response.Result != 0 && response.Result != http.StatusProcessing && response.Result != http.StatusAccepted {
			// If NAG indicates an error other than "still processing"
			return response, fmt.Errorf("NAG error fetching outcome: %s (status: %d)", response.Message, response.Result)
		}

		time.Sleep(5 * time.Second) // Polling interval
	}
}

// GetTransaction retrieves details of a specific transaction.
// This is a simplified version. Endpoint and response structure depend on NAG API.
func (a *Account) GetTransaction(blockID string, txID string) (TransactionResponse, error) {
	var response TransactionResponse
	if a.NAGURL == "" {
		response.Result = -1
		response.Message = "NAG URL not set"
		return response, fmt.Errorf(response.Message)
	}

	// Example: NAGURL/transactions/<blockID>/<txID> - This is a guess
	endpoint := fmt.Sprintf("%s/transactions/%s/%s", a.NAGURL, blockID, txID) // Placeholder
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		response.Result = -1
		response.Message = "failed to create request"
		return response, err
	}

	httpResp, err := a.Client.Do(req)
	if err != nil {
		response.Result = -1
		response.Message = "failed to get transaction"
		return response, err
	}
	defer httpResp.Body.Close()

	if err := json.NewDecoder(httpResp.Body).Decode(&response); err != nil {
		response.Result = httpResp.StatusCode
		response.Message = "failed to decode transaction response"
		return response, fmt.Errorf("%s: %w", response.Message, err)
	}

	if response.Result != 200 && response.Message == "" {
		response.Message = fmt.Sprintf("NAG returned non-success status: %d for GetTransaction", httpResp.StatusCode)
	}
	return response, nil
}

// Close resets the account fields.
func (a *Account) Close() {
	a.Address = ""
	a.Network = ""
	a.Blockchain = ""
	a.Nonce = 0
	a.NAGURL = ""
	a.NetworkNodeURL = ""
	a.Client = nil // Release the client
	fmt.Println("Account closed and reset.")
}
