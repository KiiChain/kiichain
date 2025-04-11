package params

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	evmtypes "github.com/cosmos/evm/types"
)

const (
	// Bech32Prefix defines the Bech32 prefix used for accounts
	Bech32Prefix = "kii"

	// Bech32PrefixAccAddr defines the Bech32 prefix of an account's address.
	Bech32PrefixAccAddr = Bech32Prefix
	// Bech32PrefixAccPub defines the Bech32 prefix of an account's public key.
	Bech32PrefixAccPub = Bech32Prefix + sdk.PrefixPublic
	// Bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address.
	Bech32PrefixValAddr = Bech32Prefix + sdk.PrefixValidator + sdk.PrefixOperator
	// Bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key.
	Bech32PrefixValPub = Bech32Prefix + sdk.PrefixValidator + sdk.PrefixOperator + sdk.PrefixPublic
	// Bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address.
	Bech32PrefixConsAddr = Bech32Prefix + sdk.PrefixValidator + sdk.PrefixConsensus
	// Bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key.
	Bech32PrefixConsPub = Bech32Prefix + sdk.PrefixValidator + sdk.PrefixConsensus + sdk.PrefixPublic
	// DisplayDenom defines the denomination displayed to users in client applications.
	DisplayDenom = "kii"
	// BaseDenom defines to the default denomination
	BaseDenom = "akii"
	// BaseDenomUnit defines the precision of the base denomination.
	BaseDenomUnit = 18

	// Testnet chain id
	TestnetChainID = "kiichain_1336"
	LocalChainID   = "localchain_1010"
)

// SetBech32Prefixes sets the global prefixes to be used when serializing addresses and public keys to Bech32 strings.
func SetBech32Prefixes(config *sdk.Config) {
	config.SetBech32PrefixForAccount(Bech32PrefixAccAddr, Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(Bech32PrefixValAddr, Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(Bech32PrefixConsAddr, Bech32PrefixConsPub)
}

// SetBip44CoinType sets the global coin type to be used in hierarchical deterministic wallets.
func SetBip44CoinType(config *sdk.Config) {
	config.SetCoinType(evmtypes.Bip44CoinType)
	config.SetPurpose(sdk.Purpose)                     // Shared
	config.SetFullFundraiserPath(evmtypes.BIP44HDPath) //nolint: staticcheck
}

// Init initializes all the params
func init() {
	// Get the config
	config := sdk.GetConfig()
	// Set the bech32
	SetBech32Prefixes(config)
	// Set the coin type
	SetBip44CoinType(config)
	// Seal the config
	config.Seal()
}
