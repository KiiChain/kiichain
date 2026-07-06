package types

import (
	time "time"

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

	// If total duration is zero or negative, release the full remaining amount
	totalDurationNs := schedule.EndTime.Sub(schedule.LastReleaseTime).Nanoseconds()
	if totalDurationNs <= 0 {
		return remaining, nil
	}

	// Get time parameters
	timeElapsedStamp := blockTime.Sub(schedule.LastReleaseTime)          // Time since last release
	totalDurationStamp := schedule.EndTime.Sub(schedule.LastReleaseTime) // Remaining release period

	// Guard: if block time has not advanced past the last release time (e.g.
	// clock skew or same-block re-entry), return a zero coin to avoid a
	// negative elapsed duration producing a negative proportion and a
	// downstream panic in sdk.NewCoin.
	if timeElapsedStamp.Nanoseconds() <= 0 {
		return sdk.NewCoin(schedule.TotalAmount.Denom, math.ZeroInt()), nil
	}

	// Use nanoseconds for precise duration calculations
	timeElapsedNs := math.NewInt(timeElapsedStamp.Nanoseconds()).BigInt()
	totalDurationNsBig := math.NewInt(totalDurationStamp.Nanoseconds()).BigInt()

	// Calculate linear release proportion between 0 and 1
	releaseProportion := math.LegacyNewDecFromBigInt(timeElapsedNs).Quo(math.LegacyNewDecFromBigInt(totalDurationNsBig))
	// Truncate to int, it will be a coin amt after all
	amountToRelease := math.LegacyNewDecFromInt(remaining.Amount).Mul(releaseProportion).TruncateInt()

	// Cap at remaining amount
	amountToRelease = math.MinInt(amountToRelease, remaining.Amount)

	return sdk.NewCoin(schedule.TotalAmount.Denom, amountToRelease), nil
}
