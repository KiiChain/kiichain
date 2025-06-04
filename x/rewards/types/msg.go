package types

import (
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Verify interface at compile time
var (
	_ sdk.Msg = (*MsgUpdateParams)(nil)
	_ sdk.Msg = (*MsgFundPool)(nil)
	_ sdk.Msg = (*MsgExtendReward)(nil)
)

// NewMsgUpdateParams returns a new MsgUpdateParams with the authority
// and the new params.
func NewMsgUpdateParams(authority sdk.AccAddress, params Params) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: authority.String(),
		Params:    params,
	}
}

// NewMsgFundPool returns a new MsgFundPool with a sender and
// an amount.
func NewMsgFundPool(sender sdk.AccAddress, amount sdk.Coin, endTime time.Time) *MsgFundPool {
	return &MsgFundPool{
		Sender: sender.String(),
		Amount: amount,
	}
}

// NewMsgExtendReward returns a new MsgExtendReward with the authority,
// an amount to extend and a new endTime.
func NewMsgExtendReward(authority sdk.AccAddress, extendAmount sdk.Coin, endTime time.Time) *MsgExtendReward {
	return &MsgExtendReward{
		Authority:   authority.String(),
		ExtraAmount: extendAmount,
		EndTime:     endTime,
	}
}
