package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
)

// TestMsgUpdateParamsValidate tests the Validate method of MsgUpdateParams
func TestMsgUpdateParamsValidate(t *testing.T) {
	// Prepare all the test cases
	testCases := []struct {
		name        string
		msg         *types.MsgUpdateParams
		errContains string
	}{
		{
			name: "valid - default params",
			msg:  types.NewMessageUpdateParams(authtypes.NewModuleAddress(govtypes.ModuleName).String(), types.DefaultParams()),
		},
		{
			name: "valid - custom params",
			msg:  types.NewMessageUpdateParams(authtypes.NewModuleAddress(govtypes.ModuleName).String(), types.NewParams("coin", types.DefaultMaxPriceDeviation, types.DefaultClampFactor, true)),
		},
		{
			name:        "invalid - empty authority",
			msg:         types.NewMessageUpdateParams("", types.DefaultParams()),
			errContains: "empty address string is not allowed",
		},
		{
			name:        "invalid - bad params",
			msg:         types.NewMessageUpdateParams(authtypes.NewModuleAddress(govtypes.ModuleName).String(), types.NewParams("", types.DefaultMaxPriceDeviation, types.DefaultClampFactor, true)),
			errContains: "native denom is invalid: invalid fee abstraction params",
		},
	}

	// Iterate through the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.Validate()

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
