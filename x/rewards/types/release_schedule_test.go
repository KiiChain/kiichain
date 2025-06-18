package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v2/x/rewards/types"
)

func TestRewardReleaserValidateGenesis(t *testing.T) {
	now := time.Now()
	validCoin := sdk.NewCoin("akii", math.NewInt(1000))
	invalidCoin := sdk.Coin{Denom: "invalid!", Amount: math.NewInt(-1)}

	tests := []struct {
		name     string
		schedule types.ReleaseSchedule
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid initial state",
			schedule: types.InitialReleaseSchedule(),
			wantErr:  false,
		},
		{
			name: "valid active release",
			schedule: types.ReleaseSchedule{
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
			schedule: types.ReleaseSchedule{
				TotalAmount:     invalidCoin,
				ReleasedAmount:  sdk.Coin{},
				EndTime:         time.Time{},
				LastReleaseTime: time.Time{},
				Active:          true,
			},
			wantErr: true,
			errMsg:  "invalid total amount",
		},
		{
			name: "invalid released amount",
			schedule: types.ReleaseSchedule{
				TotalAmount:     validCoin,
				ReleasedAmount:  invalidCoin,
				EndTime:         now.Add(time.Hour * 24),
				LastReleaseTime: now,
				Active:          true,
			},
			wantErr: true,
			errMsg:  "invalid released amount",
		},
		{
			name: "denom mismatch",
			schedule: types.ReleaseSchedule{
				TotalAmount:     validCoin,
				ReleasedAmount:  sdk.NewCoin("otherdenom", math.NewInt(500)),
				EndTime:         now.Add(time.Hour * 24),
				LastReleaseTime: now,
				Active:          true,
			},
			wantErr: true,
			errMsg:  "doesn't match total amount denom",
		},
		{
			name: "released exceeds total",
			schedule: types.ReleaseSchedule{
				TotalAmount:     validCoin,
				ReleasedAmount:  sdk.NewCoin("akii", math.NewInt(2000)),
				EndTime:         now.Add(time.Hour * 24),
				LastReleaseTime: now,
				Active:          true,
			},
			wantErr: true,
			errMsg:  "cannot be greater than total amount",
		},
		{
			name: "end time in past",
			schedule: types.ReleaseSchedule{
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
			schedule: types.ReleaseSchedule{
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
			schedule: types.ReleaseSchedule{
				TotalAmount:     sdk.Coin{Denom: "akii", Amount: math.NewInt(0)},
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
			schedule: types.ReleaseSchedule{
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
			err := tt.schedule.ValidateGenesis()
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
