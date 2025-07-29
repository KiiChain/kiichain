package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
)

// TestGenesisStateValidate tests the Validate method of GenesisState
func TestGenesisStateValidate(t *testing.T) {
	// Create all the test cases
	testCases := []struct {
		name         string
		genesisState *types.GenesisState
		errContains  string
	}{
		{
			name:         "valid - default genesis state",
			genesisState: types.DefaultGenesisState(),
		},
		{
			name: "valid - custom genesis state",
			genesisState: types.NewGenesisState(
				types.NewParams(
					"coin", types.DefaultMaxPriceDeviation, types.DefaultClampFactor, true),
				types.NewFeeTokenMetadataCollection(
					types.NewFeeTokenMetadata("coin", "oraclecoin", 6, types.DefaultMaxPriceDeviation, types.DefaultClampFactor),
					types.NewFeeTokenMetadata("two", "oracletwo", 18, types.DefaultMaxPriceDeviation.MulInt64(2), types.DefaultClampFactor.MulInt64(2)),
				),
			),
		},
		{
			name: "invalid - bad param",
			genesisState: types.NewGenesisState(
				types.NewParams("", types.DefaultMaxPriceDeviation, types.DefaultClampFactor, true),
				types.NewFeeTokenMetadataCollection(),
			),
			errContains: "invalid denom",
		},
		{
			name: "invalid - invalid fee token metadata",
			genesisState: types.NewGenesisState(
				types.DefaultParams(),
				types.NewFeeTokenMetadataCollection(
					types.NewFeeTokenMetadata("", "oraclecoin", 6, types.DefaultMaxPriceDeviation, types.DefaultClampFactor),
				),
			),
			errContains: "invalid fee token metadata",
		},
		{
			name: "invalid - duplicate fee token denom",
			genesisState: types.NewGenesisState(
				types.DefaultParams(),
				types.NewFeeTokenMetadataCollection(
					types.NewFeeTokenMetadata("coin", "oraclecoin", 6, types.DefaultMaxPriceDeviation, types.DefaultClampFactor),
					types.NewFeeTokenMetadata("coin", "oraclecoin2", 6, types.DefaultMaxPriceDeviation, types.DefaultClampFactor),
				),
			),
			errContains: "duplicate denom found: coin",
		},
	}

	// Iterate through the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genesisState.Validate()

			// Check the error
			if tc.errContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			}
		})
	}
}
