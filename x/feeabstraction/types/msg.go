package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate the interface implementation
var (
	_ sdk.Msg = (*MsgUpdateParams)(nil)
	_ sdk.Msg = (*MsgUpdateFeeTokens)(nil)

	// Define the types for the events
	TypeEventConvertFees           = "convert_fees"
	TypeAttributeFeePayer          = "fee_payer"
	TypeAttributeOriginalFeeAmount = "original_fee"
	TypeAttributeConvertedFee      = "converted_fee"
	TypeAttributePrice             = "price"
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

// NewMessageUpdateFeeTokens creates a new MsgUpdateFeeTokens instance
func NewMessageUpdateFeeTokens(authority string, feeTokens FeeTokenMetadataCollection) *MsgUpdateFeeTokens {
	return &MsgUpdateFeeTokens{
		Authority: authority,
		FeeTokens: feeTokens,
	}
}

// Validate performs basic validation on the MsgUpdateFeeTokens message
func (msg *MsgUpdateFeeTokens) Validate() error {
	// Validate the authority
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return err
	}

	// Validate the fee tokens
	return msg.FeeTokens.Validate()
}
