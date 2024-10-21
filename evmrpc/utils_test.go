package evmrpc_test

import (
	"context"
	"testing"

	"github.com/kiichain/kiichain3/app"
	"github.com/kiichain/kiichain3/evmrpc"
	"github.com/stretchr/testify/require"
)

func TestCheckVersion(t *testing.T) {
	testApp := app.Setup(false, false)
	k := &testApp.EvmKeeper
	ctx := testApp.GetContextForDeliverTx([]byte{}).WithBlockHeight(1)
	testApp.Commit(context.Background()) // bump store version to 1
	require.Nil(t, evmrpc.CheckVersion(ctx, k))
	ctx = ctx.WithBlockHeight(2)
	require.NotNil(t, evmrpc.CheckVersion(ctx, k))
}
