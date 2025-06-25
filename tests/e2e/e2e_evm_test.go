package e2e

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"

	"github.com/kiichain/kiichain/v2/tests/e2e/mock"
)

const (
	CounterBinary = "6080604052348015600e575f5ffd5b505f805560ce80601d5f395ff3fe6080604052348015600e575f5ffd5b50600436106030575f3560e01c80638ada066e146034578063d09de08a146048575b5f5ffd5b5f5460405190815260200160405180910390f35b604e6050565b005b60015f5f828254605f91906066565b9091555050565b8082018281125f831280158216821582161715609057634e487b7160e01b5f52601160045260245ffd5b50509291505056fea26469706673582212202eb4042585c41ce4809327e6d7a60c017098a1ca09ffc5893f6360047a5354b564736f6c634300081d0033"
)

// testEVMQueries Test EVM queries
func (s *IntegrationTestSuite) testEVMQueries(jsonRPC string) {
	s.Run("eth_blockNumber", func() {
		res, err := httpEVMPostJSON(jsonRPC, "eth_blockNumber", []interface{}{})
		s.Require().NoError(err)

		blockNumber, err := parseResultAsHex(res)
		s.Require().NoError(err)
		s.Require().True(strings.HasPrefix(blockNumber, "0x"))
	})

	s.Run("eth_chainId", func() {
		res, err := httpEVMPostJSON(jsonRPC, "eth_chainId", []interface{}{})
		s.Require().NoError(err)

		chainID, err := parseResultAsHex(res)
		s.Require().NoError(err)
		s.Require().Equal(chainID, "0x3f2")
	})

	s.Run("eth_getBalance on zero address", func() {
		res, err := httpEVMPostJSON(jsonRPC, "eth_getBalance", []interface{}{
			"0x0000000000000000000000000000000000000000", "latest",
		})
		s.Require().NoError(err)

		balance, err := parseResultAsHex(res)
		s.Require().NoError(err)
		s.Require().True(strings.HasPrefix(balance, "0x0"))
	})

	s.Run("web3_clientVersion", func() {
		res, err := httpEVMPostJSON(jsonRPC, "web3_clientVersion", []interface{}{})
		s.Require().NoError(err)

		_, ok := res["result"].(string)
		s.Require().True(ok)
	})
}

// testEVM Tests EVM send and contract usage
func (s *IntegrationTestSuite) testEVM(jsonRPC string) {
	var err error

	// Get a funded EVM account and check balance transactions
	evmAccount := s.chainA.evmAccount

	// Setup client
	client, err := ethclient.Dial(jsonRPC)
	s.Require().NoError(err)

	// Deploy contract
	s.Run("create and interact w/ contract", func() {
		// Prepare auth
		auth := setupDefaultAuth(client, evmAccount.key)

		// Deploy
		contractAddress, tx, counter, err := mock.DeployCounter(auth, client)
		s.Require().NoError(err)

		s.waitForTransaction(client, tx, evmAccount.address)
		s.T().Logf("ContractAddress : %s", contractAddress.String())

		// 6. Interact w/ contract and see changes
		tx, err = counter.Increment(auth)
		s.Require().NoError(err)
		s.waitForTransaction(client, tx, evmAccount.address)

		counterValue, err := counter.GetCounter(nil)
		s.Require().NoError(err)
		s.Require().Equal(big.NewInt(1), counterValue)
	})
}

// waitForTransaction waits until transaction is mined, requiring its success and checks reason in case of failure
func (s *IntegrationTestSuite) waitForTransaction(client *ethclient.Client, tx *geth.Transaction, sender common.Address) *geth.Receipt {
	// Wait and check tx
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	s.Require().NoError(err)

	if receipt.Status == geth.ReceiptStatusFailed {
		// Try to get the revert reason
		reason, err := getRevertReason(client, tx.Hash(), sender)
		if err != nil {
			s.T().Logf("Failed to get revert reason: %v", err)
		} else if reason != "" {
			s.T().Logf("Revert reason: %s", reason)
		}
	}
	s.Require().False(receipt.Status == geth.ReceiptStatusFailed)
	return receipt
}

// HexifyFuncAddress turns an ABI function address and turns it into a hex value
func HexifyFuncAddress(funcAddress []byte) string {
	return "0x" + hex.EncodeToString(funcAddress)
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
