package v500

import (
	"context"

	"github.com/ethereum/go-ethereum/common"

	upgradetypes "cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	erc20types "github.com/cosmos/evm/x/erc20/types"
	evmtypes "github.com/cosmos/evm/x/vm/types"

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

		// Migrate EVM info
		ctx.Logger().Info("Migrating EVM params...")
		err = MigrateEVMParams(ctx, keepers)
		if err != nil {
			return vm, err
		}

		// Run ERC20 migration
		ctx.Logger().Info("Migrating EVM info...")
		err = MigrateERC20(ctx, keepers)
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
) error {
	// In your upgrade handler
	storekeys := keepers.GetKVStoreKey()
	store := runtime.NewKVStoreService(storekeys[erc20types.StoreKey]).OpenKVStore(ctx)
	const addressLength = 42 // "0x" + 40 hex characters

	// Migrate dynamic precompiles (IBC tokens, token factory)
	ctx.Logger().Info("Migrating dynamic precompiles...")
	oldData, err := store.Get([]byte("DynamicPrecompiles"))
	if err != nil {
		return err
	}

	if len(oldData) > 0 {
		for i := 0; i < len(oldData); i += addressLength {
			address := common.HexToAddress(string(oldData[i : i+addressLength]))
			keepers.Erc20Keeper.SetDynamicPrecompile(ctx, address)
		}
		err = store.Delete([]byte("DynamicPrecompiles"))
		if err != nil {
			return err
		}
	}

	// Migrate native precompiles
	ctx.Logger().Info("Migrating Native precompiles...")
	oldData, err = store.Get([]byte("NativePrecompiles"))
	if err != nil {
		return err
	}
	if len(oldData) > 0 {
		for i := 0; i < len(oldData); i += addressLength {
			address := common.HexToAddress(string(oldData[i : i+addressLength]))
			keepers.Erc20Keeper.SetNativePrecompile(ctx, address)
		}
		err = store.Delete([]byte("NativePrecompiles"))
		if err != nil {
			return err
		}
	}

	// Add missing ERC20 param
	ctx.Logger().Info("Adding permissionless registration...")
	params := keepers.Erc20Keeper.GetParams(ctx)
	params.PermissionlessRegistration = true
	return keepers.Erc20Keeper.SetParams(ctx, params)
}

// MigrateEVMParams imports relevant old v0.1 params and sets them on new EVM param type
func MigrateEVMParams(
	ctx sdk.Context,
	keepers *keepers.AppKeepers,
) error {
	storekeys := keepers.GetKVStoreKey()
	store := runtime.NewKVStoreService(storekeys[evmtypes.StoreKey]).OpenKVStore(ctx)

	// Fetch byte data of old params
	oldData, err := store.Get(evmtypes.KeyPrefixParams)
	if err != nil {
		return err
	}

	// Read old params
	var oldParams Params
	if oldData != nil {
		if err := oldParams.Unmarshal(oldData); err != nil {
			return err
		}
	}

	// set the evm/vm params
	evmParams := evmtypes.DefaultParams()
	evmParams.EvmDenom = evmtypes.GetEVMCoinDenom()
	evmParams.ActiveStaticPrecompiles = oldParams.ActiveStaticPrecompiles
	evmParams.EVMChannels = oldParams.EVMChannels
	evmParams.AllowUnprotectedTxs = oldParams.AllowUnprotectedTxs

	return keepers.EVMKeeper.SetParams(ctx, evmParams)
}
