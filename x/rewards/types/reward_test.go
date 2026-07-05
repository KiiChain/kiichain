package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

// TestCalculateReward verifies the CalculateReward function across a comprehensive
// set of scenarios: fully-exhausted schedules, linear proportional releases at
// various elapsed fractions, partial prior releases, nanosecond-precision
// sub-second durations (regression for issue #267), and the edge case where
// EndTime equals LastReleaseTime, which must release the full remaining amount
// immediately rather than returning an error.
func TestCalculateReward(t *testing.T) {
	now := time.Now()
	denom := "akii"

	tests := []struct {
		name          string
		blockTime     time.Time
		schedule      types.ReleaseSchedule
		expectedCoin  sdk.Coin
		expectedError bool
	}{
		{
			name:      "nothing left to release",
			blockTime: now.Add(time.Hour),
			schedule: types.ReleaseSchedule{
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.NewInt(1000)),
				LastReleaseTime: now,
				EndTime:         now.Add(time.Hour * 2),
				Active:          true,
			},
			expectedCoin:  sdk.NewCoin(denom, math.ZeroInt()),
			expectedError: false,
		},
		{
			name:      "linear release - halfway",
			blockTime: now.Add(time.Hour),
			schedule: types.ReleaseSchedule{
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.ZeroInt()),
				LastReleaseTime: now,
				EndTime:         now.Add(time.Hour * 2),
				Active:          true,
			},
			expectedCoin:  sdk.NewCoin(denom, math.NewInt(500)),
			expectedError: false,
		},
		{
			name:      "linear release - one third",
			blockTime: now.Add(20 * time.Minute),
			schedule: types.ReleaseSchedule{
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(900)),
				ReleasedAmount:  sdk.NewCoin(denom, math.ZeroInt()),
				LastReleaseTime: now,
				EndTime:         now.Add(time.Hour),
				Active:          true,
			},
			expectedCoin:  sdk.NewCoin(denom, math.NewInt(300)),
			expectedError: false,
		},
		{
			name:      "partial release with existing released amount",
			blockTime: now.Add(30 * time.Minute),
			schedule: types.ReleaseSchedule{
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.NewInt(200)),
				LastReleaseTime: now,
				EndTime:         now.Add(time.Hour),
				Active:          true,
			},
			expectedCoin:  sdk.NewCoin(denom, math.NewInt(400)), // (1000-200) * (30/60)
			expectedError: false,
		},
		{
			name:      "normal case but with small fraction",
			blockTime: now.Add(3 * time.Second),
			schedule: types.ReleaseSchedule{
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(3000000000000000000)), // 3 kii
				ReleasedAmount:  sdk.NewCoin(denom, math.NewInt(2000000000000000000)),
				LastReleaseTime: now,
				EndTime:         now.Add(time.Hour * 24 * 365), // one year
				Active:          true,
			},
			// 1 kii / 10512000 (365*24*60*20) ≃ 1*10^11
			expectedCoin:  sdk.NewCoin(denom, math.NewInt(95129375951)),
			expectedError: false,
		},
		{
			name:      "last release (past end time)",
			blockTime: now.Add(time.Hour * 2),
			schedule: types.ReleaseSchedule{
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.NewInt(800)),
				LastReleaseTime: now,
				EndTime:         now.Add(time.Hour),
				Active:          true,
			},
			expectedCoin:  sdk.NewCoin(denom, math.NewInt(200)), // Cap at remaining 200
			expectedError: false,
		},
		{
			name:      "end time equal to last release releases full remaining amount",
			blockTime: now.Add(time.Hour * 2),
			schedule: types.ReleaseSchedule{
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.NewInt(800)),
				LastReleaseTime: now,
				EndTime:         now,
				Active:          true,
			},
			expectedCoin:  sdk.NewCoin(denom, math.NewInt(200)), // full remaining released
			expectedError: false,
		},
		{
			name:      "sub-second duration does not divide by zero",
			blockTime: now.Add(250 * time.Millisecond),
			schedule: types.ReleaseSchedule{
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.ZeroInt()),
				LastReleaseTime: now,
				EndTime:         now.Add(500 * time.Millisecond),
				Active:          true,
			},
			expectedCoin:  sdk.NewCoin(denom, math.NewInt(500)), // 250ms / 500ms = 50%
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := types.CalculateReward(tt.blockTime, tt.schedule)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				// Denom should never differ
				require.Equal(t, tt.expectedCoin.Denom, result.Denom)
				// Check diff within 1%
				diff := tt.expectedCoin.Amount.Sub(result.Amount)
				tolerance := tt.expectedCoin.Amount.Quo(math.NewInt(100))
				require.True(t, diff.LTE(tolerance))
			}
		})
	}
}

func TestCalculateRewardNoForcedMinimum(t *testing.T) {
	now := time.Now()
	denom := "akii"

	schedule := types.ReleaseSchedule{
		TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000000)),
		ReleasedAmount:  sdk.NewCoin(denom, math.ZeroInt()),
		LastReleaseTime: now,
		EndTime:         now.Add(time.Hour * 24 * 365 * 10),
		Active:          true,
	}

	result, err := types.CalculateReward(now.Add(time.Second), schedule)
	require.NoError(t, err)
	require.True(t, result.Amount.IsZero(), "expected zero release, got %s", result.Amount)
}
