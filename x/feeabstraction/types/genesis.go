package types

import (
	errorsmod "cosmossdk.io/errors"
)

// NewGenesisState constructs a genesis state
func NewGenesisState(params Params, feeTokens *FeeTokenMetadataCollection) *GenesisState {
	return &GenesisState{
		Params:    params,
		FeeTokens: feeTokens,
	}
}

// DefaultGenesisState returns the default genesis
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:    DefaultParams(),
		FeeTokens: &FeeTokenMetadataCollection{},
	}
}

// Validate validates the genesis state
func (gs *GenesisState) Validate() error {
	// Validate the params
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// Validate each fee token metadata and check for duplicate denoms
	denomSet := make(map[string]struct{})
	for _, token := range gs.FeeTokens.Items {
		if err := token.Validate(); err != nil {
			return err
		}
		if _, exists := denomSet[token.Denom]; exists {
			return errorsmod.Wrapf(ErrInvalidFeeTokenMetadata, "duplicate denom found: %s", token.Denom)
		}
		denomSet[token.Denom] = struct{}{}
	}

	return nil
}
