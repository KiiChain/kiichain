package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

// transfer does a IBC transfer with custom timeout options
func (p Precompile) transfer(ctx sdk.Context, method *abi.Method, stateDB vm.StateDB, args []interface{}, caller common.Address) ([]byte, error) {
	// Build and validate msg
	msg, err := NewMsgTransfer(ctx, method, caller, args)
	if err != nil {
		return nil, err
	}

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

// transferWithDefaultTimeout does a IBC transfer with default timeout options
func (p Precompile) transferWithDefaultTimeout(ctx sdk.Context, method *abi.Method, stateDB vm.StateDB, args []interface{}, caller common.Address) ([]byte, error) {
	// Build and validate message
	msg, err := p.NewMsgTransferDefaultTimeout(ctx, method, caller, args)
	if err != nil {
		return nil, err
	}

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
