package types

import (
	"cosmossdk.io/collections"
)

const (
	// ModuleName is the name of the oracle module
	ModuleName = "oracle"

	// StoreKey is the key store representation
	StoreKey = ModuleName

	// Used for routing messages to this module
	RouterKey = ModuleName

	// QuerierRoute is the route for querying data from this module
	QuerierRoute = ModuleName
)

var (
	// Defines all the keys for the oracle module
	ParamsKey                    = collections.NewPrefix(1)
	ExchangeRateKey              = collections.NewPrefix(2)
	FeederDelegationKey          = collections.NewPrefix(3)
	VotePenaltyCounterKey        = collections.NewPrefix(4)
	AggregateExchangeRateVoteKey = collections.NewPrefix(5)
	VoteTargetKey                = collections.NewPrefix(6)
	PriceSnapshotKey             = collections.NewPrefix(7)
	SpamPreventionCounter        = collections.NewPrefix(8)
)
