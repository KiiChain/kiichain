package ante_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	tenderminttypes "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authate "github.com/cosmos/cosmos-sdk/x/auth/ante"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	evmtypes "github.com/cosmos/evm/x/vm/types"

	"github.com/kiichain/kiichain/v1/ante"
	"github.com/kiichain/kiichain/v1/app/apptesting"
	"github.com/kiichain/kiichain/v1/app/helpers"
	kiiparams "github.com/kiichain/kiichain/v1/app/params"
	oracletypes "github.com/kiichain/kiichain/v1/x/oracle/types"
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
			tx, err := buildTxFromMsgs(funder, tc.msgs...)
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

// buildTxFromMsgs builds a tx from a set of messages
func buildTxFromMsgs(feePayer sdk.AccAddress, msgs ...sdk.Msg) (xauthsigning.Tx, error) {
	// Start the tx builder
	encodingConfig := kiiparams.MakeEncodingConfig()
	txBuilder := encodingConfig.TxConfig.NewTxBuilder()

	// Set the messages
	err := txBuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}

	// Set the fee payer
	txBuilder.SetFeePayer(feePayer) // Replace with actual address

	// Set gas limit and fee amount
	txBuilder.SetGasLimit(1000000)
	txBuilder.SetFeeAmount(sdk.NewCoins(feeCoin))

	return txBuilder.GetTx(), nil
}
