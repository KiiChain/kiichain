package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/x/rewards/types"
	v1beta1 "github.com/kiichain/kiichain/v7/x/rewards/types/v1beta1"
)

type legacyQuerier struct {
	Keeper
}

var _ v1beta1.QueryServer = legacyQuerier{}

// NewLegacyQuerier serves the mainnet v1beta1 queries against current state.
func NewLegacyQuerier(keeper Keeper) v1beta1.QueryServer {
	return legacyQuerier{Keeper: keeper}
}

func (q legacyQuerier) Params(ctx context.Context, _ *v1beta1.QueryParamsRequest) (*v1beta1.QueryParamsResponse, error) {
	params, err := q.Keeper.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &v1beta1.QueryParamsResponse{Params: v1beta1.Params{TokenDenom: params.TokenDenom}}, nil
}

func (legacyQuerier) ReleaseSchedule(context.Context, *v1beta1.QueryReleaseScheduleRequest) (*v1beta1.QueryReleaseScheduleResponse, error) {
	return nil, fmt.Errorf("release schedules were removed")
}

// RewardPool reports the module account bank balance as community_pool so
// older clients still see spendable funds.
func (q legacyQuerier) RewardPool(ctx context.Context, _ *v1beta1.QueryRewardPoolRequest) (*v1beta1.QueryRewardPoolResponse, error) {
	params, err := q.Keeper.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	if params.TokenDenom == "" {
		params.TokenDenom = types.DefaultParams().TokenDenom
	}

	balance := q.bankKeeper.GetBalance(ctx, q.ModuleAddress(), params.TokenDenom)
	pool := v1beta1.RewardPool{}
	if balance.IsPositive() {
		pool.CommunityPool = sdk.NewDecCoins(sdk.NewDecCoinFromCoin(balance))
	}
	return &v1beta1.QueryRewardPoolResponse{RewardPool: pool}, nil
}
