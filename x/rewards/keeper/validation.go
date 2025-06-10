package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// validateAuthority checks if address authority is valid and same as expected
func (k *Keeper) validateAuthority(authority string) error {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	if k.authority != authority {
		return errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, authority)
	}

	return nil
}

// validateAmount check if amount is a valid coin
func validateAmount(amount sdk.Coin) error {
	if err := amount.Validate(); err != nil {
		return errors.Wrap(sdkerrors.ErrInvalidCoins, amount.String())
	}

	return nil
}

// validateTime checks if time is in the past
func validateTime(endTime time.Time) error {
	if endTime.Before(time.Now()) {
		return fmt.Errorf("end time %s is not in the future", endTime)
	}

	return nil
}

// fundsAvailable checks if the asked funds are available in the pool
func (k Keeper) fundsAvailable(ctx context.Context, amount sdk.Coin) error {
	// Fetch schedule
	schedule, err := k.ReleaseSchedule.Get(ctx)
	if err != nil {
		return err
	}
	// Check if releaser is active (means some amt of the pool is promised)
	if schedule.Active {
		// Sum the promised amt to the asked funds
		amount = amount.Add(schedule.TotalAmount.Sub(schedule.ReleasedAmount))
	}

	// Get reward pool
	rewardPool, err := k.RewardPool.Get(ctx)
	if err != nil {
		return err
	}

	// Check if it is trying to use more funds than available
	poolAmount := rewardPool.CommunityPool.AmountOf(amount.Denom)
	if sdk.NewDecCoinFromCoin(amount).Amount.GT(poolAmount) {
		return fmt.Errorf("reward pool (%s) has less funds than requested (%s)", poolAmount, amount)
	}

	return nil
}
