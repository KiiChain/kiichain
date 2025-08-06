package v130

import (
	"context"

	"github.com/ethereum/go-ethereum/common"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/kiichain/kiichain/v3/app/keepers"
	utils "github.com/kiichain/kiichain/v3/app/upgrades/utils"
	"github.com/kiichain/kiichain/v3/precompiles/ibc"
	"github.com/kiichain/kiichain/v3/precompiles/wasmd"
)

// CreateUpgradeHandler creates the upgrade handler for the v1.3.0 upgrade
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// State the context and log
		ctx := sdk.UnwrapSDKContext(c)
		ctx.Logger().Info("Starting module migrations...")

		// Run the module migrations
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		// Install the new precompile
		err = utils.InstallNewPrecompiles(
			ctx,
			keepers,
			[]common.Address{
				common.HexToAddress(wasmd.WasmdPrecompileAddress),
				common.HexToAddress(ibc.IBCPrecompileAddress),
			},
		)
		if err != nil {
			return vm, err
		}

		// Log the upgrade completion
		ctx.Logger().Info("Upgrade v1.3.0 complete")
		return vm, nil
	}
}
