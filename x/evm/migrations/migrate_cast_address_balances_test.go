package migrations_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/kiichain/kiichain3/testutil/keeper"
	"github.com/kiichain/kiichain3/x/evm/migrations"
	"github.com/kiichain/kiichain3/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestMigrateCastAddressBalances(t *testing.T) {
	k := testkeeper.EVMTestApp.EvmKeeper
	ctx := testkeeper.EVMTestApp.GetContextForDeliverTx([]byte{}).WithBlockTime(time.Now())
	require.Nil(t, k.BankKeeper().MintCoins(ctx, types.ModuleName, testkeeper.UkiiCoins(100)))
	// unassociated account with funds
	kiiAddr1, evmAddr1 := testkeeper.MockAddressPair()
	require.Nil(t, k.BankKeeper().SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.AccAddress(evmAddr1[:]), testkeeper.UkiiCoins(10)))
	// associated account without funds
	kiiAddr2, evmAddr2 := testkeeper.MockAddressPair()
	k.SetAddressMapping(ctx, kiiAddr2, evmAddr2)
	// associated account with funds
	kiiAddr3, evmAddr3 := testkeeper.MockAddressPair()
	require.Nil(t, k.BankKeeper().SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.AccAddress(evmAddr3[:]), testkeeper.UkiiCoins(10)))
	k.SetAddressMapping(ctx, kiiAddr3, evmAddr3)

	require.Nil(t, migrations.MigrateCastAddressBalances(ctx, &k))

	require.Equal(t, sdk.NewInt(10), k.BankKeeper().GetBalance(ctx, sdk.AccAddress(evmAddr1[:]), "ukii").Amount)
	require.Equal(t, sdk.ZeroInt(), k.BankKeeper().GetBalance(ctx, kiiAddr1, "ukii").Amount)
	require.Equal(t, sdk.ZeroInt(), k.BankKeeper().GetBalance(ctx, sdk.AccAddress(evmAddr2[:]), "ukii").Amount)
	require.Equal(t, sdk.ZeroInt(), k.BankKeeper().GetBalance(ctx, kiiAddr2, "ukii").Amount)
	require.Equal(t, sdk.ZeroInt(), k.BankKeeper().GetBalance(ctx, sdk.AccAddress(evmAddr3[:]), "ukii").Amount)
	require.Equal(t, sdk.NewInt(10), k.BankKeeper().GetBalance(ctx, kiiAddr3, "ukii").Amount)
}
