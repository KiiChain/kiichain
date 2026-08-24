package v732

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/kiichain/kiichain/v7/app/blockedaddrs"
	"github.com/kiichain/kiichain/v7/app/keepers"
)

// CreateUpgradeHandler creates the upgrade handler for the v7.3.2 upgrade.
// After module migrations it enables the bank send restriction for the
// 22 Aug 2026 incident addresses.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)

		ctx.Logger().Info("Starting module migrations for v7.3.2...")

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Enabling bank send restriction for incident addresses...")
		blockedaddrs.Enable(ctx, keepers.GetKey(banktypes.StoreKey))

		ctx.Logger().Info("Upgrade v7.3.2 complete")
		return vm, nil
	}
}
