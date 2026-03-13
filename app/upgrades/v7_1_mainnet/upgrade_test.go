package v710mainnet_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	kiihelpers "github.com/kiichain/kiichain/v7/app/helpers"
	"github.com/kiichain/kiichain/v7/app/params"
	v710_mainnet "github.com/kiichain/kiichain/v7/app/upgrades/v7_1_mainnet"
)

const (
	wasmdPrecompileAddress = "0x0000000000000000000000000000000000001001"
	ics20PrecompileAddress = "0x0000000000000000000000000000000000000802"
	otherPrecompileAddress = "0x0000000000000000000000000000000000000803"
)

// TestEVMUpgrade verifies that EVMUpgrade correctly sets bank denom metadata
// and EVM extended denom options.
func TestEVMUpgrade(t *testing.T) {
	app, ctx := kiihelpers.SetupWithContext(t)

	err := v710_mainnet.EVMUpgrade(ctx, &app.AppKeepers)
	require.NoError(t, err)

	// Verify bank denom metadata is set for the base denom
	metadata, found := app.BankKeeper.GetDenomMetaData(ctx, params.BaseDenom)
	require.True(t, found, "denom metadata for %s should exist", params.BaseDenom)
	require.Equal(t, params.BaseDenom, metadata.Base)
	require.Equal(t, params.DisplayDenom, metadata.Display)
	require.Equal(t, "Kii", metadata.Name)
	require.Equal(t, "KII", metadata.Symbol)
	require.Len(t, metadata.DenomUnits, 2)
	require.Equal(t, params.BaseDenom, metadata.DenomUnits[0].Denom)
	require.Equal(t, uint32(0), metadata.DenomUnits[0].Exponent)
	require.Equal(t, params.DisplayDenom, metadata.DenomUnits[1].Denom)
	require.Equal(t, uint32(18), metadata.DenomUnits[1].Exponent)

	// Verify EVM params have ExtendedDenomOptions set correctly
	evmParams := app.EVMKeeper.GetParams(ctx)
	require.NotNil(t, evmParams.ExtendedDenomOptions, "ExtendedDenomOptions should be set")
	require.Equal(t, params.BaseDenom, evmParams.ExtendedDenomOptions.ExtendedDenom)
}

// TestCreateUpgradeHandler_Precompiles verifies that the upgrade handler removes the
// wasmd precompile, keeps unrelated precompiles, and adds ICS20 when it is missing.
//
// Initial state: [otherPrecompile, wasmdPrecompile] (ICS20 absent)
// Expected state: [otherPrecompile, ics20Precompile]
func TestCreateUpgradeHandler_Precompiles(t *testing.T) {
	app, ctx := kiihelpers.SetupWithContext(t)

	// Set up initial active precompiles: one precompile that stays + wasmd (no ICS20)
	evmParams := app.EVMKeeper.GetParams(ctx)
	evmParams.ActiveStaticPrecompiles = []string{otherPrecompileAddress, wasmdPrecompileAddress}
	err := app.EVMKeeper.SetParams(ctx, evmParams)
	require.NoError(t, err)

	// Run the upgrade handler
	mm := app.GetModuleManager()
	handler := v710_mainnet.CreateUpgradeHandler(mm, app.GetConfigurator(), &app.AppKeepers)
	vm, err := handler(ctx, upgradetypes.Plan{Name: v710_mainnet.UpgradeName}, mm.GetVersionMap())
	require.NoError(t, err)
	require.NotNil(t, vm)

	resultParams := app.EVMKeeper.GetParams(ctx)
	result := resultParams.ActiveStaticPrecompiles

	// wasmd must be gone
	for _, p := range result {
		require.NotEqual(t, wasmdPrecompileAddress, p, "wasmd precompile should have been removed")
	}

	// unrelated precompile must still be present
	require.Contains(t, result, otherPrecompileAddress, "unrelated precompile should be kept")

	// ICS20 must have been added back
	require.Contains(t, result, ics20PrecompileAddress, "ICS20 precompile should have been added")

	// ICS20 must appear exactly once
	count := 0
	for _, p := range result {
		if p == ics20PrecompileAddress {
			count++
		}
	}
	require.Equal(t, 1, count, "ICS20 precompile should appear exactly once")
}
