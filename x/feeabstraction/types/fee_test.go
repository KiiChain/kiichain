package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
)

// TestCalculateTokenPrice tests the CalculateTokenPrice function
func TestCalculateTokenPrice(t *testing.T) {
	// Prepare the test cases
	testCases := []struct {
		base        math.LegacyDec
		other       math.LegacyDec
		expected    math.LegacyDec
		errContains string
	}{
		{
			base:     math.LegacyNewDec(100),
			other:    math.LegacyNewDec(200),
			expected: math.LegacyNewDec(2),
		},
		{
			// Simulate a real situation, where kii is worth 0.1 USD and the asset B is worth 15 USD
			// This would mean that 1 KII = 150 TokenB
			base:     math.LegacyMustNewDecFromStr("0.1"),
			other:    math.LegacyMustNewDecFromStr("15"),
			expected: math.LegacyMustNewDecFromStr("150"),
		},
		{
			base:        math.LegacyNewDec(0),
			other:       math.LegacyNewDec(100),
			errContains: "invalid input: base or other is zero",
		},
		{
			base:        math.LegacyNewDec(100),
			other:       math.LegacyNewDec(0),
			expected:    math.LegacyDec{},
			errContains: "invalid input: base or other is zero",
		},
		{
			base:        math.LegacyNewDec(0),
			other:       math.LegacyNewDec(0),
			expected:    math.LegacyDec{},
			errContains: "invalid input: base or other is zero",
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		// Calculate the token price
		result, err := types.CalculateTokenPrice(tc.base, tc.other)

		// Check for expected error
		if tc.errContains != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errContains)
		} else {
			require.NoError(t, err)
			// Check if the result matches the expected value
			require.Equal(t, tc.expected, result)
		}
	}
}

// TestCalculateTokenAmountWithDecimals tests the CalculateTokenAmountWithDecimals function
func TestCalculateTokenAmountWithDecimals(t *testing.T) {
	// Prepare the test cases
	testCases := []struct {
		name          string
		price         math.LegacyDec
		amount        math.Int
		decimalsBase  uint64
		decimalsOther uint64
		expected      math.LegacyDec
		errContains   string
	}{
		{
			// Both tokens have 2 decimals, the price is 10, and the amount is 123
			// The expected result is 1230
			name:          "Same decimals, simple price",
			price:         math.LegacyNewDec(10),
			amount:        math.NewInt(123),
			decimalsBase:  2,
			decimalsOther: 2,
			expected:      math.LegacyMustNewDecFromStr("1230"),
		},
		{
			// Now lets imagine that we have USD with price of 10 per KII
			// and we want to convert 123 KII to USD
			// But the USD has 2 decimals and KII has 4 decimals
			// It should return 12.3
			name:          "different decimals, (Kii 4, USD 2)",
			price:         math.LegacyMustNewDecFromStr("10"),
			amount:        math.NewInt(123),
			decimalsBase:  4,
			decimalsOther: 2,
			expected:      math.LegacyMustNewDecFromStr("12.3"),
		},
		{
			// Now lets imagine the same situation as `Same decimals, simple price`
			// but with 6 decimals for USD and 18 decimals for KII
			// The result should be 1230*10^6
			name:          "different decimals, (Kii 18, USD 6)",
			price:         math.LegacyMustNewDecFromStr("10"),
			amount:        math.LegacyNewDec(123).Mul(math.LegacyNewDec(1e18)).TruncateInt(),
			decimalsBase:  18,
			decimalsOther: 6,
			expected:      math.LegacyMustNewDecFromStr("1230000000"),
		},
		{
			// Test with zero price, should return zero
			name:          "zero price",
			price:         math.LegacyNewDec(0),
			amount:        math.NewInt(123),
			decimalsBase:  2,
			decimalsOther: 2,
			expected:      math.LegacyZeroDec(),
		},
		{
			// Test with zero amount, should return zero
			name:          "zero amount",
			price:         math.LegacyNewDec(10),
			amount:        math.NewInt(0),
			decimalsBase:  2,
			decimalsOther: 2,
			expected:      math.LegacyZeroDec(),
		},
		{
			// Test with zero decimals, should return an error
			name:          "zero decimals",
			price:         math.LegacyNewDec(10),
			amount:        math.NewInt(123),
			decimalsBase:  0,
			decimalsOther: 2,
			errContains:   "invalid decimals: must be > 0",
		},
		{
			name:          "zero decimals other",
			price:         math.LegacyNewDec(10),
			amount:        math.NewInt(123),
			decimalsBase:  2,
			decimalsOther: 0,
			errContains:   "invalid decimals: must be > 0",
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate the token price with decimals
			result, err := types.CalculateTokenAmountWithDecimals(tc.price, tc.amount, tc.decimalsBase, tc.decimalsOther)

			// Check for expected error
			if tc.errContains != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
				// Check if the result matches the expected value
				require.Equal(t, tc.expected, result)
			}
		})
	}
}

// TestClampPrice tests the ClampPrice function
func TestClampPrice(t *testing.T) {
	// Prepare the test cases
	testCases := []struct {
		prevPrice   math.LegacyDec
		newPrice    math.LegacyDec
		clampFactor math.LegacyDec
		expected    math.LegacyDec
	}{
		{
			prevPrice:   math.LegacyNewDec(100),
			newPrice:    math.LegacyNewDec(110),
			clampFactor: math.LegacyMustNewDecFromStr("0.1"), // 10%
			expected:    math.LegacyNewDec(110),
		},
		{
			prevPrice:   math.LegacyNewDec(100),
			newPrice:    math.LegacyNewDec(90),
			clampFactor: math.LegacyMustNewDecFromStr("0.1"), // 10%
			expected:    math.LegacyNewDec(90),
		},
		{
			prevPrice:   math.LegacyNewDec(100),
			newPrice:    math.LegacyNewDec(80),
			clampFactor: math.LegacyMustNewDecFromStr("0.1"), // 10%
			expected:    math.LegacyNewDec(90),               // Clamped to min bound
		},
		{
			prevPrice:   math.LegacyNewDec(100),
			newPrice:    math.LegacyNewDec(120),
			clampFactor: math.LegacyMustNewDecFromStr("0.1"), // 10%
			expected:    math.LegacyNewDec(110),              // Clamped to max bound
		},
		{
			prevPrice:   math.LegacyNewDec(0), // prevPrice is zero,
			newPrice:    math.LegacyNewDec(100),
			clampFactor: math.LegacyMustNewDecFromStr("0.1"), // 10%
			expected:    math.LegacyNewDec(100),              // Should return newPrice
		},
		{
			prevPrice:   math.LegacyNewDec(100), // prevPrice is non-zero, clampFactor is zero
			newPrice:    math.LegacyNewDec(100),
			clampFactor: math.LegacyZeroDec(),
			expected:    math.LegacyNewDec(100), // Should return newPrice
		},
	}

	for _, tc := range testCases {
		result := types.ClampPrice(tc.prevPrice, tc.newPrice, tc.clampFactor)
		require.Equal(t, tc.expected, result)
	}
}
