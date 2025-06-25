package evm_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"

	"github.com/kiichain/kiichain/v2/app/apptesting"
	mock "github.com/kiichain/kiichain/v2/tests/e2e/mock"
	evmwasmbinding "github.com/kiichain/kiichain/v2/wasmbinding/evm"
	evmbindingtypes "github.com/kiichain/kiichain/v2/wasmbinding/evm/types"
	"github.com/kiichain/kiichain/v2/wasmbinding/helpers"
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
	contractAddr := deployERC20(t, ctx, app)

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
			errContains: "abi: attempting to unmarshall",
		},
		{
			name: "invalid - Empty contract address",
			query: evmbindingtypes.ERC20InformationRequest{
				Contract: "",
			},
			errContains: "abi: attempting to unmarshall an empty string",
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
	contractAddr := deployERC20(t, ctx, app)
	mintERC20(t, ctx, app, contractAddr, common.Address(actor.Bytes()), big.NewInt(100))

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
			errContains: "abi: attempting to unmarshall an empty string",
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
	contractAddr := deployERC20(t, ctx, app)
	mintERC20(t, ctx, app, contractAddr, common.Address(actor.Bytes()), big.NewInt(200))
	createERC20Allowance(t, ctx, app, contractAddr, actor, actor2, big.NewInt(100))

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
			errContains: "abi: attempting to unmarshall an empty string",
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
