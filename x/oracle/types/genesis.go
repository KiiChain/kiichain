package types

// NewGenesisState creates a new GenesisState object with the imput parameters
func NewGenesisState(params Params) *GenesisState {
	return &GenesisState{
		Params: params,
	}
}

// DefaultGenesisState creates a new genesis with the default parameters
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// ValidateGenesis executes the Validate function for an input genesis data
func ValidateGenesis(data *GenesisState) error {
	return data.Params.Validate()
}
