package keeper_test

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
	oracletypes "github.com/kiichain/kiichain/v3/x/oracle/types"
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
				Params:    types.NewParams("testcoin", "testcoin", types.DefaultClampFactor, types.DefaultFallbackNativePrice, types.DefaultTwapLookbackWindow, true),
			},
		},
		{
			name: "invalid - twap lookback window too high",
			msg: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params:    types.NewParams("testcoin", "testcoin", types.DefaultClampFactor, types.DefaultFallbackNativePrice, 1000000, true),
			},
			errContains: "Twap lookback seconds is greater than max lookback duration",
		},
		{
			name: "invalid - invalid params",
			msg: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params:    types.Params{NativeDenom: "invalid denom!"},
			},
			errContains: "native denom is invalid",
		},
		{
			name: "invalid - invalid authority",
			msg: &types.MsgUpdateParams{
				Authority: "invalid_authority",
				Params:    types.DefaultParams(),
			},
			errContains: "invalid authority address: decoding bech32 failed",
		},
		{
			name: "invalid - wrong authority",
			msg: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(types.ModuleName).String(),
				Params:    types.DefaultParams(),
			},
			errContains: "expected gov account as only signer for proposal message",
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

// TestUpdateFeeTokens tests the UpdateFeeTokens method
func (s *KeeperTestSuite) TestUpdateFeeTokens() {
	defaultFeeTokens := types.NewFeeTokenMetadataCollection(
		types.NewFeeTokenMetadata("one", "oracleone", 6, math.LegacyMustNewDecFromStr("0.01")),
		types.NewFeeTokenMetadata("two", "oracletwo", 6, math.LegacyMustNewDecFromStr("0.01")),
		types.NewFeeTokenMetadata("three", "oraclethree", 6, math.LegacyMustNewDecFromStr("0.01")))

	// Prepare all the test cases
	testCases := []struct {
		name        string
		msg         *types.MsgUpdateFeeTokens
		malleate    func(ctx sdk.Context)
		errContains string
	}{
		{
			name: "valid - valid fee tokens update",
			msg: types.NewMessageUpdateFeeTokens(
				authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				*defaultFeeTokens,
			),
			malleate: func(ctx sdk.Context) {
				// Iterate all the tokens
				for _, feeToken := range defaultFeeTokens.Items {
					// Register the token as a vote target on the oracle module
					err := s.app.OracleKeeper.VoteTarget.Set(ctx, feeToken.OracleDenom, oracletypes.Denom{Name: feeToken.OracleDenom})
					s.Require().NoError(err)
				}
			},
		},
		{
			name: "invalid - one token not registered on oracle",
			msg: types.NewMessageUpdateFeeTokens(
				authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				*defaultFeeTokens,
			),
			malleate: func(ctx sdk.Context) {
				// Register only two tokens as vote targets on the oracle module
				err := s.app.OracleKeeper.VoteTarget.Set(ctx, "one", oracletypes.Denom{Name: "one"})
				s.Require().NoError(err)
				err = s.app.OracleKeeper.VoteTarget.Set(ctx, "two", oracletypes.Denom{Name: "two"})
				s.Require().NoError(err)
			},
			errContains: "fee token denom oracleone is not registered on the oracle module",
		},
		{
			name: "invalid - invalid authority",
			msg: types.NewMessageUpdateFeeTokens(
				"",
				*defaultFeeTokens,
			),
			errContains: "invalid authority address: empty address string is not allowed",
		},
		{
			name: "invalid - wrong authority",
			msg: &types.MsgUpdateFeeTokens{
				Authority: authtypes.NewModuleAddress(types.ModuleName).String(),
				FeeTokens: *defaultFeeTokens,
			},
			errContains: "expected gov account as only signer for proposal message",
		},
		{
			name: "invalid - invalid fee tokens (bad denom)",
			msg: types.NewMessageUpdateFeeTokens(
				authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				*types.NewFeeTokenMetadataCollection(
					types.NewFeeTokenMetadata("invalid denom!", "oracleCoin", 6, math.LegacyMustNewDecFromStr("0.01")),
				),
			),
			errContains: "denom is invalid: invalid fee token metadata: invalid request",
		},
	}

	// Iterate through the test cases
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Set a cached context
			cachedCtx, _ := s.ctx.CacheContext()

			// Malleate if exists
			if tc.malleate != nil {
				tc.malleate(cachedCtx)
			}

			// Call the UpdateFeeTokens method
			_, err := s.msgServer.UpdateFeeTokens(cachedCtx, tc.msg)

			// Check for errors
			if tc.errContains != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)

				// Verify the fee tokens were updated
				tokens, err := s.keeper.FeeTokens.Get(cachedCtx)
				s.Require().NoError(err)
				s.Require().Equal(tc.msg.FeeTokens, tokens)
			}
		})
	}
}
