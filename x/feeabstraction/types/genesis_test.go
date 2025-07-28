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
			name:         "valid - custom genesis state",
			genesisState: types.NewGenesisState(types.NewParams("coin")),
		},
		{
			name:         "invalid - empty native denom",
			genesisState: types.NewGenesisState(types.NewParams("")),
			errContains:  "invalid denom",
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
