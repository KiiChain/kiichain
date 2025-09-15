package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/kiichain/kiichain/v4/x/feeabstraction/types"
)

// TestValidateParams tests the Validate method of Params
func TestValidateParams(t *testing.T) {
	// Prepare test cases
	testCases := []struct {
		name        string
		params      types.Params
		errContains string
	}{
		{
			name:   "valid - default params",
			params: types.DefaultParams(),
		},
		{
			name: "valid - custom params",
			params: types.NewParams(
				"coin",
				"oraclecoin",
				types.DefaultClampFactor,
				types.DefaultFallbackNativePrice,
				types.DefaultTwapLookbackWindow,
				true,
			),
		},
		{
			name: "invalid - empty native denom",
			params: types.NewParams(
				"",
				"oraclecoin",
				types.DefaultClampFactor,
				types.DefaultFallbackNativePrice,
				types.DefaultTwapLookbackWindow,
				true,
			),
			errContains: "native denom is invalid",
		},
		{
			name: "invalid - invalid denom",
			params: types.NewParams(
				"123",
				"oraclecoin",
				types.DefaultClampFactor,
				types.DefaultFallbackNativePrice,
				types.DefaultTwapLookbackWindow,
				true,
			),
			errContains: "native denom is invalid",
		},
		{
			name: "invalid - empty oracle denom",
			params: types.NewParams(
				"coin",
				"",
				types.DefaultClampFactor,
				types.DefaultFallbackNativePrice,
				types.DefaultTwapLookbackWindow,
				true,
			),
			errContains: "native oracle denom is invalid",
		},
		{
			name: "invalid - empty oracle denom",
			params: types.NewParams(
				"coin",
				"123",
				types.DefaultClampFactor,
				types.DefaultFallbackNativePrice,
				types.DefaultTwapLookbackWindow,
				true,
			),
			errContains: "native oracle denom is invalid",
		},
		{
			name: "invalid - negative clamp factor",
			params: types.NewParams(
				"coin",
				"oraclecoin",
				types.DefaultClampFactor.Neg(), // Negative value
				types.DefaultFallbackNativePrice,
				types.DefaultTwapLookbackWindow,
				true,
			),
			errContains: "clamp factor must be between 0 and 1",
		},
		{
			name: "invalid - clamp factor greater than 1",
			params: types.NewParams(
				"coin",
				"oraclecoin",
				types.DefaultClampFactor.Add(math.LegacyOneDec()), // Greater than 1
				types.DefaultFallbackNativePrice,
				types.DefaultTwapLookbackWindow,
				true,
			),
			errContains: "clamp factor must be between 0 and 1",
		},
		{
			name: "invalid - invalid fallback native price (negative)",
			params: types.NewParams(
				"coin",
				"oraclecoin",
				types.DefaultClampFactor,
				types.DefaultFallbackNativePrice.Neg(), // Negative value
				types.DefaultTwapLookbackWindow,
				true,
			),
			errContains: "fallback native price must be greater than 0",
		},
		{
			name: "invalid - invalid fallback native price (zero)",
			params: types.NewParams(
				"coin",
				"oraclecoin",
				types.DefaultClampFactor,
				math.LegacyZeroDec(), // Zero value
				types.DefaultTwapLookbackWindow,
				true,
			),
			errContains: "fallback native price must be greater than 0",
		},
		{
			name: "invalid - twap lookback window zero",
			params: types.NewParams(
				"coin",
				"oraclecoin",
				types.DefaultClampFactor,
				types.DefaultFallbackNativePrice, // Negative value
				0,
				true,
			),
			errContains: "twap lookback window must be greater than 0",
		},
	}

	// Iterate through the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()

			// Check the error
			if tc.errContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			}
		})
	}
}

// TestFeeTokenMetadataValidate tests the Validate method of FeeTokenMetadata
func TestFeeTokenMetadataValidate(t *testing.T) {
	// Prepare test cases
	testCases := []struct {
		name        string
		metadata    types.FeeTokenMetadata
		errContains string
	}{
		{
			name:     "valid - default metadata",
			metadata: types.NewFeeTokenMetadata("coin", "oraclecoin", 6, math.LegacyNewDec(100)),
		},
		{
			name:        "invalid - empty denom",
			metadata:    types.NewFeeTokenMetadata("", "oraclecoin", 6, math.LegacyNewDec(100)),
			errContains: "denom is invalid",
		},
		{
			name:        "invalid - empty oracle denom",
			metadata:    types.NewFeeTokenMetadata("coin", "", 6, math.LegacyNewDec(100)),
			errContains: "oracle denom is invalid",
		},
		{
			name:        "invalid - invalid denom",
			metadata:    types.NewFeeTokenMetadata("123", "oraclecoin", 6, math.LegacyNewDec(100)),
			errContains: "denom is invalid",
		},
		{
			name:        "invalid - invalid oracle denom",
			metadata:    types.NewFeeTokenMetadata("coin", "123", 6, math.LegacyNewDec(100)),
			errContains: "oracle denom is invalid",
		},
		{
			name:        "invalid - decimals zero",
			metadata:    types.NewFeeTokenMetadata("coin", "oraclecoin", 0, math.LegacyNewDec(100)),
			errContains: "decimals must be between 1 and 18",
		},
		{
			name:        "invalid - decimals greater than 18",
			metadata:    types.NewFeeTokenMetadata("coin", "oraclecoin", 19, math.LegacyNewDec(100)),
			errContains: "decimals must be between 1 and 18",
		},
		{
			name:        "invalid - negative price",
			metadata:    types.NewFeeTokenMetadata("coin", "oraclecoin", 6, math.LegacyNewDec(-100)),
			errContains: "price must be greater than 0",
		},
		{
			name:        "invalid - zero price",
			metadata:    types.NewFeeTokenMetadata("coin", "oraclecoin", 6, math.LegacyNewDec(0)),
			errContains: "price must be greater than 0",
		},
	}

	// Iterate through the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.metadata.Validate()

			if tc.errContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			}
		})
	}
}

// TestFeeTokenMetadataCollectionValidate tests the Validate method of FeeTokenMetadataCollection
func TestFeeTokenMetadataCollectionValidate(t *testing.T) {
	// Prepare test cases
	testCases := []struct {
		name        string
		collection  *types.FeeTokenMetadataCollection
		errContains string
	}{
		{
			name:       "valid - empty collection",
			collection: types.NewFeeTokenMetadataCollection(),
		},
		{
			name: "valid - single valid token",
			collection: types.NewFeeTokenMetadataCollection(
				types.NewFeeTokenMetadata("coin", "oraclecoin", 6, math.LegacyNewDec(100)),
			),
		},
		{
			name: "valid - multiple valid tokens",
			collection: types.NewFeeTokenMetadataCollection(
				types.NewFeeTokenMetadata("coin", "oraclecoin", 6, math.LegacyNewDec(100)),
				types.NewFeeTokenMetadata("two", "oracletwo", 18, math.LegacyNewDec(200)),
			),
		},
		{
			name:        "invalid - nil collection",
			collection:  nil,
			errContains: "fee token metadata collection cannot be nil",
		},
		{
			name: "invalid - invalid token in collection",
			collection: types.NewFeeTokenMetadataCollection(
				types.NewFeeTokenMetadata("", "oraclecoin", 6, math.LegacyNewDec(100)),
			),
			errContains: "denom is invalid: invalid fee token metadata",
		},
		{
			name: "invalid - duplicate denoms in collection",
			collection: types.NewFeeTokenMetadataCollection(
				types.NewFeeTokenMetadata("coin", "oraclecoin", 6, math.LegacyNewDec(100)),
				types.NewFeeTokenMetadata("coin", "oraclecoin2", 6, math.LegacyNewDec(100)),
			),
			errContains: "duplicate denom found: coin",
		},
	}
	// Iterate through the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.collection.Validate()

			if tc.errContains == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			}
		})
	}
}
