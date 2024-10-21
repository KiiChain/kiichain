package ante_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/kiichain/kiichain3/testutil/keeper"
	"github.com/kiichain/kiichain3/x/evm/ante"
	"github.com/kiichain/kiichain3/x/evm/types"
	"github.com/kiichain/kiichain3/x/evm/types/ethtx"
	"github.com/stretchr/testify/require"
)

func TestGasLimitDecorator(t *testing.T) {
	k := &testkeeper.EVMTestApp.EvmKeeper
	ctx := testkeeper.EVMTestApp.GetContextForDeliverTx([]byte{}).WithBlockTime(time.Now())
	a := ante.NewGasLimitDecorator(k)
	limitMsg, _ := types.NewMsgEVMTransaction(&ethtx.LegacyTx{GasLimit: 100})
	ctx, err := a.AnteHandle(ctx, &mockTx{msgs: []sdk.Msg{limitMsg}}, false, func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		return ctx, nil
	})
	require.Nil(t, err)
	require.Equal(t, 100, int(ctx.GasMeter().Limit()))
}
