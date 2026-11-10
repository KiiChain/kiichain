package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	v1beta1 "github.com/kiichain/kiichain/v7/x/rewards/types/v1beta1"
)

type legacyMsgServer struct {
	Keeper
}

var _ v1beta1.MsgServer = legacyMsgServer{}

// NewLegacyMsgServer returns handlers for mainnet v1beta1 messages.
// UpdateParams still applies token_denom. FundPool and ChangeSchedule are
// rejected: funds are a bank send to the module account, and schedules are gone.
func NewLegacyMsgServer(keeper Keeper) v1beta1.MsgServer {
	return &legacyMsgServer{Keeper: keeper}
}

func (k legacyMsgServer) UpdateParams(ctx context.Context, msg *v1beta1.MsgUpdateParams) (*v1beta1.MsgUpdateParamsResponse, error) {
	if err := k.validateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	if err := sdk.ValidateDenom(msg.Params.TokenDenom); err != nil {
		return nil, fmt.Errorf("invalid token denom: %w", err)
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	params.TokenDenom = msg.Params.TokenDenom
	if err := params.ValidateBasic(); err != nil {
		return nil, err
	}
	if err := k.Params.Set(ctx, params); err != nil {
		return nil, err
	}

	return &v1beta1.MsgUpdateParamsResponse{}, nil
}

func (legacyMsgServer) FundPool(context.Context, *v1beta1.MsgFundPool) (*v1beta1.MsgFundPoolResponse, error) {
	return nil, fmt.Errorf("fund-pool was removed; send tokens to the rewards module account")
}

func (legacyMsgServer) ChangeSchedule(context.Context, *v1beta1.MsgChangeSchedule) (*v1beta1.MsgChangeScheduleResponse, error) {
	return nil, fmt.Errorf("release schedules were removed; emissions follow supply_base")
}
