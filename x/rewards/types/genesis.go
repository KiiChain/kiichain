package types

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
	return gs.ReleaseSchedule.ValidateGenesis()
}
