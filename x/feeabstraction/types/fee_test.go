package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
	"github.com/stretchr/testify/require"
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

// TestTokenToMinimalDenom tests the TokenToMinimalDenom function
func TestTokenToMinimalDenom(t *testing.T) {
	// Prepare the test cases
	testCases := []struct {
		amount      math.LegacyDec
		decimals    uint64
		expected    math.Int
		errContains string
	}{
		{
			amount:   math.LegacyNewDec(1),
			decimals: 2,
			expected: math.NewInt(100),
		},
		{
			amount:   math.LegacyMustNewDecFromStr("0.1"),
			decimals: 3,
			expected: math.NewInt(100),
		},
		{
			amount:      math.LegacyNewDec(1),
			decimals:    0,
			errContains: "invalid decimals: must be > 0",
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		result, err := types.TokenToMinimalDenom(tc.amount, tc.decimals)

		if tc.errContains != "" {
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errContains)
		} else {
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		}
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
