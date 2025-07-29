package types_test

import (
	"testing"

	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
	"github.com/stretchr/testify/require"
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
			msg:  types.NewMessageUpdateParams("cosmos1...", types.DefaultParams()),
		},
		{
			name: "valid - custom params",
			msg:  types.NewMessageUpdateParams("cosmos1...", types.NewParams("coin", types.DefaultMaxPriceDeviation, types.DefaultClampFactor, true)),
		},
		{
			name:        "invalid - empty authority",
			msg:         types.NewMessageUpdateParams("", types.DefaultParams()),
			errContains: "invalid bech32 address",
		},
		{
			name:        "invalid - bad params",
			msg:         types.NewMessageUpdateParams("cosmos1...", types.NewParams("", types.DefaultMaxPriceDeviation, types.DefaultClampFactor, true)),
			errContains: "invalid denom",
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
