package evm_test

import (
	"encoding/json"
	"math/big"
	"testing"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/evm/contracts"
	erc20types "github.com/cosmos/evm/x/erc20/types"

	app "github.com/kiichain/kiichain/v2/app"
	"github.com/kiichain/kiichain/v2/app/apptesting"
	mock "github.com/kiichain/kiichain/v2/tests/e2e/mock"
	"github.com/kiichain/kiichain/v2/wasmbinding"
	evmbindingtypes "github.com/kiichain/kiichain/v2/wasmbinding/evm/types"
	"github.com/kiichain/kiichain/v2/wasmbinding/helpers"
)

// TestQueryEthCall test the EthCall query
func TestQueryEthCall(t *testing.T) {
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	reflect := helpers.InstantiateReflectContract(t, ctx, app, actor)
	require.NotEmpty(t, reflect)

	// Deploy the counter contract
	contractAddr := deployCounter(t, ctx, app)
	require.NotEmpty(t, contractAddr)
	// Increment the counter
	incrementCounter(t, ctx, app, contractAddr)

	// Prepare ABI call data for getCounter()
	counterABI, err := mock.CounterMetaData.GetAbi()
	require.NoError(t, err)

	// Prepare the input data for the getCounter function
	inputData, err := counterABI.Pack("getCounter")
	require.NoError(t, err)

	// Perform the eth_call query
	query := evmbindingtypes.Query{
		EthCall: &evmbindingtypes.EthCallRequest{
			Contract: contractAddr.Hex(),
			Data:     hexutil.Encode(inputData),
		},
	}
	resp := evmbindingtypes.EthCallRequest{}
	err = queryCustom(t, ctx, app, reflect, query, &resp)
	require.NoError(t, err)

	// Decode the response data
	resultBytes, err := hexutil.Decode(resp.Data)
	require.NoError(t, err)

	// Unpack the ABI result
	var counterValue *big.Int
	err = counterABI.UnpackIntoInterface(&counterValue, "getCounter", resultBytes)
	require.NoError(t, err)

	// Check the expected value (0 if freshly deployed)
	require.EqualValues(t, 1, counterValue.Int64())
}

// TestQueryEthCallWithError test the EthCall query with an error
func TestQueryEthCallWithError(t *testing.T) {
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	reflect := helpers.InstantiateReflectContract(t, ctx, app, actor)
	require.NotEmpty(t, reflect)

	contractAddr := deployCounter(t, ctx, app)

	// Perform the eth_call query
	query := evmbindingtypes.Query{
		EthCall: &evmbindingtypes.EthCallRequest{
			Contract: contractAddr.Hex(),
			Data:     hexutil.Encode([]byte("0xdeadbeef")), // Random data, it will revert
		},
	}
	resp := evmbindingtypes.EthCallRequest{}
	err := queryCustom(t, ctx, app, reflect, query, &resp)
	require.ErrorContains(t, err, "codespace: evm_wasmbinding, code: 1")
}

// TestQueryERC20Information test the ERC20Information query
func TestQueryERC20Information(t *testing.T) {
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	reflect := helpers.InstantiateReflectContract(t, ctx, app, actor)
	require.NotEmpty(t, reflect)

	// Deploy the counter contract
	contractAddr := deployERC20(t, ctx, app)
	require.NotEmpty(t, contractAddr)

	// Perform the ERC20Information query
	query := evmbindingtypes.Query{
		ERC20Information: &evmbindingtypes.ERC20InformationRequest{
			Contract: contractAddr.Hex(),
		},
	}
	resp := evmbindingtypes.ERC20InformationResponse{}
	err := queryCustom(t, ctx, app, reflect, query, &resp)
	require.NoError(t, err)

	// Check the expected value (0 if freshly deployed)
	require.EqualValues(t, "Test", resp.Name)
	require.EqualValues(t, "TEST", resp.Symbol)
	require.EqualValues(t, 18, resp.Decimals)
}

// TestQueryERC20Balance test the ERC20Balance query
func TestQueryERC20Balance(t *testing.T) {
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	reflect := helpers.InstantiateReflectContract(t, ctx, app, actor)
	require.NotEmpty(t, reflect)

	// Deploy the counter contract
	contractAddr := deployERC20(t, ctx, app)
	require.NotEmpty(t, contractAddr)

	// Create a new account
	account := createAccountAndRegister(t, ctx, app)
	require.NotEmpty(t, account)

	// Mint some tokens to the account
	mintAmount := big.NewInt(1000)
	mintERC20(t, ctx, app, contractAddr, account, mintAmount)

	// Perform the ERC20Balance query
	query := evmbindingtypes.Query{
		ERC20Balance: &evmbindingtypes.ERC20BalanceRequest{
			Contract: contractAddr.Hex(),
			Address:  account.Hex(),
		},
	}
	resp := evmbindingtypes.ERC20BalanceResponse{}
	err := queryCustom(t, ctx, app, reflect, query, &resp)
	require.NoError(t, err)

	// Check the expected value
	require.EqualValues(t, mintAmount.String(), resp.Balance)
}

// TestQueryERC20Allowance test the ERC20Allowance query
func TestQueryERC20Allowance(t *testing.T) {
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	reflect := helpers.InstantiateReflectContract(t, ctx, app, actor)
	require.NotEmpty(t, reflect)

	// Deploy the counter contract
	contractAddr := deployERC20(t, ctx, app)
	require.NotEmpty(t, contractAddr)

	// Create a new account
	account := createAccountAndRegister(t, ctx, app)
	require.NotEmpty(t, account)

	// Mint some tokens to the account
	mintAmount := big.NewInt(1000)
	mintERC20(t, ctx, app, contractAddr, account, mintAmount)

	// Create an allowance for the actor
	spender := common.BytesToAddress(actor.Bytes())
	createERC20Allowance(t, ctx, app, contractAddr, account, spender, mintAmount)

	// Perform the ERC20Allowance query
	query := evmbindingtypes.Query{
		ERC20Allowance: &evmbindingtypes.ERC20AllowanceRequest{
			Contract: contractAddr.Hex(),
			Owner:    account.Hex(),
			Spender:  spender.Hex(),
		},
	}
	resp := evmbindingtypes.ERC20AllowanceResponse{}
	err := queryCustom(t, ctx, app, reflect, query, &resp)
	require.NoError(t, err)

	// Check the expected value
	require.EqualValues(t, mintAmount.String(), resp.Allowance)
}

// deployCounter deploys the counter contract
func deployCounter(t *testing.T, ctx sdk.Context, app *app.KiichainApp) common.Address {
	t.Helper()
	// Select the from as the erc20 module address
	from := common.BytesToAddress(authtypes.NewModuleAddress(erc20types.ModuleName).Bytes())

	// Set the data
	counterABI, err := mock.CounterMetaData.GetAbi()
	require.NoError(t, err)
	ctorArgs, err := counterABI.Pack("")
	require.NoError(t, err)
	deployData := append(common.FromHex(mock.CounterBin), ctorArgs...)

	// Deploy the contract
	res, err := app.EVMKeeper.CallEVMWithData(ctx, from, nil, deployData, true)
	require.NoError(t, err)
	require.NotNil(t, res.Ret)

	// Derive the deployed contract address
	nonce := app.EVMKeeper.GetNonce(ctx, from)
	contractAddr := crypto.CreateAddress(from, nonce-1)
	return contractAddr
}

// incrementCounter increments the counter in the contract
func incrementCounter(t *testing.T, ctx sdk.Context, app *app.KiichainApp, contractAddr common.Address) {
	t.Helper()
	// Sender must be an account with ETH balance and nonce tracking
	from := common.BytesToAddress(authtypes.NewModuleAddress(erc20types.ModuleName).Bytes())

	// Load the ABI and pack the increment() call
	counterABI, err := mock.CounterMetaData.GetAbi()
	require.NoError(t, err)
	inputData, err := counterABI.Pack("increment")
	require.NoError(t, err)

	// Send transaction to call increment
	res, err := app.EVMKeeper.CallEVMWithData(ctx, from, &contractAddr, inputData, true)
	require.NoError(t, err)
	require.NotNil(t, res)
}

// deployERC20 deploys an ERC20 contract
func deployERC20(t *testing.T, ctx sdk.Context, app *app.KiichainApp) common.Address {
	t.Helper()
	// Select the from as the erc20 module address
	from := common.BytesToAddress(authtypes.NewModuleAddress(erc20types.ModuleName).Bytes())

	// Set the data
	erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI
	ctorArgs, err := erc20ABI.Pack("", "Test", "TEST", uint8(18))
	require.NoError(t, err)
	deployData := append(contracts.ERC20MinterBurnerDecimalsContract.Bin, ctorArgs...) //nolint:gocritic

	// Deploy the contract
	res, err := app.EVMKeeper.CallEVMWithData(ctx, from, nil, deployData, true)
	require.NoError(t, err)
	require.NotNil(t, res.Ret)

	// Derive the deployed contract address
	nonce := app.EVMKeeper.GetNonce(ctx, from)
	contractAddr := crypto.CreateAddress(from, nonce-1)
	return contractAddr
}

// mintERC20 mints an ERC20 token
func mintERC20(t *testing.T, ctx sdk.Context, app *app.KiichainApp, contractAddr common.Address, to common.Address, amount *big.Int) {
	t.Helper()
	// Sender must be an account with ETH balance and nonce tracking
	from := common.BytesToAddress(authtypes.NewModuleAddress(erc20types.ModuleName).Bytes())

	// Load the ABI and pack the mint() call
	erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI
	inputData, err := erc20ABI.Pack("mint", to, amount)
	require.NoError(t, err)

	// Send transaction to call mint
	res, err := app.EVMKeeper.CallEVMWithData(ctx, from, &contractAddr, inputData, true)
	require.NoError(t, err)
	require.NotNil(t, res)
}

// createERC20Allowance creates an ERC20 allowance
func createERC20Allowance(t *testing.T, ctx sdk.Context, app *app.KiichainApp, contractAddr common.Address, owner common.Address, spender common.Address, amount *big.Int) {
	t.Helper()
	// Load the ABI and pack the mint() call
	erc20ABI := contracts.ERC20MinterBurnerDecimalsContract.ABI
	inputData, err := erc20ABI.Pack("approve", spender, amount)
	require.NoError(t, err)

	// Send transaction to call mint
	res, err := app.EVMKeeper.CallEVMWithData(ctx, owner, &contractAddr, inputData, true)
	require.NoError(t, err)
	require.NotNil(t, res)
}

// createAccount creates the account and register using the auth module
func createAccountAndRegister(t *testing.T, ctx sdk.Context, app *app.KiichainApp) common.Address {
	t.Helper()

	// Create a new account
	randomAccount := apptesting.RandomAccountAddress()

	// Create the account in the auth module
	accountI := app.AccountKeeper.NewAccountWithAddress(ctx, randomAccount)
	app.AccountKeeper.SetAccount(ctx, accountI)

	// Return the account as common address
	return common.BytesToAddress(accountI.GetAddress().Bytes())
}

// TestQueryDenomAdmin tests the GetDenomAdmin query
type ReflectQuery struct {
	Chain *ChainRequest `json:"chain,omitempty"`
}

// ChainRequest is the request to the chain
type ChainRequest struct {
	Request wasmvmtypes.QueryRequest `json:"request"`
}

// ChainResponse is the response from the chain
type ChainResponse struct {
	Data []byte `json:"data"`
}

// queryCustom is a helper function to query the custom contract
func queryCustom(t *testing.T, ctx sdk.Context, app *app.KiichainApp, contract sdk.AccAddress, request evmbindingtypes.Query, response interface{}) error {
	t.Helper()

	// Make the request a kiichain query
	kiichainQuery := wasmbinding.KiichainQuery{
		EVM: &request,
	}

	// Marshal the request to JSON
	msgBz, err := json.Marshal(kiichainQuery)
	if err != nil {
		return err
	}
	t.Log("queryCustom1", string(msgBz))

	query := ReflectQuery{
		Chain: &ChainRequest{
			Request: wasmvmtypes.QueryRequest{Custom: msgBz},
		},
	}
	queryBz, err := json.Marshal(query)
	if err != nil {
		return err
	}
	t.Log("queryCustom3", string(queryBz))

	resBz, err := app.WasmKeeper.QuerySmart(ctx, contract, queryBz)
	if err != nil {
		return err
	}
	var resp ChainResponse
	err = json.Unmarshal(resBz, &resp)
	if err != nil {
		return err
	}
	err = json.Unmarshal(resp.Data, response)
	if err != nil {
		return err
	}

	return nil
}
