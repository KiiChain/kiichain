package v90

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/kiichain/kiichain/v7/app/keepers"
)

// CreateUpgradeHandler creates the upgrade handler for the v9.0.0 upgrade.
// This upgrade runs the module migrations shipped with the cosmos/evm v0.7.3,
// cosmos-sdk v0.54, ibc-go v11, and wasmd v0.70 dependency bumps.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	_ *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)

		ctx.Logger().Info("Starting module migrations for v9.0.0...")

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Upgrade v9.0.0 complete")
		return vm, nil
	}
}
