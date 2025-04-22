package e2e

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	goeth "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func (s *IntegrationTestSuite) testEVMQueries(jsonRCP string) {
	s.Run("eth_blockNumber", func() {
		res, err := httpEVMPostJSON(jsonRCP, "eth_blockNumber", []interface{}{})
		s.Require().NoError(err)

		blockNumber, err := parseResultAsHex(res)
		s.Require().NoError(err)
		s.Require().True(strings.HasPrefix(blockNumber, "0x"))
	})

	s.Run("eth_chainId", func() {
		res, err := httpEVMPostJSON(jsonRCP, "eth_chainId", []interface{}{})
		s.Require().NoError(err)

		chainID, err := parseResultAsHex(res)
		s.Require().NoError(err)
		s.Require().Equal(chainID, "0x3f2")
	})

	s.Run("eth_getBalance on zero address", func() {
		res, err := httpEVMPostJSON(jsonRCP, "eth_getBalance", []interface{}{
			"0x0000000000000000000000000000000000000000", "latest",
		})
		s.Require().NoError(err)

		balance, err := parseResultAsHex(res)
		s.Require().NoError(err)
		s.Require().True(strings.HasPrefix(balance, "0x0"))
	})

	s.Run("web3_clientVersion", func() {
		res, err := httpEVMPostJSON(jsonRCP, "web3_clientVersion", []interface{}{})
		s.Require().NoError(err)

		_, ok := res["result"].(string)
		s.Require().True(ok)
	})
}

func (s *IntegrationTestSuite) testEVMSend(jsonRCP string) {
	var (
		err           error
		valIdx        = 0
		c             = s.chainA
		chainEndpoint = fmt.Sprintf("http://%s", s.valResources[c.id][valIdx].GetHostPort("1317/tcp"))
	)

	// 1. Create new account
	// Make a key
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	s.Require().NoError(err)

	// Make a message to extract key, making sure we are using correct way
	msg := crypto.Keccak256([]byte("foo"))
	ethSig, _ := crypto.Sign(msg, key)
	recoveredPub, _ := crypto.Ecrecover(msg, ethSig)

	// Get pubkey, evm and cosmos address
	pubKey, _ := crypto.UnmarshalPubkey(recoveredPub)
	evmAddress := crypto.PubkeyToAddress(*pubKey)
	s.T().Logf("Newly created evm address: %s", evmAddress)
	cosmosAddress, err := PubKeyBytesToCosmosAddress(evmAddress.Bytes())
	s.Require().NoError(err)
	s.T().Logf("Newly created cosmos address: %s", cosmosAddress)

	// Get alice's cosmos and evm address
	alice, err := s.chainA.genesisAccounts[1].keyInfo.GetAddress()
	s.Require().NoError(err)
	s.T().Logf("Alice address: %s", alice)

	publicKey, err := s.chainA.genesisAccounts[1].keyInfo.GetPubKey()
	s.Require().NoError(err)
	// Make sure we are using correct generation
	aliceCosmosAddress, err := PubKeyToCosmosAddress(publicKey)
	s.Require().NoError(err)
	s.Require().Equal(alice.String(), aliceCosmosAddress)

	// Get her EVM address
	aliceEvmAddress, err := CosmosPubKeyToEVMAddress(publicKey)
	s.Require().NoError(err)
	s.T().Logf("Alice evm address : %s", aliceEvmAddress)

	// 2. Send funds for new account so it can do operations
	s.execBankSend(s.chainA, valIdx, alice.String(), cosmosAddress, tokenAmount.String(), standardFees.String(), false)

	var newBalance sdk.Coin

	// get balances of sender and recipient accounts
	s.Require().Eventually(
		func() bool {
			newBalance, err = getSpecificBalance(chainEndpoint, cosmosAddress, akiiDenom)
			s.Require().NoError(err)

			return newBalance.IsValid() && newBalance.Amount.GT(math.ZeroInt())
		},
		10*time.Second,
		5*time.Second,
	)

	s.Run("eth_getBalance on new address", func() {
		res, err := httpEVMPostJSON(jsonRCP, "eth_getBalance", []interface{}{
			evmAddress.String(), "latest",
		})
		s.Require().NoError(err)

		balance, err := parseResultAsHex(res)
		s.Require().NoError(err)
		s.T().Logf("Balance : %s", balance)
		// Balance should have something now
		s.Require().False(strings.HasPrefix(balance, "0x0"))
	})

	// Send amount
	client, err := ethclient.Dial(jsonRCP)
	amount := big.NewInt(1000000000000000000)
	tx, err := sendEVMTransaction(client, key, evmAddress, aliceEvmAddress, amount)
	s.Require().NoError(err)

	time.Sleep(time.Millisecond * 5000)

	_, receipt, err := checkTransactionByHash(client, tx.Hash())
	s.Require().NoError(err)
	s.T().Logf("Transaction status: %d\n", receipt.Status)

	s.Run("eth_getBalance on address after send", func() {
		res, err := httpEVMPostJSON(jsonRCP, "eth_getBalance", []interface{}{
			evmAddress.String(), "latest",
		})
		s.Require().NoError(err)

		balance, err := parseResultAsHex(res)
		s.Require().NoError(err)
		s.T().Logf("Balance : %s", balance)
		// Balance should have something now
		s.Require().False(strings.HasPrefix(balance, "0x0"))
	})
}

func sendEVMTransaction(
	client *ethclient.Client,
	privateKey *ecdsa.PrivateKey,
	fromAddress common.Address,
	toAddress common.Address,
	amount *big.Int,
) (*goeth.Transaction, error) {
	// Get the nonce (transaction count)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get suggested gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	// Estimate gas limit
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:  fromAddress,
		To:    &toAddress,
		Value: amount,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Create the transaction
	tx := goeth.NewTransaction(
		nonce,
		toAddress,
		amount,
		gasLimit,
		gasPrice,
		nil, // data payload (empty for simple transfers)
	)

	// Get chain ID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Sign the transaction
	signedTx, err := goeth.SignTx(tx, goeth.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send the transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx, nil
}

func checkTransactionByHash(client *ethclient.Client, txHash common.Hash) (*goeth.Transaction, *goeth.Receipt, error) {
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

// CosmosPubKeyToEVMAddress takes a cosmos key and returns a respective evm address
func CosmosPubKeyToEVMAddress(pubKey types.PubKey) (common.Address, error) {
	// Get the compressed public key bytes (33 bytes)
	pubKeyBytes := pubKey.Bytes()

	// Convert compressed public key to ECDSA format
	ethPubKey, err := crypto.DecompressPubkey(pubKeyBytes)
	if err != nil {
		return common.Address{}, fmt.Errorf("failed to decompress pubkey: %w", err)
	}

	// Generate Ethereum address from the decompressed public key
	return crypto.PubkeyToAddress(*ethPubKey), nil
}

// PubKeyToCosmosAddress converts Cosmos SDK PubKey to Bech32 address
func PubKeyToCosmosAddress(pubKey types.PubKey) (string, error) {
	// Get the address bytes from the public key
	addressBytes := pubKey.Address().Bytes()

	// Convert to Bech32 format with the given prefix
	return PubKeyBytesToCosmosAddress(addressBytes)
}

// PubKeyBytesToCosmosAddress turns given bytes into kii address
func PubKeyBytesToCosmosAddress(addressBytes []byte) (string, error) {
	// Convert to Bech32 format with the given prefix
	return bech32.ConvertAndEncode("kii", addressBytes)
}

// httpEVMPostJSON creates a post with the EVM format
func httpEVMPostJSON(url, method string, params []interface{}) (map[string]interface{}, error) {
	// Create the payload with the json format
	payload := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  params,
	}
	data, _ := json.Marshal(payload)

	// Get the response
	// Since this is a test, we can silence linting for the http call
	//nolint:gosec
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode the result
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

// parseResultAsHex parse the result as json
func parseResultAsHex(resp map[string]interface{}) (string, error) {
	if result, ok := resp["result"].(string); ok {
		return result, nil
	}
	return "", fmt.Errorf("result not found or not a string")
}
