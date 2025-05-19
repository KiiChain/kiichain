package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func (p Precompile) transfer(ctx sdk.Context, method *abi.Method, args []interface{}, caller common.Address) ([]byte, error) {
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

	return method.Outputs.Pack(true)
}

func (p Precompile) transferWithDefaultTimeout(ctx sdk.Context, method *abi.Method, args []interface{}, caller common.Address) ([]byte, error) {
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

	return method.Outputs.Pack(true)
}
