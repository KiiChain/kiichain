package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitialRewardPool returns a zero reward pool
func InitialRewardPool() RewardPool {
	return RewardPool{
		CommunityPool: sdk.DecCoins{},
		TotalReleased: sdk.Coin{},
	}
}

// ValidateGenesis validates the reward pool for a genesis state
func (rp RewardPool) ValidateGenesis() error {
	if err := rp.CommunityPool.Validate(); err != nil {
		return fmt.Errorf("invalid CommunityPool: %w", err)
	}

	if !rp.TotalReleased.IsNil() && !rp.TotalReleased.IsZero() {
		if err := rp.TotalReleased.Validate(); err != nil {
			return fmt.Errorf("invalid TotalReleased: %w", err)
		}
	}

	return nil
}
