package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
)

// BeginBlocker is called at the beginning of each block to calculate token prices for fees
func (k Keeper) BeginBlocker(ctx context.Context) error {
	// Apply telemetry metrics
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	// Unwrap the context
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Check if the module is enabled
	params, err := k.Params.Get(sdkCtx)
	if err != nil {
		return err
	}
	if !params.Enabled {
		return nil
	}

	// Calculate the token prices at the beginning of each block
	return k.CalculateFeeTokenPrices(sdkCtx)
}
