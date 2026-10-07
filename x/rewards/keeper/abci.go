package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

// BeginBlocker releases the per-block emission into the fee collector.
// Spendable funds are the rewards module account balance. A short balance pays
// only what is there; an empty balance skips the block. A failed bank send is
// logged and does not halt the chain.
func (k *Keeper) BeginBlocker(ctx sdk.Context) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get rewards params: %w", err)
	}

	if params.SupplyBase.IsNil() || params.SupplyBase.IsZero() {
		return nil
	}

	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	poolBalance := k.bankKeeper.GetBalance(ctx, moduleAddr, params.TokenDenom)
	if !poolBalance.IsPositive() {
		return nil
	}

	bondedRatio, err := k.stakingKeeper.BondedRatio(ctx)
	if err != nil {
		k.Logger(ctx).Error("failed to get bonded ratio", "error", err)
		return nil
	}

	amountToDistribute, inflation := types.CalculateReward(bondedRatio, params)
	if amountToDistribute.IsZero() {
		return nil
	}

	if amountToDistribute.Amount.GT(poolBalance.Amount) {
		amountToDistribute.Amount = poolBalance.Amount
	}

	err = k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.feeCollectorName, sdk.NewCoins(amountToDistribute))
	if err != nil {
		k.Logger(ctx).Error("failed to send rewards to fee collector", "error", err, "amount", amountToDistribute.String())
		return nil
	}

	rewardPool, err := k.RewardPool.Get(ctx)
	if err != nil {
		return err
	}
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
