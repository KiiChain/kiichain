package antedecorators_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain3/app/antedecorators"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestPriorityAnteDecorator(t *testing.T) {
	output = ""
	anteDecorators := []sdk.AnteFullDecorator{
		sdk.DefaultWrappedAnteDecorator(antedecorators.NewPriorityDecorator()),
	}
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, nil)
	chainedHandler, _ := sdk.ChainAnteDecorators(anteDecorators...)
	// test with normal priority
	newCtx, err := chainedHandler(
		ctx.WithPriority(125),
		FakeTx{},
		false,
	)
	require.NoError(t, err)
	require.Equal(t, int64(125), newCtx.Priority())
}

// func TestPriorityAnteDecoratorTooHighPriority(t *testing.T) {
// 	output = ""
// 	anteDecorators := []sdk.AnteFullDecorator{
// 		sdk.DefaultWrappedAnteDecorator(antedecorators.NewPriorityDecorator()),
// 	}
// 	ctx := sdk.NewContext(nil, tmproto.Header{}, false, nil)
// 	chainedHandler, _ := sdk.ChainAnteDecorators(anteDecorators...)
// 	// test with too high priority, should be auto capped
// 	newCtx, err := chainedHandler(
// 		ctx.WithPriority(math.MaxInt64-50),
// 		FakeTx{
// 			FakeMsgs: []sdk.Msg{
// 				&oracletypes.MsgDelegateFeedConsent{},
// 			},
// 		},
// 		false,
// 	)
// 	require.NoError(t, err)
// 	require.Equal(t, int64(math.MaxInt64-1000), newCtx.Priority())
// }
