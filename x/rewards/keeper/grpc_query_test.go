package keeper_test

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/x/rewards/keeper"
	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

func (suite *KeeperTestSuite) TestQuerierParams() {
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
			},
			expectedPass: true,
		},
		{
			name: "success - with modified params",
			setup: func() {
				modifiedParams := types.DefaultParams()
				modifiedParams.SupplyBase = math.NewInt(12345)
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
			},
			expectedPass: true,
		},
		{
			name: "success - with funds",
			setup: func() {
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

				expectedPool, err := suite.App.RewardsKeeper.RewardPool.Get(suite.Ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(expectedPool, res.RewardPool)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
