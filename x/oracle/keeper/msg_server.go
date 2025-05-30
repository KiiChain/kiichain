package keeper

import (
	"context"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kiichain/kiichain/v1/x/oracle/types"
)

type msgServer struct {
	Keeper
}

// Ensure msgServer implements the types.MsgServer interface
var _ types.MsgServer = msgServer{}

// NewMsgServer creates a new msg server instance with the oracle module's keeper as an input
func NewMsgServer(keeper Keeper) types.MsgServer {
	return msgServer{
		Keeper: keeper,
	}
}

// AggregateExchangeRateVote receive the exchange rate information, validate the feeder address (if it is allowed to perform that operation),
// then, check if the information is valid and finally add it into the exchange rate KVStore
func (ms msgServer) AggregateExchangeRateVote(ctx context.Context, msg *types.MsgAggregateExchangeRateVote) (*types.MsgAggregateExchangeRateVoteResponse, error) {
	// Get cosmos sdk context from golang context
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get the validator address who send the exchange rate from the input data
	valAddress, err := sdk.ValAddressFromBech32(msg.Validator)
	if err != nil {
		return nil, err
	}

	// convert feeder address to Account data type
	feederAddress, err := sdk.AccAddressFromBech32(msg.Feeder)
	if err != nil {
		return nil, err
	}

	// Validate feeder address
	err = ms.ValidateFeeder(sdkCtx, feederAddress, valAddress)
	if err != nil {
		return nil, err
	}

	// Convert string exchange rates to specific data types
	exchangeRates, err := types.ParseExchangeRateTuples(msg.ExchangeRates)
	if err != nil {
		return nil, errors.Wrap(sdkerrors.ErrInvalidCoins, err.Error())
	}

	// Check all denoms are in the vote target
	for _, exchangeRate := range exchangeRates {
		found, err := ms.Keeper.VoteTarget.Has(ctx, exchangeRate.Denom)
		if err != nil {
			return nil, err
		}

		// Check if found
		if !found {
			return nil, errors.Wrap(types.ErrUnknownDenom, exchangeRate.Denom)
		}
	}

	// aggregate the exchange rate prices from the feeder
	aggregateExchangeRateVote, err := types.NewAggregateExchangeRateVote(exchangeRates, valAddress)
	if err != nil {
		return nil, errors.Wrap(types.ErrAggregateVoteInvalidRate, exchangeRates.String())
	}

	err = ms.Keeper.AggregateExchangeRateVote.Set(sdkCtx, valAddress, aggregateExchangeRateVote)
	if err != nil {
		return nil, err
	}

	// Trigger events (exchange rate saved and the feeder address)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent( // Event with the exchange rate approved and added into the module
			types.EventTypeAggregateVote,
			sdk.NewAttribute(types.AttributeKeyVoter, msg.Validator),
			sdk.NewAttribute(types.AttributeKeyExchangeRates, msg.ExchangeRates),
		),
		sdk.NewEvent( // the Event with the information who send the information (the feeder address and the module name)
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Feeder),
		),
	})

	return &types.MsgAggregateExchangeRateVoteResponse{}, nil
}

// DelegateFeedConsent register a delegator address as a feeder (as a delegated address)
func (ms msgServer) DelegateFeedConsent(ctx context.Context, msg *types.MsgDelegateFeedConsent) (*types.MsgDelegateFeedConsentResponse, error) {
	// Get cosmos sdk context from golang context
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// get the validator address from the message
	validatorAddress, err := sdk.ValAddressFromBech32(msg.Operator)
	if err != nil {
		return nil, err
	}

	// Get the delegated address from the message
	delegatorAddress, err := sdk.AccAddressFromBech32(msg.Delegate)
	if err != nil {
		return nil, err
	}

	// check if the operador address is a validator (must be, because the operator is a validator)
	val, err := ms.StakingKeeper.Validator(sdkCtx, validatorAddress)
	if err != nil || val == nil {
		return nil, errors.Wrap(stakingtypes.ErrNoValidatorFound, msg.Operator)
	}

	// Assign the delegator from the validator address
	ms.SetFeederDelegation(sdkCtx, validatorAddress, delegatorAddress)

	// Trigger events (exchange rate saved and the feeder address)
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent( // the Event with the address to be registered as a delegated address (as a feeder)
			types.EventTypeFeedDelegate,
			sdk.NewAttribute(types.AttributeKeyFeeder, msg.Delegate),
		),
		sdk.NewEvent( // the Event with the information who send the information (the validator address and the module name)
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Operator),
		),
	})

	return &types.MsgDelegateFeedConsentResponse{}, nil
}
