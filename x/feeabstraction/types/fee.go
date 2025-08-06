package types

import (
	fmt "fmt"

	"cosmossdk.io/math"
)

// CalculateTokenPrice calculates the price of a fee token in terms of the base token
func CalculateTokenPrice(
	base math.LegacyDec,
	other math.LegacyDec,
) (math.LegacyDec, error) {
	// Check for zero values
	if base.IsZero() || other.IsZero() {
		return math.LegacyDec{}, fmt.Errorf("invalid input: base or other is zero")
	}

	// Get the quotient between the two tokens
	price := base.Quo(other)

	// Return the calculated price
	return price, nil
}

// CalculateTokenAmountWithDecimals calculate the amount give a price and decimals
// This handles the logic for price handling at token minimal amounts
func CalculateTokenAmountWithDecimals(
	price math.LegacyDec,
	amountAtMinimal math.Int,
	decimalsBase uint64,
	decimalsOther uint64,
) (math.LegacyDec, error) {
	// Check if the values are valid
	if decimalsBase == 0 || decimalsOther == 0 {
		return math.LegacyDec{}, fmt.Errorf("invalid decimals: must be > 0")
	}
	if amountAtMinimal.IsZero() || price.IsZero() {
		return math.LegacyZeroDec(), nil
	}

	// Calculate the minimal token to full token
	amountFull := amountAtMinimal.ToLegacyDec().Quo(math.LegacyNewDec(10).Power(decimalsBase))

	// Multiply the amount by the price
	amountInOtherFull := amountFull.Mul(price)

	// Convert the units back
	return amountInOtherFull.Mul(math.LegacyNewDec(10).Power(decimalsOther)), nil
}

// ClampPrice ensures newPrice is within ±clampFactor of prevPrice.
// If prevPrice is zero, returns newPrice unmodified.
func ClampPrice(prevPrice, newPrice, clampFactor math.LegacyDec) math.LegacyDec {
	if prevPrice.IsZero() || clampFactor.IsZero() {
		return newPrice
	}

	// Get the max and min bounds based on the clamp factor
	min := prevPrice.Mul(math.LegacyOneDec().Sub(clampFactor))
	max := prevPrice.Mul(math.LegacyOneDec().Add(clampFactor))

	// Clamp the new price within the bounds
	if newPrice.LT(min) {
		return min
	}
	// If newPrice is greater than max, return max
	if newPrice.GT(max) {
		return max
	}
	// Return the new price as it is within bounds
	return newPrice
}
