package wasmd

import (
	"embed"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"

	wasmdkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	cmn "github.com/cosmos/evm/precompiles/common"
	"github.com/cosmos/evm/x/vm/core/vm"
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
	authzKeeper authzkeeper.Keeper,
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
			AuthzKeeper:          authzKeeper,
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
			ApprovalExpiration:   cmn.DefaultExpirationDuration,
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
	// Initialize the context, db and chain data
	ctx, stateDB, snapshot, method, initialGas, args, err := p.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	// This handles any out of gas errors
	defer cmn.HandleGasError(ctx, contract, initialGas, &err)()

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
	if !contract.UseGas(cost) {
		return nil, vm.ErrOutOfGas
	}

	// Add the new journal entries to the stateDB
	if err := p.AddJournalEntries(stateDB, snapshot); err != nil {
		return nil, err
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
