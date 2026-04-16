package keepers

import (
	"bytes"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
)

// TestEVMAddressCodecStringToBytes verifies that evmAddressCodec only accepts exactly-20-byte
// accounts. A longer account (e.g. a 32-byte bech32 account) must be rejected at decode time so
// it can never reach a balance-mirroring precompile and inflate native supply.
func TestEVMAddressCodecStringToBytes(t *testing.T) {
	inner := addresscodec.NewBech32Codec("kii")
	codec := evmAddressCodec{inner}

	// A 20-byte account decodes successfully and round-trips.
	addr20 := bytes.Repeat([]byte{0xAB}, common.AddressLength)
	str20, err := inner.BytesToString(addr20)
	require.NoError(t, err)

	got, err := codec.StringToBytes(str20)
	require.NoError(t, err)
	require.Equal(t, addr20, got)

	// A 32-byte account is rejected.
	addr32 := bytes.Repeat([]byte{0xCD}, 32)
	str32, err := inner.BytesToString(addr32)
	require.NoError(t, err)

	_, err = codec.StringToBytes(str32)
	require.Error(t, err)

	// An invalid bech32 string is rejected by the wrapped codec.
	_, err = codec.StringToBytes("not-a-valid-address")
	require.Error(t, err)
}
