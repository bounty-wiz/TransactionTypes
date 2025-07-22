package main

import (
	"fmt"
	"log"
	"transactiontypes/account"

	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	_, acc2Priv := account.GetAccount(2)

	message := []byte("Login to app.xyz")
	prefixed := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)

	hash := crypto.Keccak256Hash([]byte(prefixed))

	// Sign the hash
	signature, err := crypto.Sign(hash.Bytes(), acc2Priv)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Message: %s\n", message)
	fmt.Printf("Prefixed Hash: 0x%x\n", hash.Bytes())
	fmt.Printf("Signature: 0x%x\n", signature)

	// Recover the public key
	pubKey, err := crypto.SigToPub(hash.Bytes(), signature)
	if err != nil {
		log.Fatal(err)
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	fmt.Printf("Recovered Address: %s\n", recoveredAddr.Hex())
}
