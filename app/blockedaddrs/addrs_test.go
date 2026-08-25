package blockedaddrs

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	// Sets the global "kii" bech32 prefix via its init(), same as the real
	// binary — without this, sdk.AccAddressFromBech32 rejects "kii1..."
	// addresses (default prefix is "cosmos").
	_ "github.com/kiichain/kiichain/v7/app/params"
)

func TestIsBlockedAddr(t *testing.T) {
	require.Len(t, AttackerAddrs, 22)

	for bech32Addr := range AttackerAddrs {
		require.True(t, IsBlockedAddr(bech32Addr), bech32Addr)

		accAddr := sdk.MustAccAddressFromBech32(bech32Addr)
		require.True(t, IsBlockedAccAddress(accAddr), bech32Addr)
	}
}

func TestIsBlockedAddr_HexForm(t *testing.T) {
	// kii1peafvgnleuyl20tyfwnyvtvvwwvnaujxmqe5qe, confirmed via `kiichaind
	// debug addr` against its bech32 form.
	require.True(t, IsBlockedAddr("0x0e7a96227fcf09f53d644ba6462d8c73993ef246"))
	require.True(t, IsBlockedAddr("0X0E7A96227FCF09F53D644BA6462D8C73993EF246"))
}

func TestIsBlockedAddr_NotBlocked(t *testing.T) {
	require.False(t, IsBlockedAddr("kii1c6cgjmsx0ewl6j552sp06musutmfcvxcaq4n9h"))
	require.False(t, IsBlockedAddr("0x0000000000000000000000000000000000000001"))
	require.False(t, IsBlockedAddr("not-an-address"))
}

func TestIsBlockedAccAddress_Empty(t *testing.T) {
	require.False(t, IsBlockedAccAddress(sdk.AccAddress{}))
}
