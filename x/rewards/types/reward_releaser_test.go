package types_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/kiichain/kiichain/v1/x/rewards/types"
)

func TestRewardReleaserValidateGenesis(t *testing.T) {
	now := time.Now()
	validCoin := sdk.NewCoin("akii", math.NewInt(1000))
	invalidCoin := sdk.Coin{Denom: "invalid!", Amount: math.NewInt(-1)}

	tests := []struct {
		name     string
		releaser types.RewardReleaser
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid initial state",
			releaser: types.InitialRewardReleaser(),
			wantErr:  false,
		},
		{
			name: "valid active releaser",
			releaser: types.RewardReleaser{
				TotalAmount:     validCoin,
				ReleasedAmount:  sdk.NewCoin("akii", math.NewInt(500)),
				EndTime:         now.Add(time.Hour * 24),
				LastReleaseTime: now,
				Active:          true,
			},
			wantErr: false,
		},
		{
			name: "invalid total amount",
			releaser: types.RewardReleaser{
				TotalAmount:     invalidCoin,
				ReleasedAmount:  sdk.Coin{},
				EndTime:         time.Time{},
				LastReleaseTime: time.Time{},
				Active:          false,
			},
			wantErr: true,
			errMsg:  "invalid total amount",
		},
		{
			name: "invalid released amount",
			releaser: types.RewardReleaser{
				TotalAmount:     validCoin,
				ReleasedAmount:  invalidCoin,
				EndTime:         time.Time{},
				LastReleaseTime: time.Time{},
				Active:          false,
			},
			wantErr: true,
			errMsg:  "invalid released amount",
		},
		{
			name: "denom mismatch",
			releaser: types.RewardReleaser{
				TotalAmount:     validCoin,
				ReleasedAmount:  sdk.NewCoin("otherdenom", math.NewInt(500)),
				EndTime:         time.Time{},
				LastReleaseTime: time.Time{},
				Active:          false,
			},
			wantErr: true,
			errMsg:  "doesn't match total amount denom",
		},
		{
			name: "released exceeds total",
			releaser: types.RewardReleaser{
				TotalAmount:     validCoin,
				ReleasedAmount:  sdk.NewCoin("akii", math.NewInt(2000)),
				EndTime:         time.Time{},
				LastReleaseTime: time.Time{},
				Active:          false,
			},
			wantErr: true,
			errMsg:  "cannot be greater than total amount",
		},
		{
			name: "end time in past",
			releaser: types.RewardReleaser{
				TotalAmount:     validCoin,
				ReleasedAmount:  sdk.Coin{},
				EndTime:         now.Add(-time.Hour * 24),
				LastReleaseTime: time.Time{},
				Active:          false,
			},
			wantErr: true,
			errMsg:  "cannot be in the past",
		},
		{
			name: "last release in future",
			releaser: types.RewardReleaser{
				TotalAmount:     validCoin,
				ReleasedAmount:  sdk.Coin{},
				EndTime:         time.Time{},
				LastReleaseTime: now.Add(time.Hour * 24),
				Active:          false,
			},
			wantErr: true,
			errMsg:  "cannot be in the future",
		},
		{
			name: "active with zero total",
			releaser: types.RewardReleaser{
				TotalAmount:     sdk.Coin{},
				ReleasedAmount:  sdk.Coin{},
				EndTime:         now.Add(time.Hour * 24),
				LastReleaseTime: time.Time{},
				Active:          true,
			},
			wantErr: true,
			errMsg:  "cannot have zero total amount",
		},
		{
			name: "active with zero end time",
			releaser: types.RewardReleaser{
				TotalAmount:     validCoin,
				ReleasedAmount:  sdk.Coin{},
				EndTime:         time.Time{},
				LastReleaseTime: time.Time{},
				Active:          true,
			},
			wantErr: true,
			errMsg:  "must have an end time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.releaser.ValidateGenesis()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
