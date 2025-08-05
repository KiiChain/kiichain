package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
)

// InitGenesis set the module collections though the genesis state
func (k Keeper) InitGenesis(ctx sdk.Context, gs types.GenesisState) error {
	// Validate the genesis state
	if err := gs.Validate(); err != nil {
		return err
	}

	// Set the params
	if err := k.Params.Set(ctx, gs.Params); err != nil {
		return err
	}

	// Set the fee tokens
	return k.FeeTokens.Set(ctx, *gs.FeeTokens)
}

// ExportGenesis reads the module collections and return the genesis state
func (k Keeper) ExportGenesis(ctx sdk.Context) (*types.GenesisState, error) {
	// Get the params
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	// Get the fee tokens
	feeTokens, err := k.FeeTokens.Get(ctx)
	if err != nil {
		return nil, err
	}

	// Return the genesis state
	return types.NewGenesisState(params, &feeTokens), nil
}
