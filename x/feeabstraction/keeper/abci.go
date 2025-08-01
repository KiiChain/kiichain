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

	// Apply the price calculation logic
	if err := k.CalculateFeeTokenPrices(sdkCtx); err != nil {
		return err
	}

	// Write the fee token prices to telemetry metrics
	if err := k.WriteFeeTokenPricesMetrics(sdkCtx); err != nil {
		return err
	}

	return nil
}

// WriteFeeTokenPricesMetrics writes the fee token prices to telemetry metrics
func (k Keeper) WriteFeeTokenPricesMetrics(ctx context.Context) error {
	// Get the fee token prices
	feeTokenPrices, err := k.FeeTokens.Get(ctx)
	if err != nil {
		return err
	}

	// Iterate over the fee token prices and set the gauge metrics
	for _, price := range feeTokenPrices.Items {
		// Set a module metric for each token
		telemetry.ModuleSetGauge(
			types.ModuleName,
			float32(price.Price.MustFloat64()),
			"fee_token_price",
			price.Denom,
		)
	}

	return nil
}
