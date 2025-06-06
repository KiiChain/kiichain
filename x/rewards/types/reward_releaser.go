package types

import (
	fmt "fmt"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitialRewardPool returns a zero reward pool
func InitialRewardReleaser() RewardReleaser {
	return RewardReleaser{
		TotalAmount:     sdk.Coin{},
		ReleasedAmount:  sdk.Coin{},
		EndTime:         time.Time{},
		LastReleaseTime: time.Time{},
		Active:          false,
	}
}

// ValidateGenesis validates the reward pool for a genesis state
func (rr RewardReleaser) ValidateGenesis() error {
	// Validate EndTime (zero time is allowed for genesis)
	if !rr.EndTime.IsZero() && rr.EndTime.Before(time.Now()) {
		return fmt.Errorf("end time %s cannot be in the past", rr.EndTime.String())
	}

	// Validate LastReleaseTime
	if rr.LastReleaseTime.After(time.Now()) {
		return fmt.Errorf("last release time %s cannot be in the future", rr.EndTime.String())
	}

	// Some validations just make sense if active
	if rr.Active {
		// Validate TotalAmount
		if err := rr.TotalAmount.Validate(); err != nil {
			return fmt.Errorf("invalid total amount: %w", err)
		}
		if rr.TotalAmount.IsZero() {
			return fmt.Errorf("active reward releaser cannot have zero total amount")
		}
		if rr.EndTime.IsZero() {
			return fmt.Errorf("active reward releaser must have an end time")
		}
		// Validate ReleasedAmount if not zero
		if !rr.ReleasedAmount.IsZero() {
			if err := rr.ReleasedAmount.Validate(); err != nil {
				return fmt.Errorf("invalid released amount: %w", err)
			}

			// Check ReleasedAmount doesn't exceed TotalAmount
			if rr.ReleasedAmount.Denom != rr.TotalAmount.Denom {
				return fmt.Errorf("released amount denom %s doesn't match total amount denom %s",
					rr.ReleasedAmount.Denom, rr.TotalAmount.Denom)
			}

			if rr.ReleasedAmount.Amount.GT(rr.TotalAmount.Amount) {
				return fmt.Errorf("released amount %s cannot be greater than total amount %s",
					rr.ReleasedAmount.String(), rr.TotalAmount.String())
			}
		}
	}
	return nil
}
