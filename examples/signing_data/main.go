package main

import (
	"fmt"
	"log"

	"github.com/circular-protocol/circular-go/circular_protocol_api"
	"github.com/circular-protocol/circular-go/utils"
)

func main() {
	fmt.Println("Circular Go Enterprise API - Data Signing and Certificate Example")
	fmt.Printf("Current UTC Timestamp: %s\n", utils.GetFormattedUTCTimestamp())

	// Example private key (NEVER use real private keys directly in code for production)
	const privateKeyHex = "11842e4034999297038f59f054d1794389758469070e15999837078cec243f55"
	const sampleData = "This is some sample data to be signed."

	// 1. Certificate Object Usage
	cert := &circular_protocol_api.Certificate{}
	cert.SetData(sampleData)
	fmt.Printf("Certificate Data: '%s'\n", cert.GetData())
	fmt.Printf("Certificate Size: %d bytes\n", cert.GetCertificateSize())
	jsonCert, err := cert.GetJSONCertificate()
	if err != nil {
		log.Fatalf("Failed to get JSON certificate: %v", err)
	}
	fmt.Printf("JSON Certificate: %s\n", jsonCert)

	// 2. Data Signing
	// We need an Account object to use its SignData method, even if we're not doing network ops.
	// Alternatively, SignData could be a static utility if it doesn't depend on account state.
	// For now, it's part of Account.
	account := &circular_protocol_api.Account{}
	// No need to Open or SetNetwork if only using SignData and it has no dependencies on those states.
	// However, our current SignData doesn't use Account state, so this is fine.

	fmt.Printf("Original Data: '%s'\n", sampleData)
	signature, err := account.SignData(sampleData, privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to sign data: %v", err)
	}
	fmt.Printf("Data Signature (Hex): %s\n", signature)

	// (Optional) Verify signature - this would require having the public key
	// and using a verification function (e.g., schnorr.Verify with btcec.ParsePubKey)
	// This is out of scope for this example but a good next step for a real crypto util.
	fmt.Println("Data signing example complete.")
}
