package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/feeabstraction module errors
var (
	ErrInvalidFeeTokenMetadata = errorsmod.Register(ModuleName, 1, "invalid fee token metadata")
	ErrInvalidParams           = errorsmod.Register(ModuleName, 2, "invalid fee abstraction params")
)
