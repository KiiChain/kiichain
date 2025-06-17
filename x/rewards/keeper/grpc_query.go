package keeper

import (
	"context"

	"github.com/kiichain/kiichain/v1/x/rewards/types"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

func NewQuerier(keeper Keeper) Querier {
	return Querier{Keeper: keeper}
}

// Params queries params of rewards module
func (k Querier) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params, err := k.Keeper.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: params}, nil
}

// RewardPool queries the reward pool coins
func (k Querier) RewardPool(ctx context.Context, _ *types.QueryRewardPoolRequest) (*types.QueryRewardPoolResponse, error) {
	pool, err := k.Keeper.RewardPool.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryRewardPoolResponse{RewardPool: pool}, nil
}

// ReleaseSchedule queries the schedule information
func (k Querier) ReleaseSchedule(ctx context.Context, _ *types.QueryReleaseScheduleRequest) (*types.QueryReleaseScheduleResponse, error) {
	schedule, err := k.Keeper.ReleaseSchedule.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryReleaseScheduleResponse{ReleaseSchedule: schedule}, nil
}
