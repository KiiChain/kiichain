package keeper

import (
	sdkmath "cosmossdk.io/math"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/x/rewards/types"
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

	// Reconcile CommunityPool to the actual module-account bank balance.
	// If the bank balance is lower than CommunityPool (e.g. after a faulty
	// upgrade), operating with phantom funds would cause BeginBlocker to fail
	// the bank transfer on every block and halt the chain. We cap the pool at
	// the actual bank balance so the chain starts in a self-consistent state.
	denom := data.Params.TokenDenom
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	bankBalance := k.bankKeeper.GetBalance(ctx, moduleAddr, denom)
	rewardPool, err := k.RewardPool.Get(ctx)
	if err != nil {
		panic(err)
	}

	poolAmount := rewardPool.CommunityPool.AmountOf(denom)
	bankDec := sdkmath.LegacyNewDecFromInt(bankBalance.Amount)
	if poolAmount.GT(bankDec) {
		k.Logger(ctx).Error(
			"rewards InitGenesis: CommunityPool exceeds bank balance — capping pool to prevent future chain halt",
			"denom", denom,
			"community_pool", poolAmount,
			"bank_balance", bankDec,
		)
		newPool := rewardPool.CommunityPool
		for i, coin := range newPool {
			if coin.Denom == denom {
				newPool[i].Amount = bankDec
				break
			}
		}
		rewardPool.CommunityPool = newPool
		if err := k.RewardPool.Set(ctx, rewardPool); err != nil {
			panic(err)
		}
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
