package ibc_test

import (
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	cmn "github.com/cosmos/evm/precompiles/common"
	"github.com/cosmos/evm/precompiles/testutil"

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

	// Base valid args
	validArgs := []any{
		s.keyring.GetKey(1).Addr.String(),   // receiver
		path.EndpointA.ChannelConfig.PortID, // port
		path.EndpointA.ChannelID,            // channel
		coin.Denom,                          // denom
		coin.Amount.BigInt(),                // amount
		"test memo",                         // memo
	}

	tc := []struct {
		name        string
		modifyArgs  func([]any) []any
		errContains string
	}{
		// Original test cases
		{
			name:       "valid execute",
			modifyArgs: func(args []any) []any { return args },
		},
		// Invalid number of args
		{
			name: "invalid args length",
			modifyArgs: func(args []any) []any {
				return args[:1]
			},
			errContains: "expected 6 arguments but got 1",
		},
		// Receiver validation
		{
			name: "empty receiver",
			modifyArgs: func(args []any) []any {
				args[0] = ""
				return args
			},
			errContains: "receiverAddress is not a string or empty",
		},
		{
			name: "invalid receiver - wrong type",
			modifyArgs: func(args []any) []any {
				args[0] = 12345
				return args
			},
			errContains: "receiverAddress is not a string or empty",
		},
		{
			name: "receiver too long",
			modifyArgs: func(args []any) []any {
				args[0] = strings.Repeat("a", transfertypes.MaximumReceiverLength+1)
				return args
			},
			errContains: "recipient address must not exceed",
		},
		// Port validation
		{
			name: "invalid port - empty string",
			modifyArgs: func(args []any) []any {
				args[1] = ""
				return args
			},
			errContains: "port cannot be empty",
		},
		{
			name: "invalid port - wrong type",
			modifyArgs: func(args []any) []any {
				args[1] = 1234
				return args
			},
			errContains: "port is not a string",
		},
		{
			name: "invalid port - malformed",
			modifyArgs: func(args []any) []any {
				args[1] = "invalid*port"
				return args
			},
			errContains: "channel not found",
		},
		// Channel validation
		{
			name: "invalid channel - empty string",
			modifyArgs: func(args []any) []any {
				args[2] = ""
				return args
			},
			errContains: "channelID cannot be empty",
		},
		{
			name: "invalid channel - wrong type",
			modifyArgs: func(args []any) []any {
				args[2] = 1234
				return args
			},
			errContains: "channelID is not a string",
		},
		{
			name: "invalid channel - malformed",
			modifyArgs: func(args []any) []any {
				args[2] = "invalid*channel"
				return args
			},
			errContains: "channel not found",
		},
		// Denom validation
		{
			name: "invalid denom - empty string",
			modifyArgs: func(args []any) []any {
				args[3] = ""
				return args
			},
			errContains: "invalid denom",
		},
		{
			name: "invalid denom - empty",
			modifyArgs: func(args []any) []any {
				args[3] = ""
				return args
			},
			errContains: "invalid denom",
		},
		{
			name: "invalid denom - malformed",
			modifyArgs: func(args []any) []any {
				args[3] = "invalid*denom"
				return args
			},
			errContains: "invalid coins",
		},
		{
			name: "zero amount",
			modifyArgs: func(args []any) []any {
				args[4] = big.NewInt(0)
				return args
			},
			errContains: "amount is zero",
		},
		{
			name: "negative amount",
			modifyArgs: func(args []any) []any {
				args[4] = big.NewInt(-100)
				return args
			},
			errContains: "invalid coins",
		},
		// memo validation
		{
			name: "memo too long",
			modifyArgs: func(args []any) []any {
				args[5] = strings.Repeat("a", transfertypes.MaximumMemoLength+1)
				return args
			},
			errContains: "memo must not exceed",
		},
	}

	// Loop and execute the test cases
	for _, tc := range tc {
		s.Run(tc.name, func() {
			// Apply args modification
			args := tc.modifyArgs(append([]any(nil), validArgs...))

			// Get the state db
			chainAstateDB := s.GetStateDB(s.chainA)

			// Create the contract from the precompile contract
			_, ctx := testutil.NewPrecompileContract(s.T(), s.chainA.GetContext(), sender.Addr, s.Precompile, 200000)

			// Execute the contract using the precompile
			res, err := s.Precompile.TransferWithDefaultTimeout(ctx, &method, chainAstateDB, args, sender.Addr)

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

				// Check if the data matches
				s.Require().Equal(transferEvent.Port, args[1])
				s.Require().Equal(transferEvent.Channel, args[2])
				s.Require().Equal(transferEvent.Amount, args[4])
				s.Require().Equal(transferEvent.Memo, args[5])

				// Get the next sequence to find our packet sequence
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

func (s *IBCPrecompileTestSuite) TestPrecompileTransfer() {
	// Get path and testcoin
	path := s.path
	coin := ibctesting.TestCoin

	// Get the method
	method := s.Precompile.Methods[ibcprecompile.TransferMethod] // Changed to TransferMethod

	// Get accounts
	sender := s.keyring.GetKey(0)

	// Base valid args
	validArgs := []any{
		s.keyring.GetKey(1).Addr.String(),   // receiver
		path.EndpointA.ChannelConfig.PortID, // port
		path.EndpointA.ChannelID,            // channel
		coin.Denom,                          // denom
		coin.Amount.BigInt(),                // amount
		uint64(1),                           // revisionNumber
		uint64(1000),                        // revisionHeight
		uint64(time.Now().Add(1 * time.Hour).UnixNano()), // timeoutTimestamp
		"test memo", // memo
	}

	tc := []struct {
		name        string
		modifyArgs  func([]any) []any
		errContains string
	}{
		{
			name:       "valid execute with timeout",
			modifyArgs: func(args []any) []any { return args },
		},
		{
			name: "invalid args - different than 9",
			modifyArgs: func(args []any) []any {
				return args[:1]
			},
			errContains: "expected 9 arguments but got 1",
		},
		{
			name: "invalid revision number - wrong type",
			modifyArgs: func(args []any) []any {
				args[5] = "not a number"
				return args
			},
			errContains: "revisionNumber is not a uint64",
		},
		{
			name: "invalid revision height - wrong type",
			modifyArgs: func(args []any) []any {
				args[6] = "not-a-uint"
				return args
			},
			errContains: "revisionHeight is not a uint64",
		},
		{
			name: "invalid timeout timestamp - wrong type",
			modifyArgs: func(args []any) []any {
				args[7] = "not-a-uint"
				return args
			},
			errContains: "timeoutTimestamp is not a uint64",
		},
	}

	for _, tc := range tc {
		s.Run(tc.name, func() {
			// Apply args modification
			args := tc.modifyArgs(append([]any(nil), validArgs...))

			// Get stateDB
			chainAstateDB := s.GetStateDB(s.chainA)

			// Create precompile
			_, ctx := testutil.NewPrecompileContract(s.T(), s.chainA.GetContext(), sender.Addr, s.Precompile, 200000)

			// Call transfer
			res, err := s.Precompile.Transfer(ctx, &method, chainAstateDB, args, sender.Addr)

			if tc.errContains != "" {
				s.Require().ErrorContains(err, tc.errContains)
			} else {
				s.Require().NoError(err)

				// Verify successful execution
				success, err := s.Precompile.Unpack(ibcprecompile.TransferMethod, res)
				s.Require().NoError(err)
				s.Require().True(success[0].(bool))

				// Verify event emission
				log := chainAstateDB.Logs()[0]
				event := s.Precompile.ABI.Events[ibcprecompile.EventTypeTransfer]
				s.Require().Equal(crypto.Keccak256Hash([]byte(event.Sig)), common.HexToHash(log.Topics[0].Hex()))
				s.Require().Equal(log.BlockNumber, uint64(ctx.BlockHeight()))

				// Decode the event data and check
				var transferEvent ibcprecompile.TransferEvent
				err = cmn.UnpackLog(s.Precompile.ABI, &transferEvent, ibcprecompile.EventTypeTransfer, *log)
				s.Require().NoError(err)

				// Check if the data matches
				s.Require().Equal(transferEvent.Port, args[1])
				s.Require().Equal(transferEvent.Channel, args[2])
				s.Require().Equal(transferEvent.Amount, args[4])
				s.Require().Equal(transferEvent.RevisionNumber, args[5])
				s.Require().Equal(transferEvent.RevisionHeight, args[6])
				s.Require().Equal(transferEvent.TimeoutTimestamp, args[7])
				s.Require().Equal(transferEvent.Memo, args[8])

				// Get next sequence to check packet commitment
				seq, found := s.chainA.App.GetIBCKeeper().ChannelKeeper.GetNextSequenceSend(
					s.chainA.GetContext(),
					ibctesting.TransferPort,
					path.EndpointA.ChannelID,
				)
				s.Require().True(found)

				// Verify packet commitment
				commitment := s.chainA.App.GetIBCKeeper().ChannelKeeper.GetPacketCommitment(
					s.chainA.GetContext(),
					ibctesting.TransferPort,
					path.EndpointA.ChannelID,
					seq-1,
				)
				s.Require().NotEmpty(commitment)
			}
		})
	}
}
