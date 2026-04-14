// Package v2 contains the state migration from consensus version 1 to 2.
// It backfills the secondary admin index (admin -> []denom) for all existing
// tokenfactory denoms
package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/x/tokenfactory/keeper"
)

// MigrateStore iterates every tokenfactory denom, reads its authority metadata,
// and writes the corresponding entry into the admin secondary index
func MigrateStore(ctx sdk.Context, k keeper.Keeper) error {
	logger := ctx.Logger().With("module", "tokenfactory", "migration", "v2")
	logger.Info("starting admin index backfill migration")

	iterator := k.GetAllDenomsIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		denom := string(iterator.Value())

		metadata, err := k.GetAuthorityMetadata(ctx, denom)
		if err != nil {
			return err
		}

		k.AddDenomFromAdmin(ctx, metadata.Admin, denom)
	}

	logger.Info("admin index backfill migration complete")
	return nil
}
