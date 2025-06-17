package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	oraclekeeper "github.com/kiichain/kiichain/v1/x/oracle/keeper"
)

const (
	// QueryExchangeRateMethod is the method name for exchange rate query
	QueryExchangeRateMethod = "queryExchangeRate"
	// QueryExchangeRatesMethod is the method name for exchange rates query
	QueryExchangeRatesMethod = "queryExchangeRates"
	// QueryTwaps Method is the method name for twaps query
	QueryTwapsMethod = "queryTwaps"
)

// GetExchangeRate queries the exchange rate though the oracle IOracle precompile
func (p Precompile) GetExchangeRate(ctx sdk.Context, method *abi.Method, args []any) ([]byte, error) {
	// Build the request from the arguments
	req, err := ParseGetExchangeRateArgs(args)
	if err != nil {
		return nil, err
	}

	// Start a new query service
	queryService := oraclekeeper.NewQueryServer(p.oracleKeeper)

	// Make the request
	res, err := queryService.ExchangeRate(ctx, req)
	if err != nil {
		return nil, err
	}

	// Pack the response into bytes
	return method.Outputs.Pack(
		res.OracleExchangeRate.ExchangeRate.String(),
		res.OracleExchangeRate.LastUpdate.String(),
		res.OracleExchangeRate.LastUpdateTimestamp,
	)
}

// GetExchangeRates queries the exchange rates through the oracle IOracle precompile
func (p Precompile) GetExchangeRates(ctx sdk.Context, method *abi.Method, args []any) ([]byte, error) {
	// Build the request from the arguments
	req, err := ParseGetExchangeRatesArgs(args)
	if err != nil {
		return nil, err
	}

	// Start a new query service
	queryService := oraclekeeper.NewQueryServer(p.oracleKeeper)

	// Make the request
	res, err := queryService.ExchangeRates(ctx, req)
	if err != nil {
		return nil, err
	}

	// Pack the response into bytes
	denoms := make([]string, len(res.DenomOracleExchangeRate))
	rates := make([]string, len(res.DenomOracleExchangeRate))
	lastUpdate := make([]string, len(res.DenomOracleExchangeRate))
	lastUpdateTimestamps := make([]int64, len(res.DenomOracleExchangeRate))

	// Iterate over the exchange rates and fill the slices
	for i, exchangeRate := range res.DenomOracleExchangeRate {
		denoms[i] = exchangeRate.Denom
		rates[i] = exchangeRate.OracleExchangeRate.ExchangeRate.String()
		lastUpdate[i] = exchangeRate.OracleExchangeRate.LastUpdate.String()
		lastUpdateTimestamps[i] = exchangeRate.OracleExchangeRate.LastUpdateTimestamp
	}

	// Return the packed response
	return method.Outputs.Pack(
		denoms,
		rates,
		lastUpdate,
		lastUpdateTimestamps,
	)
}

// GetTwaps queries the twaps through the oracle IOracle precompile
func (p Precompile) GetTwaps(ctx sdk.Context, method *abi.Method, args []any) ([]byte, error) {
	// Build the request from the arguments
	req, err := ParseGetTwapsArgs(args)
	if err != nil {
		return nil, err
	}

	// Start a new query service
	queryService := oraclekeeper.NewQueryServer(p.oracleKeeper)

	// Make the request
	res, err := queryService.Twaps(ctx, req)
	if err != nil {
		return nil, err
	}

	// Pack the response into bytes
	denoms := make([]string, len(res.OracleTwap))
	twaps := make([]string, len(res.OracleTwap))

	// Iterate over the twaps and fill the slices
	for i, twap := range res.OracleTwap {
		denoms[i] = twap.Denom
		twaps[i] = twap.Twap.String()
	}

	// Return the packed response
	return method.Outputs.Pack(
		denoms,
		twaps,
	)
}
