package keeper_test

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v2/x/rewards/keeper"
	"github.com/kiichain/kiichain/v2/x/rewards/types"
)

func (suite *KeeperTestSuite) TestQuerierParams() {
	// Set up default params
	defaultParams := types.DefaultParams()
	err := suite.App.RewardsKeeper.Params.Set(suite.Ctx, defaultParams)
	suite.Require().NoError(err)

	querier := keeper.NewQuerier(suite.App.RewardsKeeper)

	testCases := []struct {
		name         string
		setup        func()
		expectedPass bool
	}{
		{
			name: "success - get default params",
			setup: func() {
				// Already set up
			},
			expectedPass: true,
		},
		{
			name: "success - with modified params",
			setup: func() {
				modifiedParams := types.Params{
					TokenDenom: "modified",
				}
				err := suite.App.RewardsKeeper.Params.Set(suite.Ctx, modifiedParams)
				suite.Require().NoError(err)
			},
			expectedPass: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.setup()

			res, err := querier.Params(suite.Ctx, &types.QueryParamsRequest{})
			if tc.expectedPass {
				suite.Require().NoError(err)

				// Verify returned params match what we expect
				expectedParams, err := suite.App.RewardsKeeper.Params.Get(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(expectedParams, res.Params)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQuerierRewardPool() {
	// Initialize an empty pool
	err := suite.App.RewardsKeeper.RewardPool.Set(suite.Ctx, types.RewardPool{})
	suite.Require().NoError(err)

	querier := keeper.NewQuerier(suite.App.RewardsKeeper)

	testCases := []struct {
		name         string
		setup        func()
		expectedPass bool
	}{
		{
			name: "success - empty pool",
			setup: func() {
				// Already set up empty pool
			},
			expectedPass: true,
		},
		{
			name: "success - with funds",
			setup: func() {
				// Fund the pool
				fundMsg := types.NewMsgFundPool(
					suite.TestAccs[0],
					sdk.NewCoin("akii", math.NewInt(100000)))
				_, err := suite.msgServer.FundPool(suite.Ctx, fundMsg)
				suite.Require().NoError(err)
			},
			expectedPass: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.setup()

			res, err := querier.RewardPool(suite.Ctx, &types.QueryRewardPoolRequest{})
			if tc.expectedPass {
				suite.Require().NoError(err)

				// Verify returned pool matches what we expect
				expectedPool, err := suite.App.RewardsKeeper.RewardPool.Get(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(expectedPool, res.RewardPool)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQuerierReleaseSchedule() {
	// Initialize an empty schedule
	err := suite.App.RewardsKeeper.ReleaseSchedule.Set(suite.Ctx, types.ReleaseSchedule{})
	suite.Require().NoError(err)

	querier := keeper.NewQuerier(suite.App.RewardsKeeper)

	testCases := []struct {
		name         string
		setup        func()
		expectedPass bool
	}{
		{
			name: "success - empty schedule",
			setup: func() {
				// Already set up empty schedule
			},
			expectedPass: true,
		},
		{
			name: "success - with active schedule",
			setup: func() {
				// Set up an active schedule
				schedule := types.ReleaseSchedule{
					TotalAmount:     sdk.NewCoin("akii", math.NewInt(10000)),
					ReleasedAmount:  sdk.NewCoin("akii", math.NewInt(2000)),
					EndTime:         suite.Ctx.BlockTime().AddDate(0, 0, 7), // 1 week from now
					LastReleaseTime: suite.Ctx.BlockTime(),
					Active:          true,
				}
				err := suite.App.RewardsKeeper.ReleaseSchedule.Set(suite.Ctx, schedule)
				suite.Require().NoError(err)
			},
			expectedPass: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.setup()

			res, err := querier.ReleaseSchedule(suite.Ctx, &types.QueryReleaseScheduleRequest{})
			if tc.expectedPass {
				suite.Require().NoError(err)

				// Verify returned schedule matches what we expect
				expectedSchedule, err := suite.App.RewardsKeeper.ReleaseSchedule.Get(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(expectedSchedule, res.ReleaseSchedule)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
