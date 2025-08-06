package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/kiichain/kiichain/v3/tests/e2e/precompiles"
)

const (
	WasmdPrecompileAddress = "0x0000000000000000000000000000000000001001"
)

func (s *IntegrationTestSuite) testWasmdPrecompile() {
	// Get the first validator docker container
	valIdx := 0
	valDockerAsset := s.chainA.validators[valIdx]
	fromAddres, err := valDockerAsset.keyInfo.GetAddress()
	s.Require().NoError(err)
	from := fromAddres.String()
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))

	// Store the contract code in the docker image
	workingDirectory, err := os.Getwd()
	s.Require().NoError(err)

	srcPath := filepath.Join(workingDirectory, "../../precompiles/wasmd/testdata/counter.wasm")
	dstPath := filepath.Join(valDockerAsset.configDir(), "config", "counter.wasm")

	// Copy the file
	_, err = copyFile(srcPath, dstPath)
	s.Require().NoError(err)
	storeWasmPath := configFile("counter.wasm")

	// 1. Store the contract using the CLI
	s.WasmdStoreCLI(s.chainA, valIdx, from, storeWasmPath, standardFees.String(), false)

	// Get EVM acc
	evmAccount := s.chainA.evmAccount
	cosmosAddress, err := PubKeyBytesToCosmosAddress(evmAccount.address.Bytes())
	s.Require().NoError(err)

	// Setup evm client
	jsonRPC := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("8545/tcp"))
	client, err := ethclient.Dial(jsonRPC)
	s.Require().NoError(err)

	// Bind abigen precompile contract to address
	wasmdPrecompile, err := precompiles.NewWasmdPrecompile(common.HexToAddress(WasmdPrecompileAddress), client)
	s.Require().NoError(err)

	// 2. Instantiate the contract via precompile
	s.wasmIntantiateViaPrecompile(client, wasmdPrecompile, evmAccount, 1, `"zero"`, "counter")

	// Query contract address
	contractAddress, err := s.queryWasmContractAddressAPI(chainEndpoint, cosmosAddress, 0)
	s.Require().NoError(err)
	s.Require().NotEmpty(contractAddress)

	// Query the contract state the value should be 0
	state := s.wasmSmartQueryViaPrecompile(wasmdPrecompile, contractAddress, `"value"`)
	s.Require().Equal(string(state), `{"value":0}`)

	// 3. Execute the contract via precompile
	s.wasmExecuteViaPrecompile(client, wasmdPrecompile, evmAccount, contractAddress, `{"set": 34}`)

	// Query the contract state again the value should be 34
	state = s.wasmSmartQueryViaPrecompile(wasmdPrecompile, contractAddress, `"value"`)
	s.Require().Equal(string(state), `{"value":34}`)
}

// wasmIntantiateViaPrecompile intantiates a contract via precompile
func (s *IntegrationTestSuite) wasmIntantiateViaPrecompile(
	client *ethclient.Client,
	wasmPrecompile *precompiles.WasmdPrecompile,
	senderEvmAccount EVMAccount,
	codeID uint64,
	msg, label string,
) {
	// Instantiate contract
	s.Run("intantiating wasm contract via precompile", func() {
		// Call transfer
		tx, err := wasmPrecompile.Instantiate(
			setupDefaultAuth(client, senderEvmAccount.key),
			senderEvmAccount.address,
			codeID,               // code ID
			label,                // label
			[]byte(msg),          // msg
			[]precompiles.Coin{}, // coins
		)
		s.Require().NoError(err)

		// Wait and check tx
		s.waitForTransaction(client, tx, senderEvmAccount.address)
	})
}

// wasmExecuteViaPrecompile executes a contract function via precompile
func (s *IntegrationTestSuite) wasmExecuteViaPrecompile(
	client *ethclient.Client,
	wasmPrecompile *precompiles.WasmdPrecompile,
	senderEvmAccount EVMAccount,
	contractAddress, msg string,
) {
	// Deploy contract
	s.Run("send to IBC precompile transfer", func() {
		// Call transfer
		tx, err := wasmPrecompile.Execute(
			setupDefaultAuth(client, senderEvmAccount.key),
			contractAddress,
			[]byte(msg),          // msg
			[]precompiles.Coin{}, // coins
		)
		s.Require().NoError(err)

		// Wait and check tx
		s.waitForTransaction(client, tx, senderEvmAccount.address)
	})
}

// wasmSmartQueryViaPrecompile runs a smart wasm query via precompile
func (s *IntegrationTestSuite) wasmSmartQueryViaPrecompile(
	wasmPrecompile *precompiles.WasmdPrecompile,
	contractAddress, msg string,
) []byte {
	// Setup call options
	callOpts := &bind.CallOpts{
		Pending: false,
		Context: context.Background(),
	}

	// Call transfer
	resp, err := wasmPrecompile.QuerySmart(
		callOpts,
		contractAddress,
		[]byte(msg),
	)
	s.Require().NoError(err)
	return resp
}
