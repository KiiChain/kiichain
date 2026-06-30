package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

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
