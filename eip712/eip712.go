package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"
	"transactiontypes/account"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

const (
	// Public RPC URL for Polygon Amoy Testnet
	NodeRPCURL  = "https://polygon-amoy.drpc.org"
	AmoyChainID = 80002 // Polygon Amoy Testnet Chain ID
)

func main() {
	// 1) Connect to Amoy
	client, err := ethclient.Dial("https://polygon-amoy.drpc.org")
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	// 2) Prepare the same EIP‑712 TypedData that the user signed
	acc2Addr, acc2Priv := account.GetAccount(2)

	verifierAddr := common.HexToAddress("0xf80bb731f8ba49624dce8edb1a8188782287ff1e")

	domain := apitypes.TypedDataDomain{
		Name:              "MyDApp",
		Version:           "1",
		ChainId:           math.NewHexOrDecimal256(AmoyChainID),
		VerifyingContract: verifierAddr.Hex(),
	}
	types := apitypes.Types{
		"EIP712Domain": {
			{Name: "name", Type: "string"},
			{Name: "version", Type: "string"},
			{Name: "chainId", Type: "uint256"},
			{Name: "verifyingContract", Type: "address"},
		},
		"Permit": {
			{Name: "owner", Type: "address"},
			{Name: "spender", Type: "address"},
			{Name: "value", Type: "uint256"},
			{Name: "nonce", Type: "uint256"},
			{Name: "deadline", Type: "uint256"},
		},
	}
	deadline := big.NewInt(time.Now().Add(time.Hour).Unix())
	nonce := big.NewInt(0) // ideally fetched from the token's nonces(owner)
	message := apitypes.TypedDataMessage{
		"owner":    acc2Addr.Hex(),
		"spender":  acc2Addr.Hex(), // for verify only, can be any address
		"value":    "1000000000000000000",
		"nonce":    nonce.String(),
		"deadline": deadline.String(),
	}
	typedData := apitypes.TypedData{
		Types:       types,
		PrimaryType: "Permit",
		Domain:      domain,
		Message:     message,
	}

	// 3) Sign or supply your existing (v,r,s)
	domainSep, _ := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	msgHash, _ := typedData.HashStruct("Permit", typedData.Message)
	digest := crypto.Keccak256(
		[]byte("\x19\x01"),
		domainSep,
		msgHash,
	)
	sig, _ := crypto.Sign(digest, acc2Priv)
	r := common.BytesToHash(sig[:32])
	s := common.BytesToHash(sig[32:64])
	v := uint8(sig[64]) + 27

	// 4) ABI‑encode verifyPermit(owner,spender,value,nonce,deadline,v,r,s)
	verifierABI := `[{"inputs":[{"internalType":"address","name":"owner","type":"address"},{"internalType":"address","name":"spender","type":"address"},{"internalType":"uint256","name":"value","type":"uint256"},{"internalType":"uint256","name":"nonce","type":"uint256"},{"internalType":"uint256","name":"deadline","type":"uint256"},{"internalType":"uint8","name":"v","type":"uint8"},{"internalType":"bytes32","name":"r","type":"bytes32"},{"internalType":"bytes32","name":"s","type":"bytes32"}],"name":"verifyPermit","outputs":[{"internalType":"bool","name":"","type":"bool"}],"stateMutability":"view","type":"function"}]`
	parsed, _ := abi.JSON(strings.NewReader(verifierABI))
	calldata, _ := parsed.Pack(
		"verifyPermit",
		acc2Addr,
		acc2Addr, // spender (must match what was signed)
		big.NewInt(1e18),
		nonce,
		deadline,
		v, r, s,
	)

	// 5) Do an eth_call
	msg := ethereum.CallMsg{
		To:   &verifierAddr,
		Data: calldata,
	}
	res, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		log.Fatal(err)
	}

	// 6) Decode the bool result
	out, err := parsed.Unpack("verifyPermit", res)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Signature valid?", out[0].(bool))

	if !out[0].(bool) {
		log.Fatal("Signature verification failed")
	}

	fmt.Printf("Signature verified successfully for owner: %s\n", acc2Addr.Hex())
}
