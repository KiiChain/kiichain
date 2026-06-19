package keepers

import (
	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	evmtypes "github.com/cosmos/evm/x/vm/types"
)

// evmMsgTypeURL is the type URL of MsgEthereumTx, cached once so it can be
// compared cheaply when routing messages.
var evmMsgTypeURL = sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{})

// evmRejectingMessageRouter wraps a baseapp.MessageRouter and prevents EVM
// transactions (MsgEthereumTx) from being executed through authz. EVM messages
// carry their own signer/nonce semantics that authz grants are not designed to
// authorize, so dispatching them via authz is blocked at the routing layer.
type evmRejectingMessageRouter struct {
	baseapp.MessageRouter
}

// newEVMRejectingMessageRouter wraps the given router so that any attempt to
// route an EVM message is rejected. All non-EVM messages are delegated to the
// underlying router unchanged.
func newEVMRejectingMessageRouter(router baseapp.MessageRouter) baseapp.MessageRouter {
	return evmRejectingMessageRouter{MessageRouter: router}
}

// Handler returns the message handler for msg. For MsgEthereumTx it returns a
// handler that always fails with an unauthorized error; for every other message
// it falls through to the wrapped router.
func (r evmRejectingMessageRouter) Handler(msg sdk.Msg) baseapp.MsgServiceHandler {
	if _, ok := msg.(*evmtypes.MsgEthereumTx); ok {
		return func(sdk.Context, sdk.Msg) (*sdk.Result, error) {
			return nil, errorsmod.Wrapf(errortypes.ErrUnauthorized, "%s is not allowed to be executed through authz", evmMsgTypeURL)
		}
	}
	return r.MessageRouter.Handler(msg)
}

// HandlerByTypeURL returns the handler registered for typeURL. It returns nil
// for the EVM message type URL so that resolution fails, and otherwise defers to
// the wrapped router.
func (r evmRejectingMessageRouter) HandlerByTypeURL(typeURL string) baseapp.MsgServiceHandler {
	if typeURL == evmMsgTypeURL {
		return nil
	}
	return r.MessageRouter.HandlerByTypeURL(typeURL)
}
