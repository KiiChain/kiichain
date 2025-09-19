package v500

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	erc20types "github.com/cosmos/evm/x/erc20/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/ethereum/go-ethereum/common"

	clientkeeper "github.com/cosmos/ibc-go/v10/modules/core/02-client/keeper"
	"github.com/kiichain/kiichain/v5/app/keepers"
)

// CreateUpgradeHandler creates the upgrade handler for the v5.0.0 upgrade
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// State the context and log
		ctx := sdk.UnwrapSDKContext(c)
		ctx.Logger().Info("Starting module migrations...")

		// Run the module migrations, it will start the new module with it's init genesis
		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		// Run ERC20 migration
		MigrateERC20(ctx, keepers)

		// Remove localhost from client states IBC
		err = clientkeeper.NewMigrator(keepers.IBCKeeper.ClientKeeper).MigrateToStatelessLocalhost(ctx)
		if err != nil {
			return vm, err
		}

		// Log the upgrade completion
		ctx.Logger().Info("Upgrade v5.0.0 complete")
		return vm, nil
	}
}

// MigrateERC20 reads old dynamic and native precompiles and add them to the new storage format
func MigrateERC20(
	ctx sdk.Context,
	keepers *keepers.AppKeepers,
) {
	// In your upgrade handler
	storekeys := keepers.GetKVStoreKey()
	store := ctx.KVStore(storekeys[erc20types.StoreKey])
	const addressLength = 42 // "0x" + 40 hex characters

	// Migrate dynamic precompiles (IBC tokens, token factory)
	if oldData := store.Get([]byte("DynamicPrecompiles")); len(oldData) > 0 {
		for i := 0; i < len(oldData); i += addressLength {
			address := common.HexToAddress(string(oldData[i : i+addressLength]))
			keepers.Erc20Keeper.SetDynamicPrecompile(ctx, address)
		}
		store.Delete([]byte("DynamicPrecompiles"))
	}

	// Migrate native precompiles
	if oldData := store.Get([]byte("NativePrecompiles")); len(oldData) > 0 {
		for i := 0; i < len(oldData); i += addressLength {
			address := common.HexToAddress(string(oldData[i : i+addressLength]))
			keepers.Erc20Keeper.SetNativePrecompile(ctx, address)
		}
		store.Delete([]byte("NativePrecompiles"))
	}
}
