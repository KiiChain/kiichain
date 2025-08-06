package types

import (
	"cosmossdk.io/errors"
)

// Oracle Errors
var (
	ErrInvalidExchangeRate      = errors.Register(ModuleName, 2, "invalid exchange rate")
	ErrNoVote                   = errors.Register(ModuleName, 4, "no vote")
	ErrNoVotingPermission       = errors.Register(ModuleName, 5, "unauthorized voter")
	ErrInvalidHash              = errors.Register(ModuleName, 6, "invalid hash")
	ErrInvalidHashLength        = errors.Register(ModuleName, 7, "invalid hash length")
	ErrVerificationFailed       = errors.Register(ModuleName, 8, "hash verification failed")
	ErrNoAggregateVote          = errors.Register(ModuleName, 12, "no aggregate vote")
	ErrNoVoteTarget             = errors.Register(ModuleName, 13, "no vote target")
	ErrUnknownDenom             = errors.Register(ModuleName, 14, "unknown denom")
	ErrNoLatestPriceSnapshot    = errors.Register(ModuleName, 15, "no latest snapshot")
	ErrInvalidTwapLookback      = errors.Register(ModuleName, 16, "Twap lookback seconds is greater than max lookback duration or less than or equal to 0")
	ErrNoTwapData               = errors.Register(ModuleName, 17, "No data for the twap calculation")
	ErrParsingOracleQuery       = errors.Register(ModuleName, 18, "Error parsing KiiOracleQuery")
	ErrGettingExchangeRates     = errors.Register(ModuleName, 19, "Error while getting Exchange Rates")
	ErrEncodingExchangeRates    = errors.Register(ModuleName, 20, "Error encoding exchange rates as JSON")
	ErrGettingOracleTwaps       = errors.Register(ModuleName, 21, "Error while getting Oracle Twaps in wasmd")
	ErrEncodingOracleTwaps      = errors.Register(ModuleName, 22, "Error encoding oracle twaps as JSON")
	ErrUnknownKiiOracleQuery    = errors.Register(ModuleName, 23, "Error unknown kii oracle query")
	ErrAggregateVoteExist       = errors.Register(ModuleName, 24, "aggregate vote still present in current voting window")
	ErrAggregateVoteInvalidRate = errors.Register(ModuleName, 25, "aggregate vote has invalid exchange rate")
)
