package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/x/tokenfactory/types"
)

// InitGenesis initializes the tokenfactory module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	if genState.Params.DenomCreationFee == nil {
		genState.Params.DenomCreationFee = sdk.NewCoins()
	}

	if err := k.SetParams(ctx, genState.Params); err != nil {
		k.Logger(ctx).Error("failed to set tokenfactory params in genesis",
			"err", err,
		)
		return
	}

	for _, genDenom := range genState.GetFactoryDenoms() {
		creator, _, err := types.DeconstructDenom(genDenom.GetDenom())
		if err != nil {
			k.Logger(ctx).Error("failed to deconstruct denom in genesis",
				"denom", genDenom.GetDenom(),
				"err", err,
			)
			continue
		}

		if err := k.createDenomAfterValidation(ctx, creator, genDenom.GetDenom()); err != nil {
			k.Logger(ctx).Error("failed to create denom in genesis",
				"denom", genDenom.GetDenom(),
				"err", err,
			)
			continue
		}

		if err := k.setAuthorityMetadata(ctx, genDenom.GetDenom(), genDenom.GetAuthorityMetadata()); err != nil {
			k.Logger(ctx).Error("failed to set authority metadata in genesis",
				"denom", genDenom.GetDenom(),
				"err", err,
			)
			continue
		}
	}
}

// ExportGenesis returns the tokenfactory module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genDenoms := []types.GenesisDenom{}
	iterator := k.GetAllDenomsIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		denom := string(iterator.Value())

		authorityMetadata, err := k.GetAuthorityMetadata(ctx, denom)
		if err != nil {
			panic(err)
		}

		genDenoms = append(genDenoms, types.GenesisDenom{
			Denom:             denom,
			AuthorityMetadata: authorityMetadata,
		})
	}

	return &types.GenesisState{
		FactoryDenoms: genDenoms,
		Params:        k.GetParams(ctx),
	}
}
