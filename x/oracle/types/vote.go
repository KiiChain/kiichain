package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gopkg.in/yaml.v2"
)

// NewAggregateExchangeRateVote creates a new AggregateExchangeRateVote instance
func NewAggregateExchangeRateVote(exchangeRateTuples ExchangeRateTuples, voter sdk.ValAddress) AggregateExchangeRateVote {
	return AggregateExchangeRateVote{
		ExchangeRateTuples: exchangeRateTuples,
		Voter:              voter.String(),
	}
}

// Implement String for the AggregateExchangeRateVote (I set false on the proto file, so I have to do it manually)
func (a AggregateExchangeRateVote) String() string {
	out, _ := yaml.Marshal(a)
	return string(out)
}

// NewExchangeRateTuple creates a new ExchangeRateTuple instance
func NewExchangeRateTuple(denom string, exchangeRage sdk.Dec) ExchangeRateTuple {
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
