package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain/v7/x/rewards/keeper/invariants"
	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

// Implement KeeperInterface untuk invariants
func (k Keeper) ReleaseScheduleGet(ctx sdk.Context) (types.ReleaseSchedule, error) {
	return k.ReleaseSchedule.Get(ctx)
}

func (k Keeper) RewardPoolGet(ctx sdk.Context) (types.RewardPool, error) {
	return k.RewardPool.Get(ctx)
}

func (k Keeper) ParamsGet(ctx sdk.Context) (types.Params, error) {
	return k.Params.Get(ctx)
}

// GetAllInvariants returns all invariant checkers
func (k Keeper) GetAllInvariants() []func(ctx sdk.Context) (string, bool) {
	return []func(ctx sdk.Context) (string, bool){
		invariants.ReleaseScheduleInvariant(k),
		invariants.CommunityPoolNonNegativeInvariant(k),
		invariants.ParamsInvariant(k),
	}
}