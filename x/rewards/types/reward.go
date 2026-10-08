package types

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CalculateInflation returns the KiiChain emission rate for a bonded ratio.
// The bonded-ratio multiplier forms a distribution (bell) curve, then clamps
// to [InflationMin, InflationMax].
//
//	inflation = (1 - bondedRatio/goalBonded) * inflationRateChange * bondedRatio
func CalculateInflation(bondedRatio math.LegacyDec, p Params) math.LegacyDec {
	inflation := math.LegacyOneDec().
		Sub(bondedRatio.Quo(p.GoalBonded)).
		Mul(p.InflationRateChange).
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
// rate used. Matches cosmos-sdk x/mint: annualProvision / blocksPerYear.
// No tokens are minted; funds come from the prefunded pool.
func CalculateReward(bondedRatio math.LegacyDec, p Params) (sdk.Coin, math.LegacyDec) {
	denom := p.TokenDenom
	zero := sdk.NewCoin(denom, math.ZeroInt())

	if p.SupplyBase.IsZero() || p.BlocksPerYear == 0 {
		return zero, math.LegacyZeroDec()
	}

	inflation := CalculateInflation(bondedRatio, p)
	annualProvision := inflation.MulInt(p.SupplyBase)
	amount := annualProvision.
		QuoInt(math.NewIntFromUint64(p.BlocksPerYear)).
		TruncateInt()

	return sdk.NewCoin(denom, amount), inflation
}
