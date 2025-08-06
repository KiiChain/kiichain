package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kiichain/kiichain/v4/x/rewards/types"
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

// ChangeSchedule validates changes to the release scheduler
func (k msgServer) ChangeSchedule(ctx context.Context, msg *types.MsgChangeSchedule) (*types.MsgChangeScheduleResponse, error) {
	// Authority validation
	if err := k.validateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	// Check if schedule is sound
	schedule := msg.Schedule
	if err := k.validateSchedule(ctx, schedule); err != nil {
		return nil, fmt.Errorf("invalid schedule: %w", err)
	}

	// Check available funds
	if err := k.fundsAvailable(ctx, schedule.TotalAmount); err != nil {
		return nil, fmt.Errorf("insufficient funds: %w", err)
	}

	// Save the new schedule
	if err := k.Keeper.ReleaseSchedule.Set(ctx, schedule); err != nil {
		return nil, fmt.Errorf("failed to set release schedule: %w", err)
	}

	return &types.MsgChangeScheduleResponse{}, nil
}
