package keeper_test

import "github.com/kiichain/kiichain/v3/x/feeabstraction/types"

// TestQuerierParams tests the Params querier
func (s *KeeperTestSuite) TestQuerierParams() {
	// Define new params for the chain
	newParams := types.NewParams(
		"testcoin",
	)

	// Set the params in the keeper
	err := s.keeper.Params.Set(s.ctx, newParams)
	s.Require().NoError(err)

	// Query the params
	res, err := s.querier.Params(s.ctx, &types.QueryParamsRequest{})
	s.Require().NoError(err)

	// Check the response
	s.Require().Equal(newParams, res.Params)
}
