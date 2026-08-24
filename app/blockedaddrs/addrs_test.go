package blockedaddrs

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestAddrPairs(t *testing.T) {
	require.Len(t, AddrPairs, 40)
	for _, pair := range AddrPairs {
		require.True(t, IsBlockedAddr(pair[0]), pair[0])
		require.True(t, IsBlockedAddr(pair[1]), pair[1])
		require.True(t, IsBlockedAddr("0x"+pair[0][2:]), pair[0])

		raw, err := hex.DecodeString(pair[0][2:])
		require.NoError(t, err)
		require.True(t, IsBlockedAccAddress(sdk.AccAddress(raw)), pair[0])
	}
	require.False(t, IsBlockedAddr("0x0000000000000000000000000000000000000001"))
}
