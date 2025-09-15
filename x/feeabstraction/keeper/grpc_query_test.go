package keeper_test

import (
	"cosmossdk.io/math"

	"github.com/kiichain/kiichain/v4/x/feeabstraction/types"
)

// TestQuerierParams tests the Params querier
func (s *KeeperTestSuite) TestQuerierParams() {
	// Define new params for the chain
	newParams := types.NewParams(
		"testcoin",
		"testcoinoracle",
		types.DefaultClampFactor.MulInt64(2),
		types.DefaultFallbackNativePrice,
		types.DefaultTwapLookbackWindow,
		false,
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

// TestQuerierFeeTokens tests the FeeTokens querier
func (s *KeeperTestSuite) TestQuerierFeeTokens() {
	// Define new fee tokens
	newFeeTokens := types.NewFeeTokenMetadataCollection(
		types.NewFeeTokenMetadata("testcoin", "oracleCoin", 6, math.LegacyMustNewDecFromStr("0.01")),
	)

	// Set the fee tokens in the keeper
	err := s.keeper.FeeTokens.Set(s.ctx, *newFeeTokens)
	s.Require().NoError(err)

	// Query the fee tokens
	res, err := s.querier.FeeTokens(s.ctx, &types.QueryFeeTokensRequest{})
	s.Require().NoError(err)

	// Check the response
	s.Require().Equal(newFeeTokens, res.FeeTokens)
}
