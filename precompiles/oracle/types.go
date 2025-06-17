package oracle

import (
	"fmt"

	cmn "github.com/cosmos/evm/precompiles/common"
	oracletypes "github.com/kiichain/kiichain/v1/x/oracle/types"
)

// ParseGetExchangeRateArgs parses the arguments for the GetExchangeRate method
func ParseGetExchangeRateArgs(args []interface{}) (*oracletypes.QueryExchangeRateRequest, error) {
	// Check the number of arguments, should be 1
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	// Parse the first arg, the denom
	denom, ok := args[0].(string)
	if !ok || denom == "" {
		return nil, fmt.Errorf("invalid denom")
	}

	// Create the QueryExchangeRateRequest and return
	return &oracletypes.QueryExchangeRateRequest{
		Denom: denom,
	}, nil
}

// ParseGetExchangeRatesArgs parses the arguments for the GetExchangeRates method
func ParseGetExchangeRatesArgs(args []interface{}) (*oracletypes.QueryExchangeRatesRequest, error) {
	// Check the number of arguments, should be 0
	if len(args) != 0 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 0, len(args))
	}

	// Create the QueryExchangeRatesRequest and return
	return &oracletypes.QueryExchangeRatesRequest{}, nil
}

// ParseGetTwapsArgs parses the arguments for the GetTwaps method
func ParseGetTwapsArgs(args []interface{}) (*oracletypes.QueryTwapsRequest, error) {
	// Check the number of arguments, should be 1
	if len(args) != 1 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 1, len(args))
	}

	// Parse the second arg, the lookback period
	lookbackPeriod, ok := args[0].(int)
	if !ok || lookbackPeriod == 0 {
		return nil, fmt.Errorf("invalid lookback period")
	}

	// Create the QueryTwapsRequest and return
	return &oracletypes.QueryTwapsRequest{
		LookbackSeconds: uint64(lookbackPeriod),
	}, nil
}
