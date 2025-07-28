package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate the interface implementation
var (
	_ sdk.Msg = (*MsgUpdateParams)(nil)
)

// NewMessageUpdateParams creates a new MsgUpdateParams instance
func NewMessageUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}
