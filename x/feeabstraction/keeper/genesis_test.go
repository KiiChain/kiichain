package keeper_test

import "github.com/kiichain/kiichain/v3/x/feeabstraction/types"

// TestGenesisInitExport tests the InitGenesis and ExportGenesis
func (s *KeeperTestSuite) TestGenesisInitExport() {
	// Get the current genesis state
	genesisState, err := s.keeper.ExportGenesis(s.ctx)
	s.Require().NoError(err)

	// Check the genesis state
	s.Require().Equal(types.DefaultGenesisState(), genesisState)

	// Modify the genesis state
	genesisState.Params = types.NewParams(
		"newcoin",
		types.DefaultMaxPriceDeviation.MulInt64(2),
		types.DefaultClampFactor.MulInt64(3),
		types.DefaultFallbackNativePrice,
		types.DefaultTwapLookbackWindow,
		true,
	)

	// Apply the init genesis
	err = s.keeper.InitGenesis(s.ctx, *genesisState)
	s.Require().NoError(err)

	// Check the params after init genesis
	params, err := s.keeper.Params.Get(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(genesisState.Params, params)

	// Export the genesis state again
	exportedGenesisState, err := s.keeper.ExportGenesis(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(genesisState, exportedGenesisState)
}
