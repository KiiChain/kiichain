package keepers

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	evmtypes "github.com/cosmos/evm/x/vm/types"
)

// stubMessageRouter is a minimal baseapp.MessageRouter used to assert that the
// evmRejectingMessageRouter delegates non-EVM messages to the wrapped router.
type stubMessageRouter struct {
	handler            baseapp.MsgServiceHandler
	gotMsg             sdk.Msg
	gotTypeURL         string
	handlerCalled      bool
	handlerByURLCalled bool
}

func (s *stubMessageRouter) Handler(msg sdk.Msg) baseapp.MsgServiceHandler {
	s.handlerCalled = true
	s.gotMsg = msg
	return s.handler
}

func (s *stubMessageRouter) HandlerByTypeURL(typeURL string) baseapp.MsgServiceHandler {
	s.handlerByURLCalled = true
	s.gotTypeURL = typeURL
	return s.handler
}

func TestEVMRejectingRouterHandlerRejectsEVMMsg(t *testing.T) {
	stub := &stubMessageRouter{handler: func(sdk.Context, sdk.Msg) (*sdk.Result, error) { return &sdk.Result{}, nil }}
	router := newEVMRejectingMessageRouter(stub)

	handler := router.Handler(&evmtypes.MsgEthereumTx{})
	require.NotNil(t, handler)
	require.False(t, stub.handlerCalled, "EVM messages must not reach the wrapped router")

	_, err := handler(sdk.Context{}, &evmtypes.MsgEthereumTx{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not allowed to be executed through authz")
}

func TestEVMRejectingRouterHandlerDelegatesNonEVMMsg(t *testing.T) {
	sentinel := func(sdk.Context, sdk.Msg) (*sdk.Result, error) { return &sdk.Result{}, nil }
	stub := &stubMessageRouter{handler: sentinel}
	router := newEVMRejectingMessageRouter(stub)

	msg := &banktypes.MsgSend{}
	handler := router.Handler(msg)

	require.NotNil(t, handler)
	require.True(t, stub.handlerCalled)
	require.Equal(t, msg, stub.gotMsg)
}

func TestEVMRejectingRouterHandlerByTypeURLRejectsEVM(t *testing.T) {
	stub := &stubMessageRouter{handler: func(sdk.Context, sdk.Msg) (*sdk.Result, error) { return &sdk.Result{}, nil }}
	router := newEVMRejectingMessageRouter(stub)

	require.Nil(t, router.HandlerByTypeURL(evmMsgTypeURL))
	require.False(t, stub.handlerByURLCalled, "EVM type URL must not reach the wrapped router")
}

func TestEVMRejectingRouterHandlerByTypeURLDelegatesNonEVM(t *testing.T) {
	sentinel := func(sdk.Context, sdk.Msg) (*sdk.Result, error) { return &sdk.Result{}, nil }
	stub := &stubMessageRouter{handler: sentinel}
	router := newEVMRejectingMessageRouter(stub)

	typeURL := sdk.MsgTypeURL(&banktypes.MsgSend{})
	handler := router.HandlerByTypeURL(typeURL)

	require.NotNil(t, handler)
	require.True(t, stub.handlerByURLCalled)
	require.Equal(t, typeURL, stub.gotTypeURL)
}
