package keeper_test

import (
	"bytes"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain3/testutil/keeper"
	"github.com/stretchr/testify/require"
)

func TestSetGetAddressMapping(t *testing.T) {
	k := &keeper.EVMTestApp.EvmKeeper
	ctx := keeper.EVMTestApp.GetContextForDeliverTx([]byte{})
	kiiAddr, evmAddr := keeper.MockAddressPair()
	_, ok := k.GetEVMAddress(ctx, kiiAddr)
	require.False(t, ok)
	_, ok = k.GetKiiAddress(ctx, evmAddr)
	require.False(t, ok)
	k.SetAddressMapping(ctx, kiiAddr, evmAddr)
	foundEVM, ok := k.GetEVMAddress(ctx, kiiAddr)
	require.True(t, ok)
	require.Equal(t, evmAddr, foundEVM)
	foundKii, ok := k.GetKiiAddress(ctx, evmAddr)
	require.True(t, ok)
	require.Equal(t, kiiAddr, foundKii)
	require.Equal(t, kiiAddr, k.AccountKeeper().GetAccount(ctx, kiiAddr).GetAddress())
}

func TestDeleteAddressMapping(t *testing.T) {
	k := &keeper.EVMTestApp.EvmKeeper
	ctx := keeper.EVMTestApp.GetContextForDeliverTx([]byte{})
	kiiAddr, evmAddr := keeper.MockAddressPair()
	k.SetAddressMapping(ctx, kiiAddr, evmAddr)
	foundEVM, ok := k.GetEVMAddress(ctx, kiiAddr)
	require.True(t, ok)
	require.Equal(t, evmAddr, foundEVM)
	foundKii, ok := k.GetKiiAddress(ctx, evmAddr)
	require.True(t, ok)
	require.Equal(t, kiiAddr, foundKii)
	k.DeleteAddressMapping(ctx, kiiAddr, evmAddr)
	_, ok = k.GetEVMAddress(ctx, kiiAddr)
	require.False(t, ok)
	_, ok = k.GetKiiAddress(ctx, evmAddr)
	require.False(t, ok)
}

func TestGetAddressOrDefault(t *testing.T) {
	k := &keeper.EVMTestApp.EvmKeeper
	ctx := keeper.EVMTestApp.GetContextForDeliverTx([]byte{})
	kiiAddr, evmAddr := keeper.MockAddressPair()
	defaultEvmAddr := k.GetEVMAddressOrDefault(ctx, kiiAddr)
	require.True(t, bytes.Equal(kiiAddr, defaultEvmAddr[:]))
	defaultKiiAddr := k.GetKiiAddressOrDefault(ctx, evmAddr)
	require.True(t, bytes.Equal(defaultKiiAddr, evmAddr[:]))
}

func TestSendingToCastAddress(t *testing.T) {
	a := keeper.EVMTestApp
	ctx := a.GetContextForDeliverTx([]byte{})
	kiiAddr, evmAddr := keeper.MockAddressPair()
	castAddr := sdk.AccAddress(evmAddr[:])
	sourceAddr, _ := keeper.MockAddressPair()
	require.Nil(t, a.BankKeeper.MintCoins(ctx, "evm", sdk.NewCoins(sdk.NewCoin("ukii", sdk.NewInt(10)))))
	require.Nil(t, a.BankKeeper.SendCoinsFromModuleToAccount(ctx, "evm", sourceAddr, sdk.NewCoins(sdk.NewCoin("ukii", sdk.NewInt(5)))))
	amt := sdk.NewCoins(sdk.NewCoin("ukii", sdk.NewInt(1)))
	require.Nil(t, a.BankKeeper.SendCoinsFromModuleToAccount(ctx, "evm", castAddr, amt))
	require.Nil(t, a.BankKeeper.SendCoins(ctx, sourceAddr, castAddr, amt))
	require.Nil(t, a.BankKeeper.SendCoinsAndWei(ctx, sourceAddr, castAddr, sdk.OneInt(), sdk.OneInt()))

	a.EvmKeeper.SetAddressMapping(ctx, kiiAddr, evmAddr)
	require.NotNil(t, a.BankKeeper.SendCoinsFromModuleToAccount(ctx, "evm", castAddr, amt))
	require.NotNil(t, a.BankKeeper.SendCoins(ctx, sourceAddr, castAddr, amt))
	require.NotNil(t, a.BankKeeper.SendCoinsAndWei(ctx, sourceAddr, castAddr, sdk.OneInt(), sdk.OneInt()))
}
