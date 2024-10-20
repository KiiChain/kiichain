package migrations

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain3/x/evm/keeper"
	"github.com/kiichain/kiichain3/x/evm/types"
)

func MigrateDeliverTxHookWasmGasLimitParam(ctx sdk.Context, k *keeper.Keeper) error {
	// Fetch the v11 parameters
	keeperParams := k.GetParamsIfExists(ctx)

	// Add DeliverTxHookWasmGasLimit to with default value
	keeperParams.DeliverTxHookWasmGasLimit = types.DefaultParams().DeliverTxHookWasmGasLimit

	// Set the updated parameters back in the keeper
	k.SetParams(ctx, keeperParams)

	return nil
}
