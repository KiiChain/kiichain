package keeper

import (
	"context"

	"cosmossdk.io/errors"
	cosmosstore "cosmossdk.io/store"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	query "github.com/cosmos/cosmos-sdk/types/query"

	"github.com/kiichain/kiichain/v7/x/tokenfactory/types"
)

func (k Keeper) addDenomFromCreator(ctx context.Context, creator, denom string) {
	store := k.GetCreatorPrefixStore(sdk.UnwrapSDKContext(ctx), creator)
	store.Set([]byte(denom), []byte(denom))
}

// GetDenomsFromCreator returns a paginated list of denoms created by the given
// creator address. It has a maximum of types.MaxPageSize entries
func (k Keeper) GetDenomsFromCreator(ctx context.Context, creator string, req *query.PageRequest) ([]string, *query.PageResponse, error) {
	// Use max request size if none provided
	if req == nil {
		req = &query.PageRequest{Limit: types.MaxPageSize}
	} else if req.Limit > types.MaxPageSize {
		// If provided request size is above limit, return error
		return nil, nil, errors.Wrapf(sdkerrors.ErrInvalidRequest,
			"page size %d exceeds maximum allowed %d", req.Limit, types.MaxPageSize)
	}

	store := k.GetCreatorPrefixStore(sdk.UnwrapSDKContext(ctx), creator)

	// Append denoms from creator until pagination limit
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

func (k Keeper) GetAllDenomsIterator(ctx context.Context) cosmosstore.Iterator {
	return k.GetCreatorsPrefixStore(sdk.UnwrapSDKContext(ctx)).Iterator(nil, nil)
}
