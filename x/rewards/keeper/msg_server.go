package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kiichain/kiichain/v1/x/rewards/types"
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the distribution MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// UpdateParams validates a MsgUpdate params and sets params to be accordingly
func (k msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if err := k.validateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	if err := msg.Params.ValidateBasic(); err != nil {
		return nil, err
	}

	if err := k.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// FundPool validates a MsgFundPool and sends the funds to the module and the pool
func (k msgServer) FundPool(ctx context.Context, msg *types.MsgFundPool) (*types.MsgFundPoolResponse, error) {
	depositor, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid depositor address: %s", err)
	}

	if err := validateAmount(msg.Amount); err != nil {
		return nil, err
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	if params.TokenDenom != msg.Amount.Denom {
		return nil, fmt.Errorf("denom %s does not match expected denom: %s", msg.Amount.Denom, params.TokenDenom)
	}

	if err := k.Keeper.FundCommunityPool(ctx, msg.Amount, depositor); err != nil {
		return nil, err
	}

	return &types.MsgFundPoolResponse{}, nil
}

// ExtendReward validates changes to the release scheduler
func (k msgServer) ExtendReward(ctx context.Context, msg *types.MsgExtendReward) (*types.MsgExtendRewardResponse, error) {
	if err := k.validateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	// Validate amt
	if err := validateAmount(msg.ExtraAmount); err != nil {
		return nil, err
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	if params.TokenDenom != msg.ExtraAmount.Denom {
		return nil, fmt.Errorf("denom %s does not match expected denom: %s", msg.ExtraAmount.Denom, params.TokenDenom)
	}

	// Validate time
	// Should only time extensions be allowed? I.e do not allow reducing the time
	if err := validateTime(msg.EndTime); err != nil {
		return nil, err
	}

	// Check if funds exist (community pool funds - to be released > extra amount)
	if err := k.fundsAvailable(ctx, msg.ExtraAmount); err != nil {
		return nil, err
	}

	// Do actual work
	if err := k.Keeper.ExtendReward(ctx, msg.ExtraAmount, msg.EndTime); err != nil {
		return nil, err
	}

	return &types.MsgExtendRewardResponse{}, nil
}
