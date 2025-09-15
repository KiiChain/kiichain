package oracle

import (
	"fmt"
	"math/big"

	cmn "github.com/cosmos/evm/precompiles/common"

	oracletypes "github.com/kiichain/kiichain/v4/x/oracle/types"
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
		return nil, fmt.Errorf("invalid denom: empty or not a string")
	}

	// Regex validation: [a-z][a-z0-9/]{2,64}
	// Only allow denom starting with a-z, followed by 2-64 chars a-z, 0-9, or /
	import "regexp"
	var denomRegex = regexp.MustCompile(`^[a-z][a-z0-9/]{2,64}$`)
	if !denomRegex.MatchString(denom) {
		return nil, fmt.Errorf("invalid denom format: must match [a-z][a-z0-9/]{2,64}")
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
	lookbackPeriod, ok := args[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid lookback period")
	}

	// Create the QueryTwapsRequest and return
	return &oracletypes.QueryTwapsRequest{
		LookbackSeconds: lookbackPeriod.Uint64(),
	}, nil
}
