package blockedaddrs

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestSendRestrictionGatedByUpgrade(t *testing.T) {
	key := storetypes.NewKVStoreKey("bank")
	ctx := testutil.DefaultContext(key, storetypes.NewTransientStoreKey("transient"))
	restriction := NewSendRestriction(key)

	blocked := sdk.AccAddress(mustDecodeHex("0e7a96227fcf09f53d644ba6462d8c73993ef246"))
	allowed := sdk.AccAddress(mustDecodeHex("0000000000000000000000000000000000000001"))
	coins := sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1)))

	require.False(t, IsEnabled(ctx, key))
	_, err := restriction(ctx, blocked, allowed, coins)
	require.NoError(t, err)
	_, err = restriction(ctx, allowed, blocked, coins)
	require.NoError(t, err)

	Enable(ctx, key)
	require.True(t, IsEnabled(ctx, key))

	_, err = restriction(ctx, allowed, allowed, coins)
	require.NoError(t, err)

	_, err = restriction(ctx, blocked, allowed, coins)
	require.Error(t, err)
	require.ErrorContains(t, err, "address is blocked")

	_, err = restriction(ctx, allowed, blocked, coins)
	require.Error(t, err)
	require.ErrorContains(t, err, "address is blocked")
}

func mustDecodeHex(s string) []byte {
	bz, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return bz
}
