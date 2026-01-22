package v601

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/kiichain/kiichain/v6/app/keepers"
)

// CreateUpgradeHandler creates the upgrade handler for the v6.0.1 upgrade
// Its only purpose is to run the module migrations
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// State the context and log
		ctx := sdk.UnwrapSDKContext(c)
		ctx.Logger().Info("Starting module migrations...")

		// Get EVM module params
		evmParams := keepers.EVMKeeper.GetParams(ctx)

		// Remove ICS precompile
		newPrecompiles := []string{}
		for _, precompile := range evmParams.ActiveStaticPrecompiles {
			if precompile != "0x0000000000000000000000000000000000000802" {
				newPrecompiles = append(newPrecompiles, precompile)
			}
		}

		// Update params
		evmParams.ActiveStaticPrecompiles = newPrecompiles
		err := keepers.EVMKeeper.SetParams(ctx, evmParams)
		if err != nil {
			return vm, err
		}

		// Run the module migrations, it will start the new module with it's init genesis
		vm, err = mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		// Log the upgrade completion
		ctx.Logger().Info("Upgrade v6.0.1 complete")
		return vm, nil
	}
}
