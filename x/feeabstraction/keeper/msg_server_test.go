package keeper_test

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
)

// TestUpdateParams tests the UpdateParams method
func (s *KeeperTestSuite) TestUpdateParams() {
	// Prepare all the test cases
	testCases := []struct {
		name        string
		msg         *types.MsgUpdateParams
		errContains string
	}{
		{
			name: "valid - valid param update",
			msg: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params:    types.NewParams("testcoin"),
			},
		},
		{
			name: "invalid - invalid params",
			msg: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params:    types.Params{NativeDenom: "invalid denom!"},
			},
			errContains: "invalid denom",
		},
		{
			name: "invalid - invalid authority",
			msg: &types.MsgUpdateParams{
				Authority: "invalid_authority",
				Params:    types.NewParams("testcoin"),
			},
			errContains: "invalid authority address",
		},
	}

	// Iterate through the test cases
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Call the UpdateParams method
			_, err := s.msgServer.UpdateParams(s.ctx, tc.msg)

			// Check for errors
			if tc.errContains != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)

				// Verify the params were updated
				params, err := s.keeper.Params.Get(s.ctx)
				s.Require().NoError(err)
				s.Require().Equal(tc.msg.Params, params)
			}
		})
	}
}
