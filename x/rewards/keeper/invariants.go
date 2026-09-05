package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

// RegisterInvariants registers all rewards module invariants with the invariant
// registry. Without registration the accounting check is never executed during
// simulation runs or crisis-module invocations.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "module-accounting", ModuleAccountingInvariant(k))
}

// ModuleAccountingInvariant checks that the rewards module bank account holds at
// least the coins tracked by the CommunityPool. A deficit means phantom funds
// are recorded, which will cause BeginBlocker bank sends to fail and halt the chain.
func ModuleAccountingInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		rewardPool, err := k.RewardPool.Get(ctx)
		if err != nil {
			return sdk.FormatInvariant(
				types.ModuleName, "module-accounting",
				"failed to retrieve reward pool from state",
			), true
		}

		moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
		for _, poolCoin := range rewardPool.CommunityPool {
			required := poolCoin.Amount.TruncateInt()
			balance := k.bankKeeper.GetBalance(ctx, moduleAddr, poolCoin.Denom).Amount
			if balance.LT(required) {
				return sdk.FormatInvariant(
					types.ModuleName, "module-accounting",
					fmt.Sprintf(
						"rewards module bank balance is less than CommunityPool for denom %s: "+
							"bank=%s community_pool=%s — phantom funds will cause a chain halt",
						poolCoin.Denom, balance, required,
					),
				), true
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName, "module-accounting", "rewards module accounting is consistent",
		), false
	}
}

// ValidateModuleAccounting checks that the rewards module account holds at least the coins tracked by the CommunityPool
func (k Keeper) ValidateModuleAccounting(ctx sdk.Context) error {
	rewardPool, err := k.RewardPool.Get(ctx)
	if err != nil {
		return err
	}

	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	for _, poolCoin := range rewardPool.CommunityPool {
		required := poolCoin.Amount.TruncateInt()
		balance := k.bankKeeper.GetBalance(ctx, moduleAddr, poolCoin.Denom).Amount
		if balance.LT(required) {
			return fmt.Errorf(
				"rewards module accounting mismatch for %s: bank balance %s < community pool %s",
				poolCoin.Denom, balance, required,
			)
		}
	}

	return nil
}
