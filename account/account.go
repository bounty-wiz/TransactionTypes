package account

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	projectMarker = "transaction-types"
	key1FilePath  = "account1.key"
	key2FilePath  = "account2.key"
)

// getMigrationsPath returns the absolute path to the migrations directory.
func getPath() string {
	// Get the path of the current file
	_, filename, _, _ := runtime.Caller(0)

	// Find index of your project root folder name
	idx := strings.Index(strings.ToLower(filename), strings.ToLower(projectMarker))
	if idx == -1 {
		panic("project root folder not found in path")
	}

	// Cut path up to the project root
	rootPath := filename[:idx+len(projectMarker)]

	// Append migrations relative to root
	return filepath.Join(rootPath, "account")
}

func GetAccount(accNum int) (*common.Address, *ecdsa.PrivateKey) {
	var priv *ecdsa.PrivateKey

	path := getPath()
	var keyFilePath string
	switch accNum {
	case 1:
		keyFilePath = filepath.Join(path, key1FilePath)
	case 2:
		keyFilePath = filepath.Join(path, key2FilePath)
	default:
		log.Fatal("Invalid account number. Use 1 or 2.")
	}

	if _, err := os.Stat(keyFilePath); os.IsNotExist(err) {
		fmt.Println("Key file not found. Generating new Ethereum key...")
		priv, err = crypto.GenerateKey()
		if err != nil {
			log.Fatal("Failed to generate key:", err)
		}

		privBytes := crypto.FromECDSA(priv)
		err = os.WriteFile(keyFilePath, []byte(hex.EncodeToString(privBytes)), 0600)
		if err != nil {
			log.Fatal("Failed to write key file:", err)
		}

		fmt.Println("New key saved to", keyFilePath)
	} else {
		// Load existing key
		fmt.Println("Loading existing key from", keyFilePath)
		keyHex, err := os.ReadFile(keyFilePath)
		if err != nil {
			log.Fatal("Failed to read key file:", err)
		}

		privBytes, err := hex.DecodeString(string(keyHex))
		if err != nil {
			log.Fatal("Invalid hex in key file:", err)
		}

		priv, err = crypto.ToECDSA(privBytes)
		if err != nil {
			log.Fatal("Invalid private key:", err)
		}
	}

	// Print address and keys
	address := crypto.PubkeyToAddress(priv.PublicKey)
	pubBytes := crypto.FromECDSAPub(&priv.PublicKey)

	fmt.Println("Address:    ", address.Hex())
	fmt.Println("Public Key: ", hex.EncodeToString(pubBytes))
	fmt.Println("Private Key:", hex.EncodeToString(crypto.FromECDSA(priv)))

	return &address, priv
}
