package types

import (
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CalculateReward figures what amt to be released in the current block
// Assumes invalid values are cleared before calling, does not handle invalid blockTime/no last release
func CalculateReward(blockTime time.Time, releaser RewardReleaser) (sdk.Coin, error) {
	// Calculate remaining amount
	remaining := releaser.TotalAmount.Sub(releaser.ReleasedAmount)
	if remaining.IsZero() {
		return remaining, nil
	}

	// Get time parameters
	timeElapsedStamp := blockTime.Sub(releaser.LastReleaseTime)          // Time since last release
	totalDurationStamp := releaser.EndTime.Sub(releaser.LastReleaseTime) // Remaining release period

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

	return sdk.NewCoin(releaser.TotalAmount.Denom, amountToRelease), nil
}
