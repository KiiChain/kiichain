package keeper_test

import (
	"cosmossdk.io/math"
	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
)

// TestBeginBlocker tests the BeginBlocker of the fee abstraction module
func (s *KeeperTestSuite) TestBeginBlocker() {
	// Set the fee token prices in the keeper
	s.app.FeeAbstractionKeeper.FeeTokens.Set(s.ctx, *types.NewFeeTokenMetadataCollection(
		types.NewFeeTokenMetadata("uatom", "atom", 6, math.LegacyMustNewDecFromStr("50")),
	))

	// Get the params for the module and disable
	params, err := s.keeper.Params.Get(s.ctx)
	s.Require().NoError(err)
	params.Enabled = false
	s.Require().NoError(s.keeper.Params.Set(s.ctx, params))

	// Call the BeginBlocker
	s.Require().NoError(s.keeper.BeginBlocker(s.ctx))

	// No change is toke to the fee token (the token is still enabled)
	feeTokens, err := s.app.FeeAbstractionKeeper.FeeTokens.Get(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(feeTokens.Items, 1)
	s.Require().True(feeTokens.Items[0].Enabled)

	// Enable the module
	params.Enabled = true
	s.Require().NoError(s.keeper.Params.Set(s.ctx, params))

	// Call the BeginBlocker
	s.Require().NoError(s.app.FeeAbstractionKeeper.BeginBlocker(s.ctx))

	// Now the token should be disable due to missing twap
	feeTokens, err = s.app.FeeAbstractionKeeper.FeeTokens.Get(s.ctx)
	s.Require().NoError(err)
	s.Require().Len(feeTokens.Items, 1)
	s.Require().False(feeTokens.Items[0].Enabled)
}
