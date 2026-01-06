package ante_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tenderminttypes "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authate "github.com/cosmos/cosmos-sdk/x/auth/ante"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	evmtypes "github.com/cosmos/evm/x/vm/types"

	"github.com/kiichain/kiichain/v7/ante"
	"github.com/kiichain/kiichain/v7/app/apptesting"
	"github.com/kiichain/kiichain/v7/app/helpers"
	"github.com/kiichain/kiichain/v7/x/oracle"
	oracletypes "github.com/kiichain/kiichain/v7/x/oracle/types"
)

// Constant fee value for testing
var feeCoin = sdk.Coin{
	Denom:  "stake",
	Amount: math.NewInt(1000),
}

// TestFeelessDecorator tests the FeelessDecorator
func TestFeelessDecorator(t *testing.T) {
	// Start the app
	app := helpers.Setup(t)
	ctx := app.BaseApp.NewUncachedContext(true, tenderminttypes.Header{Height: 1, ChainID: "testing_1010-1", Time: time.Now().UTC()})

	// Create a fee payer
	funder := apptesting.RandomAccountAddress()

	// Get the chain validator address
	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	// Use the first validator as the fee payer validator
	funderVal, _ := sdk.ValAddressFromBech32(validators[0].GetOperator())

	// Fund the fee payer account
	err = app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("stake", 1000000)))
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, funder, sdk.NewCoins(sdk.NewInt64Coin("stake", 1000000)))
	require.NoError(t, err)

	// Start a new FeelessDecorator with the wrapped fee handler
	feelessDecorator := ante.NewFeelessDecorator(
		authate.NewDeductFeeDecorator(
			app.AccountKeeper,
			app.BankKeeper,
			nil,
			nil,
		),
		&app.OracleKeeper,
	)

	// Wrap into the sdk ante decorator
	anteHandler := sdk.ChainAnteDecorators(feelessDecorator)

	// Write the test cases
	testCases := []struct {
		name        string
		msgs        []sdk.Msg
		malleate    func(*testing.T, sdk.Context)
		isDeliverTx bool
		balanceDiff math.Int
	}{
		{
			name:        "No oracle message - deduct fee",
			msgs:        []sdk.Msg{banktypes.NewMsgSend(funder, funder, sdk.NewCoins(sdk.NewInt64Coin("stake", 123123)))},
			balanceDiff: feeCoin.Amount, // Expect the fee to be deducted
		},
		{
			name: "Oracle message - no fee deduction",
			msgs: []sdk.Msg{
				&oracletypes.MsgAggregateExchangeRateVote{
					ExchangeRates: "0.1stake,0.2stake",
					Feeder:        funder.String(),
					Validator:     funderVal.String(),
				},
			},
			malleate: func(t *testing.T, ctx sdk.Context) {
				t.Helper()
				// Register the validator and the feeder on the oracle keeper
				err := app.OracleKeeper.FeederDelegation.Set(ctx, funderVal, funder.String())
				require.NoError(t, err)
			},
			balanceDiff: math.ZeroInt(), // Expect no fee to be deducted
		},
		{
			name: "Oracle message - no fee deduction (deliver TX context)",
			msgs: []sdk.Msg{
				&oracletypes.MsgAggregateExchangeRateVote{
					ExchangeRates: "0.1stake,0.2stake",
					Feeder:        funder.String(),
					Validator:     funderVal.String(),
				},
			},
			isDeliverTx: true, // Set the context to deliver TX
			malleate: func(t *testing.T, ctx sdk.Context) {
				t.Helper()
				// Register the validator and the feeder on the oracle keeper
				err := app.OracleKeeper.FeederDelegation.Set(ctx, funderVal, funder.String())
				require.NoError(t, err)
			},
			balanceDiff: math.ZeroInt(), // Expect no fee to be deducted
		},
		{
			name: "Oracle and bank messages - should deduct fee",
			msgs: []sdk.Msg{
				&oracletypes.MsgAggregateExchangeRateVote{
					ExchangeRates: "0.1stake,0.2stake",
					Feeder:        funder.String(),
					Validator:     funderVal.String(),
				},
				banktypes.NewMsgSend(funder, funder, sdk.NewCoins(sdk.NewInt64Coin("stake", 123123))),
			},
			malleate: func(t *testing.T, ctx sdk.Context) {
				t.Helper()
				// Register the validator and the feeder on the oracle keeper
				err := app.OracleKeeper.FeederDelegation.Set(ctx, funderVal, funder.String())
				require.NoError(t, err)
			},
			balanceDiff: feeCoin.Amount, // Fee should be deducted for the because we have the bank message
		},
		{
			name: "Feeless double message - should deduct fee",
			msgs: []sdk.Msg{
				&oracletypes.MsgAggregateExchangeRateVote{
					ExchangeRates: "0.1stake,0.2stake",
					Feeder:        funder.String(),
					Validator:     funderVal.String(),
				},
				banktypes.NewMsgSend(funder, funder, sdk.NewCoins(sdk.NewInt64Coin("stake", 123123))),
			},
			malleate: func(t *testing.T, ctx sdk.Context) {
				t.Helper()
				// Register the validator and the feeder on the oracle keeper
				err := app.OracleKeeper.FeederDelegation.Set(ctx, funderVal, funder.String())
				require.NoError(t, err)
			},
			balanceDiff: feeCoin.Amount, // Fee should be deducted for the because we have the bank message
		},
		{
			name: "Oracle message but has voted - should deduct fee",
			msgs: []sdk.Msg{
				&oracletypes.MsgAggregateExchangeRateVote{
					ExchangeRates: "0.1stake,0.2stake",
					Feeder:        funder.String(),
					Validator:     funderVal.String(),
				},
			},
			malleate: func(t *testing.T, ctx sdk.Context) {
				t.Helper()
				// Register the validator and the feeder on the oracle keeper
				err := app.OracleKeeper.FeederDelegation.Set(ctx, funderVal, funder.String())
				require.NoError(t, err)

				// Register a vote for the validator
				err = app.OracleKeeper.AggregateExchangeRateVote.Set(ctx, funderVal, oracletypes.AggregateExchangeRateVote{
					ExchangeRateTuples: []oracletypes.ExchangeRateTuple{
						oracletypes.NewExchangeRateTuple("stake", math.LegacyNewDecWithPrec(1, 1)),
					},
					Voter: funder.String(),
				})
				require.NoError(t, err)
			},
			balanceDiff: feeCoin.Amount, // Fee should be deducted because the validator has already voted
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get a cached context
			cachedCtx, _ := ctx.CacheContext()

			// If a malleate function is provided, call it to modify the context
			if tc.malleate != nil {
				tc.malleate(t, cachedCtx)
			}

			// Build a tx from the messages
			tx, err := helpers.BuildTxFromMsgs(funder, nil, sdk.NewCoins(feeCoin), 1000000, tc.msgs...)
			require.NoError(t, err)

			// Take a sample of the user address
			balanceBefore := app.BankKeeper.GetBalance(cachedCtx, funder, "stake")

			// If the transaction is a deliver tx, set the context to deliver mode
			if tc.isDeliverTx {
				cachedCtx = cachedCtx.WithIsCheckTx(false).WithIsReCheckTx(false)
			}

			// Execute the ante handler
			_, err = anteHandler(cachedCtx, tx, false)
			require.NoError(t, err)

			// Check the balance difference, since we are using the same denom we can just subtract the amounts
			balanceAfter := app.BankKeeper.GetBalance(cachedCtx, funder, "stake")
			balanceDiff := balanceBefore.Amount.Sub(balanceAfter.Amount)

			// Check if the balance diff is as expected
			require.Equal(t, tc.balanceDiff.Int64(), balanceDiff.Int64(), "Balance difference should match expected value")
		})
	}
}

// TestVulnerabilityFeelessDoubleVoting validates AH-FL-002
// Tests non-atomic vote check allows double voting and mempool monopolization
func TestVulnerabilityFeelessDoubleVoting(t *testing.T) {
	// Start the app
	app := helpers.Setup(t)
	ctx := app.BaseApp.NewUncachedContext(true, tenderminttypes.Header{Height: 1, ChainID: "testing_1010-1", Time: time.Now().UTC()})

	// Create feeder
	feeder := apptesting.RandomAccountAddress()

	// Get validator
	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	valAddr, _ := sdk.ValAddressFromBech32(validators[0].GetOperator())

	// Fund feeder
	err = app.BankKeeper.MintCoins(ctx, evmtypes.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("stake", 1000000)))
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, evmtypes.ModuleName, feeder, sdk.NewCoins(sdk.NewInt64Coin("stake", 1000000)))
	require.NoError(t, err)

	// Register feeder
	err = app.OracleKeeper.FeederDelegation.Set(ctx, valAddr, feeder.String())
	require.NoError(t, err)

	t.Run("VULNERABILITY: Non-atomic check allows double voting", func(t *testing.T) {
		cachedCtx, _ := ctx.CacheContext()

		t.Logf("VULNERABILITY AH-FL-002 VALIDATION:")
		t.Logf("  Code location: feeless.go:100-107")
		t.Logf("")
		t.Logf("  VULNERABLE CODE:")
		t.Logf("    Line 101: _, err = gd.oracleKeeper.AggregateExchangeRateVote.Get(ctx, valAddr)")
		t.Logf("    Lines 104-106: if err != nil && errors.Is(err, collections.ErrNotFound) {")
		t.Logf("                       return true, nil  // ← RACE WINDOW")
		t.Logf("                   }")
		t.Logf("")
		t.Logf("  PROBLEM: Check and set are not atomic")
		t.Logf("")

		// Create two identical vote messages
		voteMsg1 := &oracletypes.MsgAggregateExchangeRateVote{
			ExchangeRates: "0.5stake",
			Feeder:        feeder.String(),
			Validator:     valAddr.String(),
		}
		voteMsg2 := &oracletypes.MsgAggregateExchangeRateVote{
			ExchangeRates: "0.5stake",
			Feeder:        feeder.String(),
			Validator:     valAddr.String(),
		}

		// Create feeless decorator
		feelessDecorator := ante.NewFeelessDecorator(
			authate.NewDeductFeeDecorator(app.AccountKeeper, app.BankKeeper, nil, nil),
			&app.OracleKeeper,
		)

		// Initialize the anti-spam decorator
		antiSpamDecorator := oracle.NewSpammingPreventionDecorator(&app.OracleKeeper)

		// Create the ante handler with both decorators
		anteHandler := sdk.ChainAnteDecorators(feelessDecorator, antiSpamDecorator)

		// Check 1: First vote check - should pass (no vote exists)
		tx1, err := helpers.BuildTxFromMsgs(feeder, nil, sdk.NewCoins(feeCoin), 1000000, voteMsg1)
		require.NoError(t, err)

		t.Logf("  TEST SCENARIO:")
		t.Logf("    Step 1: TX1 checks if vote exists - NOT FOUND ✓")
		_, err = anteHandler(cachedCtx, tx1, false)
		require.NoError(t, err)
		t.Logf("    Step 1 Result: TX1 antehandler passes (feeless)")

		// Check 2: Second vote check - ALSO passes because vote not yet set
		// This simulates the race condition where both txs check before either votes
		tx2, err := helpers.BuildTxFromMsgs(feeder, nil, sdk.NewCoins(feeCoin), 1000000, voteMsg2)
		require.NoError(t, err)

		t.Logf("    Step 2: TX2 checks if vote exists - STILL NOT FOUND ✓ (RACE!)")
		_, err = anteHandler(cachedCtx, tx2, false)
		require.Error(t, err)
		require.ErrorContains(t, err, "the validator has already submitted a vote at the current height=1")
	})

	t.Run("Priority manipulation demonstration", func(t *testing.T) {
		cachedCtx, _ := ctx.CacheContext()

		t.Logf("")
		t.Logf("  PRIORITY ATTACK:")
		t.Logf("    - Feeless oracle votes get MaxInt64 priority (9,223,372,036,854,775,807)")
		t.Logf("    - Regular user transactions get priority based on fee (typically 100-2000)")
		t.Logf("    - Validator can create unlimited feeless oracle votes")
		t.Logf("    - Result: Validator monopolizes block space")
		t.Logf("")

		// Create multiple vote messages to demonstrate spam
		spamVotes := make([]sdk.Msg, 10)
		for i := 0; i < 10; i++ {
			spamVotes[i] = &oracletypes.MsgAggregateExchangeRateVote{
				ExchangeRates: "0.5stake",
				Feeder:        feeder.String(),
				Validator:     valAddr.String(),
			}
		}

		feelessDecorator := ante.NewFeelessDecorator(
			authate.NewDeductFeeDecorator(app.AccountKeeper, app.BankKeeper, nil, nil),
			&app.OracleKeeper,
		)

		// Initialize the anti-spam decorator
		antiSpamDecorator := oracle.NewSpammingPreventionDecorator(&app.OracleKeeper)

		anteHandler := sdk.ChainAnteDecorators(feelessDecorator, antiSpamDecorator)

		// All spam votes pass antehandler
		passCount := 0
		for i := 0; i < 10; i++ {
			tx, err := helpers.BuildTxFromMsgs(feeder, nil, sdk.NewCoins(feeCoin), 1000000, spamVotes[i])
			require.NoError(t, err)

			_, err = anteHandler(cachedCtx, tx, false)
			if err == nil {
				passCount++
			}
		}

		// Only one should pass due to anti-spam
		if passCount != 1 {
			t.Logf("  SPAM TEST RESULT:")
			t.Logf("    - Created 10 oracle vote messages")
			t.Logf("    - All %d passed antehandler (feeless)", passCount)
			t.Logf("    - Each gets MaxInt64 priority")
			t.Logf("    - Validator can spam unlimited such txs")
			t.Logf("    - Block space monopolized by single validator")
			t.Errorf("expected only 1 message to pass the antehandler due to anti-spam decorator, but got %d", passCount)
		} else {
			t.Logf("  SPAM TEST RESULT:")
			t.Logf("	- Only 1 message was allowed to pass the antehandler due to anti-spam decorator")
		}
	})
}
