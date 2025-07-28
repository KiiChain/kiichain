package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
)

// TestValidateParams tests the Validate method of Params
func TestValidateParams(t *testing.T) {
	// Prepare test cases
	testCases := []struct {
		name        string
		params      types.Params
		errContains string
	}{
		{
			name:   "valid - default params",
			params: types.DefaultParams(),
		},
		{
			name:   "valid - custom params",
			params: types.NewParams("coin"),
		},
		{
			name:        "invalid - empty native denom",
			params:      types.NewParams(""),
			errContains: "invalid denom",
		},
		{
			name:        "invalid - invalid denom",
			params:      types.NewParams("123"),
			errContains: "invalid denom",
		},
	}

	// Iterate through the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.ValidateBasic()

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
