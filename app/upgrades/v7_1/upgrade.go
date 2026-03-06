package v710

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/kiichain/kiichain/v7/app/keepers"
)

const (
	ics20precompileAddress = "0x0000000000000000000000000000000000000802"
)

// CreateUpgradeHandler creates the upgrade handler for the v7.1.0 upgrade
// Its only purpose is to run the module migrations
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// State the context and log
		ctx := sdk.UnwrapSDKContext(c)

		ctx.Logger().Info("Removing wasmd precompile...")

		// Get EVM module params
		evmParams := keepers.EVMKeeper.GetParams(ctx)

		// Remove wasmd precompile
		newPrecompiles := []string{}
		icsExists := false
		for _, precompile := range evmParams.ActiveStaticPrecompiles {
			if precompile != "0x0000000000000000000000000000000000001001" {
				newPrecompiles = append(newPrecompiles, precompile)
			}
			if precompile == ics20precompileAddress {
				icsExists = true
			}
		}

		// Add ICS precompile back if it's not there
		if !icsExists {
			newPrecompiles = append(newPrecompiles, ics20precompileAddress)
		}

		// Update params
		evmParams.ActiveStaticPrecompiles = newPrecompiles
		err := keepers.EVMKeeper.SetParams(ctx, evmParams)
		if err != nil {
			return vm, err
		}

		ctx.Logger().Info("Starting module migrations...")

		// Run the module migrations, it will start the new module with it's init genesis
		vm, err = mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		// Log the upgrade completion
		ctx.Logger().Info("Upgrade v7.1.0 complete")
		return vm, nil
	}
}
