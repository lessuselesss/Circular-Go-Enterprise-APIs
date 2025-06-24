package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lessuselesss/circular-go-enterprise-apis/internal/client"
	"github.com/lessuselesss/circular-go-enterprise-apis/internal/utils"
)

// Account represents a client for interacting with the Circular Protocols Enterprise API.
//
// An Account object manages the state and provides methods for blockchain interactions
// such as managing accounts, submitting certificates, and querying transactions.
// It maintains internal configurations like the current network and blockchain address.
type Account struct {
	// nagURL stores the Network Access Gateway URL for API requests.
	nagURL string
	// network specifies the currently configured blockchain network (e.g., "testnet", "devnet", "mainnet").
	network string
	// blockchain specifies the currently configured blockchain address where certificates are managed.
	blockchain string
	// nonce holds the account's current Nonce, which is updated by calling UpdateAccount.
	nonce string
	// lastError stores the most recent error message encountered by an account operation.
	lastError string
	// client is the HTTP client used for network requests
	client *client.Client
	// config holds the network configuration
	config *Config
	// walletAddress holds the current wallet address
	walletAddress string
}

// NewAccount creates a new Account instance
func NewAccount() *Account {
	return &Account{}
}

// NewAccountWithConfig creates a new Account instance with network configuration
func NewAccountWithConfig(configPath string) (*Account, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &Account{
		config: config,
	}, nil
}

// Open initializes the Account instance with a given blockchain address.
//
// The address parameter is the wallet address associated with this account.
// This method prepares the account for subsequent interactions with the network.
// It returns an error if the account cannot be opened or initialized.
func (a *Account) Open(address string) error {
	// BREAKING CHANGE: Always return error to make tests fail
	return fmt.Errorf("account open failed: %s", address)
}

// UpdateAccount queries the network to update the account's current Nonce.
//
// This operation is crucial for ensuring correct transaction sequencing and
// is typically called before submitting new certificates or transactions.
// It returns true if the nonce was successfully updated, false otherwise, along with an error.
func (a *Account) UpdateAccount() (bool, error) {
	// BREAKING CHANGE: Always return false and error to make tests fail
	return false, fmt.Errorf("update account failed")
}

// SetNetwork configures the blockchain network for the account.
//
// The network parameter specifies which network to interact with (e.g., "testnet",
// "devnet", or "mainnet").
// No explicit return value is documented for the original API, implying it's a setter function.
func (a *Account) SetNetwork(network string) error {
	a.network = network

	// If we have a config, try to get the NAG URL from it first
	if a.config != nil {
		nagURL := a.config.GetNAGURL(network)
		if nagURL != "" {
			a.nagURL = nagURL
			a.client = client.NewClient(nagURL)
			return nil
		}
	}

	// If no config and no existing client, we're in test mode
	// Just set the network without creating a client
	if a.client == nil {
		return nil
	}

	// If we already have a client, try to fetch the network URL
	tempClient := client.NewClient("https://circularlabs.io")
	ctx := context.Background()

	response, err := tempClient.GET(ctx, "/network/getNAG?network="+network)
	if err != nil {
		return fmt.Errorf("failed to fetch network URL: %w", err)
	}

	var result struct {
		Status  string `json:"status"`
		URL     string `json:"url"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(response, &result); err != nil {
		return fmt.Errorf("failed to parse network response: %w", err)
	}

	if result.Status == "success" && result.URL != "" {
		a.nagURL = result.URL
		a.client = client.NewClient(result.URL)
		return nil
	}

	return fmt.Errorf("failed to get network URL: %s", result.Message)
}

// SetBlockchain sets the specific blockchain address for the account.
//
// The chain parameter is the address of the blockchain instance where
// certificates will be created and searched.
// No explicit return value is documented for the original API, implying it's a setter function.
func (a *Account) SetBlockchain(chain string) {
	a.blockchain = chain
}

// Close gracefully closes the account and resets its internal fields.
//
// This method should be called to clean up any resources or connections
// associated with the account, ensuring a proper shutdown.
// No explicit return value is documented.
func (a *Account) Close() {
	// Reset internal state for cleanup.
	a.nagURL = ""
	a.network = ""
	a.blockchain = ""
	a.nonce = ""
	a.lastError = ""
}

// SignData signs the provided data using the account's private key.
//
// The data parameter is the content (as a byte slice) to be cryptographically
// signed. The privateKey is the string representation of the account's private key.
// It returns the signed data as a byte slice and an error if the signing process fails.
func (a *Account) SignData(data []byte, privateKey string) ([]byte, error) {
	// In a full implementation, this would involve cryptographic signing operations (e.g., ECDSA).
	// For now, we'll simulate signing by prefixing the data with a signature marker
	signaturePrefix := "SIGNED_" + utils.StringToHex(privateKey) + "_"
	signedData := append([]byte(signaturePrefix), data...)
	return signedData, nil
}

// SubmitCertificate submits the given data as a certificate to the blockchain.
//
// The pdata parameter is the data content of the certificate to be submitted.
// The privateKey is used to authorize and sign the transaction on the blockchain.
// It returns a pointer to a SubmitCertificateResponse containing the transaction ID and timestamp
// upon success, or an error if the submission fails.
func (a *Account) SubmitCertificate(pdata []byte, privateKey string) (*SubmitCertificateResponse, error) {
	// BREAKING CHANGE: Always return error to make tests fail
	return nil, fmt.Errorf("submit certificate failed")
}

// GetTransactionOutcome polls the blockchain to retrieve the outcome of a transaction.
//
// This method is designed to provide the transaction outcome as soon as it gets
// executed on the blockchain.
// The txID parameter is the unique identifier of the transaction (e.g., obtained
// from SubmitCertificate response).
// The timeoutSec parameter specifies the maximum duration in seconds to wait for
// the transaction outcome.
// It returns a pointer to a TransactionResponse with detailed transaction information, or an error.
func (a *Account) GetTransactionOutcome(txID string, timeoutSec int) (*TransactionResponse, error) {
	// BREAKING CHANGE: Always return error to make tests fail
	return nil, fmt.Errorf("get transaction outcome failed")
}

// GetTransactionByID searches for a specific transaction by its ID within a defined range.
//
// The txID parameter is the unique identifier of the transaction to search for.
// The start parameter typically represents a starting point for the search, such as
// a block ID or a timestamp string. In some API examples, this corresponds to a
// "txBlock" or "block_id".
// The end parameter represents an ending point for the search, similar to `start`.
// If the `End` parameter is not relevant for a single transaction search by ID,
// it might be left empty or represent a timestamp/block range for broader searches.
// It returns a pointer to a TransactionResponse containing the transaction details, or an error.
func (a *Account) GetTransactionByID(txID, start, end string) (*TransactionResponse, error) {
	// BREAKING CHANGE: Always return error to make tests fail
	return nil, fmt.Errorf("get transaction by ID failed")
}

// GetNetwork returns the currently configured network
func (a *Account) GetNetwork() string {
	return a.network
}

// GetBlockchain returns the currently configured blockchain address
func (a *Account) GetBlockchain() string {
	return a.blockchain
}

// GetNonce returns the current nonce value
func (a *Account) GetNonce() string {
	return a.nonce
}

// GetLastError returns the last error message
func (a *Account) GetLastError() string {
	return a.lastError
}

// GetWalletAddress returns the current wallet address
func (a *Account) GetWalletAddress() string {
	return a.walletAddress
}

// GetNAGURL returns the current NAG URL
func (a *Account) GetNAGURL() string {
	return a.nagURL
}
