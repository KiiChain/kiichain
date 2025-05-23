package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGenesisState(t *testing.T) {
	// create genesis state
	params := DefaultParams()
	exchangeRateTuple := []ExchangeRateTuple{}
	feederDelegation := []FeederDelegation{}
	penaltyCounters := []PenaltyCounter{}
	aggregateExchangeRateVote := []AggregateExchangeRateVote{}
	priceSnapshot := PriceSnapshots{}
	votePenaltyCounters := []VotePenaltyCounter{}

	newGenesis := NewGenesisState(params, exchangeRateTuple, feederDelegation, penaltyCounters, aggregateExchangeRateVote, priceSnapshot, votePenaltyCounters)

	// expected result
	expected := &GenesisState{
		Params:                     params,
		ExchangeRates:              exchangeRateTuple,
		FeederDelegations:          feederDelegation,
		AggregateExchangeRateVotes: aggregateExchangeRateVote,
		PriceSnapshots:             priceSnapshot,
		VotePenaltyCounters:        votePenaltyCounters,
		PenaltyCounters:            penaltyCounters,
	}

	// validation
	require.Equal(t, expected, newGenesis)
}

func TestDefaultGenesisState(t *testing.T) {
	// expected result
	params := DefaultParams()
	exchangeRateTuple := []ExchangeRateTuple{}
	feederDelegation := []FeederDelegation{}
	penaltyCounters := []PenaltyCounter{}
	aggregateExchangeRateVote := []AggregateExchangeRateVote{}
	priceSnapshot := PriceSnapshots{}
	votePenaltyCounters := []VotePenaltyCounter{}

	expected := &GenesisState{
		Params:                     params,
		ExchangeRates:              exchangeRateTuple,
		FeederDelegations:          feederDelegation,
		AggregateExchangeRateVotes: aggregateExchangeRateVote,
		PriceSnapshots:             priceSnapshot,
		VotePenaltyCounters:        votePenaltyCounters,
		PenaltyCounters:            penaltyCounters,
	}

	// Create default genesis
	defaultGenesis := DefaultGenesisState()

	// Validation
	require.Equal(t, expected, defaultGenesis)
}

func TestValidateGenesis(t *testing.T) {
	genState := DefaultGenesisState()
	require.NoError(t, ValidateGenesis(genState))

	genState.Params.VotePeriod = 0
	require.Error(t, ValidateGenesis(genState))
}
