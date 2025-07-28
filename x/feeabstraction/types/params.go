package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v3/app/params"
)

// NewParams returns a new params instance
func NewParams(nativeDenom string) Params {
	return Params{
		NativeDenom: nativeDenom,
	}
}

// DefaultParams returns default params
func DefaultParams() Params {
	return Params{
		NativeDenom: params.BaseDenom,
	}
}

// ValidateBasic performs basic validation on distribution parameters.
func (p Params) ValidateBasic() error {
	return sdk.ValidateDenom(p.NativeDenom)
}
