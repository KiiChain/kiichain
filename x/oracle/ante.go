package oracle

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"cosmossdk.io/errors"
	"github.com/kiichain/kiichain/v1/x/oracle/keeper"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
)

// SpammingPreventionDecorator will check if
type SpammingPreventionDecorator struct {
	oracleKepper keeper.Keeper
}

// NewSpammingPreventionDecorator creates a new instance of spamming prevention decorator
func NewSpammingPreventionDecorator(keeper keeper.Keeper) SpammingPreventionDecorator {
	return SpammingPreventionDecorator{
		oracleKepper: keeper,
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
			err = spd.oracleKepper.ValidateFeeder(ctx, feederAddr, valAddr)
			if err != nil {
				return err
			}

			// check if the validator has voted on that block height
			spamPreventionHeight := spd.oracleKepper.GetSpamPreventionCounter(ctx, valAddr)
			if spamPreventionHeight == currentHeight {
				return errors.Wrap(sdkerrors.ErrConflict, fmt.Sprintf("the validator has already submitted a vote at the current height=%d", currentHeight))
			}

			// set the anti spam block height
			spd.oracleKepper.SetSpamPreventionCounter(ctx, valAddr)
			continue
		default:
			return nil
		}
	}
	return nil
}

// VoteAloneDecorator implements the AnteFullDecorator needed to be registrated as a decorator
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
