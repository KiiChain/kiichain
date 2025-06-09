package keeper_test

import (
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v1/x/rewards/types"
)

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
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.msgServer.UpdateParams(suite.Ctx, tc.msg)
			if tc.expectedPass {
				suite.Require().NoError(err)

				// Verify params were updated
				params, err := suite.App.RewardsKeeper.Params.Get(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(tc.msg.Params, params)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestFundPool() {
	// Set up default params
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

				// Verify funds were added to the pool
				pool, err := suite.App.RewardsKeeper.RewardPool.Get(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().True(pool.CommunityPool.AmountOf(defaultParams.TokenDenom).Equal((math.LegacyNewDecFromBigInt(tc.msg.Amount.Amount.BigInt()))))
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestExtendReward() {
	// Set up default params
	defaultParams := types.DefaultParams()
	err := suite.App.RewardsKeeper.Params.Set(suite.Ctx, defaultParams)
	suite.Require().NoError(err)

	// Fund the pool first
	fundMsg := types.NewMsgFundPool(
		suite.TestAccs[0],
		sdk.NewCoin(defaultParams.TokenDenom, math.NewInt(100000)))
	_, err = suite.msgServer.FundPool(suite.Ctx, fundMsg)
	suite.Require().NoError(err)

	validEndTime := time.Now().Add(time.Hour * 24) // 1 day in future

	testCases := []struct {
		name         string
		msg          *types.MsgExtendReward
		expectedPass bool
	}{
		{
			name: "valid extension",
			msg: types.NewMsgExtendReward(
				suite.App.RewardsKeeper.GetAuthority(),
				sdk.NewCoin(defaultParams.TokenDenom, math.NewInt(1000)),
				validEndTime,
			),
			expectedPass: true,
		},
		{
			name: "invalid authority",
			msg: types.NewMsgExtendReward(
				suite.TestAccs[0].String(),
				sdk.NewCoin(defaultParams.TokenDenom, math.NewInt(1000)),
				validEndTime,
			),
			expectedPass: false,
		},
		{
			name: "invalid denom",
			msg: types.NewMsgExtendReward(
				suite.App.RewardsKeeper.GetAuthority(),
				sdk.NewCoin("invalid_denom", math.NewInt(1000)),
				validEndTime,
			),
			expectedPass: false,
		},
		{
			name: "invalid time - past",
			msg: types.NewMsgExtendReward(
				suite.App.RewardsKeeper.GetAuthority(),
				sdk.NewCoin(defaultParams.TokenDenom, math.NewInt(1000)),
				time.Now().Add(-time.Hour), // Past time
			),
			expectedPass: false,
		},
		{
			name: "insufficient funds",
			msg: types.NewMsgExtendReward(
				suite.App.RewardsKeeper.GetAuthority(),
				sdk.NewCoin(defaultParams.TokenDenom, math.NewInt(100000000)), // More than in pool
				validEndTime,
			),
			expectedPass: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			_, err := suite.msgServer.ExtendReward(suite.Ctx, tc.msg)
			if tc.expectedPass {
				suite.Require().NoError(err)

				// Verify releaser was updated
				releaser, err := suite.App.RewardsKeeper.RewardReleaser.Get(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().True(releaser.Active)
				suite.Require().Equal(tc.msg.ExtraAmount, releaser.TotalAmount.Sub(releaser.ReleasedAmount))
				suite.Require().True(tc.msg.EndTime.Equal(releaser.EndTime))
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
