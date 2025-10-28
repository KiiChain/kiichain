package oracle

import (
	"embed"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cmn "github.com/cosmos/evm/precompiles/common"

	oraclekeeper "github.com/kiichain/kiichain/v5/x/oracle/keeper"
)

const (
	// OraclePrecompileAddress is the address of the oracle precompile
	OraclePrecompileAddress = "0x0000000000000000000000000000000000001003"
)

// Precompile implements the PrecompiledContract interface
var _ vm.PrecompiledContract = &Precompile{}

var (
	// Embed abi json file to the executable binary. Needed when importing as dependency.
	//
	//go:embed abi.json
	f   embed.FS
	ABI abi.ABI
)

// inits the abi for the precompile
func init() {
	var err error
	ABI, err = cmn.LoadABI(f, "abi.json")
	if err != nil {
		panic(err)
	}
}

// Precompile defines the struct for the oracle precompile
type Precompile struct {
	cmn.Precompile

	abi.ABI
	oracleKeeper oraclekeeper.Keeper
}

// NewPrecompile creates a new oracle precompile instance
func NewPrecompile(
	oracleKeeper oraclekeeper.Keeper,
) *Precompile {
	// Initialize the precompile
	return &Precompile{
		Precompile: cmn.Precompile{
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
			ContractAddress:      common.HexToAddress(OraclePrecompileAddress),
		},
		ABI:          ABI,
		oracleKeeper: oracleKeeper,
	}
}

// RequiredGas returns the required gas for the precompile
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

// Run executes the oracle precompile
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readonly bool) ([]byte, error) {
	return p.RunNativeAction(evm, contract, func(ctx sdk.Context) ([]byte, error) {
		return p.Execute(ctx, evm.StateDB, contract, readonly)
	})
}

// Execute redirects into the oracle precompile
func (p Precompile) Execute(ctx sdk.Context, stateDB vm.StateDB, contract *vm.Contract, readOnly bool) ([]byte, error) {
	method, args, err := cmn.SetupABI(p.ABI, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	var bz []byte

	// Now we call the method on the oracle keeper
	switch method.Name {
	case GetExchangeRateMethod:
		bz, err = p.GetExchangeRate(ctx, method, args)
	case GetExchangeRatesMethod:
		bz, err = p.GetExchangeRates(ctx, method, args)
	case GetTwapsMethod:
		bz, err = p.GetTwaps(ctx, method, args)
	default:
		// If default error out
		return nil, fmt.Errorf(cmn.ErrUnknownMethod, method.Name)
	}

	return bz, err
}

// IsTransaction checks if the method is a transaction
func (Precompile) IsTransaction(method *abi.Method) bool {
	// We don't have transactions
	return false
}

// Logger returns the logger for the precompile
func (p Precompile) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("evm extension", "oracle")
}
