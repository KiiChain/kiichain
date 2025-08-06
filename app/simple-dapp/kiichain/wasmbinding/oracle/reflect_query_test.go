package oracle_test

import (
	"encoding/json"
	"testing"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	app "github.com/kiichain/kiichain/v3/app"
	"github.com/kiichain/kiichain/v3/app/apptesting"
	"github.com/kiichain/kiichain/v3/wasmbinding"
	"github.com/kiichain/kiichain/v3/wasmbinding/helpers"
	oraclebindingtypes "github.com/kiichain/kiichain/v3/wasmbinding/oracle/types"
	oracletypes "github.com/kiichain/kiichain/v3/x/oracle/types"
)

// TestOracleQueries test the Oracle query
func TestOracleQueries(t *testing.T) {
	actor := apptesting.RandomAccountAddress()
	app, ctx := helpers.SetupCustomApp(t, actor)

	reflect := helpers.InstantiateReflectContract(t, ctx, app, actor)
	require.NotEmpty(t, reflect)

	// Create a rate
	err := app.OracleKeeper.ExchangeRate.Set(ctx, "uusdc", oracletypes.OracleExchangeRate{
		ExchangeRate:        math.LegacyMustNewDecFromStr("0.5"),
		LastUpdate:          math.NewIntFromUint64(1000000),
		LastUpdateTimestamp: 1000000,
	})
	require.NoError(t, err)
	// Register a price snapshot for the twaps query
	err = app.OracleKeeper.PriceSnapshot.Set(ctx, 2, oracletypes.PriceSnapshot{
		SnapshotTimestamp: 2,
		PriceSnapshotItems: []oracletypes.PriceSnapshotItem{
			{
				Denom: "uusdc",
				OracleExchangeRate: oracletypes.OracleExchangeRate{
					ExchangeRate:        math.LegacyMustNewDecFromStr("0.5"),
					LastUpdate:          math.NewIntFromUint64(1000000),
					LastUpdateTimestamp: 1000000,
				},
			},
		},
	})
	require.NoError(t, err)

	// query exchange rate
	query := oraclebindingtypes.Query{
		ExchangeRate: &oraclebindingtypes.ExchangeRateQuery{
			Denom: "uusdc",
		},
	}
	resp := oracletypes.QueryExchangeRateResponse{}
	err = queryCustom(t, ctx, app, reflect, query, &resp)
	require.NoError(t, err)

	require.EqualValues(t, resp.OracleExchangeRate.ExchangeRate, math.LegacyMustNewDecFromStr("0.5"))

	// Query all the rates
	query = oraclebindingtypes.Query{
		ExchangeRates: &oraclebindingtypes.ExchangeRatesQuery{},
	}

	respAll := oracletypes.QueryExchangeRatesResponse{}
	err = queryCustom(t, ctx, app, reflect, query, &respAll)
	require.NoError(t, err)
	require.Len(t, respAll.DenomOracleExchangeRate, 1)
	require.EqualValues(t, respAll.DenomOracleExchangeRate[0].Denom, "uusdc")

	// Query the twaps
	query = oraclebindingtypes.Query{
		Twaps: &oraclebindingtypes.TwapsQuery{
			LookbackSeconds: 10,
		},
	}

	respTwaps := oracletypes.QueryTwapsResponse{}
	err = queryCustom(t, ctx, app, reflect, query, &respTwaps)
	require.NoError(t, err)
	require.Len(t, respTwaps.OracleTwap, 1)
	require.EqualValues(t, respTwaps.OracleTwap[0].Denom, "uusdc")
}

// TestQueryDenomAdmin tests the GetDenomAdmin query
type ReflectQuery struct {
	Chain *ChainRequest `json:"chain,omitempty"`
}

// ChainRequest is the request to the chain
type ChainRequest struct {
	Request wasmvmtypes.QueryRequest `json:"request"`
}

// ChainResponse is the response from the chain
type ChainResponse struct {
	Data []byte `json:"data"`
}

// queryCustom is a helper function to query the custom contract
func queryCustom(t *testing.T, ctx sdk.Context, app *app.KiichainApp, contract sdk.AccAddress, request oraclebindingtypes.Query, response interface{}) error {
	t.Helper()

	// Make the request a kiichain query
	kiichainQuery := wasmbinding.KiichainQuery{
		Oracle: &request,
	}

	// Marshal the request to JSON
	msgBz, err := json.Marshal(kiichainQuery)
	if err != nil {
		return err
	}
	t.Log("queryCustom1", string(msgBz))

	query := ReflectQuery{
		Chain: &ChainRequest{
			Request: wasmvmtypes.QueryRequest{Custom: msgBz},
		},
	}
	queryBz, err := json.Marshal(query)
	if err != nil {
		return err
	}
	t.Log("queryCustom3", string(queryBz))

	resBz, err := app.WasmKeeper.QuerySmart(ctx, contract, queryBz)
	if err != nil {
		return err
	}
	var resp ChainResponse
	err = json.Unmarshal(resBz, &resp)
	if err != nil {
		return err
	}
	err = json.Unmarshal(resp.Data, response)
	if err != nil {
		return err
	}

	return nil
}
