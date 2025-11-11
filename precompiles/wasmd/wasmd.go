package wasmd

import (
	"embed"
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmdkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	cmn "github.com/cosmos/evm/precompiles/common"
)

const (
	// WasmdPrecompileAddress is the address of the precompile
	WasmdPrecompileAddress = "0x0000000000000000000000000000000000001001"
)

// Precompile is a struct that implements the PrecompiledContract interface
var _ vm.PrecompiledContract = &Precompile{}

// Embed the json abi to the binary
//
//go:embed abi.json
var f embed.FS

// Precompile defines the struct for the wasmd precompile
type Precompile struct {
	cmn.Precompile
	wasmdKeeper wasmdkeeper.Keeper
}

// LoadABI loads the ABI from the embedded file for the wasmd precompile
func LoadABI() (abi.ABI, error) {
	return cmn.LoadABI(f, "abi.json")
}

// NewPrecompile starts a new wasmd precompile
func NewPrecompile(
	wasmdKeeper wasmdkeeper.Keeper,
) (*Precompile, error) {
	// Load the abi
	abi, err := LoadABI()
	if err != nil {
		return nil, err
	}

	// Initialize the precompile
	precompile := &Precompile{
		Precompile: cmn.Precompile{
			ABI:                  abi,
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
		},
		wasmdKeeper: wasmdKeeper,
	}

	// Set the address of the precompile
	precompile.SetAddress(common.HexToAddress(WasmdPrecompileAddress))

	// Return the precompile
	return precompile, nil
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
	bz, err = p.run(evm, contract, readOnly)
	if err != nil {
		return cmn.ReturnRevertError(evm, err)
	}
	return bz, nil
}

// run executes the wasmd precompile (internal function)
func (p Precompile) run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	// Initialize the context, db and chain data
	ctx, stateDB, method, initialGas, args, err := p.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	// This handles any out of gas errors
	defer cmn.HandleGasError(ctx, contract, initialGas, &err)()

	// Ensure the reentrancy lock
	if err := p.ensureLock(evm.Origin, stateDB, method); err != nil {
		return nil, err
	}

	// Now we call the method based on the function
	switch method.Name {
	// Wasmd transactions
	case InstantiateMethod:
		bz, err = p.Instantiate(ctx, evm.Origin, contract, stateDB, method, args)
	case ExecuteMethod:
		bz, err = p.Execute(ctx, evm.Origin, contract, stateDB, method, args)
	// Wasmd queries
	case QueryRawMethod:
		bz, err = p.QueryRaw(ctx, method, args)
	case QuerySmartMethod:
		bz, err = p.QuerySmart(ctx, method, args)
	default:
		// If default error out
		return nil, fmt.Errorf(cmn.ErrUnknownMethod, method.Name)
	}
	if err != nil {
		return nil, err
	}

	// Check the gas cost
	cost := ctx.GasMeter().GasConsumed() - initialGas
	if !contract.UseGas(cost, nil, tracing.GasChangeCallPrecompiledContract) {
		return nil, vm.ErrOutOfGas
	}

	return bz, nil
}

// IsTransaction checks if the method is a transaction
//
// Queries are not added here
func (Precompile) IsTransaction(method *abi.Method) bool {
	// Check if the method is a transaction
	switch method.Name {
	case InstantiateMethod, ExecuteMethod:
		return true
	default:
		return false
	}
}

// Logger returns the logger for the precompile
func (p Precompile) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("evm extension", "wasmd")
}

// ensureLock ensures that a reentrancy lock is set and not broken for the target contract
// Reentrance lock is built using: precompile address, origin address, origin nonce, and method ID
//   - Args are avoided to build the lock key, since the attacker may manipulate it
//
// This is done under a transient key under Cosmos SDK's stateDB
// The lock is released at the end of the transaction by Cosmos SDK itself
// Locks are specifically done per transaction and not per contract:
// - Due to a limitation between EVM and WASM gas handling
// - And to reduce the scope of the reentrances
// - States generated as transient states are never committed and cleared at the end of the block
func (p Precompile) ensureLock(
	origin common.Address,
	stateDB vm.StateDB,
	method *abi.Method,
) error {
	// Build the lock key
	lockKey := buildReentrancyLockKey(p.Address(), origin, stateDB.GetNonce(origin), method)

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
//
//	H( "wasmd.precompile.reentrancy.lock:", precompileAddr, originAddr, originNonce[:], and methodID )
func buildReentrancyLockKey(
	precompileAddr common.Address,
	origin common.Address,
	originNonce uint64,
	method *abi.Method,
) common.Hash {
	var nonceBytes [8]byte
	binary.BigEndian.PutUint64(nonceBytes[:], originNonce)

	return crypto.Keccak256Hash(
		reentrancyPerTargetNs,
		precompileAddr.Bytes(),
		origin.Bytes(),
		nonceBytes[:],
		method.ID,
	)
}
