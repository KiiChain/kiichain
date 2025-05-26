package types

import (
	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ensure Msg interface be implemented at compile time
var (
	_ sdk.Msg = &MsgDelegateFeedConsent{}
	_ sdk.Msg = &MsgAggregateExchangeRateVote{}
)

// NewMsgAggregateExchangeRateVote creates a MsgAggregateExchangeRateVote instance
func NewMsgAggregateExchangeRateVote(exchangeRate string, feeder sdk.AccAddress, validator sdk.ValAddress) *MsgAggregateExchangeRateVote {
	return &MsgAggregateExchangeRateVote{
		ExchangeRates: exchangeRate,
		Feeder:        feeder.String(),
		Validator:     validator.String(),
	}
}

// GetSigners implements sdk.Msg interface
// Returns the signer of the transaction which is the feeder
func (msg MsgAggregateExchangeRateVote) GetSigners() []sdk.AccAddress {
	feeder, err := sdk.AccAddressFromBech32(msg.Feeder)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{feeder}
}

// ValidateBasic implements sdk.Msg interface
// ValidateBasic validates the message content (valid addresses and valid values on exchange rates)
func (msg MsgAggregateExchangeRateVote) ValidateBasic() error {
	// Check valid feeder address
	_, err := sdk.AccAddressFromBech32(msg.Feeder)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid feeder address (%s)", err)
	}

	// Check valid validator address
	_, err = sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid operator address (%s)", err)
	}

	// Check valid quantity exchange rates
	if len(msg.ExchangeRates) == 0 {
		return errors.Wrap(sdkerrors.ErrUnknownRequest, "must provide at least one oracle exchange rate")
	}

	// Check exchange rate size
	if len(msg.ExchangeRates) > 4096 {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "exchange rates string can not exceed 4096 characters")
	}

	exchangeRates, err := ParseExchangeRateTuples(msg.ExchangeRates)
	if err != nil {
		return errors.Wrap(sdkerrors.ErrInvalidCoins, "failed to parse exchange rates string cause: "+err.Error())
	}

	for _, rate := range exchangeRates {
		// Check overflow on exchange rate values
		if rate.ExchangeRate.BigInt().BitLen() > 255+math.LegacyDecimalPrecisionBits {
			return errors.Wrap(ErrInvalidExchangeRate, "overflow exchange rate")
		}
	}
	return nil
}

// NewMsgDelegateFeedConsent creates a MsgDelegateFeedConsent instance
func NewMsgDelegateFeedConsent(operatorAddress sdk.ValAddress, feederAddress sdk.AccAddress) *MsgDelegateFeedConsent {
	return &MsgDelegateFeedConsent{
		Operator: operatorAddress.String(),
		Delegate: feederAddress.String(),
	}
}

// GetSigners implements sdk.Msg interface
// Returns the signer of the transaction which is the feeder
func (msg MsgDelegateFeedConsent) GetSigners() []sdk.AccAddress {
	operator, err := sdk.ValAddressFromBech32(msg.Operator)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{sdk.AccAddress(operator)}
}

// ValidateBasic implements sdk.Msg interface
// ValidateBasic validates the message content (valid addresses)
func (msg MsgDelegateFeedConsent) ValidateBasic() error {
	// Validate operator (validator) account
	_, err := sdk.ValAddressFromBech32(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid operator address (%s)", err)
	}

	// Validate delegate address
	_, err = sdk.AccAddressFromBech32(msg.Delegate)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid delegate address (%s)", err)
	}

	return nil
}
