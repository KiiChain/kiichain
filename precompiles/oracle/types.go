package oracle

import (
	"fmt"
	"math/big"

	cmn "github.com/cosmos/evm/precompiles/common"

	oracletypes "github.com/kiichain/kiichain/v7/x/oracle/types"
)

// MaxDenomLength is the maximum allowed length for denom strings
// This prevents DoS attacks via extremely long strings
const MaxDenomLength = 256

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

	// Validate denom length to prevent DoS via oversized strings
	if len(denom) > MaxDenomLength {
		return nil, fmt.Errorf("denom too long: max %d bytes, got %d", MaxDenomLength, len(denom))
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

	// Parse the first arg, the lookback period
	lookbackPeriod, ok := args[0].(*big.Int)
	if !ok || lookbackPeriod == nil {
		return nil, fmt.Errorf("invalid lookback period")
	}

	// Validate lookback period is positive (reject zero and negative)
	if lookbackPeriod.Sign() <= 0 {
		return nil, fmt.Errorf("lookback period must be positive")
	}

	// Validate lookback period fits in uint64 (prevent overflow)
	if !lookbackPeriod.IsUint64() {
		return nil, fmt.Errorf("lookback period must fit in uint64")
	}

	// Create the QueryTwapsRequest and return
	return &oracletypes.QueryTwapsRequest{
		LookbackSeconds: lookbackPeriod.Uint64(),
	}, nil
}
