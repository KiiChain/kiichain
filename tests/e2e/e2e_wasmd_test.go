package e2e

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// testWasmdCounter runs the wasm tests
// it tests the following:
// 1. instantiate a contract
// 2. execute a contract
// 3. query a contract
func (s *IntegrationTestSuite) testWasmdCounter() {
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

	// Store the contract using the CLI
	s.WasmdStoreCLI(s.chainA, valIdx, from, storeWasmPath, standardFees.String(), false)

	// Instantiate the contract
	s.WasmdInstantiateCLI(s.chainA, valIdx, from, 1, `"zero"`, "counter", from, standardFees.String(), false)

	// Query contract address
	contractAddress, err := s.queryWasmContractAddressAPI(chainEndpoint, from, 0)
	s.Require().NoError(err)
	s.Require().NotEmpty(contractAddress)

	// Query the contract state the value should be 0
	state, err := s.queryWasmSmartQueryAPI(chainEndpoint, contractAddress, `"value"`)
	s.Require().NoError(err)
	s.Require().Equal(string(state), `{"value":0}`)

	// Execute the contract
	s.WasmdExecuteCLI(s.chainA, valIdx, from, contractAddress, `{"set": 34}`, standardFees.String(), false)

	// Query the contract state again the value should be 34
	state, err = s.queryWasmSmartQueryAPI(chainEndpoint, contractAddress, `"value"`)
	s.Require().NoError(err)
	s.Require().Equal(string(state), `{"value":34}`)
}

// WasmdExecuteCLI executes a contract using the CLI
func (s *IntegrationTestSuite) WasmdExecuteCLI(c *chain, valIdx int, from string, contractAddress, msg, fees string, expectErr bool) {
	// Create a new context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Get the execute CMD
	executeCmd := []string{
		kiichaindBinary,
		txCommand,
		"wasm",
		"execute",
		contractAddress,
		msg,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.T().Logf("%s executing wasm on host chain %s", from, c.id)
	s.executeKiichainTxCommand(ctx, c, executeCmd, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Log("successfully sent execute wasm tx")
}

// WasmdInstantiateCLI instantiates a contract using the CLI
func (s *IntegrationTestSuite) WasmdInstantiateCLI(c *chain, valIdx int, from string, codeID int, msg, label, admin, fees string, expectErr bool) {
	// Create a new context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Get the instantiate CMD
	instantiateCmd := []string{
		kiichaindBinary,
		txCommand,
		"wasm",
		"instantiate",
		fmt.Sprintf("%d", codeID),
		msg,
		fmt.Sprintf("--%s=%s", "label", label),
		fmt.Sprintf("--%s=%s", "admin", admin),
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.T().Logf("%s instantiating wasm on host chain %s", from, c.id)
	s.executeKiichainTxCommand(ctx, c, instantiateCmd, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Log("successfully sent instantiate wasm tx")
}

// WasmdStoreCLI stores a contract using the CLI
func (s *IntegrationTestSuite) WasmdStoreCLI(c *chain, valIdx int, from, contractPath, fees string, expectErr bool) {
	// Create a new context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// Get the store CMD
	storeCmd := []string{
		kiichaindBinary,
		txCommand,
		"wasm",
		"store",
		contractPath,
		fmt.Sprintf("--%s=%s", flags.FlagFrom, from),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		"--keyring-backend=test",
		"--output=json",
		"--gas=20000000",
		"-y",
	}

	// Log and create the command
	s.T().Logf("%s storing wasm on host chain %s", from, c.id)
	s.executeKiichainTxCommand(ctx, c, storeCmd, valIdx, s.defaultExecValidation(c, valIdx))
	s.T().Log("successfully sent store wasm tx")
}

// queryWasmContractAddressAPI queries the contract address using the API
func (s *IntegrationTestSuite) queryWasmContractAddressAPI(endpoint, creator string, contractIdx int) (string, error) {
	// Create the request and get the response
	body, err := httpGet(fmt.Sprintf("%s/cosmwasm/wasm/v1/contracts/creator/%s", endpoint, creator))
	if err != nil {
		return "", fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	// Unmarshal the response
	var response wasmtypes.QueryContractsByCreatorResponse
	if err = cdc.UnmarshalJSON(body, &response); err != nil {
		return "", err
	}

	// Return the contract address
	return response.ContractAddresses[contractIdx], nil
}

// queryWasmSmartQueryAPI queries the smart query API
func (s *IntegrationTestSuite) queryWasmSmartQueryAPI(endpoint, contractAddress, msg string) ([]byte, error) {
	// Encode the msg
	msgEncoded := base64.StdEncoding.EncodeToString([]byte(msg))

	// Create the request and get the response
	body, err := httpGet(fmt.Sprintf("%s/cosmwasm/wasm/v1/contract/%s/smart/%s", endpoint, contractAddress, msgEncoded))
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	// Unmarshal the response
	var response wasmtypes.QuerySmartContractStateResponse
	if err = cdc.UnmarshalJSON(body, &response); err != nil {
		return nil, err
	}

	// Return the contract address
	return response.Data, nil
}
