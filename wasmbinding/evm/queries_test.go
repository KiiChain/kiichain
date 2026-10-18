package evm_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/evm/x/vm/statedb"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	"github.com/kiichain/kiichain/v7/app/apptesting"
	mock "github.com/kiichain/kiichain/v7/tests/e2e/mock"
	evmwasmbinding "github.com/kiichain/kiichain/v7/wasmbinding/evm"
	evmbindingtypes "github.com/kiichain/kiichain/v7/wasmbinding/evm/types"
	"github.com/kiichain/kiichain/v7/wasmbinding/helpers"
)

// TestHandleEVMQuery tests the HandleEVMQuery function
func TestHandleEVMQuery(t *testing.T) {
	// Setup the app
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	// Deploy the counter contract
	contractAddr := deployCounter(t, ctx, app)
	// Increment the counter
	incrementCounter(t, ctx, app, contractAddr)

	// Prepare ABI call data for getCounter()
	counterABI, err := mock.CounterMetaData.GetAbi()
	require.NoError(t, err)

	// Prepare the input data for the getCounter function
	inputData, err := counterABI.Pack("getCounter")
	require.NoError(t, err)

	// Set all the test cases
	testCases := []struct {
		name        string
		query       evmbindingtypes.EthCallRequest
		expected    *evmbindingtypes.EthCallResponse
		errContains string
	}{
		{
			name: "Valid - getCounter",
			query: evmbindingtypes.EthCallRequest{
				Contract: contractAddr.String(),
				Data:     hexutil.Encode(inputData),
			},
			expected: &evmbindingtypes.EthCallResponse{
				Data: "0x0000000000000000000000000000000000000000000000000000000000000001",
			},
		},
		{
			name: "Valid - random data",
			query: evmbindingtypes.EthCallRequest{
				Contract: "0xfeedfacefeedfacefeedfacefeedfacefeedface",
				Data:     "0x00",
			},
			expected: &evmbindingtypes.EthCallResponse{
				Data: "0x",
			},
		},
		{
			name:        "invalid - empty call",
			query:       evmbindingtypes.EthCallRequest{},
			errContains: "empty hex string",
		},
		{
			name: "invalid - error decoding data",
			query: evmbindingtypes.EthCallRequest{
				Contract: "0x0",
				Data:     "not hex",
			},
			errContains: "hex string without 0x prefix",
		},
		{
			name: "invalid - vm execution reverted",
			query: evmbindingtypes.EthCallRequest{
				Contract: contractAddr.String(),                // Valid contract
				Data:     hexutil.Encode([]byte("0xdeadbeef")), // Random data
			},
			errContains: "execution reverted",
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start the query server
			evmQueryPlugin := evmwasmbinding.NewQueryPlugin(app.EVMKeeper)

			// Apply the query
			res, err := evmQueryPlugin.HandleEthCall(ctx, &tc.query)

			// Check the error
			if tc.errContains != "" {
				require.ErrorContains(t, err, tc.errContains)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected.Data, res.Data)
			}
		})
	}
}

// TestHandleERC20Information tests the HandleERC20Information function
func TestHandleERC20Information(t *testing.T) {
	// Setup the app
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	// Deploy a erc20 contract
	contractAddr, err := apptesting.DeployERC20(ctx, app)
	require.NoError(t, err)

	// Set all the test cases
	testCases := []struct {
		name        string
		query       evmbindingtypes.ERC20InformationRequest
		expected    *evmbindingtypes.ERC20InformationResponse
		errContains string
	}{
		{
			name: "Valid",
			query: evmbindingtypes.ERC20InformationRequest{
				Contract: contractAddr.String(),
			},
			expected: &evmbindingtypes.ERC20InformationResponse{
				Decimals:    18,
				Name:        "Test",
				Symbol:      "TEST",
				TotalSupply: "0",
			},
		},
		{
			name: "invalid - Invalid contract address format",
			query: evmbindingtypes.ERC20InformationRequest{
				Contract: "not-an-address",
			},
			errContains: "abi: attempting to unmarshal an empty string while arguments are expected",
		},
		{
			name: "invalid - Empty contract address",
			query: evmbindingtypes.ERC20InformationRequest{
				Contract: "",
			},
			errContains: "abi: attempting to unmarshal an empty string while arguments are expected",
		},
		{
			name: "invalid - Non-ERC20 contract address",
			query: evmbindingtypes.ERC20InformationRequest{
				Contract: deployCounter(t, ctx, app).String(),
			},
			errContains: "execution reverted",
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start the query server
			evmQueryPlugin := evmwasmbinding.NewQueryPlugin(app.EVMKeeper)

			// Apply the query
			res, err := evmQueryPlugin.HandleERC20Information(ctx, &tc.query)

			// Check the error
			if tc.errContains != "" {
				require.ErrorContains(t, err, tc.errContains)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, res)
			}
		})
	}
}

// TestHandleERC20Balance tests the HandleERC20Balance function
func TestHandleERC20Balance(t *testing.T) {
	// Setup the app
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	// Deploy a erc20 contract
	contractAddr, err := apptesting.DeployERC20(ctx, app)
	require.NoError(t, err)
	err = apptesting.MintERC20(ctx, app, contractAddr, common.Address(actor.Bytes()), big.NewInt(100))
	require.NoError(t, err)

	// Set all the test cases
	testCases := []struct {
		name        string
		query       evmbindingtypes.ERC20BalanceRequest
		expected    *evmbindingtypes.ERC20BalanceResponse
		errContains string
	}{
		{
			name: "Valid",
			query: evmbindingtypes.ERC20BalanceRequest{
				Contract: contractAddr.String(),
				Address:  common.Address(actor.Bytes()).String(),
			},
			expected: &evmbindingtypes.ERC20BalanceResponse{
				Balance: "100",
			},
		},
		{
			name:        "invalid - Empty contract address",
			query:       evmbindingtypes.ERC20BalanceRequest{},
			errContains: "abi: attempting to unmarshal an empty string while arguments are expected",
		},
		{
			name: "invalid - Non-ERC20 contract address",
			query: evmbindingtypes.ERC20BalanceRequest{
				Contract: deployCounter(t, ctx, app).String(),
				Address:  common.Address(actor.Bytes()).String(),
			},
			errContains: "execution reverted",
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start the query server
			evmQueryPlugin := evmwasmbinding.NewQueryPlugin(app.EVMKeeper)

			// Apply the query
			res, err := evmQueryPlugin.HandleERC20Balance(ctx, &tc.query)

			// Check the error
			if tc.errContains != "" {
				require.ErrorContains(t, err, tc.errContains)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected.Balance, res.Balance)
			}
		})
	}
}

// TestHandleERC20Allowance tests the HandleERC20Allowance function
func TestHandleERC20Allowance(t *testing.T) {
	// Setup the app
	app, ctx := helpers.SetupCustomApp(t, apptesting.RandomAccountAddress())
	actor := createAccountAndRegister(t, ctx, app)
	actor2 := createAccountAndRegister(t, ctx, app)

	// Deploy a erc20 contract
	contractAddr, err := apptesting.DeployERC20(ctx, app)
	require.NoError(t, err)
	err = apptesting.MintERC20(ctx, app, contractAddr, common.Address(actor.Bytes()), big.NewInt(200))
	require.NoError(t, err)
	err = apptesting.CreateERC20Allowance(ctx, app, contractAddr, actor, actor2, big.NewInt(100))
	require.NoError(t, err)

	// Set all the test cases
	testCases := []struct {
		name        string
		query       evmbindingtypes.ERC20AllowanceRequest
		expected    *evmbindingtypes.ERC20AllowanceResponse
		errContains string
	}{
		{
			name: "Valid",
			query: evmbindingtypes.ERC20AllowanceRequest{
				Contract: contractAddr.String(),
				Owner:    actor.String(),
				Spender:  actor2.String(),
			},
			expected: &evmbindingtypes.ERC20AllowanceResponse{
				Allowance: "100",
			},
		},
		{
			name:        "invalid - Empty contract address",
			query:       evmbindingtypes.ERC20AllowanceRequest{},
			errContains: "abi: attempting to unmarshal an empty string while arguments are expected",
		},
		{
			name: "invalid - Non-ERC20 contract address",
			query: evmbindingtypes.ERC20AllowanceRequest{
				Contract: deployCounter(t, ctx, app).String(),
				Owner:    actor.String(),
				Spender:  actor2.String(),
			},
			errContains: "execution reverted",
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start the query server
			evmQueryPlugin := evmwasmbinding.NewQueryPlugin(app.EVMKeeper)

			// Apply the query
			res, err := evmQueryPlugin.HandleERC20Allowance(ctx, &tc.query)

			// Check the error
			if tc.errContains != "" {
				require.ErrorContains(t, err, tc.errContains)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected.Allowance, res.Allowance)
			}
		})
	}
}

// TestEthCall_ConsumeGas tests that EthCall consumes gas
func TestEthCall_ConsumeGas(t *testing.T) {
	// Set the app
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	// Give to the sdk a budget
	ctx = ctx.WithGasMeter(storetypes.NewGasMeter(500_000))

	// Deploy the counter contract
	contractAddr := deployCounter(t, ctx, app)

	// Prepare ABI call data for getCounter()
	counterABI, err := mock.CounterMetaData.GetAbi()
	require.NoError(t, err)

	// Prepare the input data for the getCounter function
	inputData, err := counterABI.Pack("getCounter")
	require.NoError(t, err)

	// Start the query server
	evmQueryPlugin := evmwasmbinding.NewQueryPlugin(app.EVMKeeper)

	// Save the initial gas
	initialGas := ctx.GasMeter().GasConsumed()

	// Apply the query
	_, err = evmQueryPlugin.HandleEthCall(ctx, &evmbindingtypes.EthCallRequest{
		Contract: contractAddr.String(),
		Data:     hexutil.Encode(inputData),
	})
	require.NoError(t, err)

	// Check if gas was consumed
	finalGas := ctx.GasMeter().GasConsumed()
	require.Greater(t, finalGas, initialGas)
}

// TestEthCall_explodeGas tests gas explosion in EthCall
func TestEthCall_explodeGas(t *testing.T) {
	// Set the app
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	// Deploy the counter contract
	contractAddr := deployCounter(t, ctx, app)

	// Prepare ABI call data for explodeGas()
	counterABI, err := mock.CounterMetaData.GetAbi()
	require.NoError(t, err)

	// Prepare the input data for the explodeGas
	inputData, err := counterABI.Pack("getCounter")
	require.NoError(t, err)

	// Start the query server
	evmQueryPlugin := evmwasmbinding.NewQueryPlugin(app.EVMKeeper)

	// Give to the sdk a budget
	ctx = ctx.WithGasMeter(storetypes.NewGasMeter(1_000))

	// Apply the query
	require.PanicsWithValue(t, storetypes.ErrorOutOfGas{Descriptor: "ReadPerByte"}, func() {
		_, err = evmQueryPlugin.HandleEthCall(ctx, &evmbindingtypes.EthCallRequest{
			Contract: contractAddr.String(),
			Data:     hexutil.Encode(inputData),
		})
	})
}

// TestEthCall_ChargesEVMExecutionGas tests that EthCall charges the SDK gas meter for the EVM work,
// not only for the store reads it performs
func TestEthCall_ChargesEVMExecutionGas(t *testing.T) {
	// Set the app
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	// Install a contract that loops 100_000 times and stops, without touching storage
	// PUSH3 0x0186a0; JUMPDEST; PUSH1 1; SWAP1; SUB; DUP1; PUSH1 4; JUMPI; STOP
	loopCode := common.FromHex("620186a05b600190038060045700")
	contractAddr := common.HexToAddress("0x000000000000000000000000000000000000c0de")
	stateDB := statedb.New(ctx, app.EVMKeeper, statedb.NewEmptyTxConfig())
	stateDB.SetCode(contractAddr, loopCode)
	require.NoError(t, stateDB.Commit())

	// Measure the EVM work of the call directly on the keeper
	data := hexutil.Bytes{}
	args, err := json.Marshal(evmtypes.TransactionArgs{To: &contractAddr, Data: &data})
	require.NoError(t, err)
	refCtx, _ := ctx.CacheContext()
	refCtx = refCtx.WithGasMeter(storetypes.NewInfiniteGasMeter())
	ref, err := app.EVMKeeper.EthCall(refCtx, &evmtypes.EthCallRequest{
		Args:    args,
		GasCap:  10_000_000,
		ChainId: evmtypes.GetEthChainConfig().ChainID.Int64(),
	})
	require.NoError(t, err)
	require.Empty(t, ref.VmError)
	require.Greater(t, ref.MaxUsedGas, uint64(2_000_000))

	// Give to the sdk a budget
	ctx = ctx.WithGasMeter(storetypes.NewGasMeter(10_000_000))
	initialGas := ctx.GasMeter().GasConsumed()

	// Apply the query
	evmQueryPlugin := evmwasmbinding.NewQueryPlugin(app.EVMKeeper)
	_, err = evmQueryPlugin.HandleEthCall(ctx, &evmbindingtypes.EthCallRequest{
		Contract: contractAddr.String(),
		Data:     "0x",
	})
	require.NoError(t, err)

	// The caller must pay at least the EVM work
	require.GreaterOrEqual(t, ctx.GasMeter().GasConsumed()-initialGas, ref.MaxUsedGas)
}

// TestEthCall_ChargesGasOnRevert tests that EthCall charges the SDK gas meter when the call fails
func TestEthCall_ChargesGasOnRevert(t *testing.T) {
	// Set the app
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	// Install a contract that loops forever, so the call runs out of gas
	// JUMPDEST; PUSH1 0; JUMP
	contractAddr := common.HexToAddress("0x000000000000000000000000000000000000dead")
	stateDB := statedb.New(ctx, app.EVMKeeper, statedb.NewEmptyTxConfig())
	stateDB.SetCode(contractAddr, common.FromHex("5b600056"))
	require.NoError(t, stateDB.Commit())

	// Give to the sdk a budget
	ctx = ctx.WithGasMeter(storetypes.NewGasMeter(5_000_000))

	// Apply the query, the failed EVM execution is charged and exhausts the budget
	evmQueryPlugin := evmwasmbinding.NewQueryPlugin(app.EVMKeeper)
	require.PanicsWithValue(t, storetypes.ErrorOutOfGas{Descriptor: "evm eth_call query"}, func() {
		_, _ = evmQueryPlugin.HandleEthCall(ctx, &evmbindingtypes.EthCallRequest{
			Contract: contractAddr.String(),
			Data:     "0x",
		})
	})
}
