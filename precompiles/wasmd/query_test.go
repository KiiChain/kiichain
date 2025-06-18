package wasmd_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	wasmdprecompile "github.com/kiichain/kiichain/v2/precompiles/wasmd"
)

// TestQueryRaw is a test for the QueryRaw precompile method
func (s *WasmdPrecompileTestSuite) TestQueryRaw() {
	// Instantiate the contract
	contract := s.instantiateContract()

	// Get the method
	method := s.Precompile.Methods[wasmdprecompile.QueryRawMethod]

	// Create the test cases
	tc := []struct {
		name        string
		args        []any
		errContains string
		expValue    []byte
	}{
		{
			name:     "valid query - get the contract version",
			args:     []any{contract, []byte(`value`)},
			expValue: []byte("0"),
		},
		{
			name:     "valid query - no response (invalid address)",
			args:     []any{contract, []byte(`{"valid": "query"}`)},
			expValue: []byte{},
		},
		{
			name:        "empty args",
			args:        []any{},
			errContains: "invalid number of arguments",
		},
		{
			name:        "invalid contract address type",
			args:        []any{123, 123},
			errContains: "invalid contract address",
		},
		{
			name:        "invalid contract address bech32",
			args:        []any{"asd", 123},
			errContains: "decoding bech32 failed",
		},
		{
			name:        "invalid query msg type",
			args:        []any{contract, 123},
			errContains: "invalid query data",
		},
	}

	// Loop and execute the test cases
	for _, tc := range tc {
		s.Run(tc.name, func() {
			// Query the contract using the precompile
			res, err := s.Precompile.QueryRaw(s.Ctx, &method, tc.args)

			// Check if the result contains an error
			if tc.errContains != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)

				// Decode the response
				resUnpacked, err := s.Precompile.Unpack(wasmdprecompile.QueryRawMethod, res)
				s.Require().NoError(err)

				// Get the response bytes
				resBytes, ok := resUnpacked[0].([]byte)
				s.Require().True(ok)

				s.Require().Equal(resBytes, tc.expValue)
			}
		})
	}
}

// TestQuerySmart is a test for the QuerySmart precompile method
func (s *WasmdPrecompileTestSuite) TestQuerySmart() {
	// Instantiate the contract
	contract := s.instantiateContract()

	// Get the method
	method := s.Precompile.Methods[wasmdprecompile.QuerySmartMethod]

	// Create the test cases
	tc := []struct {
		name        string
		args        []any
		errContains string
		expValue    []byte
	}{
		{
			name:     "valid query",
			args:     []any{contract, []byte(`"value"`)},
			expValue: []byte("{\"value\":0}"),
		},
		{
			name:        "empty args",
			args:        []any{},
			errContains: "invalid number of arguments",
		},
		{
			name:        "invalid contract address type",
			args:        []any{123, 123},
			errContains: "invalid contract address",
		},
		{
			name:        "invalid contract address bech32",
			args:        []any{"asd", 123},
			errContains: "decoding bech32 failed",
		},
		{
			name:        "invalid query msg type",
			args:        []any{contract, 123},
			errContains: "invalid query data",
		},
		{
			name:        "invalid query",
			args:        []any{contract, []byte(`{"invalid": "query"}`)},
			errContains: "unknown variant `invalid`, expected `value`",
		},
	}

	// Loop and execute the test cases
	for _, tc := range tc {
		s.Run(tc.name, func() {
			// Query the contract using the precompile
			res, err := s.Precompile.QuerySmart(s.Ctx, &method, tc.args)

			// Check if the result contains an error
			if tc.errContains != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errContains)
			} else {
				s.Require().NoError(err)

				// Decode the response
				resUnpacked, err := s.Precompile.Unpack(wasmdprecompile.QuerySmartMethod, res)
				s.Require().NoError(err)

				// Get the response bytes
				resBytes, ok := resUnpacked[0].([]byte)
				s.Require().True(ok)

				s.Require().Equal(resBytes, tc.expValue)
			}
		})
	}
}

// instantiateContract is a helper function to instantiate the contract
func (s *WasmdPrecompileTestSuite) instantiateContract() string {
	s.T().Helper()

	// Instantiate the contract
	res, err := s.WasmdMsgServer.InstantiateContract(s.Ctx, &wasmtypes.MsgInstantiateContract{
		Sender: sdk.AccAddress([]byte("wasm")).String(),
		CodeID: s.CounterCodeID,
		Label:  "counter",
		Msg:    []byte(`"zero"`),
	})
	s.Require().NoError(err)

	return res.Address
}
