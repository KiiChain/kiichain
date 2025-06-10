package types

import (
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain/v1/app/params"
)

// DefaultParams returns default distribution parameters
func DefaultParams() Params {
	return Params{
		TokenDenom: params.BaseDenom, // akii base denom
	}
}

// ValidateBasic performs basic validation on distribution parameters.
func (p Params) ValidateBasic() error {
	denom := p.TokenDenom
	minDeposit, ok := math.NewIntFromString(p.GovernanceMinDeposit)
	if !ok {
		return fmt.Errorf("invalid string conversion to int on governance min deposit: %s;", p.GovernanceMinDeposit)
	}
	v := sdk.Coin{Denom: denom, Amount: minDeposit}
	if !v.IsValid() {
		return fmt.Errorf("invalid min deposit: %s", v)
	}
	return nil
}
