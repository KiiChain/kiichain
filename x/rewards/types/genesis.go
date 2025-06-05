package types

func NewGenesisState(
	params Params, rp RewardPool, releaser RewardReleaser,
) *GenesisState {
	return &GenesisState{
		Params:         params,
		RewardPool:     rp,
		RewardReleaser: releaser,
	}
}

// DefaultGenesisState returns the default genesis state of rewards.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		RewardPool:     InitialRewardPool(),
		Params:         DefaultParams(),
		RewardReleaser: InitialRewardReleaser(),
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
	return gs.RewardReleaser.ValidateGenesis()
}
