package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v1/x/rewards/types"
)

// EndBlocker calculates reward amt and sends it to the distribution pool
func (k Keeper) EndBlocker(ctx sdk.Context) error {
	// Get releaser
	releaser, err := k.RewardReleaser.Get(ctx)
	if err != nil {
		return err
	}

	// Early exit if inactive or nothing to release
	if !releaser.Active || releaser.TotalAmount.IsZero() {
		return nil
	}

	// If active and there is no previous time stamp, set it as current block's and skip this time
	if releaser.LastReleaseTime.IsZero() {
		releaser.LastReleaseTime = ctx.BlockTime()
		return k.RewardReleaser.Set(ctx, releaser)
	}

	// Calculate the amount to distribute this block
	amountToDistribute, err := CalculateReward(ctx.BlockTime(), releaser)
	if err != nil {
		return err
	}

	// If nothing to distribute, sets up as inactive for early exit next time
	if amountToDistribute.IsZero() {
		releaser.Active = false
		return k.RewardReleaser.Set(ctx, releaser)
	}

	// Get the current RewardPool from state
	rewardPool, err := k.RewardPool.Get(ctx)
	if err != nil {
		return err
	}

	// Set up coins
	coinsToDistribute := sdk.NewCoins(amountToDistribute)

	// Send to distribution pool
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, coinsToDistribute); err != nil {
		return err
	}

	// Deduct from RewardPool
	rewardPool.CommunityPool = rewardPool.CommunityPool.Sub(sdk.NewDecCoinsFromCoins(coinsToDistribute...))

	// Save change
	if err := k.RewardPool.Set(ctx, rewardPool); err != nil {
		return err
	}

	// Update releaser
	releaser.LastReleaseTime = ctx.BlockTime()
	releaser.ReleasedAmount = releaser.ReleasedAmount.Add(amountToDistribute)
	return k.RewardReleaser.Set(ctx, releaser)
}
