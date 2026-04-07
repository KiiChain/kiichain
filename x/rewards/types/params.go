package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/app/params"
)

// DefaultParams returns default rewards parameters
func DefaultParams() Params {
	return Params{
		TokenDenom: params.BaseDenom, // akii base denom
	}
}

// ValidateBasic performs basic validation on distribution parameters.
func (p Params) ValidateBasic() error {
	return sdk.ValidateDenom(p.TokenDenom)
}
