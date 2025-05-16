package ibc

import (
	"embed"
	"errors"
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clientkeeper "github.com/cosmos/ibc-go/v8/modules/core/02-client/keeper"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectionkeeper "github.com/cosmos/ibc-go/v8/modules/core/03-connection/keeper"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channelkeeper "github.com/cosmos/ibc-go/v8/modules/core/04-channel/keeper"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	cmn "github.com/cosmos/evm/precompiles/common"
	ibctransferkeeper "github.com/cosmos/evm/x/ibc/transfer/keeper"
	"github.com/cosmos/evm/x/vm/core/vm"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	// pcommon "github.com/kiichain/kiichain/precompiles/common"
)

const (
	TransferMethod                   = "transfer"
	TransferWithDefaultTimeoutMethod = "transferWithDefaultTimeout"
)

const (
	IBCPrecompileAddress = "0x0000000000000000000000000000000000001002"
)

// Embed abi json file to the executable binary. Needed when importing as dependency.
//
//go:embed abi.json
var f embed.FS

type Precompile struct {
	cmn.Precompile
	transferKeeper   ibctransferkeeper.Keeper
	evmKeeper        evmkeeper.Keeper
	clientKeeper     clientkeeper.Keeper
	connectionKeeper connectionkeeper.Keeper
	channelKeeper    channelkeeper.Keeper

	TransferID                   []byte
	TransferWithDefaultTimeoutID []byte
}

func NewPrecompile(
	transferKeeper ibctransferkeeper.Keeper,
	evmKeeper evmkeeper.Keeper,
	clientKeeper clientkeeper.Keeper,
	connectionKeeper connectionkeeper.Keeper,
	channelKeeper channelkeeper.Keeper) (*Precompile, error) {
	// Load abi
	newAbi, err := cmn.LoadABI(f, "abi.json")
	if err != nil {
		return nil, err
	}

	// Setup keepers
	p := &Precompile{
		transferKeeper:   transferKeeper,
		evmKeeper:        evmKeeper,
		clientKeeper:     clientKeeper,
		connectionKeeper: connectionKeeper,
		channelKeeper:    channelKeeper,
	}

	for name, m := range newAbi.Methods {
		switch name {
		case TransferMethod:
			p.TransferID = m.ID
		case TransferWithDefaultTimeoutMethod:
			p.TransferWithDefaultTimeoutID = m.ID
		}
	}

	// Set the address of the precompile
	p.SetAddress(common.HexToAddress(IBCPrecompileAddress))

	// Return the precompile
	return p, nil
}

// RequiredGas returns the gas required for the precompile
// This is the same implementation as the one from the EVM module pre-compiles
func (p Precompile) RequiredGas(input []byte) uint64 {
	// This is a check to avoid panic
	if len(input) < 4 {
		return 0
	}

	// Get the method ID from the first 4 bytes
	methodID := input[:4]

	// Get the method from the ABI
	method, err := p.MethodById(methodID)
	if err != nil {
		return 0
	}

	// Get the gas required for the method
	return p.Precompile.RequiredGas(input, p.IsTransaction(method))
}

// Run executes the wasmd precompile
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	// Initialize the context, db and chain data
	ctx, stateDB, snapshot, method, initialGas, args, err := p.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	// This handles any out of gas errors
	defer cmn.HandleGasError(ctx, contract, initialGas, &err)()

	// Now we call the method based on the function
	switch method.Name {
	case TransferMethod:
		bz, err = p.transfer(ctx, method, args, evm.Origin)
	case TransferWithDefaultTimeoutMethod:
		bz, err = p.transferWithDefaultTimeout(ctx, method, args, evm.Origin)
	default:
		// If default error out
		return nil, fmt.Errorf(cmn.ErrUnknownMethod, method.Name)
	}
	if err != nil {
		return nil, err
	}
	// Check the gas cost
	cost := ctx.GasMeter().GasConsumed() - initialGas
	if !contract.UseGas(cost) {
		return nil, vm.ErrOutOfGas
	}

	// Add the new journal entries to the stateDB
	if err := p.AddJournalEntries(stateDB, snapshot); err != nil {
		return nil, err
	}

	return bz, nil
}

// _ *tracing.Hooks removed. What was this doing?
func (p Precompile) Execute(ctx sdk.Context, method *abi.Method, caller common.Address, callingContract common.Address, args []interface{}, value *big.Int, readOnly bool, evm *vm.EVM, suppliedGas uint64) (ret []byte, err error) {
	if err = ValidateNonPayable(value); err != nil {
		return nil, err
	}

	if readOnly {
		return nil, errors.New("cannot call IBC precompile from staticcall")
	}

	if EVMPrecompileCalledFromDelegateCall(caller, callingContract) {
		return nil, errors.New("cannot delegatecall IBC")
	}

	switch method.Name {
	case TransferMethod:
		return p.transfer(ctx, method, args, caller)
	case TransferWithDefaultTimeoutMethod:
		return p.transferWithDefaultTimeout(ctx, method, args, caller)
	}
	return
}

func (p Precompile) EVMKeeper() evmkeeper.Keeper {
	return p.evmKeeper
}

func (p Precompile) transfer(ctx sdk.Context, method *abi.Method, args []interface{}, caller common.Address) (ret []byte, rerr error) {
	defer func() {
		if err := recover(); err != nil {
			ret = nil
			rerr = fmt.Errorf("%s", err)
			return
		}
	}()

	if err := ValidateArgsLength(args, 9); err != nil {
		rerr = err
		return
	}
	validatedArgs, err := p.validateCommonArgs(ctx, args, caller)
	if err != nil {
		rerr = err
		return
	}

	if validatedArgs.amount.Cmp(big.NewInt(0)) == 0 {
		// short circuit
		ret, rerr = method.Outputs.Pack(true)
		return
	}

	coin := sdk.Coin{
		Denom:  validatedArgs.denom,
		Amount: math.NewIntFromBigInt(validatedArgs.amount),
	}

	revisionNumber, ok := args[5].(uint64)
	if !ok {
		rerr = errors.New("revisionNumber is not a uint64")
		return
	}

	revisionHeight, ok := args[6].(uint64)
	if !ok {
		rerr = errors.New("revisionHeight is not a uint64")
		return
	}

	height := clienttypes.Height{
		RevisionNumber: revisionNumber,
		RevisionHeight: revisionHeight,
	}

	timeoutTimestamp, ok := args[7].(uint64)
	if !ok {
		rerr = errors.New("timeoutTimestamp is not a uint64")
		return
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
		rerr = err
		return
	}

	_, err = p.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), &msg)

	if err != nil {
		rerr = err
		return
	}
	ret, rerr = method.Outputs.Pack(true)
	return
}

func (p Precompile) transferWithDefaultTimeout(ctx sdk.Context, method *abi.Method, args []interface{}, caller common.Address) (ret []byte, rerr error) {
	defer func() {
		if err := recover(); err != nil {
			ret = nil
			rerr = fmt.Errorf("%s", err)
			return
		}
	}()

	if err := ValidateArgsLength(args, 6); err != nil {
		rerr = err
		return
	}
	validatedArgs, err := p.validateCommonArgs(ctx, args, caller)
	if err != nil {
		rerr = err
		return
	}

	if validatedArgs.amount.Cmp(big.NewInt(0)) == 0 {
		// short circuit
		ret, rerr = method.Outputs.Pack(true)
		return
	}

	coin := sdk.Coin{
		Denom:  validatedArgs.denom,
		Amount: math.NewIntFromBigInt(validatedArgs.amount),
	}

	connection, err := p.getChannelConnection(ctx, validatedArgs.port, validatedArgs.channelID)

	if err != nil {
		rerr = err
		return
	}

	latestConsensusHeight, err := p.getConsensusLatestHeight(ctx, *connection)
	if err != nil {
		rerr = err
		return
	}

	height, err := GetAdjustedHeight(*latestConsensusHeight)
	if err != nil {
		rerr = err
		return
	}

	timeoutTimestamp, err := p.GetAdjustedTimestamp(ctx, connection.ClientId, *latestConsensusHeight)
	if err != nil {
		rerr = err
		return
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

	err = msg.ValidateBasic()
	if err != nil {
		rerr = err
		return
	}

	_, err = p.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), &msg)

	if err != nil {
		rerr = err
		return
	}
	ret, rerr = method.Outputs.Pack(true)
	return
}

func (p Precompile) accAddressFromArg(ctx sdk.Context, arg interface{}) (sdk.AccAddress, error) {
	addr := arg.(common.Address)
	if addr == (common.Address{}) {
		return nil, errors.New("invalid addr")
	}
	kiiAddr, err := GetKiiAddressByEvmAddress(ctx, addr)
	if err != nil {
		return nil, err
	}
	return kiiAddr, nil
}

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

func (p Precompile) validateCommonArgs(ctx sdk.Context, args []interface{}, caller common.Address) (*ValidatedArgs, error) {
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

func addMemo(memoArg interface{}, transferMsg types.MsgTransfer) types.MsgTransfer {
	memo := ""
	if memoArg != nil {
		memo = memoArg.(string)
	}
	transferMsg.Memo = memo
	return transferMsg
}

func ValidateArgsLength(args []interface{}, length int) error {
	if len(args) != length {
		return fmt.Errorf("expected %d arguments but got %d", length, len(args))
	}

	return nil
}

func ValidateNonPayable(value *big.Int) error {
	if value != nil && value.Sign() != 0 {
		return errors.New("sending funds to a non-payable function")
	}

	return nil
}

func GetKiiAddressByEvmAddress(ctx sdk.Context, evmAddress common.Address) (sdk.AccAddress, error) {
	cosmosAddr := sdk.AccAddress(evmAddress.Bytes()) // Check this is working as intended
	return cosmosAddr, nil
}

func GetKiiAddressFromArg(ctx sdk.Context, arg interface{}) (sdk.AccAddress, error) {
	addr := arg.(common.Address)
	if addr == (common.Address{}) {
		return nil, errors.New("invalid addr")
	}
	return GetKiiAddressByEvmAddress(ctx, addr)
}

func GetRemainingGas(ctx sdk.Context) uint64 {
	gasMeter := ctx.GasMeter() // Verify if there is a ratio or just 1:1 between cosmos and evm
	return gasMeter.GasRemaining()
}

func EVMPrecompileCalledFromDelegateCall(caller, callingContract common.Address) bool {
	// This method blocks a lot of calls from contracts
	// What alternative is there for this?
	return callingContract != caller
}

func (Precompile) IsTransaction(method *abi.Method) bool {
	// Check if the method is a transaction
	switch method.Name {
	case TransferMethod, TransferWithDefaultTimeoutMethod:
		return true
	default:
		return false
	}
}
