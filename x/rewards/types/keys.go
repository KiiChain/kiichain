package types

import "cosmossdk.io/collections"

var (
	ParamsKey         = collections.NewPrefix(0)
	RewardPoolKey     = collections.NewPrefix(1)
	RewardReleaserKey = collections.NewPrefix(2)
)

const (
	// ModuleName defines the module name
	ModuleName = "rewards"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_rewards"
)

// KeySeparator is used to combine parts of the keys in the store
const KeySeparator = "|"
