package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kiichain/kiichain3/x/tokenfactory/types"
)

// Interface check for the query server
var _ types.QueryServer = Keeper{}

// Params implements Query/Params gRPC method.
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

// DenomsFromCreator implements Query/DenomsFromCreator gRPC method.
func (k Keeper) DenomsFromCreator(c context.Context, req *types.QueryDenomsFromCreatorRequest) (*types.QueryDenomsFromCreatorResponse, error) {
	// Unwrap the context
	ctx := sdk.UnwrapSDKContext(c)

	// Validate the request
	if req == nil || req.GetCreator() == "" {
		return nil, status.Error(codes.InvalidArgument, "creator address cannot be empty")
	}

	// Get the creator
	creator := req.GetCreator()

	// Prepare the store (this deprecates the old getDenomsFromCreator)
	prefixStore := k.GetCreatorPrefixStore(ctx, creator)

	// Paginate the response
	denoms := []string{}
	pageRes, err := query.Paginate(prefixStore, req.Pagination, func(key []byte, value []byte) error {
		denoms = append(denoms, string(value))
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Return the response
	return &types.QueryDenomsFromCreatorResponse{Denoms: denoms, Pagination: pageRes}, nil
}

// DenomMetadata implements Query/DenomMetadata gRPC method.
func (k Keeper) DenomMetadata(c context.Context, req *types.QueryDenomMetadataRequest) (*types.QueryDenomMetadataResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}

	ctx := sdk.UnwrapSDKContext(c)

	metadata, found := k.bankKeeper.GetDenomMetaData(ctx, req.Denom)
	if !found {
		return nil, status.Errorf(codes.NotFound, "client metadata for denom %s", req.Denom)
	}

	return &types.QueryDenomMetadataResponse{
		Metadata: metadata,
	}, nil
}

// DenomAllowList implements Query/DenomAllowList gRPC method.
func (k Keeper) DenomAllowList(c context.Context, req *types.QueryDenomAllowListRequest) (*types.QueryDenomAllowListResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}

	ctx := sdk.UnwrapSDKContext(c)

	allowList := k.bankKeeper.GetDenomAllowList(ctx, req.Denom)
	return &types.QueryDenomAllowListResponse{
		AllowList: allowList,
	}, nil
}
