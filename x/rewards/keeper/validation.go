package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/kiichain/kiichain/v1/x/rewards/types"
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

// validateEndTime checks if time is in the past
func validateEndTime(endTime time.Time) error {
	if endTime.Before(time.Now()) {
		return fmt.Errorf("end time %s is not in the future", endTime)
	}

	return nil
}

// fundsAvailable checks if the asked funds are available in the pool
func (k Keeper) fundsAvailable(ctx context.Context, amount sdk.Coin) error {
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

// validateSchedule checks if the asked funds are available in the pool
func (k Keeper) validateSchedule(ctx context.Context, schedule types.ReleaseSchedule) error {
	// Validate TotalAmount
	if err := validateAmount(schedule.TotalAmount); err != nil {
		return fmt.Errorf("invalid total amount: %w", err)
	}

	// Validate against module params
	params, err := k.Params.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get module params: %w", err)
	}
	if params.TokenDenom != schedule.TotalAmount.Denom {
		return fmt.Errorf("denom %s does not match expected denom: %s",
			schedule.TotalAmount.Denom, params.TokenDenom)
	}

	// Validate ReleasedAmount
	if schedule.ReleasedAmount.Denom != schedule.TotalAmount.Denom {
		return fmt.Errorf("released amount denom %s doesn't match total amount denom %s",
			schedule.ReleasedAmount.Denom, schedule.TotalAmount.Denom)
	}
	if !schedule.ReleasedAmount.IsZero() {
		if err := validateAmount(schedule.ReleasedAmount); err != nil {
			return fmt.Errorf("invalid released amount: %w", err)
		}
		if schedule.ReleasedAmount.Amount.GT(schedule.TotalAmount.Amount) {
			return fmt.Errorf("released amount %s cannot exceed total amount %s",
				schedule.ReleasedAmount, schedule.TotalAmount)
		}
	}

	// Time validations
	currentTime := time.Now()
	if schedule.EndTime.IsZero() {
		return fmt.Errorf("end time cannot be zero")
	}
	if err = validateEndTime(schedule.EndTime); err != nil {
		return err
	}

	if !schedule.LastReleaseTime.IsZero() {
		if schedule.LastReleaseTime.After(currentTime) {
			return fmt.Errorf("last release time %s cannot be in the future",
				schedule.LastReleaseTime)
		}
		if schedule.LastReleaseTime.After(schedule.EndTime) {
			return fmt.Errorf("last release time %s cannot be after end time %s",
				schedule.LastReleaseTime, schedule.EndTime)
		}
	}

	// 6. Active state consistency
	if schedule.Active {
		if schedule.TotalAmount.IsZero() {
			return fmt.Errorf("active schedule cannot have zero total amount")
		}
		if schedule.EndTime.IsZero() {
			return fmt.Errorf("active schedule must have an end time")
		}
	}
	return nil
}
