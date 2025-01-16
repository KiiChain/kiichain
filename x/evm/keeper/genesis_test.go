package keeper_test

import (
	"bytes"
	"testing"

	testkeeper "github.com/kiichain/kiichain3/testutil/keeper"
	"github.com/kiichain/kiichain3/x/evm/keeper"
	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
	k := &testkeeper.EVMTestApp.EvmKeeper
	ctx := testkeeper.EVMTestApp.GetContextForDeliverTx([]byte{})
	// coinbase address must be associated
	coinbaseKiiAddr, associated := k.GetKiiAddress(ctx, keeper.GetCoinbaseAddress())
	require.True(t, associated)
	require.True(t, bytes.Equal(coinbaseKiiAddr, k.AccountKeeper().GetModuleAddress("fee_collector")))
}
