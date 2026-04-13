package keeper

import (
	"context"

	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	query "github.com/cosmos/cosmos-sdk/types/query"

	"github.com/kiichain/kiichain/v7/x/tokenfactory/types"
)

// GetAuthorityMetadata returns the authority metadata for a specific denom
func (k Keeper) GetAuthorityMetadata(ctx context.Context, denom string) (types.DenomAuthorityMetadata, error) {
	bz := k.GetDenomPrefixStore(sdk.UnwrapSDKContext(ctx), denom).Get([]byte(types.DenomAuthorityMetadataKey))

	metadata := types.DenomAuthorityMetadata{}
	err := proto.Unmarshal(bz, &metadata)
	if err != nil {
		return types.DenomAuthorityMetadata{}, err
	}
	return metadata, nil
}

// setAuthorityMetadata stores authority metadata for a specific denom and keeps the
// admin secondary index consistent. It reads the current admin so it can remove the
// denom from the old admin's index before inserting it under the new admin
func (k Keeper) setAuthorityMetadata(ctx context.Context, denom string, metadata types.DenomAuthorityMetadata) error {
	err := metadata.Validate()
	if err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Remove denom from the current admin's index (if any metadata already exists)
	oldMetadata, err := k.GetAuthorityMetadata(sdkCtx, denom)
	if err != nil {
		return err
	}
	k.removeDenomFromAdmin(sdkCtx, oldMetadata.Admin, denom)

	store := k.GetDenomPrefixStore(sdkCtx, denom)

	bz, err := proto.Marshal(&metadata)
	if err != nil {
		return err
	}

	store.Set([]byte(types.DenomAuthorityMetadataKey), bz)

	// Insert denom under the new admin's index.
	k.addDenomFromAdmin(sdkCtx, metadata.Admin, denom)

	return nil
}

// addDenomFromAdmin inserts denom into the secondary index for admin
func (k Keeper) addDenomFromAdmin(ctx context.Context, admin, denom string) {
	k.GetAdminPrefixStore(sdk.UnwrapSDKContext(ctx), admin).Set([]byte(denom), []byte(denom))
}

// AddDenomFromAdmin is the exported form of addDenomFromAdmin
func (k Keeper) AddDenomFromAdmin(ctx context.Context, admin, denom string) {
	k.addDenomFromAdmin(ctx, admin, denom)
}

// removeDenomFromAdmin removes denom from the secondary index for admin
func (k Keeper) removeDenomFromAdmin(ctx context.Context, admin, denom string) {
	k.GetAdminPrefixStore(sdk.UnwrapSDKContext(ctx), admin).Delete([]byte(denom))
}

func (k Keeper) setAdmin(ctx context.Context, metadata types.DenomAuthorityMetadata, denom string, admin string) error {
	metadata.Admin = admin

	return k.setAuthorityMetadata(ctx, denom, metadata)
}

// GetDenomsFromAdmin returns a paginated list of denoms the given admin.
// It has maximum of types.MaxPageSize entries per request
func (k Keeper) GetDenomsFromAdmin(ctx context.Context, admin string, req *query.PageRequest) ([]string, *query.PageResponse, error) {
	if req != nil && req.Limit > types.MaxPageSize {
		return nil, nil, errors.Wrapf(sdkerrors.ErrInvalidRequest,
			"page size %d exceeds maximum allowed %d", req.Limit, types.MaxPageSize)
	}

	store := k.GetAdminPrefixStore(sdk.UnwrapSDKContext(ctx), admin)

	var denoms []string
	pageRes, err := query.Paginate(store, req, func(key, _ []byte) error {
		denoms = append(denoms, string(key))
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return denoms, pageRes, nil
}
