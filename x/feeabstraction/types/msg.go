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

// Validate performs basic validation on the MsgUpdateParams message
func (msg *MsgUpdateParams) Validate() error {
	// Validate the authority
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return err
	}

	// Validate the params
	return msg.Params.Validate()
}
