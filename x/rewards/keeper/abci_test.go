package keeper_test

import (
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

func (suite *KeeperTestSuite) TestBeginBlocker() {
	defaultParams := types.DefaultParams()
	defaultParams.SupplyBase = math.NewInt(1_000_000_000_000) // 1e12
	err := suite.App.RewardsKeeper.Params.Set(suite.Ctx, defaultParams)
	suite.Require().NoError(err)

	err = suite.App.RewardsKeeper.FundCommunityPool(
		suite.Ctx,
		sdk.NewCoin(defaultParams.TokenDenom, math.NewInt(100000)),
		suite.TestAccs[0])
	suite.Require().NoError(err)

	now := time.Now().UTC()
	denom := defaultParams.TokenDenom

	testCases := []struct {
		name                 string
		params               types.Params
		initialPool          types.RewardPool
		blockTime            time.Time
		expectTransfer       bool
		expectedChangeAmount func(bonded math.LegacyDec) sdk.Coin
		expectLastReleaseSet bool
	}{
		{
			name: "supply base zero - no action",
			params: func() types.Params {
				p := defaultParams
				p.SupplyBase = math.ZeroInt()
				return p
			}(),
			initialPool: types.RewardPool{
				CommunityPool:   sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(100000))),
				LastReleaseTime: now,
			},
			blockTime:      now.Add(time.Hour),
			expectTransfer: false,
		},
		{
			name:   "empty pool - no action",
			params: defaultParams,
			initialPool: types.RewardPool{
				CommunityPool:   sdk.DecCoins{},
				LastReleaseTime: now,
			},
			blockTime:      now.Add(time.Hour),
			expectTransfer: false,
		},
		{
			name:   "first run - sets timestamp but no distribution",
			params: defaultParams,
			initialPool: types.RewardPool{
				CommunityPool:   sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(100000))),
				LastReleaseTime: time.Time{},
			},
			blockTime:            now,
			expectTransfer:       false,
			expectLastReleaseSet: true,
		},
		{
			name:   "normal distribution",
			params: defaultParams,
			initialPool: types.RewardPool{
				CommunityPool:   sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(100000))),
				LastReleaseTime: now,
			},
			blockTime:      now.Add(time.Hour),
			expectTransfer: true,
			expectedChangeAmount: func(bonded math.LegacyDec) sdk.Coin {
				coin, _, err := types.CalculateReward(now.Add(time.Hour), now, bonded, defaultParams)
				suite.Require().NoError(err)
				coin.Amount = math.MinInt(coin.Amount, math.NewInt(100000))
				return coin
			},
		},
		{
			name:   "pool runs dry - pays remaining balance",
			params: defaultParams,
			initialPool: types.RewardPool{
				CommunityPool:   sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(10))),
				LastReleaseTime: now,
			},
			blockTime:      now.Add(365 * 24 * time.Hour),
			expectTransfer: true,
			expectedChangeAmount: func(bonded math.LegacyDec) sdk.Coin {
				return sdk.NewCoin(denom, math.NewInt(10))
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			ctx := suite.Ctx.WithBlockTime(tc.blockTime)

			err := suite.App.RewardsKeeper.Params.Set(ctx, tc.params)
			suite.Require().NoError(err)

			err = suite.App.RewardsKeeper.RewardPool.Set(ctx, tc.initialPool)
			suite.Require().NoError(err)

			feeCollectorAddr := suite.App.AccountKeeper.GetModuleAddress("fee_collector")
			initialFeeCollectorBalance := suite.App.BankKeeper.GetBalance(ctx, feeCollectorAddr, denom)

			err = suite.App.RewardsKeeper.BeginBlocker(ctx)
			suite.Require().NoError(err)

			pool, err := suite.App.RewardsKeeper.RewardPool.Get(ctx)
			suite.Require().NoError(err)

			if tc.expectLastReleaseSet {
				suite.Require().True(tc.blockTime.Equal(pool.LastReleaseTime))
				suite.Require().Equal(tc.initialPool.CommunityPool, pool.CommunityPool)
				return
			}

			if !tc.expectTransfer {
				suite.Require().True(tc.initialPool.CommunityPool.Equal(pool.CommunityPool))
				return
			}

			bonded, err := suite.App.StakingKeeper.BondedRatio(ctx)
			suite.Require().NoError(err)
			expected := tc.expectedChangeAmount(bonded)
			if expected.IsZero() {
				suite.T().Skip("calculated amount is zero for current bonded ratio; skip transfer assertion")
			}

			expectedPool := tc.initialPool.CommunityPool.Sub(sdk.NewDecCoinsFromCoins(expected))
			suite.Require().Equal(expectedPool, pool.CommunityPool)
			suite.Require().True(tc.blockTime.Equal(pool.LastReleaseTime))
			suite.Require().Equal(expected, pool.TotalReleased)

			currentFeeCollectorBalance := suite.App.BankKeeper.GetBalance(ctx, feeCollectorAddr, denom)
			suite.Require().Equal(initialFeeCollectorBalance.Add(expected), currentFeeCollectorBalance)
		})
	}
}
