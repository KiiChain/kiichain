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
	price := other.Quo(base)

	// Return the calculated price
	return price, nil
}

// TokenToMinimalDenom converts a token amount to its minimal denomination
func TokenToMinimalDenom(amount math.LegacyDec, decimals uint64) (math.Int, error) {
	// Check if the decimals are valid
	if decimals == 0 {
		return math.Int{}, fmt.Errorf("invalid decimals: must be > 0")
	}

	// Calculate the factor to convert to minimal denom
	factor := math.LegacyNewDec(10).Power(decimals)

	// Convert the amount to minimal denom
	minimalDenom := amount.Mul(factor)

	// Return the minimal denom as an integer
	return minimalDenom.TruncateInt(), nil
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
