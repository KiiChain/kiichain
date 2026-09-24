package keepers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type mockCircuitBreaker struct {
	denied map[string]bool
}

func (m mockCircuitBreaker) IsAllowed(_ context.Context, typeURL string) (bool, error) {
	return !m.denied[typeURL], nil
}

type mockStakingMsgServer struct {
	stakingtypes.UnimplementedMsgServer
	called bool
}

func (m *mockStakingMsgServer) Delegate(context.Context, *stakingtypes.MsgDelegate) (*stakingtypes.MsgDelegateResponse, error) {
	m.called = true
	return &stakingtypes.MsgDelegateResponse{}, nil
}

func TestCircuitStakingMsgServerBlocksTrippedMsgs(t *testing.T) {
	inner := &mockStakingMsgServer{}
	msgURL := sdk.MsgTypeURL(&stakingtypes.MsgDelegate{})
	server := newCircuitStakingMsgServer(inner, mockCircuitBreaker{
		denied: map[string]bool{msgURL: true},
	})

	_, err := server.Delegate(context.Background(), &stakingtypes.MsgDelegate{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "circuit breaker disables execution")
	require.False(t, inner.called)

	inner.called = false
	server = newCircuitStakingMsgServer(inner, mockCircuitBreaker{})
	_, err = server.Delegate(context.Background(), &stakingtypes.MsgDelegate{})
	require.NoError(t, err)
	require.True(t, inner.called)
}
