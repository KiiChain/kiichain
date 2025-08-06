package types

import (
	"fmt"

	"github.com/kiichain/kiichain/v4/app/params"
)

// DefaultParams returns default rewards parameters
func DefaultParams() Params {
	return Params{
		TokenDenom: params.BaseDenom, // akii base denom
	}
}

// ValidateBasic performs basic validation on distribution parameters.
func (p Params) ValidateBasic() error {
	denom := p.TokenDenom
	if denom == "" {
		return fmt.Errorf("invalid denom, empty: %s", denom)
	}
	return nil
}
