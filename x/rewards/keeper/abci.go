package keeper

import (
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

// BeginBlocker calculates the inflation-based reward amount and sends it to the fee collector.
func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// supply_base == 0 disables emissions until governance configures it
	if params.SupplyBase.IsZero() {
		return nil
	}

	rewardPool, err := k.RewardPool.Get(ctx)
	if err != nil {
		return err
	}

	poolBalance := rewardPool.CommunityPool.AmountOf(params.TokenDenom).TruncateInt()
	if !poolBalance.IsPositive() {
		// Advance the clock while empty so a later fund does not dump accrued Δt
		if !rewardPool.LastReleaseTime.IsZero() && ctx.BlockTime().After(rewardPool.LastReleaseTime) {
			rewardPool.LastReleaseTime = ctx.BlockTime()
			return k.RewardPool.Set(ctx, rewardPool)
		}
		return nil
	}

	// First active block: stamp time and skip release
	if rewardPool.LastReleaseTime.IsZero() {
		rewardPool.LastReleaseTime = ctx.BlockTime()
		return k.RewardPool.Set(ctx, rewardPool)
	}

	bondedRatio, err := k.stakingKeeper.BondedRatio(ctx)
	if err != nil {
		k.Logger(ctx).Error("failed to read bonded ratio", "error", err)
		return nil
	}

	amountToDistribute, inflation := types.CalculateReward(
		ctx.BlockTime(),
		rewardPool.LastReleaseTime,
		bondedRatio,
		params,
	)

	// Cap at remaining pool balance so emissions run until the pool is dry
	amountToDistribute.Amount = math.MinInt(amountToDistribute.Amount, poolBalance)

	if amountToDistribute.IsZero() {
		return nil
	}

	coinsToDistribute := sdk.NewCoins(amountToDistribute)

	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, coinsToDistribute); err != nil {
		k.Logger(ctx).Error("failed to send rewards to fee collector", "error", err)
		return nil
	}

	remaining, hasNeg := rewardPool.CommunityPool.SafeSub(sdk.NewDecCoinsFromCoins(coinsToDistribute...))
	if hasNeg {
		k.Logger(ctx).Error("community pool subtraction resulted in negative balance",
			"denom", amountToDistribute.Denom)
		return nil
	}
	rewardPool.CommunityPool = remaining
	rewardPool.LastReleaseTime = ctx.BlockTime()

	if rewardPool.TotalReleased.IsNil() || rewardPool.TotalReleased.IsZero() {
		rewardPool.TotalReleased = amountToDistribute
	} else {
		rewardPool.TotalReleased = rewardPool.TotalReleased.Add(amountToDistribute)
	}

	if err := k.RewardPool.Set(ctx, rewardPool); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeRewardDistributed,
		sdk.NewAttribute(types.AttributeKeyAmount, amountToDistribute.String()),
		sdk.NewAttribute(types.AttributeKeyTotalReleased, rewardPool.TotalReleased.String()),
		sdk.NewAttribute(types.AttributeKeyInflationRate, inflation.String()),
		sdk.NewAttribute(types.AttributeKeyBondedRatio, bondedRatio.String()),
	))

	k.WriteRewardMetrics(ctx, amountToDistribute, rewardPool.TotalReleased)

	return nil
}

// WriteRewardMetrics writes reward information to telemetry metrics.
// Conversion failures yield zero gauges; telemetry is best-effort.
func (k Keeper) WriteRewardMetrics(_ sdk.Context, distributed, total sdk.Coin) {
	distFloat, _ := distributed.Amount.ToLegacyDec().Float64()
	totalFloat, _ := total.Amount.ToLegacyDec().Float64()

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
