package ibc

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Default increment values
const (
	DefaultBlockIncrement = 100
	DefaultTimeIncrement  = 10 * time.Minute
)

// TransferEvent represents the solidity event that is logged
type TransferEvent struct {
	Caller           common.Address
	Denom            common.Hash
	Receiver         common.Hash
	Port             string
	Channel          string
	Amount           *big.Int
	RevisionNumber   uint64
	RevisionHeight   uint64
	TimeoutTimestamp uint64
	Memo             string
}

// NewMsgTransfer creates a new Transfer message
func NewMsgTransfer(
	ctx sdk.Context,
	method *abi.Method,
	sender common.Address,
	args []interface{},
) (*types.MsgTransfer, error) {
	if err := ValidateArgsLength(args, 9); err != nil {
		return nil, err
	}
	validatedArgs, err := validateCommonArgs(ctx, args, sender)
	if err != nil {
		return nil, err
	}

	if validatedArgs.amount.Cmp(big.NewInt(0)) == 0 {
		// short circuit
		return nil, errors.New("amount is zero, transaction is invalid")
	}

	coin := sdk.Coin{
		Denom:  validatedArgs.denom,
		Amount: math.NewIntFromBigInt(validatedArgs.amount),
	}

	revisionNumber, ok := args[5].(uint64)
	if !ok {
		return nil, errors.New("revisionNumber is not a uint64")
	}

	revisionHeight, ok := args[6].(uint64)
	if !ok {
		return nil, errors.New("revisionHeight is not a uint64")
	}

	height := clienttypes.Height{
		RevisionNumber: revisionNumber,
		RevisionHeight: revisionHeight,
	}

	timeoutTimestamp, ok := args[7].(uint64)
	if !ok {
		return nil, errors.New("timeoutTimestamp is not a uint64")
	}

	msg := types.MsgTransfer{
		SourcePort:       validatedArgs.port,
		SourceChannel:    validatedArgs.channelID,
		Token:            coin,
		Sender:           validatedArgs.senderKiiAddr.String(),
		Receiver:         validatedArgs.receiverAddressString,
		TimeoutHeight:    height,
		TimeoutTimestamp: timeoutTimestamp,
	}

	msg = addMemo(args[8], msg)

	err = msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// NewMsgTransferDefaultTimeout builds a new transfer message while collecting timeout information
func (p Precompile) NewMsgTransferDefaultTimeout(
	ctx sdk.Context,
	method *abi.Method,
	sender common.Address,
	args []interface{},
) (*types.MsgTransfer, error) {
	if err := ValidateArgsLength(args, 6); err != nil {
		return nil, err
	}
	validatedArgs, err := validateCommonArgs(ctx, args, sender)
	if err != nil {
		return nil, err
	}

	if validatedArgs.amount.Cmp(big.NewInt(0)) == 0 {
		// short circuit
		return nil, errors.New("amount is zero, transaction is invalid")
	}

	coin := sdk.Coin{
		Denom:  validatedArgs.denom,
		Amount: math.NewIntFromBigInt(validatedArgs.amount),
	}

	connection, err := p.getChannelConnection(ctx, validatedArgs.port, validatedArgs.channelID)
	if err != nil {
		return nil, err
	}

	latestConsensusHeight := p.getConsensusLatestHeight(ctx, *connection)

	timeoutTimestamp, err := p.GetAdjustedTimestamp(ctx, connection.ClientId, latestConsensusHeight)
	if err != nil {
		return nil, err
	}

	// Adjust timeout height by adding timeout height
	incrementedHeight := clienttypes.NewHeight(
		latestConsensusHeight.GetRevisionNumber(),
		latestConsensusHeight.GetRevisionHeight()+DefaultBlockIncrement,
	)

	msg := types.MsgTransfer{
		SourcePort:       validatedArgs.port,
		SourceChannel:    validatedArgs.channelID,
		Token:            coin,
		Sender:           validatedArgs.senderKiiAddr.String(),
		Receiver:         validatedArgs.receiverAddressString,
		TimeoutHeight:    incrementedHeight,
		TimeoutTimestamp: timeoutTimestamp,
	}

	msg = addMemo(args[5], msg)

	return &msg, msg.ValidateBasic()
}

// getChannelConnection gets the channel connection from the channel keeper
func (p Precompile) getChannelConnection(ctx sdk.Context, port string, channelID string) (*connectiontypes.ConnectionEnd, error) {
	channel, found := p.channelKeeper.GetChannel(ctx, port, channelID)
	if !found {
		return nil, errors.New("channel not found")
	}

	connection, found := p.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])

	if !found {
		return nil, errors.New("connection not found")
	}
	return &connection, nil
}

// getConsensusLatestHeight obtains the consensus latest height
func (p Precompile) getConsensusLatestHeight(ctx sdk.Context, connection connectiontypes.ConnectionEnd) clienttypes.Height {
	return p.clientKeeper.GetClientLatestHeight(ctx, connection.ClientId)
}

// GetAdjustedTimestamp creates default timestamp from height and unix
func (p Precompile) GetAdjustedTimestamp(ctx sdk.Context, clientID string, height clienttypes.Height) (uint64, error) {
	// Get adjusted timestamp
	timeoutTimestamp, err := p.clientKeeper.GetClientTimestampAtHeight(ctx, clientID, height)
	if err != nil {
		return 0, err
	}

	// Increment timestamp by default amt
	adjustedTime := time.Unix(0, int64(timeoutTimestamp))
	incrementedTimestamp := uint64(adjustedTime.Add(DefaultTimeIncrement).UnixNano())
	return incrementedTimestamp, nil
}

// ValidatedArgs stores common args that have been validated
type ValidatedArgs struct {
	senderKiiAddr         sdk.AccAddress
	receiverAddressString string
	port                  string
	channelID             string
	denom                 string
	amount                *big.Int
}

// validateCommonArgs validates common transfer args
func validateCommonArgs(ctx sdk.Context, args []interface{}, caller common.Address) (*ValidatedArgs, error) {
	senderKiiAddr, err := GetKiiAddressByEvmAddress(ctx, caller)
	if err != nil {
		return nil, err
	}

	receiverAddressString, ok := args[0].(string)
	if !ok || receiverAddressString == "" {
		return nil, errors.New("receiverAddress is not a string or empty")
	}

	port, ok := args[1].(string)
	if !ok {
		return nil, errors.New("port is not a string")
	}
	if port == "" {
		return nil, errors.New("port cannot be empty")
	}

	channelID, ok := args[2].(string)
	if !ok {
		return nil, errors.New("channelID is not a string")
	}
	if channelID == "" {
		return nil, errors.New("channelID cannot be empty")
	}

	denom, ok := args[3].(string)
	if !ok || denom == "" {
		return nil, errors.New("denom is not a string or is empty")
	}

	amount, ok := args[4].(*big.Int)
	if !ok {
		return nil, errors.New("amount is not a big.Int")
	}
	return &ValidatedArgs{
		senderKiiAddr:         senderKiiAddr,
		receiverAddressString: receiverAddressString,
		port:                  port,
		channelID:             channelID,
		denom:                 denom,
		amount:                amount,
	}, nil
}

// addMemo adds the memo string to the transfer
func addMemo(memoArg interface{}, transferMsg types.MsgTransfer) types.MsgTransfer {
	memo := ""
	if memoArg != nil {
		if m, ok := memoArg.(string); ok {
			memo = m
		}
	}
	transferMsg.Memo = memo
	return transferMsg
}

// ValidateArgsLength checks if the length of the args is as expected
func ValidateArgsLength(args []interface{}, length int) error {
	if len(args) != length {
		return fmt.Errorf("expected %d arguments but got %d", length, len(args))
	}

	return nil
}

// GetKiiAddressByEvmAddress transforms evm address into a kii address
func GetKiiAddressByEvmAddress(ctx sdk.Context, evmAddress common.Address) (sdk.AccAddress, error) {
	cosmosAddr := sdk.AccAddress(evmAddress.Bytes()) // Check this is working as intended
	return cosmosAddr, nil
}
