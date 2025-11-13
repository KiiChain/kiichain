package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Migrator struct {
	keeper Keeper
}

func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate1to2 migrates the params, removing fallback_native_price from the params
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	// Fetch old params, extra param won't be unmarshaled
	params, err := m.keeper.Params.Get(ctx)
	if err != nil {
		return err
	}

	// override
	return m.keeper.Params.Set(ctx, params)
}
