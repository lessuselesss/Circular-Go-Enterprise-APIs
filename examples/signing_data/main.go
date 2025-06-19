// Package main provides an example of data signing and `CCertificate` object usage
// from the `circular_protocol_api`. It demonstrates creating a certificate,
// setting its data, retrieving data in various formats, and signing arbitrary data
// using an account's private key.
package main

import (
	"fmt"
	"log"

	"github.com/circular-protocol/circular-go/circular_protocol_api"
	"github.com/circular-protocol/circular-go/utils"
)

// main function demonstrates a sequence of operations:
// 1. Creation and manipulation of a `CCertificate` object:
//   - Instantiating a new `CCertificate`.
//   - Setting data for the certificate.
//   - Retrieving and printing the certificate's data.
//   - Getting and printing the certificate's size (JSON serialized).
//   - Getting and printing the JSON representation of the certificate.
//
// 2. Data Signing using the `SignData` method from an `CEPAccount` object:
//   - Defining sample data and a private key (for example purposes only).
//   - Instantiating a `CEPAccount` (note: for `SignData` as currently implemented,
//     the account doesn't need to be opened or connected to a network if `SignData`
//     is purely cryptographic and doesn't rely on live account state like nonce).
//   - Signing the sample data.
//   - Printing the resulting hexadecimal signature.
//
// Note: This example uses a hardcoded private key for simplicity.
// In real applications, private keys must be handled securely and not embedded in code.
func main() {
	fmt.Println("Circular Go Enterprise API - Data Signing and Certificate Example")
	// Display current timestamp for reference, using the corrected utils.GetFormattedTimestamp function.
	fmt.Printf("Current Timestamp (from utils.GetFormattedTimestamp): %s\n", utils.GetFormattedTimestamp())

	// Example private key and sample data.
	// IMPORTANT: Never use real private keys directly in code for production environments.
	// This is for demonstration purposes only.
	const privateKeyHex = "11842e4034999297038f59f054d1794389758469070e15999837078cec243f55" // Example placeholder private key
	const sampleData = "This is some sample data to be signed."

	// --- 1. CCertificate Object Usage ---
	fmt.Println("\n--- CCertificate Object Usage ---")
	// Instantiate a new CCertificate object from the circular_protocol_api.
	cert := &circular_protocol_api.CCertificate{}

	// Set data for the certificate. The SetData method hex-encodes the input string.
	cert.SetData(sampleData)
	fmt.Printf("Set certificate data to (original): '%s'\n", sampleData)

	// Retrieve the data from the certificate. GetData decodes it from hex.
	certDataStr, err := cert.GetData()
	if err != nil {
		// If GetData fails (e.g., invalid hex in cert.Data), log fatal error.
		log.Fatalf("Failed to get certificate data: %v", err)
	}
	fmt.Printf("Retrieved Certificate Data (decoded): '%s'\n", certDataStr)

	// Get the size of the certificate when serialized to JSON.
	certSize, err := cert.GetCertificateSize()
	if err != nil {
		// If GetCertificateSize fails (e.g., JSON marshalling error), log fatal error.
		log.Fatalf("Failed to get certificate size: %v", err)
	}
	fmt.Printf("Certificate Size (JSON serialized): %d bytes\n", certSize)

	// Get the JSON string representation of the certificate.
	jsonCert, err := cert.GetJSONCertificate()
	if err != nil {
		// If GetJSONCertificate fails, log fatal error.
		log.Fatalf("Failed to get JSON certificate: %v", err)
	}
	fmt.Printf("JSON Certificate: %s\n", jsonCert)

	// --- 2. Data Signing ---
	fmt.Println("\n--- Data Signing ---")
	// The SignData method is part of the CEPAccount type.
	// For this example, we instantiate an account.
	// If SignData is purely cryptographic and does not depend on network state
	// (like a fetched nonce or specific account configurations), opening the account
	// or setting network/blockchain might not be strictly necessary for this specific operation.
	// The current implementation of SignData in circular_protocol_api/account.go
	// checks if account.Address is set, so it's good practice to at least Open() it.
	// However, for this specific example, we will assume SignData can be called on a new account object
	// if it primarily acts as a utility wrapper for cryptographic functions using the provided private key.
	// For robustness, one might Open the account first:
	// account := &circular_protocol_api.CEPAccount{}
	// account.Open("some-dummy-address-if-needed-by-SignData-internal-checks")
	account := &circular_protocol_api.CEPAccount{} // Create an account object.

	fmt.Printf("Original Data to Sign: '%s'\n", sampleData)

	// Sign the sample data using the account's SignData method and the private key.
	signature, err := account.SignData(sampleData, privateKeyHex)
	if err != nil {
		// If signing fails, log fatal error.
		log.Fatalf("Failed to sign data: %v", err)
	}
	fmt.Printf("Data Signature (Hex): %s\n", signature)

	// (Optional) Verify signature:
	// To verify the signature, you would typically need the corresponding public key.
	// The verification process involves:
	// 1. Parsing the public key.
	// 2. Parsing the signature.
	// 3. Hashing the original message (the same way it was hashed before signing).
	// 4. Calling a verification function (e.g., ecdsa.Verify or schnorr.Verify).
	// This is beyond the scope of this basic signing example but is a crucial part of a complete cryptographic cycle.
	fmt.Println("\nData signing example complete.")
}
