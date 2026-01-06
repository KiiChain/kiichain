package keeper

import (
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

// BeginBlocker calculates reward amt and sends it to the distribution pool
func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	// Apply telemetry metrics
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	// Get release schedule
	schedule, err := k.ReleaseSchedule.Get(ctx)
	if err != nil {
		return err
	}

	// Early exit if inactive or nothing to release
	if !schedule.Active || schedule.TotalAmount.IsZero() {
		return nil
	}

	// If active and there is no previous time stamp, set it as current block's and skip this time
	if schedule.LastReleaseTime.IsZero() {
		schedule.LastReleaseTime = ctx.BlockTime()
		return k.ReleaseSchedule.Set(ctx, schedule)
	}

	// Calculate the amount to distribute this block
	amountToDistribute, err := types.CalculateReward(ctx.BlockTime(), schedule)
	if err != nil {
		return err
	}

	// If nothing to distribute, sets up as inactive for early exit next time
	if amountToDistribute.IsZero() {
		schedule.Active = false
		return k.ReleaseSchedule.Set(ctx, schedule)
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

	// Update release schedule
	schedule.LastReleaseTime = ctx.BlockTime()
	schedule.ReleasedAmount = schedule.ReleasedAmount.Add(amountToDistribute)
	if err := k.ReleaseSchedule.Set(ctx, schedule); err != nil {
		return err
	}

	k.WriteRewardMetrics(amountToDistribute, schedule.ReleasedAmount)

	return nil
}

// WriteRewardMetrics writes reward information to telemetry metrics
func (k Keeper) WriteRewardMetrics(distributed, total sdk.Coin) {
	telemetry.ModuleSetGauge(
		types.ModuleName,
		float32(distributed.Amount.ToLegacyDec().MustFloat64()),
		"reward_amount",
	)

	telemetry.ModuleSetGauge(
		types.ModuleName,
		float32(total.Amount.ToLegacyDec().MustFloat64()),
		"total_reward_released",
	)
}
