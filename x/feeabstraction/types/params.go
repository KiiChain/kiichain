package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/kiichain/kiichain/v3/app/params"
)

var (
	// Default values for the fee abstraction parameters
	DefaultMaxPriceDeviation = math.LegacyMustNewDecFromStr("0.1") // 10%
	DefaultClampFactor       = math.LegacyMustNewDecFromStr("0.1") // 10%

)

// NewParams returns a new params instance
func NewParams(
	nativeDenom string,
	maxPriceDeviation math.LegacyDec,
	clampFactor math.LegacyDec,
	enabled bool,
) Params {
	return Params{
		NativeDenom:       nativeDenom,
		MaxPriceDeviation: maxPriceDeviation,
		ClampFactor:       clampFactor,
		Enabled:           enabled,
	}
}

// DefaultParams returns default params
func DefaultParams() Params {
	return Params{
		NativeDenom:       params.BaseDenom,
		MaxPriceDeviation: DefaultMaxPriceDeviation,
		ClampFactor:       DefaultClampFactor,
		Enabled:           true,
	}
}

// Validate performs basic validation on distribution parameters.
func (p Params) Validate() error {
	// Validate the native denom
	if err := sdk.ValidateDenom(p.NativeDenom); err != nil {
		return errorsmod.Wrap(ErrInvalidParams, "native denom is invalid")
	}

	// Validate the max price deviation
	if p.MaxPriceDeviation.IsNegative() || p.MaxPriceDeviation.GT(math.LegacyOneDec()) {
		return errorsmod.Wrap(ErrInvalidParams, "max price deviation must be between 0 and 1")
	}

	// Validate the clamp factor
	if p.ClampFactor.IsNegative() || p.ClampFactor.GT(math.LegacyOneDec()) {
		return errorsmod.Wrap(ErrInvalidParams, "clamp factor must be between 0 and 1")
	}

	return nil
}

// NewFeeTokenMetadata creates a new fee token with the given denom and address
func NewFeeTokenMetadata(
	denom, oracleDenom string,
	decimals uint32,
	price, fallbackPrice math.LegacyDec,
) FeeTokenMetadata {
	return FeeTokenMetadata{
		Denom:         denom,
		OracleDenom:   oracleDenom,
		Decimals:      decimals,
		Price:         price,
		FallbackPrice: fallbackPrice,
		Enabled:       true,
	}
}

// Validate validates the FeeTokenMetadata
func (f FeeTokenMetadata) Validate() error {
	// Validate the denom
	if err := sdk.ValidateDenom(f.Denom); err != nil {
		return err
	}
	// Validate the oracle denom
	if err := sdk.ValidateDenom(f.OracleDenom); err != nil {
		return err
	}

	// Validate the decimals, must be between bigger than 0 and less than or equal to 18
	if f.Decimals < 1 || f.Decimals > 18 {
		return errorsmod.Wrap(ErrInvalidFeeTokenMetadata, "decimals must be between 1 and 18")
	}

	// Validate the price and fallback price, must be greater than 0
	if f.Price.IsNegative() || f.Price.IsZero() {
		return errorsmod.Wrap(ErrInvalidFeeTokenMetadata, "price must be greater than 0")
	}
	if f.FallbackPrice.IsNegative() || f.FallbackPrice.IsZero() {
		return errorsmod.Wrap(ErrInvalidFeeTokenMetadata, "fallback price must be greater than 0")
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
