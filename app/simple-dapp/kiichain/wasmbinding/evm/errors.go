package evm

import (
	errorsmod "cosmossdk.io/errors"
)

const (
	ErrorCodespace = "evm_wasmbinding"
)

var (
	ErrExecutionReverted = errorsmod.Register(ErrorCodespace, 1, "execution reverted")
	ErrExecutingEthCall  = errorsmod.Register(ErrorCodespace, 2, "error executing eth call")
)
