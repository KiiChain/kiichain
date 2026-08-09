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

	now := time.Now().UTC()
	denom := defaultParams.TokenDenom

	testCases := []struct {
		name        string
		params      types.Params
		initialPool types.RewardPool
		// moduleBankBalance is the actual coins placed in the rewards module
		// account before BeginBlocker. Defaults to TruncateInt(CommunityPool)
		// when nil; set explicitly for bank-failure cases.
		moduleBankBalance    *math.Int
		blockTime            time.Time
		expectTransfer       bool
		expectedChangeAmount func(bonded math.LegacyDec) sdk.Coin
		expectLastReleaseSet bool
		expectLastReleaseAdv bool
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
			name:   "empty pool with prior release - advances last release time",
			params: defaultParams,
			initialPool: types.RewardPool{
				CommunityPool:   sdk.DecCoins{},
				LastReleaseTime: now,
			},
			blockTime:            now.Add(time.Hour),
			expectTransfer:       false,
			expectLastReleaseAdv: true,
		},
		{
			name:   "empty pool with zero last release - no action",
			params: defaultParams,
			initialPool: types.RewardPool{
				CommunityPool:   sdk.DecCoins{},
				LastReleaseTime: time.Time{},
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
			name:   "sub-unit release - zero amount skip",
			params: defaultParams,
			initialPool: types.RewardPool{
				CommunityPool:   sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(100000))),
				LastReleaseTime: now,
			},
			blockTime:      now.Add(time.Nanosecond),
			expectTransfer: false,
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
				coin, _ := types.CalculateReward(now.Add(time.Hour), now, bonded, defaultParams)
				coin.Amount = math.MinInt(coin.Amount, math.NewInt(100000))
				return coin
			},
		},
		{
			name:   "accumulates total released on subsequent release",
			params: defaultParams,
			initialPool: types.RewardPool{
				CommunityPool:   sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(100000))),
				LastReleaseTime: now,
				TotalReleased:   sdk.NewCoin(denom, math.NewInt(7)),
			},
			blockTime:      now.Add(time.Hour),
			expectTransfer: true,
			expectedChangeAmount: func(bonded math.LegacyDec) sdk.Coin {
				coin, _ := types.CalculateReward(now.Add(time.Hour), now, bonded, defaultParams)
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
		{
			name:   "bank transfer failure - logs and skips without panic",
			params: defaultParams,
			initialPool: types.RewardPool{
				// Accounting claims far more than the module account holds
				CommunityPool:   sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(1_000_000_000))),
				LastReleaseTime: now,
			},
			moduleBankBalance:    ptrInt(math.ZeroInt()),
			blockTime:            now.Add(365 * 24 * time.Hour),
			expectTransfer:       false,
			expectLastReleaseAdv: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			ctx := suite.Ctx.WithBlockTime(tc.blockTime)

			err := suite.App.RewardsKeeper.Params.Set(ctx, tc.params)
			suite.Require().NoError(err)

			wantBank := tc.initialPool.CommunityPool.AmountOf(denom).TruncateInt()
			if tc.moduleBankBalance != nil {
				wantBank = *tc.moduleBankBalance
			}
			suite.resetRewardsModuleBalance(ctx, denom, wantBank)

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

			if tc.expectLastReleaseAdv {
				suite.Require().True(tc.blockTime.Equal(pool.LastReleaseTime))
				suite.Require().True(tc.initialPool.CommunityPool.Equal(pool.CommunityPool))
				return
			}

			if !tc.expectTransfer {
				suite.Require().True(tc.initialPool.CommunityPool.Equal(pool.CommunityPool))
				suite.Require().True(tc.initialPool.LastReleaseTime.Equal(pool.LastReleaseTime))
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

			if !tc.initialPool.TotalReleased.IsNil() && !tc.initialPool.TotalReleased.IsZero() {
				suite.Require().Equal(tc.initialPool.TotalReleased.Add(expected), pool.TotalReleased)
			} else {
				suite.Require().Equal(expected, pool.TotalReleased)
			}

			currentFeeCollectorBalance := suite.App.BankKeeper.GetBalance(ctx, feeCollectorAddr, denom)
			suite.Require().Equal(initialFeeCollectorBalance.Add(expected), currentFeeCollectorBalance)
		})
	}
}

func ptrInt(v math.Int) *math.Int { return &v }

// resetRewardsModuleBalance drains the rewards module account, then funds it to want.
func (suite *KeeperTestSuite) resetRewardsModuleBalance(ctx sdk.Context, denom string, want math.Int) {
	moduleAddr := suite.App.AccountKeeper.GetModuleAddress(types.ModuleName)
	cur := suite.App.BankKeeper.GetBalance(ctx, moduleAddr, denom)
	if cur.IsPositive() {
		err := suite.App.BankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, suite.TestAccs[0], sdk.NewCoins(cur))
		suite.Require().NoError(err)
	}
	if !want.IsPositive() {
		return
	}
	coin := sdk.NewCoin(denom, want)
	suite.FundAcc(suite.TestAccs[0], sdk.NewCoins(coin))
	err := suite.App.BankKeeper.SendCoinsFromAccountToModule(ctx, suite.TestAccs[0], types.ModuleName, sdk.NewCoins(coin))
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) TestGenesisInitExport() {
	denom := types.DefaultParams().TokenDenom
	params := types.DefaultParams()
	params.SupplyBase = math.NewInt(42)

	pool := types.RewardPool{
		CommunityPool:   sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(1000))),
		LastReleaseTime: time.Now().UTC(),
		TotalReleased:   sdk.NewCoin(denom, math.NewInt(5)),
	}

	// Fund the module account so InitGenesis accounting check is quiet
	err := suite.App.RewardsKeeper.FundCommunityPool(suite.Ctx, sdk.NewCoin(denom, math.NewInt(1000)), suite.TestAccs[0])
	suite.Require().NoError(err)

	suite.App.RewardsKeeper.InitGenesis(suite.Ctx, types.GenesisState{
		Params:     params,
		RewardPool: pool,
	})

	exported := suite.App.RewardsKeeper.ExportGenesis(suite.Ctx)
	suite.Require().Equal(params, exported.Params)
	suite.Require().True(pool.CommunityPool.Equal(exported.RewardPool.CommunityPool))
	suite.Require().True(pool.LastReleaseTime.Equal(exported.RewardPool.LastReleaseTime))
	suite.Require().Equal(pool.TotalReleased, exported.RewardPool.TotalReleased)

	// Accounting mismatch is logged, not fatal
	badPool := pool
	badPool.CommunityPool = badPool.CommunityPool.Add(sdk.NewDecCoin(denom, math.NewInt(1)))
	suite.App.RewardsKeeper.InitGenesis(suite.Ctx, types.GenesisState{
		Params:     params,
		RewardPool: badPool,
	})
	suite.Require().Error(suite.App.RewardsKeeper.ValidateModuleAccounting(suite.Ctx))
}
