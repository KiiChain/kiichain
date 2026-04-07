package keeper_test

import (
	"errors"
	"time"

	"cosmossdk.io/collections"
	math "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/app/params"
	"github.com/kiichain/kiichain/v7/x/feeabstraction/types"
	oracletypes "github.com/kiichain/kiichain/v7/x/oracle/types"
)

var (
	defaultKii  = types.NewFeeTokenMetadata(params.BaseDenom, params.DisplayDenom, 18)
	defaultAtom = types.NewFeeTokenMetadata("uatom", "atom", 6)
	defaultSol  = types.NewFeeTokenMetadata("usol", "sol", 18)
)

// TestCalculateFeeTokenPrices tests the CalculateFeeTokenPrices function
func (s *KeeperTestSuite) TestCalculateFeeTokenPrices() {
	// Prepare the test cases
	testCases := []struct {
		name        string
		malleate    func(ctx sdk.Context) sdk.Context
		postCheck   func(ctx sdk.Context)
		errContains string
	}{
		{
			name: "empty sets",
		},
		{
			name: "module disabled whithout kii price",
			postCheck: func(ctx sdk.Context) {
				// Check params is disabled
				params, err := s.app.FeeAbstractionKeeper.Params.Get(ctx)
				s.Require().NoError(err)
				s.Require().False(params.Enabled)
			},
		},
		{
			name: "calculate fee token prices",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Mock oracle twaps
				startTime := ctx.BlockTime()
				ctx = s.createTwaps(ctx, startTime, math.LegacyMustNewDecFromStr("0.01"), 100, "kii")
				ctx = s.createTwaps(ctx, startTime, math.LegacyMustNewDecFromStr("0.5"), 100, "atom")

				// Set the fee token prices in the keeper
				err := s.app.FeeAbstractionKeeper.FeeTokens.Set(ctx, *types.NewFeeTokenMetadataCollection(
					types.NewFeeTokenMetadata("uatom", "atom", 6),
					defaultKii,
				))
				s.Require().NoError(err)
				// Set the initial uatom price for clamping reference
				err = s.app.FeeAbstractionKeeper.TokenPrices.Set(ctx, "uatom", math.LegacyMustNewDecFromStr("50"))
				s.Require().NoError(err)

				return ctx
			},
			postCheck: func(ctx sdk.Context) {
				// All tokens are still enabled
				feeTokens, err := s.app.FeeAbstractionKeeper.FeeTokens.Get(ctx)
				s.Require().NoError(err)
				for _, token := range feeTokens.Items {
					s.Require().True(token.Enabled, "Expected token to be enabled: %s", token.Denom)
				}

				// Calculate the token from one to the other
				expectedPrice := math.LegacyMustNewDecFromStr("0.5").Quo(math.LegacyMustNewDecFromStr("0.01"))
				expectedPriceMin := expectedPrice.Mul(math.LegacyMustNewDecFromStr("0.8")) // Allow 20% variance
				expectedPriceMax := expectedPrice.Mul(math.LegacyMustNewDecFromStr("1.2")) // Allow 20% variance

				// Check the price of the uatom token (should be in expected 20%)
				atomPrice, err := s.app.FeeAbstractionKeeper.TokenPrices.Get(ctx, "uatom")
				s.Require().NoError(err)
				s.Require().True(
					atomPrice.GTE(expectedPriceMin) && atomPrice.LTE(expectedPriceMax),
				)
			},
		},
		{
			name: "all tokens are disabled due to no twap",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Enable fee abstraction by setting kii price
				startTime := ctx.BlockTime()
				ctx = s.createTwaps(ctx, startTime, math.LegacyMustNewDecFromStr("0.5"), 100, "kii")

				// Set the fee tokens in the keeper without twaps
				err := s.app.FeeAbstractionKeeper.FeeTokens.Set(ctx, *types.NewFeeTokenMetadataCollection(
					defaultSol,
					defaultAtom,
				))
				s.Require().NoError(err)

				return ctx
			},
			postCheck: func(ctx sdk.Context) {
				// Check that the fee tokens are all disabled
				feeTokens, err := s.app.FeeAbstractionKeeper.FeeTokens.Get(ctx)
				s.Require().NoError(err)
				for _, token := range feeTokens.Items {
					s.Require().False(token.Enabled, "Expected token to be disabled: %s", token.Denom)
				}
			},
		},
		{
			name: "partial tokens disabled due to no twap",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Mock oracle twaps
				startTime := ctx.BlockTime()
				ctx = s.createTwaps(ctx, startTime, math.LegacyMustNewDecFromStr("0.4"), 100, "kii")
				ctx = s.createTwaps(ctx, startTime, math.LegacyMustNewDecFromStr("0.5"), 100, "atom")

				// Set the fee tokens in the keeper
				err := s.app.FeeAbstractionKeeper.FeeTokens.Set(ctx, *types.NewFeeTokenMetadataCollection(
					defaultSol,
					defaultAtom,
				))
				s.Require().NoError(err)

				return ctx
			},
			postCheck: func(ctx sdk.Context) {
				// Check that the fee tokens are partially disabled
				feeTokens, err := s.app.FeeAbstractionKeeper.FeeTokens.Get(ctx)
				s.Require().NoError(err)
				for _, token := range feeTokens.Items {
					if token.Denom == "uatom" {
						s.Require().True(token.Enabled, "Expected uatom to be enabled")
					} else {
						s.Require().False(token.Enabled, "Expected token to be disabled: %s", token.Denom)
					}
				}
			},
		},
		{
			name: "token enabled but price is zero",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Mock oracle twaps
				startTime := ctx.BlockTime()
				ctx = s.createTwaps(ctx, startTime, math.LegacyMustNewDecFromStr("0.4"), 100, "kii")
				ctx = s.createTwaps(ctx, startTime, math.LegacyMustNewDecFromStr("0.5"), 100, "atom")

				// Set the fee token prices in the keeper with zero price
				err := s.app.FeeAbstractionKeeper.FeeTokens.Set(ctx, *types.NewFeeTokenMetadataCollection(
					defaultKii,
					types.NewFeeTokenMetadata("uatom", "atom", 6),
				))
				s.Require().NoError(err)

				return ctx
			},
			postCheck: func(ctx sdk.Context) {
				// Price is set as normal
				feeTokens, err := s.app.FeeAbstractionKeeper.FeeTokens.Get(ctx)
				s.Require().NoError(err)
				s.Require().Len(feeTokens.Items, 2)
				atomPrice, err := s.app.FeeAbstractionKeeper.TokenPrices.Get(ctx, "uatom")
				s.Require().NoError(err)
				s.Require().NotEqualValues(math.LegacyZeroDec(), atomPrice)
				s.Require().True(feeTokens.Items[1].Enabled)
			},
		},
		{
			name: "partial tokens disabled, not price update",
			malleate: func(ctx sdk.Context) sdk.Context {
				// Mock oracle twaps
				startTime := ctx.BlockTime()
				ctx = s.createTwaps(ctx, startTime, math.LegacyMustNewDecFromStr("0.4"), 100, "kii")
				ctx = s.createTwaps(ctx, startTime, math.LegacyMustNewDecFromStr("0.5"), 100, "atom")
				ctx = s.createTwaps(ctx, startTime, math.LegacyMustNewDecFromStr("0.5"), 100, "sol")

				// Set the fee token prices in the keeper
				err := s.app.FeeAbstractionKeeper.FeeTokens.Set(ctx, *types.NewFeeTokenMetadataCollection(
					types.FeeTokenMetadata{
						Denom:       "uatom",
						OracleDenom: "atom",
						Decimals:    6,
						Enabled:     false, // Token disabled
					},
					defaultSol,
				))
				s.Require().NoError(err)
				// Set the stored price for the disabled token (should remain untouched)
				err = s.app.FeeAbstractionKeeper.TokenPrices.Set(ctx, "uatom", math.LegacyMustNewDecFromStr("50"))
				s.Require().NoError(err)
				// Set the initial price for usol so we can verify it changed
				err = s.app.FeeAbstractionKeeper.TokenPrices.Set(ctx, "usol", math.LegacyOneDec())
				s.Require().NoError(err)

				return ctx
			},
			postCheck: func(ctx sdk.Context) {
				// Check that the fee tokens are partially disabled
				feeTokens, err := s.app.FeeAbstractionKeeper.FeeTokens.Get(ctx)
				s.Require().NoError(err)
				s.Require().Len(feeTokens.Items, 2)

				// Check the first token is disabled and price is untouched
				s.Require().False(feeTokens.Items[0].Enabled)
				atomPrice, err := s.app.FeeAbstractionKeeper.TokenPrices.Get(ctx, "uatom")
				s.Require().NoError(err)
				s.Require().Equal(math.LegacyMustNewDecFromStr("50"), atomPrice)

				// Check the second token is enabled and price has changed
				s.Require().True(feeTokens.Items[1].Enabled)
				solPrice, err := s.app.FeeAbstractionKeeper.TokenPrices.Get(ctx, "usol")
				s.Require().NoError(err)
				s.Require().NotEqual(math.LegacyOneDec(), solPrice)
			},
		},
	}

	// Iterate through the test cases
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Set a cached context
			cachedCtx, _ := s.ctx.CacheContext()

			// Malleate the context
			if tc.malleate != nil {
				cachedCtx = tc.malleate(cachedCtx)
			}

			// Call the function under test
			err := s.keeper.CalculateFeeTokenPrices(cachedCtx)
			if tc.errContains != "" {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)
			}

			// Post check the context
			if tc.postCheck != nil {
				tc.postCheck(cachedCtx)
			}
		})
	}
}

// CreateTwaps for oracle keeper tests
func (s *KeeperTestSuite) createTwaps(
	ctx sdk.Context,
	startTime time.Time,
	startRate math.LegacyDec,
	steps int, //nolint:unparam
	denom string,
) sdk.Context {
	s.T().Helper()
	// Get initial values for the context
	height := ctx.BlockHeight()

	for i := 0; i < steps; i++ {
		// Each snapshot 3 seconds apart
		snapshotTime := startTime.Add(time.Second * 3 * time.Duration(i))
		snapshotHeight := height + int64(i)

		snapshotItems := oracletypes.PriceSnapshotItems{
			oracletypes.NewPriceSnapshotItem(denom, oracletypes.OracleExchangeRate{
				ExchangeRate:        startRate,
				LastUpdate:          math.NewInt(snapshotHeight),
				LastUpdateTimestamp: snapshotTime.Unix(),
			}),
		}

		snapshot, err := s.app.OracleKeeper.PriceSnapshot.Get(ctx, snapshotTime.Unix())
		if errors.Is(err, collections.ErrNotFound) {
			// Create the snapshot
			snapshot = oracletypes.NewPriceSnapshot(
				snapshotTime.Unix(),
				snapshotItems,
			)
		} else {
			s.Require().NoError(err)
			snapshot.PriceSnapshotItems = append(snapshot.PriceSnapshotItems, snapshotItems...)
		}

		// Set the snapshot in the keeper
		err = s.app.OracleKeeper.PriceSnapshot.Set(ctx, snapshot.SnapshotTimestamp, snapshot)
		s.Require().NoError(err)

		// Vary the price slightly for each step by 0.1%
		startRate = startRate.Mul(math.LegacyMustNewDecFromStr("1.001"))

		// Advance the context
		ctx = ctx.WithBlockHeight(snapshotHeight).WithBlockTime(snapshotTime)
	}

	// Set the token as a vote target
	err := s.app.OracleKeeper.VoteTarget.Set(ctx, denom, oracletypes.Denom{Name: denom})
	s.Require().NoError(err)

	return ctx
}
