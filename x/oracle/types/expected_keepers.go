package types

import (
	context "context"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// StakingKeeper is expected keeper for staking module, because I need to handle
// reward and slashink on my oracle module
type StakingKeeper interface {
	Validator(ctx context.Context, address sdk.ValAddress) (stakingtypes.ValidatorI, error)                                           //Retrieves a validator's information
	TotalBondedTokens(ctx context.Context) (math.Int, error)                                                                          // Retrieves total staked tokens (useful for slashing calculations)
	Slash(ctx context.Context, consAddr sdk.ConsAddress, infractionHeight, power int64, slashFactor math.LegacyDec) (math.Int, error) // Slashes a validator or delegate who fails to vote in the oracle
	Jail(ctx context.Context, consAddr sdk.ConsAddress) error                                                                         // Jail a validator or delegator
	ValidatorsPowerStoreIterator(ctx context.Context) (corestore.Iterator, error)                                                     // Used to computing validator rankings or total power
	MaxValidators(ctx context.Context) (uint32, error)                                                                                // Return the maximum amount of bonded validators
	PowerReduction(ctx context.Context) (res math.Int)                                                                                //Returns the power reduction factor,
}

// AccountKeeper is expected keeper for auth module, because I need to handle
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress                                //Ensures the oracle module has an account
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI // Retrieves detailed account information
	SetModuleAccount(ctx context.Context, macc sdk.ModuleAccountI)              // Creates a module account
}

// BankKeeper is expected keeper for bank module, because I need to handle
// coins, get balance, receive and send coins
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin                                           //Check the oracle module account balance by denom
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins                                                    // Check the oracle module account balance all denom
	SendCoinsFromModuleToModule(ctx context.Context, senderModule string, recipientModule string, amount sdk.Coins) error // Transfer tokens between module accounts (e.g., moving slashed tokens)
	GetDenomMetaData(ctx context.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx context.Context, denomMetaData banktypes.Metadata)
}
