package keeper_test

import (
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v3/x/rewards/types"
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
		initialSchedule      types.ReleaseSchedule
		initialPool          sdk.DecCoins
		blockTime            time.Time
		expectedChange       bool
		expectedSchedule     types.ReleaseSchedule
		expectedChangeAmount sdk.Coin
	}{
		{
			name: "inactive schedule - no action",
			initialSchedule: types.ReleaseSchedule{
				Active: false,
			},
			blockTime:      now.Add(time.Hour),
			expectedChange: false,
		},
		{
			name: "zero total amount - no action",
			initialSchedule: types.ReleaseSchedule{
				Active:      true,
				TotalAmount: sdk.NewCoin(denom, math.ZeroInt()),
			},
			blockTime:      now.Add(time.Hour),
			expectedChange: false,
		},
		{
			name: "first run - sets timestamp but no distribution",
			initialSchedule: types.ReleaseSchedule{
				Active:          true,
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.ZeroInt()),
				LastReleaseTime: time.Time{},
				EndTime:         now.Add(time.Hour * 2),
			},
			blockTime: now,
			expectedSchedule: types.ReleaseSchedule{
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
			initialSchedule: types.ReleaseSchedule{
				Active:          true,
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.ZeroInt()),
				LastReleaseTime: now,
				EndTime:         now.Add(time.Hour * 2),
			},
			initialPool: sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(1000))),
			blockTime:   now.Add(time.Hour),
			expectedSchedule: types.ReleaseSchedule{
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
			initialSchedule: types.ReleaseSchedule{
				Active:          true,
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.NewInt(900)),
				LastReleaseTime: now,
				EndTime:         now.Add(time.Hour),
			},
			initialPool: sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(100))),
			blockTime:   now.Add(time.Hour),
			expectedSchedule: types.ReleaseSchedule{
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
			initialSchedule: types.ReleaseSchedule{
				Active:          true,
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.NewInt(1000)),
				LastReleaseTime: now.Add(time.Hour),
				EndTime:         now.Add(time.Hour),
			},
			initialPool: sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(100))),
			blockTime:   now.Add(time.Hour * 2),
			expectedSchedule: types.ReleaseSchedule{
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

			// Set initial schedule state
			err := suite.App.RewardsKeeper.ReleaseSchedule.Set(ctx, tc.initialSchedule)
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

			// Execute BeginBlocker
			err = suite.App.RewardsKeeper.BeginBlocker(ctx)
			suite.Require().NoError(err)

			if tc.expectedChange {
				// Verify schedule state
				schedule, err := suite.App.RewardsKeeper.ReleaseSchedule.Get(ctx)
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expectedSchedule.Active, schedule.Active)
				suite.Require().Equal(tc.expectedSchedule.ReleasedAmount, schedule.ReleasedAmount)
				suite.Require().Equal(tc.expectedSchedule.TotalAmount, schedule.TotalAmount)
				suite.Require().True(tc.expectedSchedule.LastReleaseTime.Equal(schedule.LastReleaseTime))
				suite.Require().True(tc.expectedSchedule.EndTime.Equal(schedule.EndTime))

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
