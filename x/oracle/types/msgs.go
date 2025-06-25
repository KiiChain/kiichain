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
	_ sdk.Msg = &MsgUpdateParams{}
)

// NewMsgAggregateExchangeRateVote creates a MsgAggregateExchangeRateVote instance
func NewMsgAggregateExchangeRateVote(exchangeRate string, feeder sdk.AccAddress, validator sdk.ValAddress) *MsgAggregateExchangeRateVote {
	return &MsgAggregateExchangeRateVote{
		ExchangeRates: exchangeRate,
		Feeder:        feeder.String(),
		Validator:     validator.String(),
	}
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
func NewMsgDelegateFeedConsent(validatorOwner sdk.AccAddress, feederAddress sdk.AccAddress) *MsgDelegateFeedConsent {
	return &MsgDelegateFeedConsent{
		ValidatorOwner: validatorOwner.String(),
		Delegate:       feederAddress.String(),
	}
}

// ValidateBasic implements sdk.Msg interface
// ValidateBasic validates the message content (valid addresses)
func (msg MsgDelegateFeedConsent) ValidateBasic() error {
	// Validate the validator owner address
	_, err := sdk.AccAddressFromBech32(msg.ValidatorOwner)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid validator owner address (%s)", err)
	}

	// Validate delegate address
	_, err = sdk.AccAddressFromBech32(msg.Delegate)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid delegate address (%s)", err)
	}

	return nil
}
