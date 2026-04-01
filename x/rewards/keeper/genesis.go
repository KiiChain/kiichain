package keeper

import (
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

	// Consistency check: warn if the module bank balance for the configured denom
	// does not match the CommunityPool accounting. A mismatch means coins were
	// moved without going through FundCommunityPool and will eventually cause
	// BeginBlocker to return an error when it tries to distribute.
	denom := data.Params.TokenDenom
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	bankBalance := k.bankKeeper.GetBalance(ctx, moduleAddr, denom)
	poolAmount := data.RewardPool.CommunityPool.AmountOf(denom).TruncateInt()
	if !bankBalance.Amount.Equal(poolAmount) {
		k.Logger(ctx).Error(
			"rewards module bank balance does not match CommunityPool accounting at genesis",
			"denom", denom,
			"bank_balance", bankBalance.Amount,
			"community_pool", poolAmount,
		)
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
