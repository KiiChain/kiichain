package v730

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/kiichain/kiichain/v7/app/keepers"
)

// CreateUpgradeHandler creates the upgrade handler for the v7.3.0 upgrade.
// This upgrade coordinates the switch to the EVM v0.6.1-fork.1 dependency, whose
// changes are state-machine-breaking. No custom state migrations are needed, so
// the handler only runs pending module migrations.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	_ *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)

		ctx.Logger().Info("Starting module migrations for v7.3.0...")

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Upgrade v7.3.0 complete")
		return vm, nil
	}
}
