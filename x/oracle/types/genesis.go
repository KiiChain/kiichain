package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
)

// NewGenesisState creates a new GenesisState object with the imput parameters
func NewGenesisState(params Params, exchangeRateTuple []ExchangeRateTuple, feederDelegation []FeederDelegation,
	penaltyCounters []PenaltyCounter, aggregateExchangeRateVote []AggregateExchangeRateVote, priceSnapshot PriceSnapshots, votePenaltyCounters []VotePenaltyCounter) *GenesisState {
	return &GenesisState{
		Params:                     params,
		ExchangeRates:              exchangeRateTuple,
		FeederDelegations:          feederDelegation,
		PenaltyCounters:            penaltyCounters,
		AggregateExchangeRateVotes: aggregateExchangeRateVote,
		PriceSnapshots:             priceSnapshot,
		VotePenaltyCounters:        votePenaltyCounters,
	}
}

// DefaultGenesisState creates a new genesis with the default parameters
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:                     DefaultParams(),
		ExchangeRates:              []ExchangeRateTuple{},
		FeederDelegations:          []FeederDelegation{},
		PenaltyCounters:            []PenaltyCounter{},
		AggregateExchangeRateVotes: []AggregateExchangeRateVote{},
		PriceSnapshots:             PriceSnapshots{},
		VotePenaltyCounters:        []VotePenaltyCounter{},
	}
}

// ValidateGenesis executes the Validate function for an input genesis data
func ValidateGenesis(data *GenesisState) error {
	return data.Params.Validate()
}

// GetGenesisStateFromAppState returns the x/oracle genesisState
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	// Unmarshal current genesis state
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}
