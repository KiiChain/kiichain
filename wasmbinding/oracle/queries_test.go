package oracle_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/kiichain/kiichain/v1/app/apptesting"
	"github.com/kiichain/kiichain/v1/wasmbinding/helpers"
	"github.com/kiichain/kiichain/v1/wasmbinding/oracle"
	oraclebindingtypes "github.com/kiichain/kiichain/v1/wasmbinding/oracle/types"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
)

// TestHandleOracleQuery tests the HandleOracleQuery function of the oracle module
func TestHandleOracleQuery(t *testing.T) {
	// Setup the app
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	// Create two rates
	err := app.OracleKeeper.ExchangeRate.Set(ctx, "uusdc", types.OracleExchangeRate{
		ExchangeRate:        math.LegacyMustNewDecFromStr("0.5"),
		LastUpdate:          math.NewIntFromUint64(1000000),
		LastUpdateTimestamp: 1000000,
	})
	require.NoError(t, err)
	err = app.OracleKeeper.ExchangeRate.Set(ctx, "akii", types.OracleExchangeRate{
		ExchangeRate:        math.LegacyMustNewDecFromStr("125.2"),
		LastUpdate:          math.NewIntFromUint64(2000000),
		LastUpdateTimestamp: 2000000,
	})
	require.NoError(t, err)

	// Register a price snapshot for the twaps query
	err = app.OracleKeeper.PriceSnapshot.Set(ctx, 2, types.PriceSnapshot{
		SnapshotTimestamp: 2,
		PriceSnapshotItems: []types.PriceSnapshotItem{
			{
				Denom: "uusdc",
				OracleExchangeRate: types.OracleExchangeRate{
					ExchangeRate:        math.LegacyMustNewDecFromStr("0.5"),
					LastUpdate:          math.NewIntFromUint64(1000000),
					LastUpdateTimestamp: 1000000,
				},
			},
		},
	})
	require.NoError(t, err)

	// Set all the test cases
	testCases := []struct {
		name        string
		query       oraclebindingtypes.Query
		expected    []byte
		errContains string
	}{
		{
			name: "Valid - exchange rate",
			query: oraclebindingtypes.Query{
				ExchangeRate: &oraclebindingtypes.ExchangeRateQuery{
					Denom: "uusdc",
				},
			},
			expected: []byte(`{"oracle_exchange_rate":{"exchange_rate":"0.500000000000000000","last_update":"1000000","last_update_timestamp":1000000}}`),
		},
		{
			name: "Invalid - exchange rate empty denom",
			query: oraclebindingtypes.Query{
				ExchangeRate: &oraclebindingtypes.ExchangeRateQuery{
					Denom: "",
				},
			},
			errContains: "invalid request: empty denom",
		},
		{
			name: "Invalid - exchange rate bad request",
			query: oraclebindingtypes.Query{
				ExchangeRate: &oraclebindingtypes.ExchangeRateQuery{
					Denom: "unknown",
				},
			},
			errContains: "not found",
		},
		{
			name: "valid - exchange rates",
			query: oraclebindingtypes.Query{
				ExchangeRates: &oraclebindingtypes.ExchangeRatesQuery{},
			},
			expected: []byte(`{"denom_oracle_exchange_rate":[{"denom":"akii","oracle_exchange_rate":{"exchange_rate":"125.200000000000000000","last_update":"2000000","last_update_timestamp":2000000}},{"denom":"uusdc","oracle_exchange_rate":{"exchange_rate":"0.500000000000000000","last_update":"1000000","last_update_timestamp":1000000}}]}`),
		},
		{
			name: "valid - twaps",
			query: oraclebindingtypes.Query{
				Twaps: &oraclebindingtypes.TwapsQuery{
					LookbackSeconds: 1000,
				},
			},
			expected: []byte(`{"oracle_twap":[{"denom":"uusdc","twap":"0.500000000000000000","lookback_seconds":1000}]}`),
		},
		{
			name: "valid - twaps",
			query: oraclebindingtypes.Query{
				Twaps: &oraclebindingtypes.TwapsQuery{
					LookbackSeconds: 0,
				},
			},
			errContains: "Twap lookback seconds is greater than max lookback",
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start the query plugin
			queryPlugin := oracle.NewQueryPlugin(app.OracleKeeper)

			// Handle the query
			bz, err := queryPlugin.HandleOracleQuery(ctx, tc.query)

			// Check for errors
			if tc.errContains != "" {
				require.ErrorContains(t, err, tc.errContains)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, bz)
			}
		})
	}
}
