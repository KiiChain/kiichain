package keeper_test

import (
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v1/x/rewards/types"
)

func (suite *KeeperTestSuite) TestEndBlocker() {
	// Set up default params
	defaultParams := types.DefaultParams()
	err := suite.App.RewardsKeeper.Params.Set(suite.Ctx, defaultParams)
	suite.Require().NoError(err)

	// Fund the reward pool first
	err = suite.App.RewardsKeeper.FundCommunityPool(
		suite.Ctx,
		sdk.NewCoin(defaultParams.TokenDenom, math.NewInt(100000)),
		suite.TestAccs[0])
	suite.Require().NoError(err)

	now := time.Now()
	denom := defaultParams.TokenDenom

	testCases := []struct {
		name                 string
		initialReleaser      types.RewardReleaser
		initialPool          sdk.DecCoins
		blockTime            time.Time
		expectedChange       bool
		expectedReleaser     types.RewardReleaser
		expectedChangeAmount sdk.Coin
	}{
		{
			name: "inactive releaser - no action",
			initialReleaser: types.RewardReleaser{
				Active: false,
			},
			blockTime:      now.Add(time.Hour),
			expectedChange: false,
		},
		{
			name: "zero total amount - no action",
			initialReleaser: types.RewardReleaser{
				Active:      true,
				TotalAmount: sdk.NewCoin(denom, math.ZeroInt()),
			},
			blockTime:      now.Add(time.Hour),
			expectedChange: false,
		},
		{
			name: "first run - sets timestamp but no distribution",
			initialReleaser: types.RewardReleaser{
				Active:          true,
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.ZeroInt()),
				LastReleaseTime: time.Time{},
				EndTime:         now.Add(time.Hour * 2),
			},
			blockTime: now,
			expectedReleaser: types.RewardReleaser{
				Active:          true,
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.ZeroInt()),
				LastReleaseTime: now,
				EndTime:         now.Add(time.Hour * 2),
			},
			expectedChange:       true,
			expectedChangeAmount: sdk.NewCoin(denom, math.ZeroInt()),
		},
		{
			name: "normal distribution - partial release",
			initialReleaser: types.RewardReleaser{
				Active:          true,
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.ZeroInt()),
				LastReleaseTime: now,
				EndTime:         now.Add(time.Hour * 2),
			},
			initialPool: sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(1000))),
			blockTime:   now.Add(time.Hour),
			expectedReleaser: types.RewardReleaser{
				Active:          true,
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.NewInt(500)),
				LastReleaseTime: now.Add(time.Hour),
				EndTime:         now.Add(time.Hour * 2),
			},
			expectedChange:       true,
			expectedChangeAmount: sdk.NewCoin(denom, math.NewInt(500)),
		},
		{
			name: "final distribution",
			initialReleaser: types.RewardReleaser{
				Active:          true,
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.NewInt(900)),
				LastReleaseTime: now,
				EndTime:         now.Add(time.Hour),
			},
			initialPool: sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(100))),
			blockTime:   now.Add(time.Hour),
			expectedReleaser: types.RewardReleaser{
				Active:          true,
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.NewInt(1000)),
				LastReleaseTime: now.Add(time.Hour),
				EndTime:         now.Add(time.Hour),
			},
			expectedChange:       true,
			expectedChangeAmount: sdk.NewCoin(denom, math.NewInt(100)),
		},
		{
			name: "no more distribution - set as inactive",
			initialReleaser: types.RewardReleaser{
				Active:          true,
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.NewInt(1000)),
				LastReleaseTime: now.Add(time.Hour),
				EndTime:         now.Add(time.Hour),
			},
			initialPool: sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(100))),
			blockTime:   now.Add(time.Hour * 2),
			expectedReleaser: types.RewardReleaser{
				Active:          false,
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.NewInt(1000)),
				LastReleaseTime: now.Add(time.Hour),
				EndTime:         now.Add(time.Hour),
			},
			expectedChange:       true,
			expectedChangeAmount: sdk.NewCoin(denom, math.ZeroInt()),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Setup initial state
			ctx := suite.Ctx.WithBlockTime(tc.blockTime)

			// Set initial releaser state
			err := suite.App.RewardsKeeper.RewardReleaser.Set(ctx, tc.initialReleaser)
			suite.Require().NoError(err)

			// Set initial pool state if needed
			if !tc.initialPool.Empty() {
				err := suite.App.RewardsKeeper.RewardPool.Set(ctx, types.RewardPool{
					CommunityPool: tc.initialPool,
				})
				suite.Require().NoError(err)
			}

			// Get initial fee collector balance
			feeCollectorAddr := suite.App.AccountKeeper.GetModuleAddress("fee_collector")
			initialFeeCollectorBalance := suite.App.BankKeeper.GetBalance(ctx, feeCollectorAddr, denom)

			// Execute EndBlocker
			err = suite.App.RewardsKeeper.EndBlocker(ctx)
			suite.Require().NoError(err)

			if tc.expectedChange {
				// Verify releaser state
				releaser, err := suite.App.RewardsKeeper.RewardReleaser.Get(ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expectedReleaser.Active, releaser.Active)
				suite.Require().Equal(tc.expectedReleaser.ReleasedAmount, releaser.ReleasedAmount)
				suite.Require().Equal(tc.expectedReleaser.TotalAmount, releaser.TotalAmount)
				suite.Require().True(tc.expectedReleaser.LastReleaseTime.Equal(releaser.LastReleaseTime))
				suite.Require().True(tc.expectedReleaser.EndTime.Equal(releaser.EndTime))

				// If expecting transfer
				if !tc.expectedChangeAmount.IsZero() {
					// Verify reward pool deduction
					rewardPool, err := suite.App.RewardsKeeper.RewardPool.Get(ctx)
					suite.Require().NoError(err)
					expectedPool := tc.initialPool.Sub(sdk.NewDecCoinsFromCoins(tc.expectedChangeAmount))
					suite.Require().Equal(expectedPool, rewardPool.CommunityPool)

					// Verify fee collector balance change
					currentFeeCollectorBalance := suite.App.BankKeeper.GetBalance(ctx, feeCollectorAddr, denom)
					expectedBalance := initialFeeCollectorBalance.Add(tc.expectedChangeAmount)
					suite.Require().Equal(expectedBalance, currentFeeCollectorBalance)
				}
			}
		})
	}
}
