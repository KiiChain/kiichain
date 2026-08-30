package types

import "fmt"

// NewGenesisState constructs a genesis state
func NewGenesisState(params Params, rp RewardPool) *GenesisState {
	return &GenesisState{
		Params:     params,
		RewardPool: rp,
	}
}

// DefaultGenesisState returns the default genesis state of rewards.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		RewardPool: InitialRewardPool(),
		Params:     DefaultParams(),
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

	tokenDenom := gs.Params.TokenDenom
	for _, coin := range gs.RewardPool.CommunityPool {
		if coin.Denom != tokenDenom {
			return fmt.Errorf("community pool coin denom %s does not match token denom %s",
				coin.Denom, tokenDenom)
		}
	}

	if !gs.RewardPool.TotalReleased.IsNil() && !gs.RewardPool.TotalReleased.IsZero() &&
		gs.RewardPool.TotalReleased.Denom != tokenDenom {
		return fmt.Errorf("total released denom %s does not match token denom %s",
			gs.RewardPool.TotalReleased.Denom, tokenDenom)
	}

	return nil
}
