package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/x/tokenfactory/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := k.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}

func (k Keeper) DenomAuthorityMetadata(ctx context.Context, req *types.QueryDenomAuthorityMetadataRequest) (*types.QueryDenomAuthorityMetadataResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	authorityMetadata, err := k.GetAuthorityMetadata(sdkCtx, req.GetDenom())
	if err != nil {
		return nil, err
	}

	return &types.QueryDenomAuthorityMetadataResponse{AuthorityMetadata: authorityMetadata}, nil
}

func (k Keeper) DenomsFromCreator(ctx context.Context, req *types.QueryDenomsFromCreatorRequest) (*types.QueryDenomsFromCreatorResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	denoms, pageRes, err := k.GetDenomsFromCreator(sdkCtx, req.GetCreator(), req.GetPagination())
	if err != nil {
		return nil, err
	}
	return &types.QueryDenomsFromCreatorResponse{Denoms: denoms, Pagination: pageRes}, nil
}

func (k Keeper) DenomsFromAdmin(ctx context.Context, req *types.QueryDenomsFromAdminRequest) (*types.QueryDenomsFromAdminResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	denoms, pageRes, err := k.GetDenomsFromAdmin(sdkCtx, req.GetAdmin(), req.GetPagination())
	if err != nil {
		return nil, err
	}
	return &types.QueryDenomsFromAdminResponse{Denoms: denoms, Pagination: pageRes}, nil
}
