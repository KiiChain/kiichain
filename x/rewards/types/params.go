package types

import (
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/app/params"
)

// DefaultParams returns default rewards parameters.
// SupplyBase defaults to zero => zero emissions until governance sets it.
func DefaultParams() Params {
	return Params{
		TokenDenom:   params.BaseDenom,                 // akii
		GoalBonded:   math.LegacyNewDecWithPrec(67, 2), // 0.67
		InflationMin: math.LegacyZeroDec(),             // 0.00
		InflationMax: math.LegacyNewDecWithPrec(20, 2), // 0.20
		SupplyBase:   math.ZeroInt(),                   // TBD via governance
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

	return nil
}
