// Package v2 migrates x/rewards from consensus version 1 to 2.
//
// Version 2 replaces ReleaseSchedule emissions with inflation-curve params on
// RewardPool. This migration deletes the obsolete schedule key and backfills
// default inflation params when upgrading an existing chain.
package v2

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/x/rewards/keeper"
	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

// legacyReleaseSchedulePrefix is the collections prefix formerly used by
// ReleaseSchedule (removed in consensus version 2).
var legacyReleaseSchedulePrefix = collections.NewPrefix(2)

// MigrateStore deletes obsolete ReleaseSchedule state and ensures params /
// reward-pool fields required by the inflation-based emitter are populated.
func MigrateStore(ctx sdk.Context, k keeper.Keeper) error {
	logger := ctx.Logger().With("module", "rewards", "migration", "v2")
	logger.Info("starting rewards v2 migration")

	if err := deleteLegacyReleaseSchedule(ctx, k); err != nil {
		return err
	}

	if err := migrateParams(ctx, k); err != nil {
		return err
	}

	if err := migrateRewardPool(ctx, k); err != nil {
		return err
	}

	logger.Info("rewards v2 migration complete")
	return nil
}

func deleteLegacyReleaseSchedule(ctx sdk.Context, k keeper.Keeper) error {
	store := k.StoreService().OpenKVStore(ctx)
	return store.Delete(legacyReleaseSchedulePrefix)
}

func migrateParams(ctx sdk.Context, k keeper.Keeper) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		// No params yet; genesis/init will set defaults.
		return nil
	}

	defaults := types.DefaultParams()

	// v1 only persisted token_denom. Missing/zero goal_bonded means the
	// inflation-curve params were never configured — install defaults and keep
	// the existing denom. supply_base stays 0 so emissions remain off until gov.
	if params.GoalBonded.IsNil() || !params.GoalBonded.IsPositive() {
		tokenDenom := params.TokenDenom
		if tokenDenom == "" {
			tokenDenom = defaults.TokenDenom
		}
		params = defaults
		params.TokenDenom = tokenDenom
		params.SupplyBase = math.ZeroInt()
		return k.Params.Set(ctx, params)
	}

	if params.InflationMin.IsNil() {
		params.InflationMin = defaults.InflationMin
	}
	if params.InflationMax.IsNil() || params.InflationMax.LT(params.InflationMin) {
		params.InflationMax = defaults.InflationMax
	}
	if params.SupplyBase.IsNil() {
		params.SupplyBase = math.ZeroInt()
	}
	if params.InflationRateChange.IsNil() || !params.InflationRateChange.IsPositive() {
		params.InflationRateChange = defaults.InflationRateChange
	}
	if params.BlocksPerYear == 0 {
		params.BlocksPerYear = defaults.BlocksPerYear
	}
	return k.Params.Set(ctx, params)
}

func migrateRewardPool(ctx sdk.Context, k keeper.Keeper) error {
	pool, err := k.RewardPool.Get(ctx)
	if err != nil {
		return nil
	}

	if pool.TotalReleased.IsNil() {
		pool.TotalReleased = sdk.Coin{}
		return k.RewardPool.Set(ctx, pool)
	}
	return nil
}
