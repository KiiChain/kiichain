package types

import (
	"fmt"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CalculateReward figures what amt to be released in the current block
// Assumes invalid values are cleared before calling, does not handle invalid blockTime/no last release
func CalculateReward(blockTime time.Time, schedule ReleaseSchedule) (sdk.Coin, error) {
	// Calculate remaining amount
	remaining := schedule.TotalAmount.Sub(schedule.ReleasedAmount)
	if remaining.IsZero() {
		return remaining, nil
	}

	// If total duration would be 0, there would be a div by 0
	if schedule.EndTime.Equal(schedule.LastReleaseTime) {
		return sdk.Coin{}, fmt.Errorf("end time is equal to last release and would do a division by 0. EndTime: %s", schedule.EndTime)
	}

	// Get time parameters
	timeElapsedStamp := blockTime.Sub(schedule.LastReleaseTime)          // Time since last release
	totalDurationStamp := schedule.EndTime.Sub(schedule.LastReleaseTime) // Remaining release period

	// Convert to big int, using truncated seconds
	timeElapsed, err := math.NewDecFromInt64(int64(timeElapsedStamp.Seconds())).BigInt()
	if err != nil {
		return sdk.Coin{}, err
	}
	totalDuration, err := math.NewDecFromInt64(int64(totalDurationStamp.Seconds())).BigInt()
	if err != nil {
		return sdk.Coin{}, err
	}

	// Calculate linear release proportion between 0 and 1
	releaseProportion := math.LegacyNewDecFromBigInt(timeElapsed).Quo(math.LegacyNewDecFromBigInt(totalDuration))
	// Truncate to int, it will be a coin amt after all
	amountToRelease := math.LegacyNewDecFromInt(remaining.Amount).Mul(releaseProportion).TruncateInt()

	// Cap at remaining amount
	amountToRelease = math.MinInt(amountToRelease, remaining.Amount)

	return sdk.NewCoin(schedule.TotalAmount.Denom, amountToRelease), nil
}
