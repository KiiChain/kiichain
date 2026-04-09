package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/app/params"
)

var (
	// Default values for the fee abstraction parameters
	DefaultClampFactor        = math.LegacyMustNewDecFromStr("0.10") // 10%
	DefaultTwapLookbackWindow = uint64(120)                          // 120 seconds (2 minutes)
)

// NewParams returns a new params instance
func NewParams(
	nativeDenom, nativeOracleDenom string,
	clampFactor math.LegacyDec,
	twapLookbackWindow uint64,
	enabled bool,
) Params {
	return Params{
		NativeDenom:        nativeDenom,
		NativeOracleDenom:  nativeOracleDenom,
		ClampFactor:        clampFactor,
		Enabled:            enabled,
		TwapLookbackWindow: twapLookbackWindow,
	}
}

// DefaultParams returns default params
func DefaultParams() Params {
	return Params{
		NativeDenom:        params.BaseDenom,
		NativeOracleDenom:  params.DisplayDenom,
		ClampFactor:        DefaultClampFactor,
		TwapLookbackWindow: DefaultTwapLookbackWindow,
		Enabled:            true,
	}
}

// Validate performs basic validation on distribution parameters.
func (p Params) Validate() error {
	// Validate the native denom
	if err := sdk.ValidateDenom(p.NativeDenom); err != nil {
		return errorsmod.Wrap(ErrInvalidParams, "native denom is invalid")
	}

	// Validate the native oracle denom
	if err := sdk.ValidateDenom(p.NativeOracleDenom); err != nil {
		return errorsmod.Wrap(ErrInvalidParams, "native oracle denom is invalid")
	}

	// Validate the clamp factor
	if p.ClampFactor.IsNegative() || p.ClampFactor.GT(math.LegacyOneDec()) {
		return errorsmod.Wrap(ErrInvalidParams, "clamp factor must be between 0 and 1")
	}

	// Validate the twap lookback window
	if p.TwapLookbackWindow == 0 {
		return errorsmod.Wrap(ErrInvalidParams, "twap lookback window must be greater than 0")
	}

	return nil
}

// NewFeeTokenMetadata creates a new fee token with the given denom and address
func NewFeeTokenMetadata(
	denom, oracleDenom string,
	decimals uint32,
	price math.LegacyDec,
) FeeTokenMetadata {
	return FeeTokenMetadata{
		Denom:       denom,
		OracleDenom: oracleDenom,
		Decimals:    decimals,
		Price:       price,
		Enabled:     true,
	}
}

// Validate validates the FeeTokenMetadata
func (f FeeTokenMetadata) Validate() error {
	// Validate the denom
	if err := sdk.ValidateDenom(f.Denom); err != nil {
		return errorsmod.Wrap(ErrInvalidFeeTokenMetadata, "denom is invalid")
	}
	// Validate the oracle denom
	if err := sdk.ValidateDenom(f.OracleDenom); err != nil {
		return errorsmod.Wrap(ErrInvalidFeeTokenMetadata, "oracle denom is invalid")
	}

	// Validate the decimals, must be between bigger than 0 and less than or equal to 18
	if f.Decimals < 1 || f.Decimals > 18 {
		return errorsmod.Wrap(ErrInvalidFeeTokenMetadata, "decimals must be between 1 and 18")
	}

	// Validate the price, must be greater than 0
	if f.Price.IsNegative() || f.Price.IsZero() {
		return errorsmod.Wrap(ErrInvalidFeeTokenMetadata, "price must be greater than 0")
	}

	return nil
}

// NewFeeTokenMetadataCollection creates a new FeeTokenMetadataCollection
func NewFeeTokenMetadataCollection(feeTokens ...FeeTokenMetadata) *FeeTokenMetadataCollection {
	return &FeeTokenMetadataCollection{
		Items: feeTokens,
	}
}

// Validate validates the FeeTokenMetadataCollection
func (c *FeeTokenMetadataCollection) Validate() error {
	// Check if the collection is nil
	if c == nil {
		return errorsmod.Wrap(ErrInvalidFeeTokenMetadata, "fee token metadata collection cannot be nil")
	}

	// Validate each fee token metadata and check for duplicates
	denomSet := make(map[string]struct{})
	for _, token := range c.Items {
		if err := token.Validate(); err != nil {
			return err
		}
		if _, exists := denomSet[token.Denom]; exists {
			return errorsmod.Wrapf(ErrInvalidFeeTokenMetadata, "duplicate denom found: %s", token.Denom)
		}
		denomSet[token.Denom] = struct{}{}
	}

	return nil
}

// NewUpdateTokenMetadata creates a new fee token with the given denom and decimals
func NewUpdateTokenMetadata(
	denom, oracleDenom string,
	decimals uint32,
) UpdateTokenMetadata {
	return UpdateTokenMetadata{
		Denom:       denom,
		OracleDenom: oracleDenom,
		Decimals:    decimals,
		Enabled:     true,
	}
}

// Validate validates the UpdateTokenMetadata
func (f UpdateTokenMetadata) Validate() error {
	// Validate the denom
	if err := sdk.ValidateDenom(f.Denom); err != nil {
		return errorsmod.Wrap(ErrInvalidFeeTokenMetadata, "denom is invalid")
	}
	// Validate the oracle denom
	if err := sdk.ValidateDenom(f.OracleDenom); err != nil {
		return errorsmod.Wrap(ErrInvalidFeeTokenMetadata, "oracle denom is invalid")
	}

	// Validate the decimals, must be between bigger than 0 and less than or equal to 18
	if f.Decimals < 1 || f.Decimals > 18 {
		return errorsmod.Wrap(ErrInvalidFeeTokenMetadata, "decimals must be between 1 and 18")
	}

	return nil
}

// NewUpdateTokenMetadataCollection creates a new UpdateTokenMetadataCollection
func NewUpdateTokenMetadataCollection(tokens ...UpdateTokenMetadata) *UpdateTokenMetadataCollection {
	return &UpdateTokenMetadataCollection{
		Items: tokens,
	}
}

// Validate validates the UpdateTokenMetadataCollection
func (c *UpdateTokenMetadataCollection) Validate() error {
	// Check if the collection is nil
	if c == nil {
		return errorsmod.Wrap(ErrInvalidFeeTokenMetadata, "fee token metadata collection cannot be nil")
	}

	// Validate each fee token metadata and check for duplicates
	denomSet := make(map[string]struct{})
	for _, token := range c.Items {
		if err := token.Validate(); err != nil {
			return err
		}
		if _, exists := denomSet[token.Denom]; exists {
			return errorsmod.Wrapf(ErrInvalidFeeTokenMetadata, "duplicate denom found: %s", token.Denom)
		}
		denomSet[token.Denom] = struct{}{}
	}

	return nil
}
