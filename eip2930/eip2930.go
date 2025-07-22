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
	NodeRPCURL  = "https://polygon-amoy.drpc.org"
	AmoyChainID = 80002   // Polygon Amoy Testnet Chain ID
	ValueToSend = 0.01e18 // 0.01 ETH
)

func main() {
	acc2Addr, acc2Priv := account.GetAccount(2)
	to := lo.ToPtr(common.HexToAddress("0x0fd9e8d3af1aaee056eb9e802c3a762a667b1904"))

	ctx := context.Background()
	client, err := ethclient.Dial(NodeRPCURL)
	if err != nil {
		log.Fatal("Failed to connect to Ethereum node:", err)
	}

	// Nonce
	nonce, err := client.PendingNonceAt(ctx, lo.FromPtr(acc2Addr))
	if err != nil {
		log.Fatal("Failed to fetch nonce:", err)
	}

	// Gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal("Failed to fetch gas price:", err)
	}

	accessList := types.AccessList{
		{
			Address:     common.Address(common.HexToAddress("0x0fd9e8d3af1aaee056eb9e802c3a762a667b1904")),
			StorageKeys: []common.Hash{common.HexToHash("0xf9a42dc9f268c1720e130da18118febc65e0ca534e035a0e39d30cf8daea5f0a"), common.HexToHash("0x0c2d31ae2b93233fa550fc5df04cd7b0b742c0821a2494f91cd79ca74a9e2e48")},
		},
	}

	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:       *acc2Addr,
		To:         to,
		Data:       common.FromHex("0xa9059cbb0000000000000000000000008056361b1c1361436D61D187d761233b42d1c20e000000000000000000000000000000000000000000000000016345785D8A0000"),
		AccessList: accessList,
	})
	if err != nil {
		log.Fatal("Failed to fetch gas limit:", err)
	}

	chainID := big.NewInt(AmoyChainID) // Use your chain's ID (80002 = Polygon Mumbai Testnet)

	// Construct AccessListTx
	txData := types.AccessListTx{
		ChainID:    chainID,
		Nonce:      nonce,
		GasPrice:   gasPrice,
		Gas:        gasLimit,
		To:         to,
		Data:       common.FromHex("0xa9059cbb0000000000000000000000008056361b1c1361436D61D187d761233b42d1c20e000000000000000000000000000000000000000000000000016345785D8A0000"),
		AccessList: accessList,
	}
	tx := types.NewTx(&txData)

	// Sign it
	signedTx, err := types.SignTx(tx, types.NewEIP2930Signer(chainID), acc2Priv)
	if err != nil {
		log.Fatal("Failed to sign tx:", err)
	}

	// Broadcast
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		log.Fatal("Broadcast failed:", err)
	}

	fmt.Println("Access List Transaction Sent!")
	fmt.Println("Tx hash:", signedTx.Hash().Hex())

	// Optionally wait for mining
	time.Sleep(10 * time.Second)
	receipt, err := client.TransactionReceipt(ctx, signedTx.Hash())
	if err != nil {
		fmt.Println("Tx not mined yet.")
	} else {
		fmt.Println("Mined in block:", receipt.BlockNumber)
	}
}
