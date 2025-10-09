package wasmd

import (
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmdkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
)

const (
	// InstantiateMethod is the method name for instantiating a contract
	InstantiateMethod = "instantiate"
	// ExecuteMethod is the method name for executing a contract
	ExecuteMethod = "execute"
)

// reentrancyPerTargetNs is the namespace for the reentrancy lock per target contract
var reentrancyPerTargetNs = []byte("wasmd.precompile.reentrancy.lock:")

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

	// Ensure the reentrancy lock
	if err := p.ensureLock(origin, stateDB, method); err != nil {
		return nil, err
	}

	// Log the call
	p.Logger(ctx).Debug(
		"tx called",
		"method", method.Name,
		"args", fmt.Sprintf(
			"{ contract: %s, sender: %s }",
			msg.Contract, msg.Sender,
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

// ensureLock ensures that a reentrancy lock is set and not broken for the target contract
// Reentrance lock is built using: precompile address, origin and origin nonce
//   - Args are avoided to build the lock key, since the attacker may manipulate it
// This is done under a transient key under Cosmos SDK's stateDB
// The lock is released at the end of the transaction by Cosmos SDK itself
func (p Precompile) ensureLock(
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
) error {
	// Build the lock key
	lockKey := buildReentrancyLockKey(p.Address(), origin, stateDB.GetNonce(origin))

	// Check if already locked
	if stateDB.GetTransientState(p.Address(), lockKey) == common.BytesToHash([]byte{1}) {
		return fmt.Errorf(
			"reentrancy detected in precompile %s, method %s",
			p.Address().Hex(), method.Name,
		)
	}

	// Set lock
	stateDB.SetTransientState(p.Address(), lockKey, common.BytesToHash([]byte{1}))
	return nil
}

// buildReentrancyLockKey builds a deterministic lock key:
//	H( "wasmd.precompile.reentrancy.lock:", precompileAddr, originAddr, originNonce[8] )
func buildReentrancyLockKey(
	precompileAddr common.Address,
	origin common.Address,
	originNonce uint64,
) common.Hash {
	var nonceBytes [8]byte
	binary.BigEndian.PutUint64(nonceBytes[:], originNonce)

	return crypto.Keccak256Hash(
		reentrancyPerTargetNs,
		precompileAddr.Bytes(),
		origin.Bytes(),
		nonceBytes[:],
	)
}
