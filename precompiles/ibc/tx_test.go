package ibc_test

import (
	"math/big"

	cmn "github.com/cosmos/evm/precompiles/common"
	"github.com/cosmos/evm/precompiles/testutil"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ibcprecompile "github.com/kiichain/kiichain/v1/precompiles/ibc"
)

// TestPrecompileTransferWithDefaultTimeout calls IBC precompile transfer with default timeout
func (s *IBCPrecompileTestSuite) TestPrecompileTransferWithDefaultTimeout() {
	// Get path and testcoin
	path := s.path
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
	amount := coin.Amount.BigInt()
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

		// Argument length validation
		{
			name: "invalid args - different than 6",
			args: []any{
				"invalid",
			},
			errContains: "expected 6 arguments but got 1",
		},

		// Receiver validation
		{
			name: "invalid receiver - empty string",
			args: []any{
				"",
				port,
				channel,
				denom,
				amount,
				memo,
			},
			errContains: "receiverAddress is not a string or empty",
		},
		{
			name: "invalid receiver - wrong type",
			args: []any{
				12345, // not a string
				port,
				channel,
				denom,
				amount,
				memo,
			},
			errContains: "receiverAddress is not a string or empty",
		},

		// Port validation
		{
			name: "invalid port - empty string",
			args: []any{
				receiver,
				"",
				channel,
				denom,
				amount,
				memo,
			},
			errContains: "port cannot be empty",
		},
		{
			name: "invalid port - wrong type",
			args: []any{
				receiver,
				12345, // not a string
				channel,
				denom,
				amount,
				memo,
			},
			errContains: "port is not a string",
		},

		// Channel validation
		{
			name: "invalid channel - empty string",
			args: []any{
				receiver,
				port,
				"",
				denom,
				amount,
				memo,
			},
			errContains: "channelID cannot be empty",
		},
		{
			name: "invalid channel - wrong type",
			args: []any{
				receiver,
				port,
				12345, // not a string
				denom,
				amount,
				memo,
			},
			errContains: "channelID is not a string",
		},

		// Denom validation
		{
			name: "invalid denom - empty string",
			args: []any{
				receiver,
				port,
				channel,
				"",
				amount,
				memo,
			},
			errContains: "invalid denom",
		},

		// Zero amount
		{
			name: "Invalid amount - zero value",
			args: []any{
				receiver,
				port,
				channel,
				denom,
				big.NewInt(0),
				memo,
			},
			errContains: "Amount is zero",
		},
		{
			name: "invalid amount - wrong type",
			args: []any{
				receiver,
				port,
				channel,
				denom,
				"not-a-bigint", // not a *big.Int
				memo,
			},
			errContains: "amount is not a big.Int",
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
				event := s.Precompile.ABI.Events[ibcprecompile.EventTypeTransfer]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight()))

				// Decode the event data and check
				var transferEvent ibcprecompile.TransferEvent
				err = cmn.UnpackLog(s.Precompile.ABI, &transferEvent, ibcprecompile.EventTypeTransfer, *log)
				s.Require().NoError(err)

				// Check if the data match
				s.Require().Equal(transferEvent.Amount, amount)
				s.Require().Equal(transferEvent.Port, port)

				// Check package commitment
				// Get the next sequence to find our packet
				seq, found := s.chainA.App.GetIBCKeeper().ChannelKeeper.GetNextSequenceSend(
					s.chainA.GetContext(),
					ibctesting.TransferPort,
					s.path.EndpointA.ChannelID,
				)
				s.Require().True(found)
				s.Require().Greater(seq, uint64(0), "sequence should increment")

				// Check packet commitment exists
				commitment := s.chainA.App.GetIBCKeeper().ChannelKeeper.GetPacketCommitment(
					s.chainA.GetContext(),
					ibctesting.TransferPort,
					s.path.EndpointA.ChannelID,
					seq-1, // The packet we just sent
				)
				s.Require().NotEmpty(commitment, "packet commitment should exist")
			}
		})
	}
}
