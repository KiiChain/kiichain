package ibc_test

import (
	cmn "github.com/cosmos/evm/precompiles/common"
	"github.com/cosmos/evm/precompiles/testutil"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ibcprecompile "github.com/kiichain/kiichain/v1/precompiles/ibc"
)

// TestPrecompileTransferWithDefaultTimeout calls IBC precompile transfer with default timeout
func (s *IBCPrecompileTestSuite) TestPrecompileTransferWithDefaultTimeout() {
	// Get and setup path
	path := s.path
	s.coordinator.Setup(path)

	// Get up test coin
	coin := ibctesting.TestCoin

	// Get the method
	method := s.Precompile.Methods[ibcprecompile.TransferWithDefaultTimeoutMethod]

	// Get an account from the keyring
	sender := s.keyring.GetKey(0)

	// Prepare valid info
	receiver := s.keyring.GetKey(1).Addr.String()
	port := path.EndpointA.ChannelConfig.PortID
	channel := path.EndpointA.ChannelID
	denom := coin.Denom
	amount := coin.Amount
	memo := "test"

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
				receiver,
				port,
				channel,
				denom,
				amount,
				memo,
			},
			expectedResData: []byte{},
		},
		{
			name: "invalid args - different than 6",
			args: []any{
				"invalid",
			},
			errContains: "invalid number of arguments; expected 6; got: 1",
		},
	}

	// Loop and execute the test cases
	for _, tc := range tc {
		s.Run(tc.name, func() {
			// Get the state db
			chainAstateDB := s.GetStateDB(s.chainA)

			// Create the contract from the precompile contract
			_, ctx := testutil.NewPrecompileContract(s.T(), s.chainA.GetContext(), sender.Addr, s.Precompile, 200000)

			// Execute the contract using the precompile
			res, err := s.Precompile.TransferWithDefaultTimeout(ctx, &method, chainAstateDB, tc.args, sender.Addr)

			// Check if the error contains the expected string
			if tc.errContains != "" {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)

				// Unpack the result
				success, err := s.Precompile.Unpack(ibcprecompile.TransferWithDefaultTimeoutMethod, res)
				s.Require().NoError(err)

				// The response data must match the expected data
				successCall, ok := success[0].(bool)
				s.Require().True(ok)
				s.Require().True(successCall)

				// Check if events were emitted
				log := chainAstateDB.Logs()[0] // Always zero index, since the db is initialized per test
				event := s.Precompile.ABI.Events[ibcprecompile.TransferWithDefaultTimeoutMethod]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight()))

				// Decode the event data and check
				var executeEvent ibcprecompile.TransferEvent
				err = cmn.UnpackLog(s.Precompile.ABI, &executeEvent, ibcprecompile.EventTypeTransfer, *log)
				s.Require().NoError(err)

				// Check if the data match

				// Check the event value
				s.Require().Equal(tc.expectedResData, executeEvent.Data)
			}
		})
	}
}
