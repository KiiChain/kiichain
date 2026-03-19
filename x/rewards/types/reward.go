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

	// fix(rewards): use nanosecond precision instead of Seconds() to avoid truncation
	// that causes division by zero when duration is between 1-999ms (issue #267)
	totalDurationNs := schedule.EndTime.Sub(schedule.LastReleaseTime).Nanoseconds()
	if totalDurationNs <= 0 {
		return sdk.Coin{}, fmt.Errorf("end time is equal to or before last release time and would cause a division by 0. EndTime: %s, LastReleaseTime: %s", schedule.EndTime, schedule.LastReleaseTime)
	}

	// Get time parameters
	timeElapsedStamp := blockTime.Sub(schedule.LastReleaseTime)          // Time since last release
	totalDurationStamp := schedule.EndTime.Sub(schedule.LastReleaseTime) // Remaining release period

	// Convert to nanoseconds (instead of Seconds) to avoid precision loss truncation
	// that caused division by zero when duration < 1 second (issue #267)
	timeElapsedNs := math.NewInt(timeElapsedStamp.Nanoseconds()).BigInt()
	totalDurationNsBig := math.NewInt(totalDurationStamp.Nanoseconds()).BigInt()

	// Calculate linear release proportion between 0 and 1
	releaseProportion := math.LegacyNewDecFromBigInt(timeElapsedNs).Quo(math.LegacyNewDecFromBigInt(totalDurationNsBig))
	// Truncate to int, it will be a coin amt after all
	amountToRelease := math.LegacyNewDecFromInt(remaining.Amount).Mul(releaseProportion).TruncateInt()

	// fix(rewards): Do NOT force minimum 1 coin unconditionally (issue #265)
	// The original code always released at least 1 coin per block, which caused severe
	// over-distribution for long schedules where the proportional release is < 1 coin/block.
	// Example: 1M tokens over 1 year at 6s blocks = ~0.19 tokens/block → truncates to 0 per block
	// but forces 1, draining the pool 5x faster than planned.
	//
	// Fix: Only enforce the minimum-1-coin guard when we are in the final distribution window
	// (i.e. this is the last block before EndTime), ensuring at least the last coin is released.
	isLastWindow := blockTime.Equal(schedule.EndTime) || blockTime.After(schedule.EndTime)
	if isLastWindow && amountToRelease.IsZero() && !remaining.IsZero() {
		amountToRelease = math.NewInt(1)
	}

	// Cap at remaining amount
	amountToRelease = math.MinInt(amountToRelease, remaining.Amount)

	// If nothing to release this block, return zero coin without error
	if amountToRelease.IsZero() {
		return sdk.NewCoin(schedule.TotalAmount.Denom, math.ZeroInt()), nil
	}

	return sdk.NewCoin(schedule.TotalAmount.Denom, amountToRelease), nil
}
