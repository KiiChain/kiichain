package oracle

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkacltypes "github.com/cosmos/cosmos-sdk/types/accesscontrol"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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
				return sdkerrors.Wrap(sdkerrors.ErrAlreadyExists, fmt.Sprintf("the validator has already submitted a vote at the current height=%d", currentHeight))
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

// AnteDeps implements the AnteFullDecorator interface, required to register SpammingPreventionDecorator as decorator
func (spd SpammingPreventionDecorator) AnteDeps(txDeps []sdkacltypes.AccessOperation, tx sdk.Tx, txIndex int, next sdk.AnteDepGenerator) ([]sdkacltypes.AccessOperation, error) {
	deps := []sdkacltypes.AccessOperation{} // Here I will store the dependencies

	// Iterate over all messages inside the transaction
	for _, msg := range tx.GetMsgs() {
		switch msg := msg.(type) {

		// Process the aggregate exchange rate messages
		case *types.MsgAggregateExchangeRateVote:
			valAddrs, _ := sdk.ValAddressFromBech32(msg.Feeder)
			deps = append(deps, []sdkacltypes.AccessOperation{
				// validate feeder
				{
					ResourceType:       sdkacltypes.ResourceType_KV_ORACLE_FEEDERS,
					AccessType:         sdkacltypes.AccessType_READ,
					IdentifierTemplate: hex.EncodeToString(types.GetFeederDelegationKey(valAddrs)),
				},

				// Validate the validator exists
				{
					ResourceType:       sdkacltypes.ResourceType_KV_STAKING_VALIDATOR,
					AccessType:         sdkacltypes.AccessType_READ,
					IdentifierTemplate: hex.EncodeToString(stakingtypes.GetValidatorKey(valAddrs)),
				},

				// Check exchange rate exists
				{
					ResourceType:       sdkacltypes.ResourceType_KV_ORACLE_AGGREGATE_VOTES,
					AccessType:         sdkacltypes.AccessType_READ,
					IdentifierTemplate: hex.EncodeToString(types.GetAggregateExchangeRateVoteKey(valAddrs)),
				},
			}...)
		default:
			continue
		}
	}

	// add the new dependencies (deps) with the previous ones (txDeps) and are passed to the next decorator
	return next(append(txDeps, deps...), tx, txIndex)
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
		return ctx, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "oracle votes cannot be in the same tx as other messages")
	}

	// Continue on the next decorator
	return next(ctx, tx, simulate)
}

// AnteDeps implements the AnteDepDecorator interface
// AnteDeps collects the dependencies the vote alone decorator needs
func (VoteAloneDecorator) AnteDeps(txDeps []sdkacltypes.AccessOperation, tx sdk.Tx, txIndex int, next sdk.AnteDepGenerator) (newTxDeps []sdkacltypes.AccessOperation, err error) {
	// requires no dependencies
	return next(txDeps, tx, txIndex)
}
