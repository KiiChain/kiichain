package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/kiichain/kiichain3/x/oracle/types"
)

// Keeper manages the oracle module's state
type Keeper struct {
	storeKey   sdk.StoreKey         // storage key to access the module's state
	cdc        codec.BinaryCodec    // Codec for binary serialization
	paramSpace paramstypes.Subspace // Manages the module's parameters
}

func NewKeeper(storeKey sdk.StoreKey, cdc codec.BinaryCodec, paramSpace paramstypes.Subspace) Keeper {
	// Ensure paramstore is properly initialized
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:   storeKey,
		cdc:        cdc,
		paramSpace: paramSpace,
	}
}

func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var params types.Params

	k.paramSpace.GetParamSet(ctx, &params)
	return params
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
