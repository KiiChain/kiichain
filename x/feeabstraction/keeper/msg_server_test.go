package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	erc20types "github.com/cosmos/evm/x/erc20/types"

	"github.com/kiichain/kiichain/v7/app/apptesting"
	"github.com/kiichain/kiichain/v7/x/feeabstraction/types"
	oracletypes "github.com/kiichain/kiichain/v7/x/oracle/types"
)

// TestUpdateParams tests the UpdateParams method
func (s *KeeperTestSuite) TestUpdateParams() {
	// registerOracleTarget registers a denom as an oracle vote target and adds it to the whitelist
	registerOracleTarget := func(ctx sdk.Context, denom string) {
		err := s.app.OracleKeeper.VoteTarget.Set(ctx, denom, oracletypes.Denom{Name: denom})
		s.Require().NoError(err)

		oracleParams, err := s.app.OracleKeeper.Params.Get(ctx)
		s.Require().NoError(err)
		oracleParams.Whitelist = append(oracleParams.Whitelist, oracletypes.Denom{Name: denom})
		err = s.app.OracleKeeper.Params.Set(ctx, oracleParams)
		s.Require().NoError(err)
	}

	// Prepare all the test cases
	testCases := []struct {
		name        string
		msg         *types.MsgUpdateParams
		malleate    func(ctx sdk.Context)
		errContains string
	}{
		{
			name: "valid - valid param update",
			msg: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params:    types.NewParams("testcoin", "testcoin", types.DefaultClampFactor, types.DefaultTwapLookbackWindow, true),
			},
			malleate: func(ctx sdk.Context) {
				registerOracleTarget(ctx, "testcoin")
			},
		},
		{
			name: "invalid - twap lookback window too high",
			msg: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params:    types.NewParams("testcoin", "testcoin", types.DefaultClampFactor, 1000000, true),
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
		{
			name: "invalid - native oracle denom not a vote target",
			msg: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params:    types.NewParams("testcoin", "testcoin", types.DefaultClampFactor, types.DefaultTwapLookbackWindow, true),
			},
			errContains: "native oracle denom testcoin is not registered as an oracle vote target",
		},
		{
			name: "invalid - native oracle denom not in whitelist",
			msg: &types.MsgUpdateParams{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Params:    types.NewParams("testcoin", "testcoin", types.DefaultClampFactor, types.DefaultTwapLookbackWindow, true),
			},
			malleate: func(ctx sdk.Context) {
				// Register as vote target but do NOT add to whitelist
				err := s.app.OracleKeeper.VoteTarget.Set(ctx, "testcoin", oracletypes.Denom{Name: "testcoin"})
				s.Require().NoError(err)
			},
			errContains: "native oracle denom testcoin is not in the oracle whitelist",
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

			// Call the UpdateParams method
			_, err := s.msgServer.UpdateParams(cachedCtx, tc.msg)

			// Check for errors
			if tc.errContains != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)

				// Verify the params were updated
				params, err := s.keeper.Params.Get(cachedCtx)
				s.Require().NoError(err)
				s.Require().Equal(tc.msg.Params, params)
			}
		})
	}
}

// TestUpdateFeeTokens tests the UpdateFeeTokens method
func (s *KeeperTestSuite) TestUpdateFeeTokens() {
	defaultFeeTokens := types.NewUpdateTokenMetadataCollection(
		types.NewUpdateTokenMetadata("one", "oracleone", 6),
		types.NewUpdateTokenMetadata("two", "oracletwo", 6),
		types.NewUpdateTokenMetadata("three", "oraclethree", 6))

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
				Tokens:    *defaultFeeTokens,
			},
			errContains: "expected gov account as only signer for proposal message",
		},
		{
			name: "invalid - invalid fee tokens (bad denom)",
			msg: types.NewMessageUpdateFeeTokens(
				authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				*types.NewUpdateTokenMetadataCollection(
					types.NewUpdateTokenMetadata("invalid denom!", "oracleCoin", 6),
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
				s.Require().Len(tokens.Items, len(tc.msg.Tokens.Items))
				for i, token := range tokens.Items {
					s.Require().Equal(tc.msg.Tokens.Items[i].Denom, token.Denom)
					s.Require().Equal(tc.msg.Tokens.Items[i].OracleDenom, token.OracleDenom)
					s.Require().Equal(tc.msg.Tokens.Items[i].Decimals, token.Decimals)
					s.Require().True(token.Price.IsZero(), "expected price to be 0 for denom %s, got %s", token.Denom, token.Price)
				}
			}
		})
	}
}

// TestUpdateFeeTokensDecimalsMismatch tests that MsgUpdateFeeTokens rejects
// proposals where FeeTokenMetadata.Decimals disagrees with the existing
// token records
func (s *KeeperTestSuite) TestUpdateFeeTokensDecimalsMismatch() {
	// registerOracleTarget registers a denom as a vote target on the oracle module.
	registerOracleTarget := func(ctx sdk.Context, denom string) {
		err := s.app.OracleKeeper.VoteTarget.Set(ctx, denom, oracletypes.Denom{Name: denom})
		s.Require().NoError(err)
	}

	govAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	testCases := []struct {
		name        string
		malleate    func(ctx sdk.Context) sdk.Context
		token       types.UpdateTokenMetadata
		errContains string
	}{
		// ERC20
		{
			name: "erc20 - correct decimals accepted",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Configure ERC20 w/ 18 decimals
				erc20Addr, err := apptesting.DeployERC20(ctx, s.app)
				s.Require().NoError(err)
				_, err = s.app.Erc20Keeper.RegisterERC20(ctx, &erc20types.MsgRegisterERC20{
					Signer:         govAddr,
					Erc20Addresses: []string{erc20Addr.Hex()},
				})
				s.Require().NoError(err)
				denom := "erc20:" + erc20Addr.Hex()
				registerOracleTarget(ctx, "oracleerc20")
				ctx = ctx.WithValue("erc20denom", denom)
				return ctx
			},
			token:       types.NewUpdateTokenMetadata("placeholder", "oracleerc20", 18),
			errContains: "",
		},
		{
			name: "erc20 - wrong decimals rejected",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Configure ERC20 w/ 18 decimals
				erc20Addr, err := apptesting.DeployERC20(ctx, s.app)
				s.Require().NoError(err)
				_, err = s.app.Erc20Keeper.RegisterERC20(ctx, &erc20types.MsgRegisterERC20{
					Signer:         govAddr,
					Erc20Addresses: []string{erc20Addr.Hex()},
				})
				s.Require().NoError(err)
				denom := "erc20:" + erc20Addr.Hex()
				registerOracleTarget(ctx, "oracleerc20")
				ctx = ctx.WithValue("erc20denom", denom)
				return ctx
			},
			// DeployERC20 deploys with 18 decimals; we deliberately register 6.
			token:       types.NewUpdateTokenMetadata("placeholder", "oracleerc20", 6),
			errContains: "declared decimals 6 does not match contract decimals 18",
		},
		// Bank path
		{
			name: "bank - correct decimals accepted",
			malleate: func(ctx sdk.Context) sdk.Context {
				s.app.BankKeeper.SetDenomMetaData(ctx, banktypes.Metadata{
					Base:    "ucoin",
					Display: "coin",
					DenomUnits: []*banktypes.DenomUnit{
						{Denom: "ucoin", Exponent: 0},
						{Denom: "coin", Exponent: 6},
					},
				})
				registerOracleTarget(ctx, "oraclecoin")
				return ctx
			},
			token:       types.NewUpdateTokenMetadata("ucoin", "oraclecoin", 6),
			errContains: "",
		},
		{
			name: "bank - wrong decimals rejected",
			malleate: func(ctx sdk.Context) sdk.Context {
				s.app.BankKeeper.SetDenomMetaData(ctx, banktypes.Metadata{
					Base:    "ucoin",
					Display: "coin",
					DenomUnits: []*banktypes.DenomUnit{
						{Denom: "ucoin", Exponent: 0},
						{Denom: "coin", Exponent: 6},
					},
				})
				registerOracleTarget(ctx, "oraclecoin")
				return ctx
			},
			// Metadata says 6; we deliberately declare 18.
			token:       types.NewUpdateTokenMetadata("ucoin", "oraclecoin", 18),
			errContains: "declared decimals 18 does not match bank metadata decimals 6",
		},
		{
			name: "bank - no metadata registered, check skipped",
			malleate: func(ctx sdk.Context) sdk.Context {
				registerOracleTarget(ctx, "oracleunknown")
				return ctx
			},
			token:       types.NewUpdateTokenMetadata("unknowncoin", "oracleunknown", 8),
			errContains: "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cachedCtx, _ := s.ctx.CacheContext()

			if tc.malleate != nil {
				cachedCtx = tc.malleate(cachedCtx)
			}

			// For ERC20 cases the denom depends on deployed address
			token := tc.token
			if denom, ok := cachedCtx.Value("erc20denom").(string); ok && denom != "" {
				token.Denom = denom
			}

			msg := types.NewMessageUpdateFeeTokens(
				govAddr,
				*types.NewUpdateTokenMetadataCollection(token),
			)

			_, err := s.msgServer.UpdateFeeTokens(cachedCtx, msg)

			if tc.errContains != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
