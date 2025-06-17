package oracle

import (
	"encoding/json"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	oraclebindingtypes "github.com/kiichain/kiichain/v1/wasmbinding/oracle/types"
	oraclekeeper "github.com/kiichain/kiichain/v1/x/oracle/keeper"
	oracletypes "github.com/kiichain/kiichain/v1/x/oracle/types"
)

// QueryPlugin is the query plugin object for the oracle queries
type QueryPlugin struct {
	oracleKeeper      oraclekeeper.Keeper
	oracleQueryServer oraclekeeper.QueryServer
}

// NewQueryPlugin returns a new query plugin
func NewQueryPlugin(oracleKeeper oraclekeeper.Keeper) *QueryPlugin {
	// Start the requery server
	oracleQueryServer := oraclekeeper.NewQueryServer(oracleKeeper)

	// Return the query plugin
	return &QueryPlugin{
		oracleKeeper:      oracleKeeper,
		oracleQueryServer: oracleQueryServer,
	}
}

// HandleOracleQuery is a custom querier for the oracle module
func (qp *QueryPlugin) HandleOracleQuery(ctx sdk.Context, oracleQuery oraclebindingtypes.Query) ([]byte, error) {
	// Match the query under the module
	switch {
	// The query is an exchange rate query
	case oracleQuery.ExchangeRate != nil:
		// Apply the request
		exchangeRate, err := qp.HandleExchangeRate(ctx, *oracleQuery.ExchangeRate)
		if err != nil {
			return nil, err
		}

		// Marshal the response
		bz, err := json.Marshal(exchangeRate)
		if err != nil {
			return nil, err
		}
		return bz, nil

	// The query is an exchange rates query
	case oracleQuery.ExchangeRates != nil:
		// Apply the request
		exchangeRates, err := qp.HandleExchangeRates(ctx)
		if err != nil {
			return nil, err
		}

		// Marshal the response
		bz, err := json.Marshal(exchangeRates)
		if err != nil {
			return nil, err
		}
		return bz, nil

	// The query is a twaps query
	case oracleQuery.Twaps != nil:
		twaps, err := qp.HandleTwaps(ctx, *oracleQuery.Twaps)
		if err != nil {
			return nil, err
		}

		bz, err := json.Marshal(twaps)
		if err != nil {
			return nil, err
		}

		return bz, nil

	default:
		return nil, wasmvmtypes.UnsupportedRequest{Kind: "unknown oracle query variant"}
	}
}

// HandleExchangeRate handles the exchange rate query
func (qp *QueryPlugin) HandleExchangeRate(ctx sdk.Context, query oraclebindingtypes.ExchangeRateQuery) (*oracletypes.QueryExchangeRateResponse, error) {
	// Validate the query
	if query.Denom == "" {
		return nil, wasmvmtypes.InvalidRequest{Err: "empty denom"}
	}

	// Get the exchange rate from the keeper
	exchangeRate, err := qp.oracleQueryServer.ExchangeRate(
		ctx,
		&oracletypes.QueryExchangeRateRequest{
			Denom: query.Denom,
		},
	)
	if err != nil {
		return nil, err
	}

	// Return the response
	return exchangeRate, nil
}

// HandleExchangeRates handles the exchange rates query
func (qp *QueryPlugin) HandleExchangeRates(ctx sdk.Context) (*oracletypes.QueryExchangeRatesResponse, error) {
	// Get the exchange rates from the keeper
	exchangeRates, err := qp.oracleQueryServer.ExchangeRates(
		ctx,
		&oracletypes.QueryExchangeRatesRequest{},
	)
	if err != nil {
		return nil, err
	}

	// Return the response
	return exchangeRates, nil
}

// HandleTwaps handles the twaps query
func (qp *QueryPlugin) HandleTwaps(ctx sdk.Context, query oraclebindingtypes.TwapsQuery) (*oracletypes.QueryTwapsResponse, error) {
	// Get the twaps from the keeper
	twaps, err := qp.oracleQueryServer.Twaps(
		ctx,
		&oracletypes.QueryTwapsRequest{
			LookbackSeconds: query.LookbackSeconds,
		},
	)
	if err != nil {
		return nil, err
	}

	// Return the response
	return twaps, nil
}
