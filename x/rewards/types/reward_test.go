package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

func TestCalculateInflation(t *testing.T) {
	params := types.DefaultParams()

	tests := []struct {
		name        string
		bondedRatio math.LegacyDec
		want        math.LegacyDec
	}{
		{
			name:        "zero bonded ratio -> inflation min",
			bondedRatio: math.LegacyZeroDec(),
			want:        params.InflationMin,
		},
		{
			name:        "at goal bonded -> zero before clamp, returns min",
			bondedRatio: params.GoalBonded,
			want:        params.InflationMin,
		},
		{
			name:        "above goal bonded -> negative raw, clamped to min",
			bondedRatio: math.LegacyNewDecWithPrec(80, 2),
			want:        params.InflationMin,
		},
		{
			name:        "peak at goalBonded/2",
			bondedRatio: params.GoalBonded.QuoInt64(2),
			// peak = rateChange * goalBonded / 4 = 0.13 * 0.67 / 4
			want: types.InflationRateChange.Mul(params.GoalBonded).QuoInt64(4),
		},
		{
			name:        "mid curve below peak",
			bondedRatio: math.LegacyNewDecWithPrec(20, 2),
			want: math.LegacyOneDec().
				Sub(math.LegacyNewDecWithPrec(20, 2).Quo(params.GoalBonded)).
				Mul(types.InflationRateChange).
				Mul(math.LegacyNewDecWithPrec(20, 2)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := types.CalculateInflation(tc.bondedRatio, params)
			require.True(t, tc.want.Equal(got), "want %s got %s", tc.want, got)
		})
	}
}

func TestCalculateInflationClampsToMax(t *testing.T) {
	params := types.DefaultParams()
	params.InflationMax = math.LegacyNewDecWithPrec(1, 2) // 0.01, below curve peak

	bonded := params.GoalBonded.QuoInt64(2)
	got := types.CalculateInflation(bonded, params)
	require.True(t, params.InflationMax.Equal(got))
}

func TestCalculateReward(t *testing.T) {
	params := types.DefaultParams()
	params.SupplyBase = math.NewInt(1_000_000_000_000) // 1e12

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	bonded := params.GoalBonded.QuoInt64(2) // peak
	inflation := types.CalculateInflation(bonded, params)

	t.Run("zero supply base", func(t *testing.T) {
		p := params
		p.SupplyBase = math.ZeroInt()
		coin, inf, err := types.CalculateReward(now.Add(time.Hour), now, bonded, p)
		require.NoError(t, err)
		require.True(t, coin.IsZero())
		require.True(t, inf.IsZero())
	})

	t.Run("non-positive elapsed", func(t *testing.T) {
		coin, inf, err := types.CalculateReward(now, now, bonded, params)
		require.NoError(t, err)
		require.True(t, coin.IsZero())
		require.True(t, inflation.Equal(inf))
	})

	t.Run("one year elapsed releases annual provision", func(t *testing.T) {
		coin, inf, err := types.CalculateReward(
			now.Add(time.Duration(types.SecondsPerYear)*time.Second),
			now,
			bonded,
			params,
		)
		require.NoError(t, err)
		require.True(t, inflation.Equal(inf))

		expected := inflation.MulInt(params.SupplyBase).TruncateInt()
		require.True(t, expected.Equal(coin.Amount), "want %s got %s", expected, coin.Amount)
		require.Equal(t, params.TokenDenom, coin.Denom)
	})

	t.Run("half year is half annual provision", func(t *testing.T) {
		coin, _, err := types.CalculateReward(
			now.Add(time.Duration(types.SecondsPerYear/2)*time.Second),
			now,
			bonded,
			params,
		)
		require.NoError(t, err)

		annual := inflation.MulInt(params.SupplyBase).TruncateInt()
		require.True(t, coin.Amount.LTE(annual))
		require.True(t, coin.Amount.GTE(annual.QuoRaw(2).Sub(math.NewInt(1))))
	})
}
