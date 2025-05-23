package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// StakingKeeper is expected keeper for staking module, because I need to handle
// reward and slashink on my oracle module
type StakingKeeper interface {
	Validator(ctx sdk.Context, address sdk.ValAddress) stakingtypes.ValidatorI //Retrieves a validator's information
	TotalBondedTokens(ctx sdk.Context) sdk.Int                                 // Retrieves total staked tokens (useful for slashing calculations)
	Slash(sdk.Context, sdk.ConsAddress, int64, int64, sdk.Dec)                 // Slashes a validator or delegate who fails to vote in the oracle
	Jail(ctx sdk.Context, consAddr sdk.ConsAddress)                            // Jail a validator or delegator
	ValidatorsPowerStoreIterator(ctx sdk.Context) sdk.Iterator                 // Used to computing validator rankings or total power
	MaxValidators(ctx sdk.Context) uint32                                      // Return the maximum amount of bonded validators
	PowerReduction(ctx sdk.Context) (res sdk.Int)                              //Returns the power reduction factor,
}

// AccountKeeper is expected keeper for auth module, because I need to handle
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress                                  //Ensures the oracle module has an account
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI // Retrieves detailed account information
	SetModuleAccount(ctx sdk.Context, macc authtypes.ModuleAccountI)              // Creates a module account
}

// BankKeeper is expected keeper for bank module, because I need to handle
// coins, get balance, receive and send coins
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin                                           //Check the oracle module account balance by denom
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins                                                    // Check the oracle module account balance all denom
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule string, recipientModule string, amount sdk.Coins) error // Transfer tokens between module accounts (e.g., moving slashed tokens)
	GetDenomMetaData(ctx sdk.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata)
}
