package e2e

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"

	"github.com/kiichain/kiichain/v2/tests/e2e/mock"
)

const (
	CounterBinary = "6080604052348015600e575f5ffd5b505f805560ce80601d5f395ff3fe6080604052348015600e575f5ffd5b50600436106030575f3560e01c80638ada066e146034578063d09de08a146048575b5f5ffd5b5f5460405190815260200160405180910390f35b604e6050565b005b60015f5f828254605f91906066565b9091555050565b8082018281125f831280158216821582161715609057634e487b7160e01b5f52601160045260245ffd5b50509291505056fea26469706673582212202eb4042585c41ce4809327e6d7a60c017098a1ca09ffc5893f6360047a5354b564736f6c634300081d0033"
)

// testEVMQueries Test EVM queries
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

// testEVM Tests EVM send and contract usage
func (s *IntegrationTestSuite) testEVM(jsonRCP string) {
	var (
		err           error
		valIdx        = 0
		c             = s.chainA
		chainEndpoint = fmt.Sprintf("http://%s", s.valResources[c.id][valIdx].GetHostPort("1317/tcp"))
	)

	// Get a funded EVM account and check balance transactions
	key, _ := s.setupEVMwithFunds(jsonRCP, chainEndpoint, valIdx)

	// Setup client
	client, err := ethclient.Dial(jsonRCP)
	s.Require().NoError(err)

	// Deploy contract
	s.Run("create and interact w/ contract", func() {
		// Prepare auth
		auth, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(1010))
		if err != nil {
			log.Fatal(err)
		}

		// Set optional params
		auth.Value = big.NewInt(0)
		auth.GasLimit = uint64(3000000) // gas limit
		auth.GasPrice, _ = client.SuggestGasPrice(context.Background())

		// Deploy
		contractAddress, tx, counter, err := mock.DeployCounter(auth, client)
		s.Require().NoError(err)

		s.waitForTransaction(client, tx)
		s.T().Logf("ContractAddress : %s", contractAddress.String())

		// 6. Interact w/ contract and see changes
		tx, err = counter.Increment(auth)
		s.Require().NoError(err)
		s.waitForTransaction(client, tx)

		counterValue, err := counter.GetCounter(nil)
		s.Require().NoError(err)
		s.Require().Equal(big.NewInt(1), counterValue)
	})
}

// setupEVMwithFunds sets up a new EVM account and sends funds from Alice to it, checking balance changes
func (s *IntegrationTestSuite) setupEVMwithFunds(jsonRCP, chainEndpoint string, valIdx int) (*ecdsa.PrivateKey, common.Address) {
	// 1. Create new account
	// Make a key
	key, err := crypto.HexToECDSA("88cbead91aee890d27bf06e003ade3d4e952427e88f88d31d61d3ef5e5d54305")
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

	// 2. Send funds via cosmos for new account so it can do operations
	s.execBankSend(s.chainA, valIdx, alice.String(), cosmosAddress, tokenAmount.String(), standardFees.String(), false)

	var newBalance sdk.Coin
	// Get balances of sender and recipient accounts
	s.Require().Eventually(
		func() bool {
			// Get balance via cosmos
			newBalance, err = getSpecificBalance(chainEndpoint, cosmosAddress, akiiDenom)
			s.Require().NoError(err)

			// Balance should already have some coin
			return newBalance.IsValid() && newBalance.Amount.GT(math.ZeroInt())
		},
		10*time.Second,
		5*time.Second,
	)

	s.Run("eth_getBalance on new address", func() {
		// Get balance via evm
		res, err := httpEVMPostJSON(jsonRCP, "eth_getBalance", []interface{}{
			evmAddress.String(), "latest",
		})
		s.Require().NoError(err)

		balance, err := parseResultAsHex(res)
		s.Require().NoError(err)
		s.T().Logf("Balance : %s", balance)

		// Balance should have something
		s.Require().False(strings.HasPrefix(balance, "0x0"))
	})

	// 3. Send via evm
	client, err := ethclient.Dial(jsonRCP)
	amount := big.NewInt(1000000000000000000)
	receipt, err := sendEVM(client, key, evmAddress, aliceEvmAddress, amount)
	s.Require().NoError(err)
	s.T().Logf("Transaction status: %d\n", receipt.Status)

	// 4. check changes
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
	return key, evmAddress
}

// waitForTransaction waits until transaction is mined, requiring its success
func (s *IntegrationTestSuite) waitForTransaction(client *ethclient.Client, tx *geth.Transaction) *geth.Receipt {
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	s.Require().NoError(err)
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
