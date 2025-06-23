package api

// SubmitCertificateResponse represents the outcome of a certificate submission.
// It provides the transaction ID and timestamp upon successful submission.
type SubmitCertificateResponse struct {
	Result   int `json:"Result"` // Result code, 200 for success.
	Response struct {
		TxID      string `json:"TxID"`      // The unique transaction ID generated for the certificate.
		Timestamp string `json:"Timestamp"` // The UTC timestamp of the certificate submission.
	} `json:"Response"`
	Node    string `json:"Node"`    // The address of the node that handled the submission.
	Message string `json:"message"` // An optional message, typically present on error (Result != 200).
}

// TransactionResponse represents the detailed outcome of a transaction on the blockchain.
//
// This structure is used for responses from methods like GetTransactionOutcome and
// GetTransactionByID, containing comprehensive details about the transaction
// including blockchain identifiers, fees, status, and timestamps.
type TransactionResponse struct {
	Result   int    `json:"Result"`   // HTTP-like status code indicating the operation's success or failure.
	Response struct {
		BlockID       string  `json:"BlockID"`       // The identifier of the block in which the transaction was recorded.
		BroadcastFee  float64 `json:"BroadcastFee"`  // The fee incurred for broadcasting the transaction.
		DeveloperFee  float64 `json:"DeveloperFee"`  // Any developer fees associated with the transaction.
		From          string  `json:"From"`          // The blockchain address from which the transaction originated.
		GasLimit      float64 `json:"GasLimit"`      // The gas limit set for the transaction.
		ID            string  `json:"ID"`            // The unique identifier of the transaction.
		Instructions  int     `json:"Instructions"`  // The number of instructions processed by the transaction.
		NagFee        float64 `json:"NagFee"`        // The Network Access Gateway fee.
		NodeID        string  `json:"NodeID"`        // The ID of the node that processed the transaction.
		Nonce         string  `json:"Nonce"`         // The nonce value of the account at the time of the transaction.
		OSignature    string  `json:"OSignature"`    // The original signature of the transaction.
		Payload       string  `json:"Payload"`       // The hexadecimal representation of the data payload.
		ProcessingFee float64 `json:"ProcessingFee"` // The fee for processing the transaction.
		ProtocolFee   float64 `json:"ProtocolFee"`   // The protocol fee.
		Status        string  `json:"Status"`        // The execution status of the transaction (e.g., "Executed").
		Timestamp     string  `json:"Timestamp"`     // The UTC timestamp when the transaction occurred.
		To            string  `json:"To"`            // The blockchain address to which the transaction was sent.
		Type          string  `json:"Type"`          // The type of transaction (e.g., "C_TYPE_CERTIFICATE").
	} `json:"Response"`
	Node    string `json:"Node"`    // The address of the node that handled the request.
	Message string `json:"message"` // An optional message, typically present on error (Result != 200).
}