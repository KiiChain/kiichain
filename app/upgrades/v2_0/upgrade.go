package v200

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/kiichain/kiichain/v1/app/keepers"
	rewardtypes "github.com/kiichain/kiichain/v1/x/rewards/types"
)

// CreateUpgradeHandler creates the upgrade handler for the v2.0.0 upgrade
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// State the context and log
		ctx := sdk.UnwrapSDKContext(c)
		ctx.Logger().Info("Starting module migrations...")

		// Set the initial version for the new module to 1
		if vm["rewards"] == 0 {
			vm["rewards"] = 1
		}

		// Run the module migrations
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		// Set default params
		if err := keepers.RewardsKeeper.Params.Set(ctx, rewardtypes.DefaultParams()); err != nil {
			return vm, err
		}

		// Log the upgrade completion
		ctx.Logger().Info("Upgrade v2.0.0 complete")
		return vm, nil
	}
}
