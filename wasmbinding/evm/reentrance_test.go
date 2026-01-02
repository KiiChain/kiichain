package evm_test

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	erc20types "github.com/cosmos/evm/x/erc20/types"

	app "github.com/kiichain/kiichain/v7/app"
	"github.com/kiichain/kiichain/v7/app/apptesting"
	"github.com/kiichain/kiichain/v7/wasmbinding/helpers"
)

// Test the EVM -> Wasm(precompile) -> EVM round-trip
func TestBridge_Reentrancy(t *testing.T) {
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	// Deploy the EVM ReentrancyHarness
	reflect, hAddr := prepareTest(t, ctx, app, actor)

	// Ping pack
	ping, _ := helpers.ReentrancyABI.Pack(
		"ping",
		big.NewInt(1024), // Burn a low amount of gas
	)

	// Prepare the inner Wasm→EVM EthCall query payload
	internalCall := helpers.BuildReflectChainQueryForEthCall(t, hAddr, "0x"+common.Bytes2Hex(ping))

	// Prepare the ABI call data, we will call reentranceQuery()
	// With the same data
	internalInput, err := helpers.ReentrancyABI.Pack(
		"reentranceQuery",
		// The reflect contract address
		reflect.String(),
		// The payload is the inner Wasm→EVM EthCall query
		internalCall,
	)
	require.NoError(t, err)

	// Execute one EVM call
	from := common.BytesToAddress(authtypes.NewModuleAddress(erc20types.ModuleName).Bytes())
	ret, err := app.EVMKeeper.CallEVMWithData(ctx, from, &hAddr, internalInput, false, nil)
	require.NoError(t, err)
	require.NotNil(t, ret.Ret)

	// Unpack the output
	var output *big.Int
	err = helpers.ReentrancyABI.UnpackIntoInterface(&output, "ping", ret.Ret)
	require.NoError(t, err)
	require.Equal(t, int64(32), output.Int64())
}

// Test the EVM -> Wasm(precompile) -> EVM round-trip with multiple calls
func TestBridge_Reentrancy_MultipleCalls(t *testing.T) {
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	// Deploy the EVM ReentrancyHarness
	reflect, hAddr := prepareTest(t, ctx, app, actor)

	// Ping pack
	ping, _ := helpers.ReentrancyABI.Pack(
		"ping",
		big.NewInt(1024), // Burn a low amount of gas
	)

	// Prepare the inner Wasm→EVM EthCall query payload
	internalCall := helpers.BuildReflectChainQueryForEthCall(t, hAddr, "0x"+common.Bytes2Hex(ping))

	// Prepare the ABI call data, we will call reentranceQuery()
	// With the same data
	internalInput, err := helpers.ReentrancyABI.Pack(
		"spin",
		// The reflect contract address
		reflect.String(),
		// The payload is the inner Wasm→EVM EthCall query
		internalCall,
		big.NewInt(2), // Number of calls
	)
	require.NoError(t, err)

	// Execute one EVM call
	from := common.BytesToAddress(authtypes.NewModuleAddress(erc20types.ModuleName).Bytes())
	ret, err := app.EVMKeeper.CallEVMWithData(ctx, from, &hAddr, internalInput, false, nil)

	// We expect an error here due to reentrancy protection
	require.Error(t, err)

	// The reentrance message is in the return data
	require.NotNil(t, ret.Ret)
	require.Contains(t, string(ret.Ret), "reentrancy detected in precompile")
}

// Test the EVM -> Wasm(precompile) -> EVM round-trip with excessive gas consumption
func TestBridge_Reentrancy_ExcessiveGas(t *testing.T) {
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	// Deploy the EVM ReentrancyHarness
	reflect, hAddr := prepareTest(t, ctx, app, actor)

	// Ping pack
	ping, _ := helpers.ReentrancyABI.Pack(
		"ping",
		big.NewInt(1_000_000_000), // Burn an excessive amount of gas
	)

	// Prepare the inner Wasm→EVM EthCall query payload
	internalCall := helpers.BuildReflectChainQueryForEthCall(t, hAddr, "0x"+common.Bytes2Hex(ping))

	// Prepare the ABI call data, we will call reentranceQuery()
	// With the same data
	internalInput, err := helpers.ReentrancyABI.Pack(
		"reentranceQuery",
		// The reflect contract address
		reflect.String(),
		// The payload is the inner Wasm→EVM EthCall query
		internalCall,
	)
	require.NoError(t, err)

	// Execute one EVM call
	from := common.BytesToAddress(authtypes.NewModuleAddress(erc20types.ModuleName).Bytes())
	ret, err := app.EVMKeeper.CallEVMWithData(ctx, from, &hAddr, internalInput, false, nil)

	// We expect an error here
	require.Error(t, err)
	// The out of gas message is in the return data
	require.NotNil(t, ret.Ret)
	require.Contains(t, string(ret.Ret), "Querier contract error")
}

// prepareTest is a helper to prepare the test environment
func prepareTest(t *testing.T, ctx sdk.Context, app *app.KiichainApp, actor sdk.AccAddress) (sdk.AccAddress, common.Address) {
	t.Helper()

	// Register the wasmd precompile
	err := app.EVMKeeper.EnableStaticPrecompiles(ctx, common.HexToAddress("0x0000000000000000000000000000000000001001"))
	require.NoError(t, err)

	// Instantiate the existing Reflect CosmWasm contract
	reflect := helpers.InstantiateReflectContract(t, ctx, app, actor)
	require.NotEmpty(t, reflect)

	// Deploy the EVM ReentrancyHarness
	hAddr := helpers.DeployReentrancy(t, ctx, app)

	return reflect, hAddr
}
