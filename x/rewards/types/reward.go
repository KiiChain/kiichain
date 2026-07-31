package types

import (
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SecondsPerYear is the reference year length (365 days) used to convert an
// annual emission rate into a per-block provision.
const SecondsPerYear = 31_536_000

// InflationRateChange is the max annual swing of the emission rate.
var InflationRateChange = math.LegacyNewDecWithPrec(13, 2) // 0.13

// CalculateInflation returns the KiiChain emission rate for a bonded ratio.
// The bonded-ratio multiplier forms a distribution (bell) curve, then clamps
// to [InflationMin, InflationMax].
//
//	inflation = (1 - bondedRatio/goalBonded) * inflationRateChange * bondedRatio
func CalculateInflation(bondedRatio math.LegacyDec, p Params) math.LegacyDec {
	inflation := math.LegacyOneDec().
		Sub(bondedRatio.Quo(p.GoalBonded)).
		Mul(InflationRateChange).
		Mul(bondedRatio)

	if inflation.LT(p.InflationMin) {
		return p.InflationMin
	}
	if inflation.GT(p.InflationMax) {
		return p.InflationMax
	}
	return inflation
}

// CalculateReward returns the amount to release this block and the inflation
// rate used. The annual provision (inflation * supplyBase) is prorated by the
// wall-clock time since the last release. No tokens are minted; funds come
// from the prefunded pool.
func CalculateReward(
	blockTime time.Time,
	lastReleaseTime time.Time,
	bondedRatio math.LegacyDec,
	p Params,
) (sdk.Coin, math.LegacyDec) {
	denom := p.TokenDenom
	zero := sdk.NewCoin(denom, math.ZeroInt())

	if p.SupplyBase.IsZero() {
		return zero, math.LegacyZeroDec()
	}

	inflation := CalculateInflation(bondedRatio, p)
	annualProvision := inflation.MulInt(p.SupplyBase)

	elapsedNs := blockTime.Sub(lastReleaseTime).Nanoseconds()
	if elapsedNs <= 0 {
		return zero, inflation
	}

	const nsPerYear = int64(SecondsPerYear) * 1_000_000_000
	amount := annualProvision.
		MulInt64(elapsedNs).
		QuoInt64(nsPerYear).
		TruncateInt()

	return sdk.NewCoin(denom, amount), inflation
}
