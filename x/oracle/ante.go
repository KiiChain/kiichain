package oracle

import (
	"fmt"

	"github.com/pkg/errors"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kiichain/kiichain/v7/x/oracle/keeper"
	"github.com/kiichain/kiichain/v7/x/oracle/types"
)

// SpammingPreventionDecorator will check if
type SpammingPreventionDecorator struct {
	oracleKeeper *keeper.Keeper
}

// NewSpammingPreventionDecorator creates a new instance of spamming prevention decorator
func NewSpammingPreventionDecorator(keeper *keeper.Keeper) SpammingPreventionDecorator {
	return SpammingPreventionDecorator{
		oracleKeeper: keeper,
	}
}

// AnteHandle is the handler that checks if the transaction's validator has voted previously
func (spd SpammingPreventionDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if ctx.IsReCheckTx() {
		return next(ctx, tx, simulate)
	}

	if !simulate && ctx.IsCheckTx() {
		err := spd.CheckOracleSpamming(ctx, tx.GetMsgs())
		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}

// CheckOracleSpamming checks whether the msgs are spamming purpose or not
func (spd SpammingPreventionDecorator) CheckOracleSpamming(ctx sdk.Context, msgs []sdk.Msg) error {
	currentHeight := ctx.BlockHeight()

	for _, msg := range msgs {
		switch msg := msg.(type) {
		case *types.MsgAggregateExchangeRateVote:
			// validate a valid feeder address
			feederAddr, err := sdk.AccAddressFromBech32(msg.Feeder)
			if err != nil {
				return err
			}

			// validate a valid validator address
			valAddr, err := sdk.ValAddressFromBech32(msg.Validator)
			if err != nil {
				return err
			}

			// validate the feeder delegation is valid
			err = spd.oracleKeeper.ValidateFeeder(ctx, feederAddr, valAddr)
			if err != nil {
				return err
			}

			// SECURITY FIX: Check voting period instead of just block height
			// Get oracle parameters to determine voting period
			params, err := spd.oracleKeeper.Params.Get(ctx)
			if err != nil {
				return err
			}
			
			// SECURITY FIX: Prevent panic if VotePeriod is 0
			if params.VotePeriod == 0 {
				return errors.Wrap(sdkerrors.ErrInvalidRequest, "oracle vote period cannot be zero")
			}
			
			// Calculate current voting period
			currentVotingPeriod := currentHeight / int64(params.VotePeriod)
			
			// Check if validator has voted in this voting period
			spamPreventionHeight, err := spd.oracleKeeper.SpamPreventionCounter.Get(ctx, valAddr)
			if err != nil {
				// If the error is not found, its because the validator has not voted before
				if errors.Is(err, collections.ErrNotFound) {
					spamPreventionHeight = -1
				} else {
					return err
				}
			}
			
			// Calculate voting period of last vote
			lastVotingPeriod := spamPreventionHeight / int64(params.VotePeriod)
			
			if spamPreventionHeight != -1 && lastVotingPeriod == currentVotingPeriod {
				return errors.Wrap(sdkerrors.ErrConflict, fmt.Sprintf("the validator has already submitted a vote in voting period=%d (current height=%d)", currentVotingPeriod, currentHeight))
			}

			// set the anti spam block height
			err = spd.oracleKeeper.SetSpamPreventionCounterWithDefault(ctx, valAddr)
			if err != nil {
				return err
			}
			continue
		default:
			return nil
		}
	}
	return nil
}

// VoteAloneDecorator implements the AnteFullDecorator needed to be registered as a decorator
type VoteAloneDecorator struct{}

// NewVoteAloneDecorator returns a new instance of VoteAloneDecorator
func NewVoteAloneDecorator() VoteAloneDecorator {
	return VoteAloneDecorator{}
}

// AnteHandle implements the AnteDecorator interfaces on the
// AnteHandler checks
func (VoteAloneDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	oracleVote := false
	otherMsg := false

	// Iterate over all messages on the transaction
	for _, msg := range tx.GetMsgs() {
		switch msg.(type) {
		case *types.MsgAggregateExchangeRateVote:
			oracleVote = true
		default:
			otherMsg = true
		}
	}

	if oracleVote && otherMsg {
		return ctx, errors.Wrap(sdkerrors.ErrInvalidRequest, "oracle votes cannot be in the same tx as other messages")
	}

	// Continue on the next decorator
	return next(ctx, tx, simulate)
}
