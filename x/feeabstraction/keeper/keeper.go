package keeper

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
)

// FeePrice is a generate
// TODO: implement me and the module
type FeePrice struct {
	Denom string
	Price math.LegacyDec
}

// Params define the module parameters
// TODO: implement me and the module
type Params struct {
	DefaultDenom string
}

// This is a mocked keeper for the fee abstraction module
// This must be implemented
// TODO: implement me and the module
type Keeper struct {
	erc20Keeper types.Erc20Keeper
	bankKeeper  types.BankKeeper
}

// NewKeeper creates a new instance of the Keeper
// TODO: implement me and the module
func NewKeeper(erc20Keeper types.Erc20Keeper, bankKeeper types.BankKeeper) Keeper {
	return Keeper{
		erc20Keeper: erc20Keeper,
		bankKeeper:  bankKeeper,
	}
}

// GetFeeAbstractionTokens returns the fee abstraction tokens
// For now we use a mocked value
// TODO: implement me and the module
func (k Keeper) GetFeePrices(ctx sdk.Context) ([]FeePrice, error) {
	return []FeePrice{
		{
			Denom: "erc20/0x28B081E92CEf14492Ba90fdBCeCbe0693F9a1f8d",
			Price: math.LegacyNewDecFromInt(math.NewInt(10)), // 10 uatom = 1 kii
		},
	}, nil
}

// GetParams returns the module params
// TODO: implement me and the module
func (k Keeper) GetParams(ctx sdk.Context) Params {
	return Params{
		DefaultDenom: "akii",
	}
}
