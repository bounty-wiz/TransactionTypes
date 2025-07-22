package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"
	"transactiontypes/account"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/samber/lo"
)

const (
	// Public RPC URL for Polygon Amoy Testnet
	NodeRPCURL  = "https://polygon-amoy.drpc.org"
	GasLimit    = 21000   // Standard gas limit for a simple ETH transfer
	AmoyChainID = 80002   // Polygon Amoy Testnet Chain ID
	ValueToSend = 0.01e18 // 0.01 ETH
)

func main() {
	acc1Addr, acc1Priv := account.GetAccount(1)
	acc2Addr, _ := account.GetAccount(2)

	ctx := context.Background()
	client, err := ethclient.Dial(NodeRPCURL)
	if err != nil {
		log.Fatal("Failed to connect to Ethereum node:", err)
	}

	nonce, err := client.PendingNonceAt(ctx, lo.FromPtr(acc1Addr))
	if err != nil {
		log.Fatal("Failed to fetch nonce:", err)
	}

	// Get suggested gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal("Failed to fetch gas price:", err)
	}

	// Create a types.LegacyTx
	tx := types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      GasLimit,
		To:       acc2Addr, // Send to account 2
		Value:    big.NewInt(ValueToSend),
		Data:     []byte{},
	}

	// Convert it into a full types.Transaction object
	legacyTx := types.NewTx(&tx)

	chainID := big.NewInt(AmoyChainID) // Use your chain's ID (80002 = Polygon Mumbai Testnet)
	// Note: Although this is a legacy transaction, we still need to sign it with the chain ID for EIP-155 compatibility.
	// Because most of the nodes protect against replay attacks by requiring the chain ID in the signature.
	// Nodes that not protect against this will be able to get signed transaction with:
	// types.SignTx(tx, types.HomesteadSigner{}, priv)
	signedTx, err := types.SignTx(legacyTx, types.NewEIP155Signer(chainID), acc1Priv)
	if err != nil {
		log.Fatal("Failed to sign transaction:", err)
	}

	// Broadcast the transaction
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		log.Fatal("Broadcast failed:", err)
	}

	fmt.Println("Transaction sent!")
	fmt.Println("Tx hash:", signedTx.Hash().Hex())

	// Optionally wait for inclusion
	time.Sleep(10 * time.Second)
	receipt, err := client.TransactionReceipt(ctx, signedTx.Hash())
	if err != nil {
		fmt.Println("Tx not mined yet.")
	} else {
		fmt.Println("Tx mined in block:", receipt.BlockNumber)
	}
}
