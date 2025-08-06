package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Verify interface at compile time
var (
	_ sdk.Msg = (*MsgUpdateParams)(nil)
	_ sdk.Msg = (*MsgFundPool)(nil)
	_ sdk.Msg = (*MsgChangeSchedule)(nil)
)

// NewMsgUpdateParams returns a new MsgUpdateParams with the authority
// and the new params.
func NewMsgUpdateParams(authority string, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority,
		Params:    params,
	}
}

// NewMsgFundPool returns a new MsgFundPool with a sender and
// an amount.
func NewMsgFundPool(sender sdk.AccAddress, amount sdk.Coin) *MsgFundPool {
	return &MsgFundPool{
		Sender: sender.String(),
		Amount: amount,
	}
}

// NewMsgChangeSchedule returns a new MsgChangeSchedule with the authority,
// and a new schedule.
func NewMsgChangeSchedule(authority string, schedule ReleaseSchedule) *MsgChangeSchedule {
	return &MsgChangeSchedule{
		Authority: authority,
		Schedule:  schedule,
	}
}
