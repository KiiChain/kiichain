package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain3/x/oracle/types"
)

// GetParams returns the Oracle module's params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	params := types.Params{}
	k.paramSpace.GetParamSet(ctx, &params)
	return params
}
