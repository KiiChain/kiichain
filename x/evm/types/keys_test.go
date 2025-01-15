package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain3/x/evm/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestEVMAddressToKiiAddressKey(t *testing.T) {
	evmAddr := common.HexToAddress("0x1234567890abcdef1234567890abcdef12345678")
	expectedPrefix := types.EVMAddressToKiiAddressKeyPrefix
	key := types.EVMAddressToKiiAddressKey(evmAddr)

	require.Equal(t, expectedPrefix[0], key[0], "Key prefix for evm address to kii address key is incorrect")
	require.Equal(t, append(expectedPrefix, evmAddr.Bytes()...), key, "Generated key format is incorrect")
}

func TestKiiAddressToEVMAddressKey(t *testing.T) {
	kiiAddr := sdk.AccAddress("kii1234567890abcdef1234567890abcdef12345678")
	expectedPrefix := types.KiiAddressToEVMAddressKeyPrefix
	key := types.KiiAddressToEVMAddressKey(kiiAddr)

	require.Equal(t, expectedPrefix[0], key[0], "Key prefix for kii address to evm address key is incorrect")
	require.Equal(t, append(expectedPrefix, kiiAddr...), key, "Generated key format is incorrect")
}

func TestStateKey(t *testing.T) {
	evmAddr := common.HexToAddress("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	expectedPrefix := types.StateKeyPrefix
	key := types.StateKey(evmAddr)

	require.Equal(t, expectedPrefix[0], key[0], "Key prefix for state key is incorrect")
	require.Equal(t, append(expectedPrefix, evmAddr.Bytes()...), key, "Generated key format is incorrect")
}

func TestBlockBloomKey(t *testing.T) {
	height := int64(123456)
	key := types.BlockBloomKey(height)

	require.Equal(t, types.BlockBloomPrefix[0], key[0], "Key prefix for block bloom key is incorrect")
}
