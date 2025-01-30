package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// TestNewAggregateExchangeRateVote tests the creation of AggregateExchangeRateVote
func TestNewAggregateExchangeRateVote(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name               string
		exchangeRateTuples ExchangeRateTuples
		voter              sdk.ValAddress
		expected           AggregateExchangeRateVote
	}{
		{
			name: "Valid inputs with multiple exchanges rates (tuples)",
			exchangeRateTuples: ExchangeRateTuples{
				{Denom: "BTC/USD", ExchangeRate: sdk.NewDec(45000)},
				{Denom: "ETH/USD", ExchangeRate: sdk.NewDec(3000)},
			},
			voter: sdk.ValAddress([]byte("validator1")),
			expected: AggregateExchangeRateVote{
				ExchangeRateTuples: ExchangeRateTuples{
					{Denom: "BTC/USD", ExchangeRate: sdk.NewDec(45000)},
					{Denom: "ETH/USD", ExchangeRate: sdk.NewDec(3000)},
				},
				Voter: sdk.ValAddress([]byte("validator1")).String(),
			},
		},
		{
			name:               "Empty exchange rate (tuples)",
			exchangeRateTuples: ExchangeRateTuples{},
			voter:              sdk.ValAddress([]byte("validator2")),
			expected: AggregateExchangeRateVote{
				ExchangeRateTuples: ExchangeRateTuples{},
				Voter:              sdk.ValAddress([]byte("validator2")).String(),
			},
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NewAggregateExchangeRateVote(tc.exchangeRateTuples, tc.voter)
			require.Equal(t, tc.expected, result, "NewAggregateExchangeRateVote() did not return expected result")
		})
	}
}

func TestNewExchangeRateTuple(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name         string
		denom        string
		exchangeRage sdk.Dec
		expected     ExchangeRateTuple
	}{
		{
			name:         "Valid inputs",
			denom:        "BTC/USD",
			exchangeRage: sdk.NewDec(45000),
			expected: ExchangeRateTuple{
				Denom:        "BTC/USD",
				ExchangeRate: sdk.NewDec(45000),
			},
		},

		{
			name:         "Empty exchange rate",
			denom:        "",
			exchangeRage: sdk.NewDec(3000),
			expected: ExchangeRateTuple{
				Denom:        "",
				ExchangeRate: sdk.NewDec(3000),
			},
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NewExchangeRateTuple(tc.denom, tc.exchangeRage)
			require.Equal(t, tc.expected, result, "NewAggregateExchangeRateVote() did not return expected result")
		})
	}
}
