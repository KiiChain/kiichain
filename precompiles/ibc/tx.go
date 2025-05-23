package ibc

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Transfer does a IBC transfer with custom timeout options
func (p Precompile) Transfer(ctx sdk.Context, method *abi.Method, stateDB vm.StateDB, args []interface{}, caller common.Address) ([]byte, error) {
	// Build and validate msg
	msg, err := NewMsgTransfer(ctx, method, caller, args)
	if err != nil {
		return nil, err
	}

	// Log the call
	p.logTransfer(ctx, method, msg)

	// Transfer
	_, err = p.transferKeeper.Transfer(ctx, msg)
	if err != nil {
		return nil, err
	}

	// Emit the event
	err = p.EmitEventTransfer(ctx, stateDB, caller, msg.Receiver, msg.Token.Denom, msg.SourcePort, msg.SourceChannel, msg.Token.Amount.BigInt(), msg.TimeoutHeight, msg.TimeoutTimestamp)
	if err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

// TransferWithDefaultTimeout does a IBC transfer with default timeout options
func (p Precompile) TransferWithDefaultTimeout(ctx sdk.Context, method *abi.Method, stateDB vm.StateDB, args []interface{}, caller common.Address) ([]byte, error) {
	// Build and validate message
	msg, err := p.NewMsgTransferDefaultTimeout(ctx, method, caller, args)
	if err != nil {
		return nil, err
	}

	// Log the call
	p.logTransfer(ctx, method, msg)

	// Transfer
	_, err = p.transferKeeper.Transfer(ctx, msg)
	if err != nil {
		return nil, err
	}

	// Emit the event
	err = p.EmitEventTransfer(ctx, stateDB, caller, msg.Receiver, msg.Token.Denom, msg.SourcePort, msg.SourceChannel, msg.Token.Amount.BigInt(), msg.TimeoutHeight, msg.TimeoutTimestamp)
	if err != nil {
		return nil, err
	}

	return method.Outputs.Pack(true)
}

func (p Precompile) logTransfer(ctx sdk.Context, method *abi.Method, msg *types.MsgTransfer) {
	p.Logger(ctx).Debug(
		"tx called",
		"method", method.Name,
		"args", fmt.Sprintf(
			"{ sender: %s, receiver: %s, port: %s, channel: %s, token: %s%s, heght: %d of number %d, timeoutStamp: %d, memo: %s }",
			msg.Sender, msg.Receiver, msg.SourcePort, msg.SourceChannel, msg.Token.Amount, msg.Token.Denom,
			msg.TimeoutHeight.RevisionHeight, msg.TimeoutHeight.RevisionNumber, msg.TimeoutTimestamp, msg.Memo,
		),
	)
}
