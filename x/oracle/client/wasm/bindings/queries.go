package bindings

import "github.com/kiichain/kiichain/v2/x/oracle/types"

// KiiOracleQuery represents the querier with the query functions
type KiiOracleQuery struct {
	// queries the oracle exchange rates
	ExchangeRates *types.QueryExchangeRatesRequest `json:"exchange_rates,omitempty"`
	// queries the oracle TWAPs
	OracleTwaps *types.QueryTwapsRequest `json:"oracle_twaps,omitempty"`
	// queries the actives assets
	Actives *types.QueryActivesRequest `json:"actives,omitempty"`
	// queries the price history
	PriceSnapshotHistory *types.QueryPriceSnapshotHistoryRequest `json:"price_snapshot_history,omitempty"`
	// queries the feeder delegated of a validator
	FeederDelegation *types.QueryFeederDelegationRequest `json:"feeder_delegation,omitempty"`
	// queries the penalty counter of a validator
	VotePenaltyCounter *types.QueryVotePenaltyCounterRequest `json:"vote_penalty_counter,omitempty"`
}
