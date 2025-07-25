package types

// NewGenesisState constructs a genesis state
func NewGenesisState(params Params) *GenesisState {
	return &GenesisState{
		Params: params,
	}
}

// DefaultGenesisState returns the default genesis
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate validates the genesis state
func (gs *GenesisState) Validate() error {
	return gs.Params.ValidateBasic()
}
