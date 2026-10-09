package v2

import (
	"fmt"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"

	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/store/v2/prefix"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Migrate clears legacy pending packet markers. The old markers were keyed by
// (channelOrClientId, sequence) and did not contain the denom, so they cannot be
// migrated into the new denom-scoped collections state.
//
// Migrate intentionally does not compensate RateLimit.Flow values for packets
// that were pending at upgrade time. Without the denom, the migration cannot
// identify which rate limit to adjust, so that pre-upgrade flow remains charged
// until the next quota reset.
func Migrate(ctx sdk.Context, storeService corestore.KVStoreService) error {
	adapter := runtime.KVStoreAdapter(storeService.OpenKVStore(ctx))

	if err := clearPrefixStore(prefix.NewStore(adapter, types.PendingSendPacketPrefix)); err != nil {
		return fmt.Errorf("clearing legacy pending send packet entries: %w", err)
	}
	if err := clearPrefixStore(prefix.NewStore(adapter, types.PendingReceivePacketPrefix)); err != nil {
		return fmt.Errorf("clearing legacy pending receive packet entries: %w", err)
	}

	return nil
}

func clearPrefixStore(store prefix.Store) error {
	iterator := store.Iterator(nil, nil)

	keys := make([][]byte, 0)
	for ; iterator.Valid(); iterator.Next() {
		keys = append(keys, append([]byte(nil), iterator.Key()...))
	}
	if err := iterator.Close(); err != nil {
		return err
	}

	for _, key := range keys {
		store.Delete(key)
	}

	return nil
}
