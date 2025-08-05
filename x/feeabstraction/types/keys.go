package types

import "cosmossdk.io/collections"

// Defines all the KV keys for the collections
var (
	ParamsKey    = collections.NewPrefix(0)
	FeeTokensKey = collections.NewPrefix(1)
)

const (
	// ModuleName define the module name
	ModuleName = "feeabstraction"

	// StoreKey defines the module store key
	StoreKey = ModuleName

	// RouterKey is the message route
	RouterKey = ModuleName

	// QuerierRoute defines the module query routing key
	QuerierRoute = ModuleName
)
