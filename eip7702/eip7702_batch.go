package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"
	"transactiontypes/account"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
)

const (
	// Public RPC URL for Polygon Amoy Testnet
	NodeRPCURL  = "https://polygon-amoy.drpc.org"
	AmoyChainID = 80002 // Polygon Amoy Testnet Chain ID
)

func main() {
	ctx := context.Background()
	client, err := ethclient.Dial(NodeRPCURL)
	if err != nil {
		log.Fatal("RPC connection failed:", err)
	}

	// Load account
	acc2Addr, acc2Priv := account.GetAccount(2)
	acc1Addr, acc1Priv := account.GetAccount(1)

	to := common.HexToAddress("0x87581c71b3693062f4d3e34617c3919ec1abf39b")

	// Define contract and parameters
	moduleAddr := common.HexToAddress("0x4f9c96915a9ce8cd5eb11a2c35ab587fc97d5126")

	froms := []common.Address{
		*acc1Addr,
		*acc2Addr,
	}

	// Build calldata
	contractAbiJson := `[
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"internalType": "address",
				"name": "from",
				"type": "address"
			}
		],
		"name": "PingStart",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"internalType": "address",
				"name": "from",
				"type": "address"
			}
		],
		"name": "PingSuccess",
		"type": "event"
	},
	{
		"inputs": [
			{
				"internalType": "address[]",
				"name": "froms",
				"type": "address[]"
			}
		],
		"name": "triggerPings",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`
	parsedAbi, _ := abi.JSON(strings.NewReader(contractAbiJson))
	data, err := parsedAbi.Pack("triggerPings", froms)
	if err != nil {
		log.Fatal("ABI pack error:", err)
	}

	// Nonce and gas
	baseNonce2, err := client.PendingNonceAt(ctx, *acc2Addr)
	if err != nil {
		log.Fatal("Nonce fetch failed:", err)
	}
	nonce2 := baseNonce2 + 1

	nonce1, err := client.PendingNonceAt(ctx, *acc1Addr)
	if err != nil {
		log.Fatal("Nonce fetch failed:", err)
	}

	gasTipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		log.Fatal("Failed to fetch gas tip cap:", err)
	}
	baseFee, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal("Failed to fetch base fee:", err)
	}
	gasFeeCap := new(big.Int).Add(baseFee, gasTipCap)

	// Create EIP-712-style signature for delegation
	sig2, err := signEIP7702Delegation(acc2Priv, AmoyChainID, moduleAddr, nonce2)
	if err != nil {
		log.Fatal("Signature failed:", err)
	}

	sig1, err := signEIP7702Delegation(acc1Priv, AmoyChainID, moduleAddr, nonce1)
	if err != nil {
		log.Fatal("Signature failed:", err)
	}

	r2 := new(big.Int).SetBytes(sig2[:32])
	s2 := new(big.Int).SetBytes(sig2[32:64])
	v2 := uint8(sig2[64])

	r1 := new(big.Int).SetBytes(sig1[:32])
	s1 := new(big.Int).SetBytes(sig1[32:64])
	v1 := uint8(sig1[64])

	// Build EIP-7702 TxWithDelegation
	delegation := types.SetCodeTx{
		ChainID:   uint256.NewInt(AmoyChainID),
		Nonce:     baseNonce2,
		GasTipCap: uint256.MustFromBig(gasTipCap),
		GasFeeCap: uint256.MustFromBig(gasFeeCap),
		Gas:       120000,
		To:        to,
		Data:      data,
		AuthList: []types.SetCodeAuthorization{
			{
				ChainID: *uint256.NewInt(AmoyChainID),
				Address: moduleAddr,
				Nonce:   nonce1,
				R:       *uint256.MustFromBig(r1),
				S:       *uint256.MustFromBig(s1),
				V:       v1,
			},
			{
				ChainID: *uint256.NewInt(AmoyChainID),
				Address: moduleAddr,
				Nonce:   nonce2,
				R:       *uint256.MustFromBig(r2),
				S:       *uint256.MustFromBig(s2),
				V:       v2,
			},
		},
	}

	fullTx := types.NewTx(&delegation)
	signedTx, err := types.SignTx(fullTx, types.LatestSignerForChainID(big.NewInt(AmoyChainID)), acc2Priv)
	if err != nil {
		log.Fatal("Signing failed:", err)
	}

	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		log.Fatal("Tx failed:", err)
	}

	fmt.Println("EIP-7702 Tx sent:", signedTx.Hash().Hex())

	time.Sleep(10 * time.Second)
	receipt, err := client.TransactionReceipt(ctx, signedTx.Hash())
	if err != nil {
		fmt.Println("Waiting...")
	} else {
		fmt.Println("Tx mined in block", receipt.BlockNumber)
	}
}

// signEIP7702Delegation creates a hash of (chainID, from, nonce) and signs it
func signEIP7702Delegation(priv *ecdsa.PrivateKey, chainID int64, from common.Address, nonce uint64) ([]byte, error) {
	// Encode [chain_id, address, nonce] in RLP
	msgPayload, err := rlp.EncodeToBytes([]interface{}{
		big.NewInt(chainID),
		from,
		big.NewInt(int64(nonce)),
	})
	if err != nil {
		return nil, err
	}

	// Prepend MAGIC 0x05
	prefixed := append([]byte{0x05}, msgPayload...)

	// Hash
	msgHash := crypto.Keccak256Hash(prefixed)
	return crypto.Sign(msgHash.Bytes(), priv)
}
