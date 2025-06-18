package wasmd_test

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cmn "github.com/cosmos/evm/precompiles/common"
	"github.com/cosmos/evm/precompiles/testutil"

	wasmdprecompile "github.com/kiichain/kiichain/v2/precompiles/wasmd"
)

// TestInstantiate is a test for the Instantiate precompile method
func (s *WasmdPrecompileTestSuite) TestInstantiate() {
	// Get the method
	method := s.Precompile.Methods[wasmdprecompile.InstantiateMethod]

	// Get a account from the keyring
	account := s.keyring.GetKey(0)

	// Create the test cases
	tc := []struct {
		name            string
		args            []any
		errContains     string
		expectedResData []byte
	}{
		{
			name: "valid instantiation",
			args: []any{
				account.Addr,
				s.CounterCodeID,
				"Perfectly fine label",
				[]byte(`"zero"`),
				[]cmn.Coin{},
			},
			expectedResData: []byte{},
		},
		{
			name: "invalid args - different than 5",
			args: []any{
				"invalid",
			},
			errContains: "invalid number of arguments; expected 5; got: 1",
		},
		{
			name: "invalid args - invalid type for admin",
			args: []any{
				"invalid",
				"invalid",
				"invalid",
				"invalid",
				"invalid",
			},
			errContains: "invalid admin address",
		},
		{
			name: "invalid args - invalid type for code id",
			args: []any{
				account.Addr,
				"invalid",
				"invalid",
				"invalid",
				"invalid",
			},
			errContains: "invalid code ID",
		},
		{
			name: "invalid args - invalid type for label",
			args: []any{
				account.Addr,
				s.CounterCodeID,
				123,
				"invalid",
				"invalid",
			},
			errContains: "invalid label",
		},
		{
			name: "invalid args - invalid type for init msg",
			args: []any{
				account.Addr,
				s.CounterCodeID,
				"Perfectly fine label",
				"invalid",
				"invalid",
			},
			errContains: "invalid init message",
		},
		{
			name: "invalid args - invalid type for funds",
			args: []any{
				account.Addr,
				s.CounterCodeID,
				"Perfectly fine label",
				[]byte(`"zero"`),
				"invalid",
			},
			errContains: "expected slice, got string",
		},
		{
			name: "invalid cosmwasm msg",
			args: []any{
				account.Addr,
				s.CounterCodeID,
				"Perfectly fine label",
				[]byte(`invalid`),
				[]cmn.Coin{},
			},
			errContains: "payload msg: invalid",
		},
		{
			name: "invalid cosmwasm call",
			args: []any{
				account.Addr,
				s.CounterCodeID,
				"Perfectly fine label",
				[]byte(`{"invalid": "call"}`),
				[]cmn.Coin{},
			},
			errContains: "Error parsing into type counter::msg::CounterInitMsg: unknown variant",
		},
	}

	// Loop and execute the test cases
	for _, tc := range tc {
		s.Run(tc.name, func() {
			// Get the state db
			stateDB := s.GetStateDB()

			// Create the contract from the precompile contract
			contract, ctx := testutil.NewPrecompileContract(s.T(), s.Ctx, account.Addr, s.Precompile, 200000)

			// Execute the contract using the precompile
			res, err := s.Precompile.Instantiate(ctx, account.Addr, contract, stateDB, &method, tc.args)

			// Check if the error contains the expected string
			if tc.errContains != "" {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)

				// Unpack the result
				success, err := s.Precompile.Unpack(wasmdprecompile.InstantiateMethod, res)
				s.Require().NoError(err)

				// Check if the call was a success
				successCall, ok := success[0].(bool)
				s.Require().True(ok)
				s.Require().True(successCall)

				// Check if events were emitted
				log := stateDB.Logs()[0] // Always zero index, since the db is initialized per test
				event := s.Precompile.ABI.Events[wasmdprecompile.EventTypeContractInstantiated]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.Ctx.BlockHeight()))

				// Decode the event data and check
				var instantiateEvent wasmdprecompile.ContractInstantiatedEvent
				err = cmn.UnpackLog(s.Precompile.ABI, &instantiateEvent, wasmdprecompile.EventTypeContractInstantiated, *log)
				s.Require().NoError(err)

				// Check if the data match
				s.Require().Equal(account.Addr, instantiateEvent.Caller)
				s.Require().Equal(tc.args[1], instantiateEvent.CodeID)
				s.Require().Equal(tc.expectedResData, instantiateEvent.Data)

				// Get the contract from the response
				contractAddr := instantiateEvent.ContractAddress

				// Now we can query the keeper and check if the contract was created
				admin, ok := tc.args[0].(common.Address)
				s.Require().True(ok)
				contractInfo := s.App.WasmKeeper.GetContractInfo(s.Ctx, sdk.MustAccAddressFromBech32(contractAddr))
				s.Require().NotNil(contractInfo)
				s.Require().Equal(sdk.AccAddress(admin.Bytes()).String(), contractInfo.Admin)
				s.Require().Equal(s.CounterCodeID, tc.args[1])
				s.Require().Equal(tc.args[2], contractInfo.Label)
			}
		})
	}
}

// TestExecute is a test for the Execute precompile method
func (s *WasmdPrecompileTestSuite) TestExecute() {
	// Instantiate the contract
	contractAddr := s.instantiateContract()

	// Get the method
	method := s.Precompile.Methods[wasmdprecompile.ExecuteMethod]

	// Get a account from the keyring
	account := s.keyring.GetKey(0)

	// Create the test cases
	tc := []struct {
		name            string
		args            []any
		errContains     string
		expectedResData []byte
	}{
		{
			name: "valid execute",
			args: []any{
				contractAddr,
				[]byte(`{"set": 34}`),
				[]cmn.Coin{},
			},
			expectedResData: []byte{},
		},
		{
			name: "invalid args - different than 3",
			args: []any{
				"invalid",
			},
			errContains: "invalid number of arguments; expected 3; got: 1",
		},
		{
			name: "invalid args - invalid type for contract address",
			args: []any{
				123,
				"invalid",
				"invalid",
			},
			errContains: "invalid contract address",
		},
		{
			name: "invalid args - invalid bech32 address",
			args: []any{
				"invalid",
				"invalid",
				"invalid",
			},
			errContains: "invalid contract address: decoding bech32 failed",
		},
		{
			name: "invalid args - invalid execute msg type",
			args: []any{
				contractAddr,
				123,
				"invalid",
			},
			errContains: "invalid execute message",
		},
		{
			name: "invalid args - invalid funds type",
			args: []any{
				contractAddr,
				[]byte(`"invalid"`),
				"invalid",
			},
			errContains: "expected slice, got string",
		},
		{
			name: "invalid args - invalid execute msg",
			args: []any{
				contractAddr,
				[]byte(`{`),
				[]cmn.Coin{},
			},
			errContains: "payload msg: invalid",
		},
		{
			name: "invalid cosmwasm call",
			args: []any{
				contractAddr,
				[]byte(`"invalid"`),
				[]cmn.Coin{},
			},
			errContains: "Error parsing into type counter::msg::CounterExecMsg: unknown variant",
		},
	}

	// Loop and execute the test cases
	for _, tc := range tc {
		s.Run(tc.name, func() {
			// Get the state db
			stateDB := s.GetStateDB()

			// Create the contract from the precompile contract
			contract, ctx := testutil.NewPrecompileContract(s.T(), s.Ctx, account.Addr, s.Precompile, 200000)

			// Execute the contract using the precompile
			res, err := s.Precompile.Execute(ctx, account.Addr, contract, stateDB, &method, tc.args)

			// Check if the error contains the expected string
			if tc.errContains != "" {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)

				// Unpack the result
				success, err := s.Precompile.Unpack(wasmdprecompile.ExecuteMethod, res)
				s.Require().NoError(err)

				// The response data must match the expected data
				successCall, ok := success[0].(bool)
				s.Require().True(ok)
				s.Require().True(successCall)

				// Now we can query the keeper and check if the contract was executed
				contractAccAddress, err := sdk.AccAddressFromBech32(contractAddr)
				s.Require().NoError(err)
				res, err := s.App.WasmKeeper.QuerySmart(s.Ctx, contractAccAddress, []byte(`"value"`))
				s.Require().NoError(err)
				s.Require().Equal([]byte(`{"value":34}`), res)

				// Check if events were emitted
				log := stateDB.Logs()[0] // Always zero index, since the db is initialized per test
				event := s.Precompile.ABI.Events[wasmdprecompile.EventTypeContractExecuted]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(s.Ctx.BlockHeight()))

				// Decode the event data and check
				var executeEvent wasmdprecompile.ContractExecutedEvent
				err = cmn.UnpackLog(s.Precompile.ABI, &executeEvent, wasmdprecompile.EventTypeContractExecuted, *log)
				s.Require().NoError(err)

				// Check if the data match
				s.Require().Equal(crypto.Keccak256Hash([]byte(contractAddr)), executeEvent.ContractAddress)
				s.Require().Equal(account.Addr, executeEvent.Caller)

				// Check the event value
				s.Require().Equal(tc.expectedResData, executeEvent.Data)
			}
		})
	}
}
