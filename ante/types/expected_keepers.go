package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// FeeAbstractionKeeper defines the required interface for the Fee Abstraction module
type FeeAbstractionKeeper interface {
	ConvertNativeFee(ctx sdk.Context, account sdk.AccAddress, fees sdk.Coins) (sdk.Coins, error)
}
