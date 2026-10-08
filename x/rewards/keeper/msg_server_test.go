package keeper_test

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

// TestUpdateParams test changes to the params of the module
func (suite *KeeperTestSuite) TestUpdateParams() {
	testCases := []struct {
		name         string
		msg          *types.MsgUpdateParams
		expectedPass bool
	}{
		{
			name: "valid authority",
			msg: types.NewMsgUpdateParams(
				suite.App.RewardsKeeper.GetAuthority(),
				types.DefaultParams(),
			),
			expectedPass: true,
		},
		{
			name: "invalid authority",
			msg: types.NewMsgUpdateParams(
				suite.TestAccs[0].String(),
				types.DefaultParams(),
			),
			expectedPass: false,
		},
		{
			name: "invalid params - empty denom",
			msg: types.NewMsgUpdateParams(
				suite.App.RewardsKeeper.GetAuthority(),
				types.Params{TokenDenom: ""},
			),
			expectedPass: false,
		},
		{
			name: "valid params with supply base",
			msg: types.NewMsgUpdateParams(
				suite.App.RewardsKeeper.GetAuthority(),
				func() types.Params {
					p := types.DefaultParams()
					p.SupplyBase = math.NewInt(1_000_000)
					return p
				}(),
			),
			expectedPass: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.msgServer.UpdateParams(suite.Ctx, tc.msg)
			if tc.expectedPass {
				suite.Require().NoError(err)

				params, err := suite.App.RewardsKeeper.Params.Get(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(tc.msg.Params, params)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

// TestFundPool tests funding the pool
func (suite *KeeperTestSuite) TestFundPool() {
	defaultParams := types.DefaultParams()
	err := suite.App.RewardsKeeper.Params.Set(suite.Ctx, defaultParams)
	suite.Require().NoError(err)

	testCases := []struct {
		name         string
		msg          *types.MsgFundPool
		expectedPass bool
	}{
		{
			name: "valid funding",
			msg: types.NewMsgFundPool(
				suite.TestAccs[0],
				sdk.NewCoin(defaultParams.TokenDenom, math.NewInt(1000))),
			expectedPass: true,
		},
		{
			name: "invalid sender",
			msg: types.NewMsgFundPool(
				sdk.AccAddress{},
				sdk.NewCoin(defaultParams.TokenDenom, math.NewInt(1000))),
			expectedPass: false,
		},
		{
			name: "invalid denom",
			msg: types.NewMsgFundPool(
				suite.TestAccs[0],
				sdk.NewCoin("invalid_denom", math.NewInt(1000))),
			expectedPass: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.msgServer.FundPool(suite.Ctx, tc.msg)
			if tc.expectedPass {
				suite.Require().NoError(err)

				pool, err := suite.App.RewardsKeeper.RewardPool.Get(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().True(pool.CommunityPool.AmountOf(defaultParams.TokenDenom).Equal((math.LegacyNewDecFromBigInt(tc.msg.Amount.Amount.BigInt()))))
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
