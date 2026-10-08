package types

import (
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/app/params"
)

// DefaultBlocksPerYear assumes a 2s block time (app TimeoutCommit):
// 60 * 60 * 8766 / 2 = 15_778_800.
const DefaultBlocksPerYear uint64 = 60 * 60 * 8766 / 2

// DefaultParams returns default rewards parameters.
// SupplyBase defaults to zero => zero emissions until governance sets it.
// SupplyBase is a notional emission-scale base (not chain total supply).
func DefaultParams() Params {
	return Params{
		TokenDenom:          params.BaseDenom,                 // akii
		GoalBonded:          math.LegacyNewDecWithPrec(67, 2), // 0.67
		InflationMin:        math.LegacyZeroDec(),             // 0.00
		InflationMax:        math.LegacyNewDecWithPrec(20, 2), // 0.20
		SupplyBase:          math.ZeroInt(),                   // TBD via governance
		InflationRateChange: math.LegacyNewDecWithPrec(13, 2), // 0.13
		BlocksPerYear:       DefaultBlocksPerYear,
	}
}

// ValidateBasic performs basic validation on rewards parameters.
func (p Params) ValidateBasic() error {
	if err := sdk.ValidateDenom(p.TokenDenom); err != nil {
		return err
	}

	if !p.GoalBonded.IsPositive() || p.GoalBonded.GT(math.LegacyOneDec()) {
		return fmt.Errorf("goalBonded must be in (0,1], got %s", p.GoalBonded)
	}

	if p.InflationMin.IsNegative() {
		return fmt.Errorf("inflationMin cannot be negative, got %s", p.InflationMin)
	}

	if p.InflationMax.LT(p.InflationMin) {
		return fmt.Errorf("inflationMax %s < inflationMin %s", p.InflationMax, p.InflationMin)
	}

	if p.SupplyBase.IsNegative() {
		return fmt.Errorf("supplyBase cannot be negative, got %s", p.SupplyBase)
	}

	if !p.InflationRateChange.IsPositive() || p.InflationRateChange.GT(math.LegacyOneDec()) {
		return fmt.Errorf("inflationRateChange must be in (0,1], got %s", p.InflationRateChange)
	}

	if p.BlocksPerYear == 0 {
		return fmt.Errorf("blocksPerYear must be positive, got %d", p.BlocksPerYear)
	}

	return nil
}
