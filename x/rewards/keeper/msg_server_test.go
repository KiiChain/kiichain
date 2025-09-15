package keeper_test

import (
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v4/x/rewards/types"
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

// TestFundPool tests funding the pool
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

// TestChangeSchedule tests changes to the release schedule
func (suite *KeeperTestSuite) TestChangeSchedule() {
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

	// Get module authority
	authority := suite.App.RewardsKeeper.GetAuthority()

	// Valid base schedule
	validEndTime := time.Now().Add(time.Hour * 24)
	validSchedule := types.ReleaseSchedule{
		TotalAmount:     sdk.NewCoin(defaultParams.TokenDenom, math.NewInt(50000)),
		ReleasedAmount:  sdk.NewCoin(defaultParams.TokenDenom, math.NewInt(0)),
		EndTime:         validEndTime,
		LastReleaseTime: time.Time{},
		Active:          true,
	}

	testCases := []struct {
		name           string
		authority      string
		modifySchedule func(types.ReleaseSchedule) types.ReleaseSchedule
		expectedPass   bool
	}{
		{
			name:      "valid schedule",
			authority: authority,
			modifySchedule: func(s types.ReleaseSchedule) types.ReleaseSchedule {
				return s // unchanged valid schedule
			},
			expectedPass: true,
		},
		{
			name:      "invalid authority",
			authority: suite.TestAccs[0].String(),
			modifySchedule: func(s types.ReleaseSchedule) types.ReleaseSchedule {
				return s // authority is checked before schedule validation
			},
			expectedPass: false,
		},
		{
			name:      "invalid denom",
			authority: authority,
			modifySchedule: func(s types.ReleaseSchedule) types.ReleaseSchedule {
				s.TotalAmount.Denom = "invalid"
				s.ReleasedAmount.Denom = "invalid"
				return s
			},
			expectedPass: false,
		},
		{
			name:      "zero total amount",
			authority: authority,
			modifySchedule: func(s types.ReleaseSchedule) types.ReleaseSchedule {
				s.TotalAmount.Amount = math.NewInt(0)
				return s
			},
			expectedPass: false,
		},
		{
			name:      "negative amount",
			authority: authority,
			modifySchedule: func(s types.ReleaseSchedule) types.ReleaseSchedule {
				s.TotalAmount.Amount = math.NewInt(-100)
				return s
			},
			expectedPass: false,
		},
		{
			name:      "released exceeds total",
			authority: authority,
			modifySchedule: func(s types.ReleaseSchedule) types.ReleaseSchedule {
				s.ReleasedAmount.Amount = math.NewInt(60000)
				return s
			},
			expectedPass: false,
		},
		{
			name:      "denom mismatch",
			authority: authority,
			modifySchedule: func(s types.ReleaseSchedule) types.ReleaseSchedule {
				s.ReleasedAmount.Denom = "otherdenom"
				return s
			},
			expectedPass: false,
		},
		{
			name:      "end time in past",
			authority: authority,
			modifySchedule: func(s types.ReleaseSchedule) types.ReleaseSchedule {
				s.EndTime = time.Now().Add(-time.Hour)
				return s
			},
			expectedPass: false,
		},
		{
			name:      "last release in future",
			authority: authority,
			modifySchedule: func(s types.ReleaseSchedule) types.ReleaseSchedule {
				s.LastReleaseTime = time.Now().Add(time.Hour)
				return s
			},
			expectedPass: false,
		},
		{
			name:      "last release after end time",
			authority: authority,
			modifySchedule: func(s types.ReleaseSchedule) types.ReleaseSchedule {
				s.LastReleaseTime = s.EndTime.Add(time.Hour)
				return s
			},
			expectedPass: false,
		},
		{
			name:      "active with zero end time",
			authority: authority,
			modifySchedule: func(s types.ReleaseSchedule) types.ReleaseSchedule {
				s.EndTime = time.Time{}
				return s
			},
			expectedPass: false,
		},
		{
			name:      "insufficient funds",
			authority: authority,
			modifySchedule: func(s types.ReleaseSchedule) types.ReleaseSchedule {
				s.TotalAmount.Amount = math.NewInt(200000) // More than we funded
				return s
			},
			expectedPass: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Create modified schedule
			modifiedSchedule := tc.modifySchedule(validSchedule)

			// Create message
			msg := &types.MsgChangeSchedule{
				Authority: tc.authority,
				Schedule:  modifiedSchedule,
			}

			_, err := suite.msgServer.ChangeSchedule(suite.Ctx, msg)
			if tc.expectedPass {
				suite.Require().NoError(err)

				// Verify schedule was updated
				storedSchedule, err := suite.App.RewardsKeeper.ReleaseSchedule.Get(suite.Ctx)
				suite.Require().NoError(err)
				// Check individually cause times can be utc vs local but same stamp
				suite.Require().Equal(modifiedSchedule.Active, storedSchedule.Active)
				suite.Require().Equal(modifiedSchedule.ReleasedAmount, storedSchedule.ReleasedAmount)
				suite.Require().Equal(modifiedSchedule.TotalAmount, storedSchedule.TotalAmount)
				suite.Require().True(modifiedSchedule.LastReleaseTime.Equal(storedSchedule.LastReleaseTime))
				suite.Require().True(modifiedSchedule.EndTime.Equal(storedSchedule.EndTime))
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
