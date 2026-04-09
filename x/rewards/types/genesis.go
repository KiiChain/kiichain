package types

import "fmt"

// NewGenesisState constructs a genesis state
func NewGenesisState(
	params Params, rp RewardPool, release ReleaseSchedule,
) *GenesisState {
	return &GenesisState{
		Params:          params,
		RewardPool:      rp,
		ReleaseSchedule: release,
	}
}

// DefaultGenesisState returns the default genesis state of rewards.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		RewardPool:      InitialRewardPool(),
		Params:          DefaultParams(),
		ReleaseSchedule: InitialReleaseSchedule(),
	}
}

// Validate validates the genesis state of rewards genesis input
func (gs *GenesisState) Validate() error {
	if err := gs.Params.ValidateBasic(); err != nil {
		return err
	}

	if err := gs.RewardPool.ValidateGenesis(); err != nil {
		return err
	}

	if err := gs.ReleaseSchedule.ValidateGenesis(); err != nil {
		return err
	}

	// Enforce denom consistency: all pool coins and the active schedule must use TokenDenom.
	tokenDenom := gs.Params.TokenDenom
	for _, coin := range gs.RewardPool.CommunityPool {
		if coin.Denom != tokenDenom {
			return fmt.Errorf("community pool coin denom %s does not match token denom %s",
				coin.Denom, tokenDenom)
		}
	}

	if gs.ReleaseSchedule.Active && gs.ReleaseSchedule.TotalAmount.Denom != tokenDenom {
		return fmt.Errorf("release schedule total amount denom %s does not match token denom %s",
			gs.ReleaseSchedule.TotalAmount.Denom, tokenDenom)
	}

	return nil
}
