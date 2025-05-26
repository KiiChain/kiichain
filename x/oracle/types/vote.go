package types

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v2"
)

// NewAggregateExchangeRateVote creates a new AggregateExchangeRateVote instance
func NewAggregateExchangeRateVote(exchangeRateTuples ExchangeRateTuples, voter sdk.ValAddress) (AggregateExchangeRateVote, error) {
	// Iterate over the exchangeRateTuples and validate all exchangeRate are higher than zero
	for _, exchangeRate := range exchangeRateTuples {
		if !exchangeRate.ExchangeRate.IsPositive() {
			return AggregateExchangeRateVote{}, fmt.Errorf("exchange rate for denom %s must be greater than zero, got %s", exchangeRate.Denom, exchangeRate.ExchangeRate.String())
		}
	}

	return AggregateExchangeRateVote{
		ExchangeRateTuples: exchangeRateTuples,
		Voter:              voter.String(),
	}, nil
}

// Implement String for the AggregateExchangeRateVote (I set false on the proto file, so I have to do it manually)
func (a AggregateExchangeRateVote) String() string {
	out, _ := yaml.Marshal(a)
	return string(out)
}

// NewExchangeRateTuple creates a new ExchangeRateTuple instance
func NewExchangeRateTuple(denom string, exchangeRage math.LegacyDec) ExchangeRateTuple {
	return ExchangeRateTuple{
		Denom:        denom,
		ExchangeRate: exchangeRage,
	}
}

// String implements stringify
func (v ExchangeRateTuple) String() string {
	out, _ := yaml.Marshal(v)
	return string(out)
}

// ExchangeRateTuples represent an array of ExchangeRateTuple on params.go
type ExchangeRateTuples []ExchangeRateTuple

func (tuples ExchangeRateTuples) String() string {
	out, _ := yaml.Marshal(tuples)
	return string(out)
}

// String implements stringify
func (o OracleExchangeRate) String() string {
	out, _ := yaml.Marshal(o)
	return string(out)
}

// ParseExchangeRateTuples parses from exchangeRate string tuple to ExchangeRateTuples{} data type
func ParseExchangeRateTuples(exchangeRateStr string) (ExchangeRateTuples, error) {
	// Remove innecesaries spaces. i.e: " BTC:45000 , ETH:3000 " -> "BTC:45000 , ETH:3000"
	exchangeRateStr = strings.TrimSpace(exchangeRateStr)
	if len(exchangeRateStr) == 0 {
		return nil, nil
	}

	// Separate elements by the comma
	tupleStrs := strings.Split(exchangeRateStr, ",")
	exchangeTuples := make(ExchangeRateTuples, len(tupleStrs)) // the parsed elements will be stored here
	duplicateCheckMap := make(map[string]bool)                 // map to track duplicate

	//Iterate each element ["BTC:45000", "ETH:3000", "COP:4500"]...
	for i, tupleStr := range tupleStrs {
		decCoin, err := sdk.ParseDecCoin(tupleStr) // convert decimal coin to string
		if err != nil {
			return nil, err
		}

		// convert each string rate into ExchangeRateTuple{} data type
		exchangeTuples[i] = ExchangeRateTuple{
			Denom:        decCoin.Denom,
			ExchangeRate: decCoin.Amount,
		}

		// Check duplicate
		_, ok := duplicateCheckMap[decCoin.Denom]
		if ok {
			return nil, fmt.Errorf("duplicate denom %s", decCoin.Denom)
		}

		duplicateCheckMap[decCoin.Denom] = true
	}
	return exchangeTuples, nil
}

// DenomOracleExchangeRatePairs represents an array of DenomOracleExchangeRate on query.go
type DenomOracleExchangeRatePairs []DenomOracleExchangeRate

// String implements stringify for DenomOracleExchangeRatePairs
func (o DenomOracleExchangeRatePairs) String() string {
	out, _ := yaml.Marshal(o)
	return string(out)
}
