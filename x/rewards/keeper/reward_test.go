package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v1/x/rewards/keeper"
	"github.com/kiichain/kiichain/v1/x/rewards/types"
)

func TestCalculateReward(t *testing.T) {
	now := time.Now()
	denom := "akii"

	tests := []struct {
		name          string
		blockTime     time.Time
		releaser      types.RewardReleaser
		expectedCoin  sdk.Coin
		expectedError bool
	}{
		{
			name:      "nothing left to release",
			blockTime: now.Add(time.Hour),
			releaser: types.RewardReleaser{
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
			releaser: types.RewardReleaser{
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
			releaser: types.RewardReleaser{
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
			releaser: types.RewardReleaser{
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
			releaser: types.RewardReleaser{
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
			releaser: types.RewardReleaser{
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
			name:      "zero time duration",
			blockTime: now,
			releaser: types.RewardReleaser{
				TotalAmount:     sdk.NewCoin(denom, math.NewInt(1000)),
				ReleasedAmount:  sdk.NewCoin(denom, math.ZeroInt()),
				LastReleaseTime: now,
				EndTime:         now,
				Active:          true,
			},
			expectedCoin:  sdk.NewCoin(denom, math.ZeroInt()),
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := keeper.CalculateReward(tt.blockTime, tt.releaser)

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
