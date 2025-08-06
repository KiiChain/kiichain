package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v4/x/rewards/types"
)

// InitGenesis sets rewards information from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	if err := k.RewardPool.Set(ctx, data.RewardPool); err != nil {
		panic(err)
	}

	if err := k.Params.Set(ctx, data.Params); err != nil {
		panic(err)
	}

	if err := k.ReleaseSchedule.Set(ctx, data.ReleaseSchedule); err != nil {
		panic(err)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	rewardPool, err := k.RewardPool.Get(ctx)
	if err != nil {
		panic(err)
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	releaseSchedule, err := k.ReleaseSchedule.Get(ctx)
	if err != nil {
		panic(err)
	}

	return types.NewGenesisState(params, rewardPool, releaseSchedule)
}
