package types

import "fmt"

// InitialRewardPool returns a zero reward pool. Spendable funds live on the
// module account, not in this state.
func InitialRewardPool() RewardPool {
	return RewardPool{}
}

// ValidateGenesis validates the reward pool for a genesis state.
func (rp RewardPool) ValidateGenesis() error {
	if !rp.TotalReleased.IsNil() && !rp.TotalReleased.IsZero() {
		if err := rp.TotalReleased.Validate(); err != nil {
			return fmt.Errorf("invalid TotalReleased: %w", err)
		}
	}

	return nil
}
