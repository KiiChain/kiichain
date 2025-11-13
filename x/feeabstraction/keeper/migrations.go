package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain/v5/x/feeabstraction/types"
)

type Migrator struct {
	keeper Keeper
}

func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	// storeServie := runtime.NewKVStoreService(appKeepers.keys[feeabstractiontypes.StoreKey])
	// sb := collections.NewSchemaBuilder(m.keeper.storeService)

	// // Fetch old params, that have one less field
	// schema := collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[v1.Params](m.keeper.cdc))

	oldParams, err := m.keeper.Params.Get(ctx)
	if err != nil {
		return err
	}

	newParams := types.Params{
		NativeDenom:        oldParams.NativeDenom,
		NativeOracleDenom:  oldParams.NativeOracleDenom,
		Enabled:            oldParams.Enabled,
		ClampFactor:        oldParams.ClampFactor,
		TwapLookbackWindow: oldParams.TwapLookbackWindow,
	}

	// Set them up, the extra field should not matter
	return m.keeper.Params.Set(ctx, newParams)
}
