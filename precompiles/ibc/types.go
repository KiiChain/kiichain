package ibc

import (
	"errors"
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

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
		_, rerr := method.Outputs.Pack(true)
		return nil, rerr
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
		return nil, errors.New("Amount is zero, transaction is invalid")
	}

	coin := sdk.Coin{
		Denom:  validatedArgs.denom,
		Amount: math.NewIntFromBigInt(validatedArgs.amount),
	}

	connection, err := p.getChannelConnection(ctx, validatedArgs.port, validatedArgs.channelID)

	if err != nil {
		return nil, err
	}

	latestConsensusHeight, err := p.getConsensusLatestHeight(ctx, *connection)
	if err != nil {
		return nil, err
	}

	height, err := GetAdjustedHeight(*latestConsensusHeight)
	if err != nil {
		return nil, err
	}

	timeoutTimestamp, err := p.GetAdjustedTimestamp(ctx, connection.ClientId, *latestConsensusHeight)
	if err != nil {
		return nil, err
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
func (p Precompile) getConsensusLatestHeight(ctx sdk.Context, connection connectiontypes.ConnectionEnd) (*clienttypes.Height, error) {
	clientState, found := p.clientKeeper.GetClientState(ctx, connection.ClientId)

	if !found {
		return nil, errors.New("could not get the client state")
	}

	latestHeight := clientState.GetLatestHeight()
	return &clienttypes.Height{
		RevisionNumber: latestHeight.GetRevisionNumber(),
		RevisionHeight: latestHeight.GetRevisionHeight(),
	}, nil
}

// GetAdjustedHeight calculates the default timeout height
func GetAdjustedHeight(latestConsensusHeight clienttypes.Height) (clienttypes.Height, error) {
	defaultTimeoutHeight, err := clienttypes.ParseHeight(types.DefaultRelativePacketTimeoutHeight)
	if err != nil {
		return clienttypes.Height{}, err
	}

	absoluteHeight := latestConsensusHeight
	absoluteHeight.RevisionNumber += defaultTimeoutHeight.RevisionNumber
	absoluteHeight.RevisionHeight += defaultTimeoutHeight.RevisionHeight
	return absoluteHeight, nil
}

// GetAdjustedTimestamp creates default timestamp from height and unix
func (p Precompile) GetAdjustedTimestamp(ctx sdk.Context, clientId string, height clienttypes.Height) (uint64, error) {
	consensusState, found := p.clientKeeper.GetClientConsensusState(ctx, clientId, height)
	var consensusStateTimestamp uint64
	if found {
		consensusStateTimestamp = consensusState.GetTimestamp()
	}

	defaultRelativePacketTimeoutTimestamp := types.DefaultRelativePacketTimeoutTimestamp
	blockTime := ctx.BlockTime().UnixNano()
	if blockTime > 0 {
		now := uint64(blockTime)
		if now > consensusStateTimestamp {
			return now + defaultRelativePacketTimeoutTimestamp, nil
		} else {
			return consensusStateTimestamp + defaultRelativePacketTimeoutTimestamp, nil
		}
	} else {
		return 0, errors.New("block time is not greater than Jan 1st, 1970 12:00 AM")
	}
}

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

	denom := args[3].(string)
	if denom == "" {
		return nil, errors.New("invalid denom")
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
		memo = memoArg.(string)
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
