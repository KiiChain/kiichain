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
			want: math.LegacyNewDecWithPrec(13, 2).Mul(params.GoalBonded).QuoInt64(4),
		},
		{
			name:        "mid curve below peak",
			bondedRatio: math.LegacyNewDecWithPrec(20, 2),
			want: math.LegacyOneDec().
				Sub(math.LegacyNewDecWithPrec(20, 2).Quo(params.GoalBonded)).
				Mul(math.LegacyNewDecWithPrec(13, 2)).
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

// spreadsheetInflationCases are golden vectors from
// "Inflation Calculation (2).xlsx" (goalBonded=0.67, inflationRateChange=0.13).
// wantInflation is the KiiChain column; wantAnnual is TruncateInt(inflation * 1.8e9).
var spreadsheetInflationCases = []struct {
	bonded       string
	wantInflation string
	wantAnnual   int64
}{
	{"0", "0.000000000000000000", 0},
	{"0.02", "0.002522388059701493", 4_540_298},
	{"0.04", "0.004889552238805970", 8_801_194},
	{"0.06", "0.007101492537313433", 12_782_686},
	{"0.08", "0.009158208955223881", 16_484_776},
	{"0.1", "0.011059701492537313", 19_907_462},
	{"0.12", "0.012805970149253731", 23_050_746},
	{"0.14", "0.014397014925373134", 25_914_626},
	{"0.16", "0.015832835820895522", 28_499_104},
	{"0.18", "0.017113432835820896", 30_804_179},
	{"0.2", "0.018238805970149254", 32_829_850},
	{"0.22", "0.019208955223880597", 34_576_119},
	{"0.24", "0.020023880597014925", 36_042_985},
	{"0.26", "0.020683582089552239", 37_230_447},
	{"0.28", "0.021188059701492537", 38_138_507},
	{"0.3", "0.021537313432835821", 38_767_164},
	{"0.32", "0.021731343283582090", 39_116_417},
	{"0.34", "0.021770149253731343", 39_186_268}, // spreadsheet peak
	{"0.36", "0.021653731343283582", 38_976_716},
	{"0.38", "0.021382089552238806", 38_487_761},
	{"0.4", "0.020955223880597015", 37_719_402},
	{"0.42", "0.020373134328358209", 36_671_641},
	{"0.44", "0.019635820895522388", 35_344_477},
	{"0.46", "0.018743283582089552", 33_737_910},
	{"0.48", "0.017695522388059702", 31_851_940}, // LegacyDec rounding (sheet shows ...701)
	{"0.5", "0.016492537313432836", 29_686_567},
	{"0.52", "0.015134328358208955", 27_241_791},
	{"0.54", "0.013620895522388060", 24_517_611},
	{"0.56", "0.011952238805970149", 21_514_029},
	{"0.58", "0.010128358208955224", 18_231_044},
	{"0.6", "0.008149253731343284", 14_668_656},
	{"0.62", "0.006014925373134328", 10_826_865},
	{"0.64", "0.003725373134328358", 6_705_671},
	{"0.66", "0.001280597014925373", 2_305_074},
	{"0.68", "0.000000000000000000", 0}, // above goalBonded
	{"0.8", "0.000000000000000000", 0},
	{"1.0", "0.000000000000000000", 0},
}

func TestCalculateInflationSpreadsheet(t *testing.T) {
	params := types.DefaultParams()
	require.True(t, params.GoalBonded.Equal(math.LegacyMustNewDecFromStr("0.67")))

	for _, tc := range spreadsheetInflationCases {
		t.Run("bonded_"+tc.bonded, func(t *testing.T) {
			got := types.CalculateInflation(math.LegacyMustNewDecFromStr(tc.bonded), params)
			want := math.LegacyMustNewDecFromStr(tc.wantInflation)
			require.True(t, want.Equal(got), "bonded=%s want %s got %s", tc.bonded, want, got)
		})
	}
}

func TestCalculateRewardSpreadsheetAnnual(t *testing.T) {
	// Spreadsheet "Total inflation (KiiChain)" column uses supply_base = 1.8e9
	params := types.DefaultParams()
	params.SupplyBase = math.NewInt(1_800_000_000)

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	oneYearLater := now.Add(time.Duration(types.SecondsPerYear) * time.Second)

	for _, tc := range spreadsheetInflationCases {
		t.Run("bonded_"+tc.bonded, func(t *testing.T) {
			bonded := math.LegacyMustNewDecFromStr(tc.bonded)
			coin, inf := types.CalculateReward(oneYearLater, now, bonded, params)

			wantInf := math.LegacyMustNewDecFromStr(tc.wantInflation)
			require.True(t, wantInf.Equal(inf), "inflation want %s got %s", wantInf, inf)
			require.True(t, math.NewInt(tc.wantAnnual).Equal(coin.Amount),
				"annual want %d got %s", tc.wantAnnual, coin.Amount)
			require.Equal(t, params.TokenDenom, coin.Denom)
		})
	}
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
		coin, inf := types.CalculateReward(now.Add(time.Hour), now, bonded, p)
		require.True(t, coin.IsZero())
		require.True(t, inf.IsZero())
	})

	t.Run("non-positive elapsed", func(t *testing.T) {
		coin, inf := types.CalculateReward(now, now, bonded, params)
		require.True(t, coin.IsZero())
		require.True(t, inflation.Equal(inf))
	})

	t.Run("one year elapsed releases annual provision", func(t *testing.T) {
		coin, inf := types.CalculateReward(
			now.Add(time.Duration(types.SecondsPerYear)*time.Second),
			now,
			bonded,
			params,
		)
		require.True(t, inflation.Equal(inf))

		expected := inflation.MulInt(params.SupplyBase).TruncateInt()
		require.True(t, expected.Equal(coin.Amount), "want %s got %s", expected, coin.Amount)
		require.Equal(t, params.TokenDenom, coin.Denom)
	})

	t.Run("half year is half annual provision", func(t *testing.T) {
		coin, _ := types.CalculateReward(
			now.Add(time.Duration(types.SecondsPerYear/2)*time.Second),
			now,
			bonded,
			params,
		)

		annual := inflation.MulInt(params.SupplyBase).TruncateInt()
		require.True(t, coin.Amount.LTE(annual))
		require.True(t, coin.Amount.GTE(annual.QuoRaw(2).Sub(math.NewInt(1))))
	})
}
