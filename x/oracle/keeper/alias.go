package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v2/x/oracle/types"
)

// GetOracleAccount returns the module name stores on the auth module (to know that the oracle have an account)
func (k Keeper) GetOracleAccount(ctx sdk.Context) sdk.ModuleAccountI {
	return k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
}
