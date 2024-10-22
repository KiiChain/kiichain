package migrations_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	testkeeper "github.com/kiichain/kiichain3/testutil/keeper"
	"github.com/kiichain/kiichain3/utils"
	"github.com/kiichain/kiichain3/x/evm/keeper"
	"github.com/kiichain/kiichain3/x/evm/migrations"
	"github.com/kiichain/kiichain3/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestMigrateERCNativePointers(t *testing.T) {
	k := testkeeper.EVMTestApp.EvmKeeper
	ctx := testkeeper.EVMTestApp.GetContextForDeliverTx([]byte{}).WithBlockTime(time.Now())
	var pointerAddr common.Address
	require.Nil(t, k.RunWithOneOffEVMInstance(ctx, func(e *vm.EVM) error {
		a, err := k.UpsertERCNativePointer(ctx, e, "test", utils.ERCMetadata{Name: "name", Symbol: "symbol", Decimals: 6})
		pointerAddr = a
		return err
	}, func(s1, s2 string) {}))
	require.Nil(t, migrations.MigrateERCNativePointers(ctx, &k))
	// address should stay the same
	addr, _, _ := k.GetERC20NativePointer(ctx, "test")
	require.Equal(t, pointerAddr, addr)
}

func TestMigrateERCCW20Pointers(t *testing.T) {
	k := testkeeper.EVMTestApp.EvmKeeper
	ctx := testkeeper.EVMTestApp.GetContextForDeliverTx([]byte{}).WithBlockTime(time.Now())
	var pointerAddr common.Address
	require.Nil(t, k.RunWithOneOffEVMInstance(ctx, func(e *vm.EVM) error {
		a, err := k.UpsertERCCW20Pointer(ctx, e, "test", utils.ERCMetadata{Name: "name", Symbol: "symbol"})
		pointerAddr = a
		return err
	}, func(s1, s2 string) {}))
	require.Nil(t, migrations.MigrateERCCW20Pointers(ctx, &k))
	// address should stay the same
	addr, _, _ := k.GetERC20CW20Pointer(ctx, "test")
	require.Equal(t, pointerAddr, addr)
}

func TestMigrateERCCW721Pointers(t *testing.T) {
	k := testkeeper.EVMTestApp.EvmKeeper
	ctx := testkeeper.EVMTestApp.GetContextForDeliverTx([]byte{}).WithBlockTime(time.Now())
	var pointerAddr common.Address
	require.Nil(t, k.RunWithOneOffEVMInstance(ctx, func(e *vm.EVM) error {
		a, err := k.UpsertERCCW721Pointer(ctx, e, "test", utils.ERCMetadata{Name: "name", Symbol: "symbol"})
		pointerAddr = a
		return err
	}, func(s1, s2 string) {}))
	require.Nil(t, migrations.MigrateERCCW721Pointers(ctx, &k))
	// address should stay the same
	addr, _, _ := k.GetERC721CW721Pointer(ctx, "test")
	require.Equal(t, pointerAddr, addr)
}

func TestMigrateCWERC20Pointers(t *testing.T) {
	k := testkeeper.EVMTestApp.EvmKeeper
	ctx := testkeeper.EVMTestApp.GetContextForDeliverTx([]byte{}).WithBlockTime(time.Now())
	require.Nil(t, migrations.StoreCWPointerCode(ctx, &k, true, false))
	msgServer := keeper.NewMsgServerImpl(&k)
	res, err := msgServer.RegisterPointer(sdk.WrapSDKContext(ctx), &types.MsgRegisterPointer{
		PointerType: types.PointerType_ERC20,
		ErcAddress:  "0x0000000000000000000000000000000000000000",
	})
	require.Nil(t, err)
	require.Nil(t, migrations.MigrateCWERC20Pointers(ctx, &k))
	// address should stay the same
	addr, _, _ := k.GetCW20ERC20Pointer(ctx, common.HexToAddress("0x0000000000000000000000000000000000000000"))
	require.Equal(t, res.PointerAddress, addr.String())
}

func TestMigrateCWERC721Pointers(t *testing.T) {
	k := testkeeper.EVMTestApp.EvmKeeper
	ctx := testkeeper.EVMTestApp.GetContextForDeliverTx([]byte{}).WithBlockTime(time.Now())
	require.Nil(t, migrations.StoreCWPointerCode(ctx, &k, false, true))
	msgServer := keeper.NewMsgServerImpl(&k)
	res, err := msgServer.RegisterPointer(sdk.WrapSDKContext(ctx), &types.MsgRegisterPointer{
		PointerType: types.PointerType_ERC721,
		ErcAddress:  "0x0000000000000000000000000000000000000000",
	})
	require.Nil(t, err)
	require.Nil(t, migrations.MigrateCWERC721Pointers(ctx, &k))
	// address should stay the same
	addr, _, _ := k.GetCW721ERC721Pointer(ctx, common.HexToAddress("0x0000000000000000000000000000000000000000"))
	require.Equal(t, res.PointerAddress, addr.String())
}
