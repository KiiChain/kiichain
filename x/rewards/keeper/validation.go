package keeper

import (
\t"context"
\t"fmt"
\t"time"

\t"cosmossdk.io/errors"

\tsdk "github.com/cosmos/cosmos-sdk/types"
\tsdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
\tgovtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

\t"github.com/kiichain/kiichain/v7/x/rewards/types"
)

// validateAuthority checks if address authority is valid and same as expected
func (k *Keeper) validateAuthority(authority string) error {
\tif _, err := sdk.AccAddressFromBech32(authority); err != nil {
\t	return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
\t}

\tif k.authority != authority {
\t	return errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, authority)
\t}

\treturn nil
}

// validateAmount check if amount is a valid coin
func validateAmount(amount sdk.Coin) error {
\tif err := amount.Validate(); err != nil {
\t	return errors.Wrap(sdkerrors.ErrInvalidCoins, amount.String())
\t}

\treturn nil
}

// validateEndTime checks if time is in the past
func validateEndTime(ctx sdk.Context, endTime time.Time) error {
\tif endTime.Before(ctx.BlockTime()) {
\t	return fmt.Errorf("end time %s is not in the future", endTime)
\t}

\treturn nil
}

// fundsAvailable checks if the asked funds are available in the pool
func (k Keeper) fundsAvailable(ctx context.Context, amount sdk.Coin) error {
\trewardPool, err := k.RewardPool.Get(ctx)
\tif err != nil {
\t	return err
\t}
\tpoolAmount := rewardPool.CommunityPool.AmountOf(amount.Denom)
\tif sdk.NewDecCoinFromCoin(amount).Amount.GT(poolAmount) {
\t	return fmt.Errorf("reward pool (%s) has less funds than requested (%s)", poolAmount, amount)
\t}
\treturn nil
}

// validateSchedule checks if the asked funds are available in the pool
func (k Keeper) validateSchedule(ctx sdk.Context, schedule types.ReleaseSchedule) error {
\tif err := validateAmount(schedule.TotalAmount); err != nil {
\t	return fmt.Errorf("invalid total amount: %w", err)
\t}

\tparams, err := k.Params.Get(ctx)
\tif err != nil {
\t	return fmt.Errorf("failed to get module params: %w", err)
\t}
\tif params.TokenDenom != schedule.TotalAmount.Denom {
\t	return fmt.Errorf("denom %s does not match expected denom: %s",
\t		schedule.TotalAmount.Denom, params.TokenDenom)
\t}

\t// Validate ReleasedAmount only when non-zero. The denom check was previously
\t// run unconditionally, which rejected zero-value coins (sdk.Coin{} with empty
\t// denom) used to represent a fresh schedule with no released tokens yet.
\tif !schedule.ReleasedAmount.IsZero() {
\t	if schedule.ReleasedAmount.Denom != schedule.TotalAmount.Denom {
\t		return fmt.Errorf("released amount denom %s doesn't match total amount denom %s",
\t			schedule.ReleasedAmount.Denom, schedule.TotalAmount.Denom)
\t	}
\t	if err := validateAmount(schedule.ReleasedAmount); err != nil {
\t		return fmt.Errorf("invalid released amount: %w", err)
\t	}
\t	if schedule.ReleasedAmount.Amount.GT(schedule.TotalAmount.Amount) {
\t		return fmt.Errorf("released amount %s cannot exceed total amount %s",
\t			schedule.ReleasedAmount, schedule.TotalAmount)
\t	}
\t}

\tif schedule.EndTime.IsZero() {
\t	return fmt.Errorf("end time cannot be zero")
\t}
\tif err = validateEndTime(ctx, schedule.EndTime); err != nil {
\t	return err
\t}

\tcurrentTime := ctx.BlockTime()
\tif !schedule.LastReleaseTime.IsZero() {
\t	if schedule.LastReleaseTime.After(currentTime) {
\t		return fmt.Errorf("last release time %s cannot be in the future", schedule.LastReleaseTime)
\t	}
\t	if schedule.LastReleaseTime.After(schedule.EndTime) {
\t		return fmt.Errorf("last release time %s cannot be after end time %s",
\t			schedule.LastReleaseTime, schedule.EndTime)
\t	}
\t}

\tif schedule.Active {
\t	if schedule.TotalAmount.IsZero() {
\t		return fmt.Errorf("active schedule cannot have zero total amount")
\t	}
\t	if schedule.EndTime.IsZero() {
\t		return fmt.Errorf("active schedule must have an end time")
\t	}
\t}
\treturn nil
}
