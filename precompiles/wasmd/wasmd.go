package wasmd

import (
	"embed"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

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

var (
	// Embed abi json file to the executable binary. Needed when importing as dependency.
	//
	//go:embed abi.json
	f   embed.FS
	ABI abi.ABI
)

// Load ABI on init
func init() {
	var err error
	ABI, err = cmn.LoadABI(f, "abi.json")
	if err != nil {
		panic(err)
	}
}

// Precompile defines the struct for the wasmd precompile
type Precompile struct {
	cmn.Precompile

	abi.ABI
	wasmdKeeper wasmdkeeper.Keeper
}

// NewPrecompile starts a new wasmd precompile
func NewPrecompile(
	wasmdKeeper wasmdkeeper.Keeper,
) *Precompile {
	// Initialize the precompile
	return &Precompile{
		Precompile: cmn.Precompile{
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
			ContractAddress:      common.HexToAddress(WasmdPrecompileAddress),
		},
		ABI:         ABI,
		wasmdKeeper: wasmdKeeper,
	}
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
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readonly bool) ([]byte, error) {
	return p.RunNativeAction(evm, contract, func(ctx sdk.Context) ([]byte, error) {
		return p.Execute(ctx, evm.StateDB, evm.Origin, contract, readonly)
	})
}

func (p Precompile) Execute(ctx sdk.Context, stateDB vm.StateDB, caller common.Address, contract *vm.Contract, readOnly bool) ([]byte, error) {
	method, args, err := cmn.SetupABI(p.ABI, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	var bz []byte

	// Now we call the method based on the function
	switch method.Name {
	// Wasmd transactions
	case InstantiateMethod:
		bz, err = p.Instantiate(ctx, caller, contract, stateDB, method, args)
	case ExecuteMethod:
		bz, err = p.ExecuteWasm(ctx, caller, contract, stateDB, method, args)
	// Wasmd queries
	case QueryRawMethod:
		bz, err = p.QueryRaw(ctx, method, args)
	case QuerySmartMethod:
		bz, err = p.QuerySmart(ctx, method, args)
	default:
		// If default error out
		return nil, fmt.Errorf(cmn.ErrUnknownMethod, method.Name)
	}

	return bz, err
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
