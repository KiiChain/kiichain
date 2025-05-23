package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName is the name of the oracle module
	ModuleName = "oracle"

	// StoreKey is the key store representation
	StoreKey = ModuleName

	// In-memory store (for temporary data, not persistent)
	MemStoreKey = "oracle_mem"

	// Used for routing messages to this module
	RouterKey = ModuleName

	// Used for handling queries
	QuerierRoute = ModuleName
)

var (
	// Prefixes to store the data
	ExchangeRateKey              = []byte{0x01} // The latest exchange rate for each token (e.g., "BTC/USD" price)
	FeederDelegationKey          = []byte{0x02} // The account delegated to submit oracle votes for a validator
	VotePenaltyCounterKey        = []byte{0x03} // Tracks missed vote counts for validators
	AggregateExchangeRateVoteKey = []byte{0x04} // Stores the exchange rate votes submitted by validators
	VoteTargetKey                = []byte{0x05} // Stores the list of assets that validators must submit votes for
	PriceSnapshotKey             = []byte{0x06} // Stores historical price snapshots at specific timestamps
	SpamPreventionCounter        = []byte{0x07} // Stores repeated submissions by validator.
)

// GetExchangeRateKey returns the key to search the latest exchange rate by denom
// e.g = "BTC/USD" -> GetExchangeRateKey -> [0x01]["BTC/USD"]
func GetExchangeRateKey(denom string) []byte {
	return append(ExchangeRateKey, []byte(denom)...)
}

// GetFeederDelegationKey returns the key to search the address of the delegated account by validator
func GetFeederDelegationKey(valAddr sdk.ValAddress) []byte {
	return append(FeederDelegationKey, address.MustLengthPrefix(valAddr)...)
}

// GetVotePenaltyCounterKey returns the key to search the vote penality counters by validator address
func GetVotePenaltyCounterKey(valAddr sdk.ValAddress) []byte {
	return append(VotePenaltyCounterKey, address.MustLengthPrefix(valAddr)...)
}

// GetAggregateExchangeRateVoteKey returns the key to search the exchange rate votes submitted by validator address
func GetAggregateExchangeRateVoteKey(valAddr sdk.ValAddress) []byte {
	return append(AggregateExchangeRateVoteKey, address.MustLengthPrefix(valAddr)...)
}

// GetSpamPreventionCounterKey returns the key to search the spam prevention counter by validator address
func GetSpamPreventionCounterKey(valAddr sdk.ValAddress) []byte {
	return append(SpamPreventionCounter, address.MustLengthPrefix(valAddr)...)
}

// GetVoteTargetKey returns the key to search the exchange rate by its name
func GetVoteTargetKey(denom string) []byte {
	return append(VoteTargetKey, []byte(denom)...)
}

// ExtractDenomFromVoteTargetKey extracts the denom from a value result
// e.g: [0x05]["BTC/USD"] -> ExtractDenomFromVoteTargetKey -> "BTC/USD"
func ExtractDenomFromVoteTargetKey(key []byte) string {
	denom := string(key[1:])
	return denom
}

// GetPriceSnapshotKey returns the key to search the price snapshot by timestamp
func GetPriceSnapshotKey(timestamp uint64) []byte {
	timestampKey := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampKey, timestamp)
	return append(PriceSnapshotKey, timestampKey...)
}
