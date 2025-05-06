package e2e

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// sendEVM does a send native coin via EVM
func sendEVM(
	client *ethclient.Client,
	privateKey *ecdsa.PrivateKey,
	fromAddress common.Address,
	toAddress common.Address,
	amount *big.Int,
) (geth.Receipt, error) {
	// Get the nonce (transaction count)
	return EVMTransaction(client, privateKey, fromAddress, toAddress, amount, nil)
}

// EVMTransaction builds up an evm_sendTransaction and returns its receipt
func EVMTransaction(
	client *ethclient.Client,
	privateKey *ecdsa.PrivateKey,
	fromAddress common.Address,
	toAddress common.Address,
	amount *big.Int,
	contractBinary []byte,
) (geth.Receipt, error) {
	// Get the nonce (transaction count)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return geth.Receipt{}, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get suggested gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return geth.Receipt{}, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Get chain ID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return geth.Receipt{}, fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Create the transaction
	tx := geth.NewTransaction(
		nonce,
		toAddress,
		amount,
		1500000, // gas limit
		gasPrice,
		contractBinary, // contract bytes
	)

	// Sign the transaction
	signedTx, err := geth.SignTx(tx, geth.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return geth.Receipt{}, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send the transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return geth.Receipt{}, fmt.Errorf("failed to send transaction: %w", err)
	}

	time.Sleep(time.Millisecond * 5000)

	_, receipt, err := checkTransactionByHash(client, signedTx.Hash())
	if err != nil {
		return geth.Receipt{}, fmt.Errorf("failed to get receipt: %w", err)
	}
	if receipt.Status == 0 {
		return geth.Receipt{}, fmt.Errorf("receipt status is 0")
	}
	return *receipt, nil
}

// checkTransactionByHash returns a receipt from a given tx hash
func checkTransactionByHash(client *ethclient.Client, txHash common.Hash) (*geth.Transaction, *geth.Receipt, error) {
	// Get the transaction details
	tx, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// If transaction is still pending
	if isPending {
		return tx, nil, fmt.Errorf("transaction is still pending")
	}

	// Get the transaction receipt
	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		return tx, nil, fmt.Errorf("failed to get receipt: %w", err)
	}

	return tx, receipt, nil
}
