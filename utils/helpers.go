/*
 * Author: Danny De Novi (dannydenovi29@gmail.com)
 * Date: 12/10/2024
 * License: MIT
 */

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

/* SendRequest sends an HTTP POST request with JSON data
* @param data interface{} - data to send
* @param nagFunction string - function to call
* @param nagURL string - URL to call
* @return map[string]interface{} - response
* @return error - error
 */
// SendRequest performs an HTTP POST request and handles the JSON response
func SendRequest(data interface{}, nagFunction string, nagURL string) map[string]interface{} {
	url := nagURL + nagFunction

	// Convert the data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		response := map[string]interface{}{
			"Result":   "500",
			"Response": "Wrong JSON format",
		}

		return response
	}

	// Create the request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		response := map[string]interface{}{
			"Result":   "500",
			"Response": "Error during the creation of the request",
		}

		return response
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		response := map[string]interface{}{
			"Result":   "500",
			"Response": "Error during the sending of the request",
		}

		return response
	}
	defer resp.Body.Close()

	// Read the response and decode it
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		response := map[string]interface{}{
			"Result":   "500",
			"Response": "Error during the reading of the response",
		}

		return response
	}

	if resp.StatusCode != http.StatusOK {
		response := map[string]interface{}{
			"Result":   "500",
			"Response": "Error during the decoding of the response",
		}

		return response
	}

	var response map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		response := map[string]interface{}{
			"Result":   "500",
			"Response": "Error during the decoding of the response",
		}

		return response
	}

	return response
}

// PadNumber pads a number with leading zeros to number less than 10
func PadNumber(number int) string {
	if number < 10 {
		return fmt.Sprintf("0%d", number)
	}
	return fmt.Sprintf("%d", number)
}

// Generate formatted timestamp in the format YYYY-MM-DD-HH:MM:SS
func GetFormattedTimestamp() string {
	t := time.Now()
	return fmt.Sprintf("%d:%s:%s-%s:%s:%s", t.Year(), PadNumber(int(t.Month())), PadNumber(t.Day()), PadNumber(t.Hour()), PadNumber(t.Minute()), PadNumber(t.Second()))
}

// ECSignature defines the structure for DER encoded signature
type ECSignature struct {
	R, S *big.Int
}

/* // SignMessage firma un messaggio con una chiave privata in formato hex e restituisce la firma in formato hex.
* @param message string - message to sign
* @param privateKey string - private key in hex format
* @return string - signature in hex format
 */

func SignMessage(message string, privateKey string) map[string]interface{} {

	bytesPrivateKey, err := hex.DecodeString(privateKey)
	if err != nil {
		return map[string]interface{}{
			"Result":   "500",
			"Response": "Error during the decoding of the private key",
		}
	}

	privKey := secp256k1.PrivKeyFromBytes(bytesPrivateKey)

	messageHash := chainhash.HashB([]byte(message))
	r, s, err := ecdsa.Sign(rand.Reader, privKey.ToECDSA(), messageHash)
	if err != nil {
		return map[string]interface{}{
			"Result":   "500",
			"Response": "Error during the signing of the message",
		}
	}

	derSignature, err := asn1.Marshal(ECDSASignature{R: r, S: s})

	if err != nil {
		return map[string]interface{}{
			"Result":   "500",
			"Response": "Error during the encoding of the signature",
		}
	}

	stringDERSignature := hex.EncodeToString(derSignature)
	return map[string]interface{}{"Signature": stringDERSignature, "R": r, "S": s}
}

type ECDSASignature struct {
	R, S *big.Int
}

// StringToHex converts a string to a hexadecimal string
func StringToHex(str string) string {
	return hex.EncodeToString([]byte(str))
}

// HexToString coneverts a hexadecimal string to a string
func HexToString(hexStr string) (string, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// HexFix removes the prefix "0x"
func HexFix(word interface{}) string {
	switch v := word.(type) {
	case int:
		return fmt.Sprintf("%x", v)
	case string:
		if len(v) >= 2 && v[:2] == "0x" {
			return v[2:]
		}
		return v
	default:
		return ""
	}
}

// Sha256 calculates the SHA-256 hash of a string
func Sha256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func bytesToHex(b []byte) string {
	return hex.EncodeToString(b)
}

func VerifySignature(publicKey string, message string, signature string) bool {
	if len(publicKey) != 130 {
		fmt.Println("Invalid public key length")
		return false
	}

	// Remove the 0x prefix from the public key
	remove0x := HexFix(publicKey)

	// Hash the message to verify with SHA256
	msgHash := sha256.Sum256([]byte(message))

	// Decode the public key from hex
	publicKeyBytes, err := hex.DecodeString(remove0x)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// Parse the public key
	pubKey, err := secp256k1.ParsePubKey(publicKeyBytes)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// Convert btcec.PublicKey to ecdsa.PublicKey
	ecdsaPubKey := ecdsa.PublicKey{
		Curve: secp256k1.S256(),
		X:     pubKey.X(),
		Y:     pubKey.Y(),
	}

	// Decode the signature from hex
	signatureBytes, err := hex.DecodeString(signature)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// Parse the signature from DER format
	var ecsig ECDSASignature
	_, err = asn1.Unmarshal(signatureBytes, &ecsig)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// Verify the signature
	return ecdsa.Verify(&ecdsaPubKey, msgHash[:], ecsig.R, ecsig.S)
}

func GetPublicKey(privateKey string) string {
	privKeyBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		return ""
	}

	privKey := secp256k1.PrivKeyFromBytes(privKeyBytes)
	pubKey := privKey.PubKey()

	return hex.EncodeToString(pubKey.SerializeUncompressed())
}

func GetKeysFromString(seedPhrase string) (map[string]string, error) {
	// Generate private key using RFC 6979
	hashHex := Sha256(seedPhrase)
	hash, err := hex.DecodeString(hashHex)
	if err != nil {
		return nil, err
	}

	privateKey := secp256k1.PrivKeyFromBytes(hash)
	publicKey := privateKey.PubKey()

	// Encode private key in DER format
	privateKeyDER := privateKey.Serialize()

	// Encode public key in DER format
	publicKeyDER := publicKey.SerializeUncompressed()

	// Generate address from public key
	addressDER := Sha256(bytesToHex(publicKey.SerializeUncompressed()))

	// Create a map to hold the results
	result := map[string]string{
		"privateKey": bytesToHex(privateKeyDER),
		"publicKey":  bytesToHex(publicKeyDER),
		"address":    addressDER,
		"seedPhrase": seedPhrase,
	}

	return result, nil
}

// NAG FUNCTIONS

const TEST_CONTRACT = "Circular_TestContract_"
const CALL_CONTRACT = "Circular_CallContract_"
const CHECK_WALLET = "Circular_CheckWallet_"
const GET_WALLET = "Circular_GetWallet_"
const GET_LATEST_TRANSACTIONS = "Circular_GetLatestTransactions_"
const GET_WALLET_BALANCE = "Circular_GetWalletBalance_"
const REGISTER_WALLET = "Circular_RegisterWallet_"
const GET_DOMAIN = "Circular_GetDomain_"
const GET_ASSET_LIST = "Circular_GetAssetList_"
const GET_ASSET = "Circular_GetAsset_"
const GET_ASSET_SUPPLY = "Circular_GetAssetSupply_"
const GET_VOUCHER = "Circular_GetVoucher_"
const GET_BLOCK_RANGE = "Circular_GetBlockRange_"
const GET_BLOCK = "Circular_GetBlock_"
const GET_BLOCK_COUNT = "Circular_GetBlockCount_"
const GET_ANALYTICS = "Circular_GetAnalytics_"
const GET_BLOCKCHAINS = "Circular_GetBlockchains_"
const GET_PENDING_TRANSACTION = "Circular_GetPendingTransaction_"
const GET_TRANSACTION_BY_ID = "Circular_GetTransactionbyID_"
const GET_TRANSACTION_BY_NODE = "Circular_GetTransactionbyNode_"
const GET_TRANSACTIONS_BY_ADDRESS = "Circular_GetTransactionbyAddress_"
const GET_TRANSACTION_BY_DATE = "Circular_GetTransactionbyDate_"
const SEND_TRANSACTION = "Circular_AddTransaction_"
const GET_WALLET_NONCE = "Circular_GetWalletNonce_"
