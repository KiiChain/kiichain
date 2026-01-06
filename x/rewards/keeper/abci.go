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

	k.WriteRewardMetrics(ctx, amountToDistribute, schedule.ReleasedAmount)

	return nil
}

// WriteRewardMetrics writes reward information to telemetry metrics
func (k Keeper) WriteRewardMetrics(ctx sdk.Context, distributed, total sdk.Coin) {
	// Forgo telemetry on conversion errors
	distFloat, err := distributed.Amount.ToLegacyDec().Float64()
	if err != nil {
		k.Logger(ctx).Error("failed to convert distributed amount to float64",
			"error", err,
			"distributed", distributed.String(),
			"denom", distributed.Denom,
		)
		return
	}
	totalFloat, err := total.Amount.ToLegacyDec().Float64()
	if err != nil {
		k.Logger(ctx).Error("failed to convert total release amount to float64",
			"error", err,
			"total", total.String(),
			"denom", total.Denom,
		)
		return
	}

	telemetry.ModuleSetGauge(
		types.ModuleName,
		float32(distFloat),
		"reward_released",
	)

	telemetry.ModuleSetGauge(
		types.ModuleName,
		float32(totalFloat),
		"total_reward_released",
	)
}
