package wasmd

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	cmn "github.com/cosmos/evm/precompiles/common"
)

// ContractInstantiatedEvent is the event emitted when a contract is instantiated
type ContractInstantiatedEvent struct {
	Caller          common.Address
	CodeID          uint64
	ContractAddress string
	Data            []byte
}

// ContractExecutedEvent is the event emitted when a contract is executed
type ContractExecutedEvent struct {
	ContractAddress common.Hash
	Caller          common.Address
	Data            []byte
}

// ParseQueryRawArgs parses the arguments for the raw query method
func ParseQueryRawArgs(args []interface{}) (*wasmtypes.QueryRawContractStateRequest, error) {
	// Check the number of arguments, should be 2
	if len(args) != 2 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	// Parse the first arg, the contract address
	contractAddr, ok := args[0].(string)
	if !ok || contractAddr == "" {
		return nil, fmt.Errorf("invalid contract address")
	}

	// Check if the addr is a valid bech32 address
	if _, err := sdk.AccAddressFromBech32(contractAddr); err != nil {
		return nil, fmt.Errorf("invalid contract address: %w", err)
	}

	// Parse the second arg, the query data
	queryData, ok := args[1].([]byte)
	if !ok || len(queryData) == 0 {
		return nil, fmt.Errorf("invalid query data")
	}

	// Create the QueryRawContractStateRequest and return
	return &wasmtypes.QueryRawContractStateRequest{
		Address:   contractAddr,
		QueryData: queryData,
	}, nil
}

// ParseQuerySmartArgs parses the arguments for the smart query method
func ParseQuerySmartArgs(args []interface{}) (*wasmtypes.QuerySmartContractStateRequest, error) {
	// Check the number of arguments, should be 2
	if len(args) != 2 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 2, len(args))
	}

	// Parse the first arg, the contract address
	contractAddr, ok := args[0].(string)
	if !ok || contractAddr == "" {
		return nil, fmt.Errorf("invalid contract address")
	}

	// Check if the addr is a valid bech32 address
	if _, err := sdk.AccAddressFromBech32(contractAddr); err != nil {
		return nil, fmt.Errorf("invalid contract address: %w", err)
	}

	// Parse the second arg, the query data
	queryData, ok := args[1].([]byte)
	if !ok || len(queryData) == 0 {
		return nil, fmt.Errorf("invalid query data")
	}

	// Create the QuerySmartContractStateRequest and return
	return &wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddr,
		QueryData: queryData,
	}, nil
}

// NewMsgInstantiate creates a new instantiate message from args
func NewMsgInstantiate(
	sender common.Address,
	args []interface{},
) (*wasmtypes.MsgInstantiateContract, error) {
	// Check the number of arguments, should be 5
	if len(args) != 5 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 5, len(args))
	}

	// Parse the first arg, the admin
	admin, ok := args[0].(common.Address)
	if !ok {
		return nil, fmt.Errorf("invalid admin address")
	}

	// Parse the second arg, the code ID
	codeID, ok := args[1].(uint64)
	if !ok {
		return nil, fmt.Errorf("invalid code ID")
	}

	// Parse the third arg, the label
	label, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("invalid label")
	}

	// Parse the fourth arg, the init message
	msg, ok := args[3].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid init message")
	}

	// Parse the fifth arg, the funds
	funds, err := ConvertEVMCoinsToSDKCoins(args[4])
	if err != nil {
		return nil, fmt.Errorf("invalid funds: %w", err)
	}

	// Parse the sender
	senderAccAddress := sdk.AccAddress(sender.Bytes())
	// Parse the admin address
	adminAccAddress := sdk.AccAddress(admin.Bytes())

	// Get the message
	msgInstantiate := &wasmtypes.MsgInstantiateContract{
		Sender: senderAccAddress.String(),
		Admin:  adminAccAddress.String(),
		CodeID: codeID,
		Label:  label,
		Msg:    msg,
		Funds:  funds,
	}

	// Create the MsgInstantiateContract and return
	return msgInstantiate, msgInstantiate.ValidateBasic()
}

// NewMsgExecute creates a new execute message from args
func NewMsgExecute(
	sender common.Address,
	args []interface{},
) (*wasmtypes.MsgExecuteContract, error) {
	// Check the number of arguments, should be 3
	if len(args) != 3 {
		return nil, fmt.Errorf(cmn.ErrInvalidNumberOfArgs, 3, len(args))
	}

	// Parse the first arg, the contract address
	contractAddr, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("invalid contract address")
	}

	// Check if the addr is a valid bech32 address
	if _, err := sdk.AccAddressFromBech32(contractAddr); err != nil {
		return nil, fmt.Errorf("invalid contract address: %w", err)
	}

	// Parse the second arg, the execute message
	msg, ok := args[1].([]byte)
	if !ok {
		return nil, fmt.Errorf("invalid execute message")
	}

	// Parse the third arg, the funds
	funds, err := ConvertEVMCoinsToSDKCoins(args[2])
	if err != nil {
		return nil, fmt.Errorf("invalid funds: %w", err)
	}

	// Parse the sender
	senderAccAddress := sdk.AccAddress(sender.Bytes())

	// Get the message
	msgExecute := &wasmtypes.MsgExecuteContract{
		Sender:   senderAccAddress.String(),
		Contract: contractAddr,
		Msg:      msg,
		Funds:    funds,
	}

	// Create the MsgExecuteContract and return
	return msgExecute, msgExecute.ValidateBasic()
}

// ConvertEVMCoinsToSDKCoins converts a slice of EVM coins to SDK coins
func ConvertEVMCoinsToSDKCoins(input any) ([]sdk.Coin, error) {
	v := reflect.ValueOf(input)

	// Check if the input is a slice
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expected slice, got %T", input)
	}

	// Create the response
	var coins []sdk.Coin
	for i := 0; i < v.Len(); i++ {
		// Get the item
		item := v.Index(i)

		// Get their fields
		denom := item.FieldByName("Denom").String()
		amount := item.FieldByName("Amount").Interface().(*big.Int)

		// Add to the array as a new coin
		coins = append(coins, sdk.NewCoin(denom, math.NewIntFromBigInt(amount)))
	}

	// Return the response
	return coins, nil
}
