package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Verify interface at compile time
var (
	_ sdk.Msg = (*MsgUpdateParams)(nil)
)

// NewMsgUpdateParams returns a new MsgUpdateParams with the authority
// and the new params.
func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}
