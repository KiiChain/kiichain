package wasmd

import (
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cmn "github.com/cosmos/evm/precompiles/common"
	"github.com/ethereum/go-ethereum/common"
)

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
	if !ok || admin == (common.Address{}) {
		return nil, fmt.Errorf("invalid admin address")
	}

	// Parse the second arg, the code ID
	codeID, ok := args[1].(uint64)
	if !ok || codeID == 0 {
		return nil, fmt.Errorf("invalid code ID")
	}

	// Parse the third arg, the label
	label, ok := args[2].(string)
	if !ok || label == "" {
		return nil, fmt.Errorf("invalid label")
	}

	// Parse the fourth arg, the init message
	msg, ok := args[3].([]byte)
	if !ok || len(msg) == 0 {
		return nil, fmt.Errorf("invalid init message")
	}

	// Parse the fifth arg, the funds
	funds, ok := args[4].([]sdk.Coin)
	if !ok || len(funds) == 0 {
		return nil, fmt.Errorf("invalid funds")
	}

	// Parse the sender
	senderAccAddress := sdk.AccAddress(sender.Bytes())
	// Parse the admin address
	adminAccAddress := sdk.AccAddress(admin.Bytes())

	// Create the MsgInstantiateContract and return
	return &wasmtypes.MsgInstantiateContract{
		Sender: senderAccAddress.String(),
		Admin:  adminAccAddress.String(),
		CodeID: codeID,
		Label:  label,
		Msg:    msg,
		Funds:  funds,
	}, nil
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
	if !ok || contractAddr == "" {
		return nil, fmt.Errorf("invalid contract address")
	}

	// Parse the second arg, the execute message
	msg, ok := args[1].([]byte)
	if !ok || len(msg) == 0 {
		return nil, fmt.Errorf("invalid execute message")
	}

	// Parse the third arg, the funds
	funds, ok := args[2].([]sdk.Coin)
	if !ok || len(funds) == 0 {
		return nil, fmt.Errorf("invalid funds")
	}

	// Parse the sender
	senderAccAddress := sdk.AccAddress(sender.Bytes())

	// Create the MsgExecuteContract and return
	return &wasmtypes.MsgExecuteContract{
		Sender:   senderAccAddress.String(),
		Contract: contractAddr,
		Msg:      msg,
		Funds:    funds,
	}, nil
}
