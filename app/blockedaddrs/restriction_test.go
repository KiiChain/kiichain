package blockedaddrs

import (
	"context"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestSendRestriction(t *testing.T) {
	blocked := sdk.AccAddress(mustDecodeHex("0e7a96227fcf09f53d644ba6462d8c73993ef246"))
	allowed := sdk.AccAddress(mustDecodeHex("0000000000000000000000000000000000000001"))
	coins := sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1)))
	ctx := sdk.Context{}.WithContext(context.Background())

	_, err := SendRestriction(ctx, allowed, allowed, coins)
	require.NoError(t, err)

	_, err = SendRestriction(ctx, blocked, allowed, coins)
	require.Error(t, err)
	require.ErrorContains(t, err, "address is blocked")

	_, err = SendRestriction(ctx, allowed, blocked, coins)
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
