package evm_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	mock "github.com/kiichain/kiichain/v1/tests/e2e/mock"
	evmwasmbinding "github.com/kiichain/kiichain/v1/wasmbinding/evm"
	evmbindingtypes "github.com/kiichain/kiichain/v1/wasmbinding/evm/types"

	"github.com/kiichain/kiichain/v1/wasmbinding/helpers"
	"github.com/stretchr/testify/require"
)

// TestHandleEVMQuery tests the HandleEVMQuery function
func TestHandleEVMQuery(t *testing.T) {
	// Setup the app
	actor := helpers.RandomAccountAddress()
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
		query       evmbindingtypes.EthCall
		expected    *evmbindingtypes.EthCallResponse
		errContains string
	}{
		{
			name: "Valid - getCounter",
			query: evmbindingtypes.EthCall{
				Contract: contractAddr.String(),
				Data:     hexutil.Encode(inputData),
			},
			expected: &evmbindingtypes.EthCallResponse{
				Data: "0x0000000000000000000000000000000000000000000000000000000000000001",
			},
		},
		{
			name: "Valid - random data",
			query: evmbindingtypes.EthCall{
				Contract: "0xfeedfacefeedfacefeedfacefeedfacefeedface",
				Data:     "0x00",
			},
			expected: &evmbindingtypes.EthCallResponse{
				Data: "0x",
			},
		},
		{
			name:        "invalid - empty call",
			query:       evmbindingtypes.EthCall{},
			errContains: "empty hex string",
		},
		{
			name: "invalid - error decoding data",
			query: evmbindingtypes.EthCall{
				Contract: "0x0",
				Data:     "not hex",
			},
			errContains: "hex string without 0x prefix",
		},
		{
			name: "invalid - vm execution reverted",
			query: evmbindingtypes.EthCall{
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
