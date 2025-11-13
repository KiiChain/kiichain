package wasmd_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v6/precompiles/wasmd"
)

// EVMCoin represents a coin in EVM format
type EVMCoin struct {
	Denom  string
	Amount *big.Int
}

// TestConvertEVMCoinsToSDKCoins tests the conversion from EVM coins to SDK coins
func TestConvertEVMCoinsToSDKCoins(t *testing.T) {
	// Prepare the test cases
	testCases := []struct {
		name          string
		input         any
		expected      []sdk.Coin
		errorContains string
	}{
		{
			name:     "valid input with multiple coins",
			input:    []EVMCoin{{Denom: "akii", Amount: big.NewInt(100)}, {Denom: "btc", Amount: big.NewInt(200)}},
			expected: []sdk.Coin{{Denom: "akii", Amount: math.NewInt(100)}, {Denom: "btc", Amount: math.NewInt(200)}},
		},
		{
			name:     "valid coin (duplicate denoms)",
			input:    []EVMCoin{{Denom: "akii", Amount: big.NewInt(100)}, {Denom: "akii", Amount: big.NewInt(200)}},
			expected: []sdk.Coin{{Denom: "akii", Amount: math.NewInt(300)}},
		},
		{
			name:     "valid coin (zero amount)",
			input:    []EVMCoin{{Denom: "akii", Amount: big.NewInt(0)}},
			expected: []sdk.Coin{},
		},
		{
			name:          "invalid input type (not a slice)",
			input:         "invalid input",
			errorContains: "expected slice, got string",
		},
		{
			name:          "invalid input type (slice of ints)",
			input:         []int{1, 2, 3},
			errorContains: "expected slice of structs, got slice of []int",
		},
		{
			name: "invalid amount type in struct",
			input: []struct {
				Denom  string
				Amount math.Int
			}{
				{Denom: "akii", Amount: math.NewInt(100)},
				{Denom: "btc", Amount: math.NewInt(200)},
			},
			errorContains: "amount field must be a *big.Int, got math.Int",
		},
		{
			name:          "invalid input type as SDK.Coin slice",
			input:         []sdk.Coin{{Denom: "akii", Amount: math.NewInt(100)}, {Denom: "btc", Amount: math.NewInt(200)}},
			expected:      []sdk.Coin{{Denom: "akii", Amount: math.NewInt(100)}, {Denom: "btc", Amount: math.NewInt(200)}},
			errorContains: "amount field must be a *big.Int, got math.Int",
		},
		{
			name:          "struct missing Denom field",
			input:         []struct{ Amount *big.Int }{{Amount: big.NewInt(100)}},
			errorContains: "struct missing Denom field",
		},
		{
			name:          "struct missing Amount field",
			input:         []struct{ Denom string }{{Denom: "akii"}},
			errorContains: "struct missing Amount field",
		},
		{
			name: "Denom field not a string",
			input: []struct {
				Denom  int
				Amount *big.Int
			}{{Denom: 123, Amount: big.NewInt(100)}},
			errorContains: "denom field must be a string, got int",
		},
		{
			name: "Amount field not a *big.Int",
			input: []struct {
				Denom  string
				Amount int
			}{{Denom: "akii", Amount: 100}},
			errorContains: "amount field must be a *big.Int, got int",
		},
		{
			name:     "empty slice",
			input:    []EVMCoin{},
			expected: []sdk.Coin{},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: []sdk.Coin{},
		},
		{
			name:          "invalid coin (negative amount)",
			input:         []EVMCoin{{Denom: "akii", Amount: big.NewInt(-100)}},
			errorContains: "invalid coins",
		},
		{
			name:          "invalid coin (empty denom)",
			input:         []EVMCoin{{Denom: "", Amount: big.NewInt(100)}},
			errorContains: "invalid coins",
		},
		{
			name:          "invalid coin (nil amount)",
			input:         []EVMCoin{{Denom: "akii", Amount: nil}},
			errorContains: "amount field must be a *big.Int, got *big.Int",
		},
	}

	// Execute the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Execute the conversion
			result, err := wasmd.ConvertEVMCoinsToSDKCoins(tc.input)

			// Check for expected errors
			if tc.errorContains != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.errorContains)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, result)
			}
		},
		)
	}
}
