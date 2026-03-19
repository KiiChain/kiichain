package types

import (
	context "context"

	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// StakingKeeper is expected keeper for staking module, because I need to handle
// reward and slashink on my oracle module
type StakingKeeper interface {
	Validator(ctx context.Context, address sdk.ValAddress) (stakingtypes.ValidatorI, error)                                               // Retrieves a validator's information
	TotalBondedTokens(ctx context.Context) (math.Int, error)                                                                              // Retrieves total staked tokens (useful for slashing calculations)
	Jail(ctx context.Context, consAddr sdk.ConsAddress) error                                                                             // Jail validators
	ValidatorsPowerStoreIterator(ctx context.Context) (corestore.Iterator, error)                                                         // Used to computing validator rankings or total power
	MaxValidators(ctx context.Context) (uint32, error)                                                                                    // Return the maximum amount of bonded validators
	PowerReduction(ctx context.Context) (res math.Int)                                                                                    // Returns the power reduction factor
	GetValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (stakingtypes.Validator, error)                                 // Retrieves a validator by consensus address
	RemoveValidatorTokens(ctx context.Context, validator stakingtypes.Validator, tokensToRemove math.Int) (stakingtypes.Validator, error) // Removes tokens from a validator without burning
	TokensFromConsensusPower(ctx context.Context, power int64) math.Int                                                                   // Converts consensus power to token amount
	BondDenom(ctx context.Context) (string, error)                                                                                        // Returns the bond denomination
}

// DistributionKeeper is the expected keeper for the distribution module
type DistributionKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error // Sends coins to the community pool
}

// AccountKeeper is expected keeper for auth module, because I need to handle
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress                                // Ensures the oracle module has an account
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI // Retrieves detailed account information
	SetModuleAccount(ctx context.Context, macc sdk.ModuleAccountI)              // Creates a module account
	AddressCodec() address.Codec
}

// BankKeeper is expected keeper for bank module, because I need to handle
// coins, get balance, receive and send coins
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin                                           // Check the oracle module account balance by denom
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins                                                    // Check the oracle module account balance all denom
	SendCoinsFromModuleToModule(ctx context.Context, senderModule string, recipientModule string, amount sdk.Coins) error // Transfer tokens between module accounts (e.g., moving slashed tokens)
	GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata)
}
