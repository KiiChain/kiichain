package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/kiichain/kiichain/v7/x/feeabstraction/types"
	oracletypes "github.com/kiichain/kiichain/v7/x/oracle/types"
)

// CalculateFeeTokenPrices returns the price of the fee token in terms of the base token
func (k Keeper) CalculateFeeTokenPrices(ctx sdk.Context) error {
	// Get the params
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// Get the twaps for the tokens
	twaps, err := k.oracleKeeper.CalculateTwaps(ctx, params.TwapLookbackWindow)
	if err != nil {
		// Log or emit telemetry for monitoring
		k.Logger(ctx).Warn("failed to calculate TWAPs", "msg", err)
		// If we have an error we can set that the twp is zero
		twaps = oracletypes.OracleTwaps{}
	}

	// Parse the twaps into a map for easier access
	twapPriceMap := make(map[string]math.LegacyDec)
	for _, twap := range twaps {
		twapPriceMap[twap.Denom] = twap.Twap
	}

	// Find the price for the base token
	baseTokenPrice, ok := twapPriceMap[params.NativeOracleDenom]
	if !ok {
		// Disable fee abstraction if there is no pricing
		k.Logger(ctx).Debug("%s has no price, feeabstraction disabled", params.NativeOracleDenom)
		params.Enabled = false
		return k.Params.Set(ctx, params)
	}

	// Iterate all the tokens
	updateTokens, err := k.calculatePriceTokens(
		ctx,
		twapPriceMap,
		baseTokenPrice,
		params.ClampFactor,
	)
	if err != nil {
		return err
	}

	// Save the updated tokens
	return k.FeeTokens.Set(ctx, *types.NewFeeTokenMetadataCollection(updateTokens...))
}

// calculatePriceTokens calculates the price of each fee token in terms of the base token
func (k Keeper) calculatePriceTokens(
	ctx sdk.Context,
	twapPriceMap map[string]math.LegacyDec,
	baseTokenPrice math.LegacyDec,
	clampFactor math.LegacyDec,
) ([]types.FeeTokenMetadata, error) {
	// Get all the fee tokens
	feeTokens, err := k.FeeTokens.Get(ctx)
	if err != nil {
		return nil, err
	}

	// Iterate through the fee tokens and calculate their prices
	updateTokens := make([]types.FeeTokenMetadata, 0, len(feeTokens.Items))
	for _, token := range feeTokens.Items {
		// Check if the token is enabled
		if !token.Enabled {
			updateTokens = append(updateTokens, token)
			continue
		}

		// Missing TWAP, fallback to zero
		tokenPrice, ok := twapPriceMap[token.OracleDenom]
		if !ok {
			tokenPrice = math.LegacyZeroDec()
		}

		// If the token price is zero, we disable the token for safety
		if tokenPrice.IsZero() {
			// Log or emit telemetry for monitoring
			k.Logger(ctx).Warn("token price is zero, disabling token", "denom", token.Denom)
			// Disable the token
			token.Enabled = false
			if err := k.TokenPrices.Set(ctx, token.Denom, math.LegacyZeroDec()); err != nil {
				return nil, err
			}
			updateTokens = append(updateTokens, token)
			continue
		}

		// Calculate the price of the token in terms of the base token
		price, err := types.CalculateTokenPrice(baseTokenPrice, tokenPrice)
		if err != nil {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "error calculating token price for denom %s: %v", token.Denom, err)
		}

		// Get the previous stored price for clamping
		prevPrice, err := k.TokenPrices.Get(ctx, token.Denom)
		if err != nil {
			// If price doesn't exist, take it as 0
			prevPrice = math.LegacyZeroDec()
		}

		// Apply clamping
		price = types.ClampPrice(prevPrice, price, clampFactor)

		// Store the updated price
		if err := k.TokenPrices.Set(ctx, token.Denom, price); err != nil {
			return nil, err
		}
		updateTokens = append(updateTokens, token)
	}

	// Return the updated tokens
	return updateTokens, nil
}
