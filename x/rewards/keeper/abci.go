package keeper

import (
	"fmt"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

// haltSchedule logs err, marks the release schedule inactive, and returns nil so
// the error never reaches FinalizeBlock and halts the chain
func (k Keeper) haltSchedule(ctx sdk.Context, schedule types.ReleaseSchedule, err error) error {
	k.Logger(ctx).Error("halting release schedule due to error", "error", err)
	schedule.Active = false
	if setErr := k.ReleaseSchedule.Set(ctx, schedule); setErr != nil {
		return setErr
	}
	return nil
}

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
		return k.haltSchedule(ctx, schedule, err)
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

	// Verify the CommunityPool has sufficient balance for the denom before transferring
	poolAmount := rewardPool.CommunityPool.AmountOf(amountToDistribute.Denom)
	if math.LegacyNewDecFromInt(amountToDistribute.Amount).GT(poolAmount) {
		return k.haltSchedule(ctx, schedule, fmt.Errorf("community pool has insufficient balance for %s: pool has %s, need %s",
			amountToDistribute.Denom, poolAmount, amountToDistribute.Amount))
	}

	// Send to distribution pool
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, coinsToDistribute); err != nil {
		return err
	}

	// Deduct from RewardPool using SafeSub so a divergence cannot panic the chain.
	remaining, hasNeg := rewardPool.CommunityPool.SafeSub(sdk.NewDecCoinsFromCoins(coinsToDistribute...))
	if hasNeg {
		return k.haltSchedule(ctx, schedule, fmt.Errorf("community pool subtraction resulted in negative balance for denom %s", amountToDistribute.Denom))
	}
	rewardPool.CommunityPool = remaining

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
