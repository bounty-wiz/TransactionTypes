package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"
	"transactiontypes/account"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/samber/lo"
)

const (
	// Public RPC URL for Polygon Amoy Testnet
	NodeRPCURL  = "https://polygon-amoy.drpc.org"
	AmoyChainID = 80002 // Polygon Amoy Testnet Chain ID
)

func main() {
	acc2Addr, acc2Priv := account.GetAccount(2)
	to := lo.ToPtr(common.HexToAddress("0x0fd9e8d3af1aaee056eb9e802c3a762a667b1904"))

	ctx := context.Background()
	client, err := ethclient.Dial(NodeRPCURL)
	if err != nil {
		log.Fatal("Failed to connect to Ethereum node:", err)
	}

	nonce, err := client.PendingNonceAt(ctx, lo.FromPtr(acc2Addr))
	if err != nil {
		log.Fatal("Failed to fetch nonce:", err)
	}

	feeHistory, err := client.FeeHistory(ctx, 5, nil, nil)
	if err != nil {
		log.Fatal("Failed to fetch gas price:", err)
	}

	GasTipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		log.Fatal("Failed to fetch gas price:", err)
	}

	// base fee from most recent block
	latestBaseFee := feeHistory.BaseFee[len(feeHistory.BaseFee)-1]

	// GasFeeCap = baseFee + tip
	// Note: Here we use only the latest base fee + 12% of the previous base fee,
	// but you can also use the average or median from the fee history.
	bufferedBaseFee := new(big.Int).Mul(latestBaseFee, big.NewInt(112))
	bufferedBaseFee.Div(bufferedBaseFee, big.NewInt(100))

	// Final GasFeeCap = bufferedBaseFee + GasTipCap
	GasFeeCap := new(big.Int).Add(bufferedBaseFee, GasTipCap)

	chainID := big.NewInt(AmoyChainID) // Use your chain's ID (80002 = Polygon Mumbai Testnet)

	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From: *acc2Addr,
		To:   to,
		Data: common.FromHex("0xa9059cbb0000000000000000000000008056361b1c1361436D61D187d761233b42d1c20e000000000000000000000000000000000000000000000000016345785D8A0000"),
	})
	if err != nil {
		log.Fatal("Failed to fetch gas limit:", err)
	}

	tx := types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: GasTipCap,
		GasFeeCap: GasFeeCap,
		Gas:       gasLimit,
		To:        to,
		Data:      common.FromHex("0xa9059cbb0000000000000000000000008056361b1c1361436D61D187d761233b42d1c20e000000000000000000000000000000000000000000000000016345785D8A0000"),
	}

	// Convert it into a full types.Transaction object
	eip1559Tx := types.NewTx(&tx)

	signedTx, err := types.SignTx(eip1559Tx, types.LatestSignerForChainID(chainID), acc2Priv)
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
