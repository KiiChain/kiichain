package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"
	oracletypes "github.com/kiichain/kiichain/v3/x/oracle/types"
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
		// If we have an error we can set that the twp is zero
		twaps = oracletypes.OracleTwaps{}
	}

	// Parse the twaps into a map for easier access
	twapPriceMap := make(map[string]math.LegacyDec)
	for _, twap := range twaps {
		twapPriceMap[twap.Denom] = twap.Twap
	}

	// Find the price for the base token
	baseTokenPrice, ok := twapPriceMap[params.NativeDenom]
	if !ok {
		baseTokenPrice = params.FallbackNativePrice
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
		tokenPrice := twapPriceMap[token.OracleDenom]

		// If the token price is zero, we disable the token for safety
		if tokenPrice.IsZero() {
			token.Enabled = false
			token.Price = math.LegacyZeroDec()
			updateTokens = append(updateTokens, token)
			continue
		}

		// Calculate the price of the token in terms of the base token
		price, err := types.CalculateTokenPrice(baseTokenPrice, tokenPrice)
		if err != nil {
			return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "error calculating token price for denom %s: %v", token.Denom, err)
		}

		// Apply clamping
		price = types.ClampPrice(token.Price, price, clampFactor)

		// Update the token price
		token.Price = price
		updateTokens = append(updateTokens, token)
	}

	// Return the updated tokens
	return updateTokens, nil
}
