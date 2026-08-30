package keeper_test

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

func (suite *KeeperTestSuite) TestBeginBlocker() {
	defaultParams := types.DefaultParams()
	defaultParams.SupplyBase = math.NewInt(1_000_000_000_000) // 1e12
	// Positive floor so transfer cases stay deterministic regardless of the
	// test app's live bonded ratio (curve would otherwise clamp to 0 at/above goal).
	defaultParams.InflationMin = math.LegacyNewDecWithPrec(1, 2) // 0.01
	err := suite.App.RewardsKeeper.Params.Set(suite.Ctx, defaultParams)
	suite.Require().NoError(err)

	denom := defaultParams.TokenDenom

	testCases := []struct {
		name        string
		params      types.Params
		initialPool types.RewardPool
		// moduleBankBalance is the actual coins placed in the rewards module
		// account before BeginBlocker. Defaults to TruncateInt(CommunityPool)
		// when nil; set explicitly for bank-failure cases.
		moduleBankBalance    *math.Int
		expectTransfer       bool
		expectedChangeAmount func(bonded math.LegacyDec) sdk.Coin
	}{
		{
			name: "supply base zero - no action",
			params: func() types.Params {
				p := defaultParams
				p.SupplyBase = math.ZeroInt()
				return p
			}(),
			initialPool: types.RewardPool{
				CommunityPool: sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(100000))),
			},
			expectTransfer: false,
		},
		{
			name:   "empty pool - no action",
			params: defaultParams,
			initialPool: types.RewardPool{
				CommunityPool: sdk.DecCoins{},
			},
			expectTransfer: false,
		},
		{
			name:   "normal distribution",
			params: defaultParams,
			initialPool: types.RewardPool{
				CommunityPool: sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(100000))),
			},
			expectTransfer: true,
			expectedChangeAmount: func(bonded math.LegacyDec) sdk.Coin {
				coin, _ := types.CalculateReward(bonded, defaultParams)
				coin.Amount = math.MinInt(coin.Amount, math.NewInt(100000))
				return coin
			},
		},
		{
			name:   "accumulates total released on subsequent release",
			params: defaultParams,
			initialPool: types.RewardPool{
				CommunityPool: sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(100000))),
				TotalReleased: sdk.NewCoin(denom, math.NewInt(7)),
			},
			expectTransfer: true,
			expectedChangeAmount: func(bonded math.LegacyDec) sdk.Coin {
				coin, _ := types.CalculateReward(bonded, defaultParams)
				coin.Amount = math.MinInt(coin.Amount, math.NewInt(100000))
				return coin
			},
		},
		{
			name:   "pool runs dry - pays remaining balance",
			params: defaultParams,
			initialPool: types.RewardPool{
				CommunityPool: sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(1))),
			},
			expectTransfer: true,
			expectedChangeAmount: func(bonded math.LegacyDec) sdk.Coin {
				uncapped, _ := types.CalculateReward(bonded, defaultParams)
				suite.Require().True(uncapped.Amount.GT(math.NewInt(1)),
					"uncapped per-block reward must exceed remaining pool; got %s", uncapped.Amount)
				return sdk.NewCoin(denom, math.NewInt(1))
			},
		},
		{
			name:   "bank transfer failure - logs and skips without panic",
			params: defaultParams,
			initialPool: types.RewardPool{
				// Accounting claims far more than the module account holds
				CommunityPool: sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(1_000_000_000))),
			},
			moduleBankBalance: ptrInt(math.ZeroInt()),
			expectTransfer:    false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			ctx := suite.Ctx

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

			if !tc.expectTransfer {
				suite.Require().True(tc.initialPool.CommunityPool.Equal(pool.CommunityPool))
				suite.Require().True(totalReleasedUnchanged(tc.initialPool.TotalReleased, pool.TotalReleased))
				suite.Require().True(initialFeeCollectorBalance.Equal(
					suite.App.BankKeeper.GetBalance(ctx, feeCollectorAddr, denom),
				))
				return
			}

			bonded, err := suite.App.StakingKeeper.BondedRatio(ctx)
			suite.Require().NoError(err)
			expected := tc.expectedChangeAmount(bonded)
			suite.Require().False(expected.IsZero(), "expected transfer amount must be positive")

			suite.Require().True(tc.initialPool.CommunityPool.Sub(sdk.NewDecCoinsFromCoins(expected)).Equal(pool.CommunityPool))

			finalFeeCollectorBalance := suite.App.BankKeeper.GetBalance(ctx, feeCollectorAddr, denom)
			suite.Require().True(finalFeeCollectorBalance.Equal(initialFeeCollectorBalance.Add(expected)))

			if tc.initialPool.TotalReleased.IsNil() || tc.initialPool.TotalReleased.IsZero() {
				suite.Require().True(pool.TotalReleased.Equal(expected))
			} else {
				suite.Require().True(pool.TotalReleased.Equal(tc.initialPool.TotalReleased.Add(expected)))
			}
		})
	}
}

func ptrInt(v math.Int) *math.Int { return &v }

func totalReleasedUnchanged(before, after sdk.Coin) bool {
	beforeZero := before.IsNil() || before.IsZero()
	afterZero := after.IsNil() || after.IsZero()
	if beforeZero && afterZero {
		return true
	}
	return before.Equal(after)
}

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
		CommunityPool: sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(1000))),
		TotalReleased: sdk.NewCoin(denom, math.NewInt(5)),
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
