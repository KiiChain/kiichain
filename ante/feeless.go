package ante

import (
	"errors"
	"math"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"

	oraclekeeper "github.com/kiichain/kiichain/v1/x/oracle/keeper"
	oracletypes "github.com/kiichain/kiichain/v1/x/oracle/types"
)

// FeelessDecorator defines a decorator that allows feeless transaction based on conditions
type FeelessDecorator struct {
	// feeDecorator is the SDK fee decorator that deducts fees from the fee payer
	feeDecorator sdk.AnteDecorator
	// oracleKeeper is one of the modules that has feeless transactions
	oracleKeeper *oraclekeeper.Keeper
}

// Type assertion for the FeelessDecorator
var _ sdk.AnteDecorator = FeelessDecorator{}

// NewFeelessDecorator creates a new FeelessDecorator
func NewFeelessDecorator(feeDecorator sdk.AnteDecorator, oracleKeeper *oraclekeeper.Keeper) FeelessDecorator {
	return FeelessDecorator{
		feeDecorator: feeDecorator,
		oracleKeeper: oracleKeeper,
	}
}

// AnteHandle executes the antehandler logic for feeless transactions
// This checks if the transaction is gasless and if so, it skips the fee deduction
// This also force the TX to take
func (gd FeelessDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// Check if the transaction is gasless
	isFeeless, err := gd.IsTxFeeless(ctx, tx)
	if err != nil {
		return ctx, err
	}

	// If feeless, ignore fee deduction
	if isFeeless {
		ctx = ctx.WithPriority(math.MaxInt64)
		return next(ctx, tx, simulate)
	}

	// Go to the next ante handler using the No Consumption Gas Meter
	return gd.feeDecorator.AnteHandle(ctx, tx, simulate, next)
}

// IsTxFeeless checks if the transaction is feeless
func (gd FeelessDecorator) IsTxFeeless(ctx sdk.Context, tx sdk.Tx) (bool, error) {
	// Check if the transaction has exactly one message
	// If it has any amount different than one, we can return that its not gasless
	// This protects against DDoS attacks where a transaction has multiple messages
	if len(tx.GetMsgs()) != 1 {
		return false, nil
	}

	// Iterate all the msgs on the tx
	for _, msg := range tx.GetMsgs() {
		switch m := msg.(type) {
		case *oracletypes.MsgAggregateExchangeRateVote:
			// Check if the message message is a feeless message
			return gd.MsgAggregateExchangeRateVoteIsFeeless(ctx, m)
		default:
			// We can return that its not gasless
			return false, nil
		}
	}

	return false, nil
}

// MsgAggregateExchangeRateVoteIsFeeless checks if the MsgAggregateExchangeRateVote is feeless
// A feeless MsgAggregateExchangeRateVote is one that has not been casted yet
// and the feeder is allowed to vote for the validator
func (gd FeelessDecorator) MsgAggregateExchangeRateVoteIsFeeless(ctx sdk.Context, msg *oracletypes.MsgAggregateExchangeRateVote) (bool, error) {
	// Validate the feeder address
	feederAddr, err := sdk.AccAddressFromBech32(msg.Feeder)
	if err != nil {
		return false, err
	}

	// Validate the validator address
	valAddr, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return false, err
	}

	// Validate if the feeder is allowed to vote
	err = gd.oracleKeeper.ValidateFeeder(ctx, feederAddr, valAddr)
	if err != nil {
		return false, err
	}

	// Check if a vote was already casted
	_, err = gd.oracleKeeper.AggregateExchangeRateVote.Get(ctx, valAddr)

	// If the error is not nil and the error is not found means that the vote was not casted yet,
	if err != nil && errors.Is(err, collections.ErrNotFound) {
		// This means that the vote is feeless
		return true, nil
	}

	// Reaching this point means that the data exists or there is an error other than not found
	return false, err
}
