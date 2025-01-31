package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochTypes "github.com/kiichain/kiichain3/x/epoch/types"
)

// BeforeEpochStart is a hook that's ran after an epoch starts
func (k Keeper) BeforeEpochStart(_ sdk.Context, _ epochTypes.Epoch) {}

// AfterEpochEnd is a hook that's ran after an epoch ends
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epoch epochTypes.Epoch) {
	// Get the latest minter
	latestMinter, err := k.GetOrUpdateLatestMinter(ctx, epoch)
	if err != nil {
		// We can panic, it's common for hooks to panic
		// The panic is captured on epoch hooks
		panic(err)
	}

	// Get the coinsToMint
	coinsToMint, err := latestMinter.GetReleaseAmountToday(epoch.CurrentEpochStartTime.UTC())
	if err != nil {
		panic(err)
	}

	// Get the remaining mint amount
	if coinsToMint.IsZero() || latestMinter.GetRemainingMintAmount() == 0 {
		k.Logger(ctx).Debug("No coins to mint", "minter", latestMinter)
	}

	// mint coins, update supply
	if err := k.MintCoins(ctx, coinsToMint); err != nil {
		// We can panic, it's common for hooks to panic
		// The panic is captured, logged and handled on epoch hooks
		panic(err)
	}
	// send the minted coins to the fee collector account
	if err := k.AddCollectedFees(ctx, coinsToMint); err != nil {
		// We can panic, it's common for hooks to panic
		// The panic is captured, logged and handled on epoch hooks
		panic(err)
	}

	// Released successfully, decrement the remaining amount by the daily release amount and update minter
	amountMinted := coinsToMint.AmountOf(latestMinter.GetDenom())
	latestMinter.RecordSuccessfulMint(ctx, epoch, amountMinted.Uint64())
	k.Logger(ctx).Info("Minted coins", "minter", latestMinter, "amount", coinsToMint.String())
	k.SetMinter(ctx, latestMinter)
}

// Hooks is the hook struct for the mint module
type Hooks struct {
	k Keeper
}

// Assert the interface
var _ epochTypes.EpochHooks = Hooks{}

// Return the wrapper struct.
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// Epoch hook for before epoch start
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epoch epochTypes.Epoch) {
	h.k.BeforeEpochStart(ctx, epoch)
}

// Epoch hook for before after start
func (h Hooks) AfterEpochEnd(ctx sdk.Context, epoch epochTypes.Epoch) {
	h.k.AfterEpochEnd(ctx, epoch)
}
