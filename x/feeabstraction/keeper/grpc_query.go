package keeper

import (
	"context"

	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
)

// Interface assertion for the querier
var _ types.QueryServer = Querier{}

// Querier is the Querier wrapper for the keeper
type Querier struct {
	Keeper
}

// NewQuerier returns a new querier
func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// Params queries the params of the module
func (q Querier) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	// Get the params from the keeper
	params, err := q.Keeper.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	// Return the response with the params
	return &types.QueryParamsResponse{Params: params}, nil
}
