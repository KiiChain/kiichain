package oracle_test

import (
	"math/big"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	oracleprecompile "github.com/kiichain/kiichain/v1/precompiles/oracle"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
)

type ExchangeRateResponse struct {
	ExchangeRate        string `json:"exchange_rate"`
	LastUpdate          string `json:"last_update"`
	LastUpdateTimestamp int64  `json:"last_update_timestamp"`
}

type ExchangeRatesResponse struct {
	Denom               string `json:"denom"`
	ExchangeRate        string `json:"exchange_rate"`
	LastUpdate          string `json:"last_update"`
	LastUpdateTimestamp int64  `json:"last_update_timestamp"`
}

type TwapsResponse struct {
	Denom string `json:"denom"`
	Twap  string `json:"twap"`
}

// TestGetExchangeRate tests the GetExchangeRate method of the oracle precompile
func (s *OraclePrecompileTestSuite) TestGetExchangeRate() {
	// Get the method
	method := s.Precompile.Methods[oracleprecompile.GetExchangeRateMethod]

	// Store a exchange rate for testing
	err := s.App.OracleKeeper.ExchangeRate.Set(s.Ctx, "ATOM", types.OracleExchangeRate{
		ExchangeRate:        math.LegacyMustNewDecFromStr("0.5"),
		LastUpdate:          math.NewInt(123),
		LastUpdateTimestamp: 1234,
	})
	s.Require().NoError(err)

	// Create the test cases
	tc := []struct {
		name        string
		args        []any
		errContains string
		expValue    ExchangeRateResponse
	}{
		{
			name: "valid query - get exchange rate",
			args: []any{"ATOM"},
			expValue: ExchangeRateResponse{
				ExchangeRate:        "0.500000000000000000",
				LastUpdate:          "123",
				LastUpdateTimestamp: 1234,
			},
		},
		{
			name:        "invalid currency",
			args:        []any{"INVALID"},
			errContains: "not found",
		},
		{
			name:        "invalid number of arguments",
			args:        []any{},
			errContains: "invalid number of arguments",
		},
	}

	// Loop and execute the test cases
	for _, tc := range tc {
		s.Run(tc.name, func() {
			res, err := s.Precompile.GetExchangeRate(s.Ctx, &method, tc.args)
			if tc.errContains != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)

				// Decode the response
				resUnpacked, err := s.Precompile.Unpack(oracleprecompile.GetExchangeRateMethod, res)
				s.Require().NoError(err)

				// Check the response
				require.Equal(s.T(), 3, len(resUnpacked))
				s.Require().Equal(tc.expValue.ExchangeRate, resUnpacked[0])
				s.Require().Equal(tc.expValue.LastUpdate, resUnpacked[1])
				s.Require().Equal(tc.expValue.LastUpdateTimestamp, resUnpacked[2])
			}
		})
	}
}

// TestGetExchangeRates tests the GetExchangeRates method of the oracle precompile
func (s *OraclePrecompileTestSuite) TestGetExchangeRates() {
	// Get the method
	method := s.Precompile.Methods[oracleprecompile.GetExchangeRatesMethod]

	// Store some exchange rates for testing
	err := s.App.OracleKeeper.ExchangeRate.Set(s.Ctx, "ATOM", types.OracleExchangeRate{
		ExchangeRate:        math.LegacyMustNewDecFromStr("0.5"),
		LastUpdate:          math.NewInt(123),
		LastUpdateTimestamp: 1234,
	})
	s.Require().NoError(err)
	err = s.App.OracleKeeper.ExchangeRate.Set(s.Ctx, "KII", types.OracleExchangeRate{
		ExchangeRate:        math.LegacyMustNewDecFromStr("1.0"),
		LastUpdate:          math.NewInt(456),
		LastUpdateTimestamp: 5678,
	})
	s.Require().NoError(err)

	// Create the test cases
	tc := []struct {
		name        string
		args        []any
		errContains string
		expValue    []ExchangeRatesResponse
	}{
		{
			name: "valid query - get exchange rates",
			args: []any{},
			expValue: []ExchangeRatesResponse{
				{Denom: "ATOM", ExchangeRate: "0.500000000000000000", LastUpdate: "123", LastUpdateTimestamp: 1234},
				{Denom: "KII", ExchangeRate: "1.000000000000000000", LastUpdate: "456", LastUpdateTimestamp: 5678},
			},
		},
		{
			name:        "invalid number of arguments",
			args:        []any{"extra"},
			errContains: "invalid number of arguments",
		},
	}

	for _, tc := range tc {
		s.Run(tc.name, func() {
			res, err := s.Precompile.GetExchangeRates(s.Ctx, &method, tc.args)
			if tc.errContains != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)

				resUnpacked, err := s.Precompile.Unpack(oracleprecompile.GetExchangeRatesMethod, res)
				s.Require().NoError(err)

				for i, exp := range tc.expValue {
					s.Require().Equal(exp.Denom, resUnpacked[0].([]string)[i])
					s.Require().Equal(exp.ExchangeRate, resUnpacked[1].([]string)[i])
					s.Require().Equal(exp.LastUpdate, resUnpacked[2].([]string)[i])
					s.Require().Equal(big.NewInt(exp.LastUpdateTimestamp), resUnpacked[3].([]*big.Int)[i])
				}
			}
		})
	}
}

// TestGetTwaps tests the GetTwaps method of the oracle precompile
func (s *OraclePrecompileTestSuite) TestGetTwaps() {
	// Get the method
	method := s.Precompile.Methods[oracleprecompile.GetTwapsMethod]

	// Register a price snapshot for the twaps query
	err := s.App.OracleKeeper.PriceSnapshot.Set(s.Ctx, 2, types.PriceSnapshot{
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
	require.NoError(s.T(), err)

	// Create the test cases
	tc := []struct {
		name        string
		args        []any
		errContains string
		expValue    []TwapsResponse
	}{
		{
			name: "valid query - get twaps",
			args: []any{2},
			expValue: []TwapsResponse{
				{Denom: "uusdc", Twap: "0.500000000000000000"},
			},
		},
		{
			name:        "invalid query - zero lookback period",
			args:        []any{"extra"},
			errContains: "invalid lookback period",
		},
		{
			name:        "invalid number of arguments",
			args:        []any{},
			errContains: "invalid number of arguments",
		},
	}

	for _, tc := range tc {
		s.Run(tc.name, func() {
			res, err := s.Precompile.GetTwaps(s.Ctx, &method, tc.args)
			if tc.errContains != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)

				resUnpacked, err := s.Precompile.Unpack(oracleprecompile.GetTwapsMethod, res)
				s.Require().NoError(err)

				for i, exp := range tc.expValue {
					s.Require().Equal(exp.Denom, resUnpacked[0].([]string)[i])
					s.Require().Equal(exp.Twap, resUnpacked[1].([]string)[i])
				}
			}
		})
	}
}
