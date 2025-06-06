package types

import (
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultParams returns default distribution parameters
func DefaultParams() Params {
	return Params{
		GovernanceMinDeposit: "1000000000000000000000", // 1000 kii
		TokenDenom:           "akii",                   // akii base denom
	}
}

// ValidateBasic performs basic validation on distribution parameters.
func (p Params) ValidateBasic() error {
	denom := p.TokenDenom
	minDeposit, err := math.NewIntFromString(p.GovernanceMinDeposit)
	if err {
		return fmt.Errorf("invalid string conversion to int on governance min deposit: %s", p.GovernanceMinDeposit)
	}
	v := sdk.Coin{Denom: denom, Amount: minDeposit}
	if v.IsValid() {
		return fmt.Errorf("invalid min deposit: %s", v)
	}
	return nil
}
