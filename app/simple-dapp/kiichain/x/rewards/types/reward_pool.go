package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitialRewardPool returns a zero reward pool
func InitialRewardPool() RewardPool {
	return RewardPool{
		CommunityPool: sdk.DecCoins{},
	}
}

// ValidateGenesis validates the reward pool for a genesis state
func (rp RewardPool) ValidateGenesis() error {
	if rp.CommunityPool.IsAnyNegative() {
		return fmt.Errorf("negative CommunityPool in distribution fee pool, is %v",
			rp.CommunityPool)
	}

	return nil
}
