package wasmd

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmdkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	"github.com/cosmos/evm/x/vm/core/vm"
)

const (
	// InstantiateMethod is the method name for instantiating a contract
	InstantiateMethod = "instantiate"
	// ExecuteMethod is the method name for executing a contract
	ExecuteMethod = "execute"
)

// Instantiate executes wasmd instantiate from the precompile
func (p Precompile) Instantiate(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Create the instantiate message
	msg, err := NewMsgInstantiate(origin, args)
	if err != nil {
		return nil, err
	}

	// Log the call
	p.Logger(ctx).Debug(
		"tx called",
		"method", method.Name,
		"args", fmt.Sprintf(
			"{ admin: %s, code_id: %d, sender: %s }",
			msg.Admin, msg.CodeID, msg.Sender,
		),
	)

	// Initialize the message server
	msgSrv := wasmdkeeper.NewMsgServerImpl(&p.wasmdKeeper)

	// Call the instantiate method
	res, err := msgSrv.InstantiateContract(ctx, msg)
	if err != nil {
		return nil, err
	}

	// Emit the event
	err = p.EmitEventContractInstantiated(ctx, stateDB, origin, msg.CodeID, res.Address, res.Data)
	if err != nil {
		return nil, err
	}

	// Return the response
	return method.Outputs.Pack(true)
}

// Execute executes wasmd execute from the precompile
func (p Precompile) Execute(
	ctx sdk.Context,
	origin common.Address,
	contract *vm.Contract,
	stateDB vm.StateDB,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	// Create the execute message
	msg, err := NewMsgExecute(origin, args)
	if err != nil {
		return nil, err
	}

	// Log the call
	p.Logger(ctx).Debug(
		"tx called",
		"method", method.Name,
		"args", fmt.Sprintf(
			"{ contract: %s, msg: %s, sender: %s }",
			msg.Contract, msg.Msg, msg.Sender,
		),
	)

	// Initialize the message server
	msgSrv := wasmdkeeper.NewMsgServerImpl(&p.wasmdKeeper)

	// Call the instantiate method
	res, err := msgSrv.ExecuteContract(ctx, msg)
	if err != nil {
		return nil, err
	}

	// Emit the event
	err = p.EmitEventContractExecuted(ctx, stateDB, msg.Contract, origin, res.Data)
	if err != nil {
		return nil, err
	}

	// Return the response
	return method.Outputs.Pack(true)
}
