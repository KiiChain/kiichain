package oracle

import (
	"embed"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"

	cmn "github.com/cosmos/evm/precompiles/common"
	"github.com/cosmos/evm/x/vm/core/vm"

	oraclekeeper "github.com/kiichain/kiichain/v4/x/oracle/keeper"
)

const (
	// OraclePrecompileAddress is the address of the oracle precompile
	OraclePrecompileAddress = "0x0000000000000000000000000000000000001003"
)

// Precompile implements the PrecompiledContract interface
var _ vm.PrecompiledContract = &Precompile{}

// Embed the json abi to the binary
//
//go:embed abi.json
var f embed.FS

// Precompile defines the struct for the oracle precompile
type Precompile struct {
	cmn.Precompile
	oracleKeeper oraclekeeper.Keeper
}

// LoadABI loads the ABI from the embedded file for the oracle precompile
func LoadABI() (abi.ABI, error) {
	return cmn.LoadABI(f, "abi.json")
}

// NewPrecompile creates a new oracle precompile instance
func NewPrecompile(
	oracleKeeper oraclekeeper.Keeper,
	authzKeeper authzkeeper.Keeper,
) (*Precompile, error) {
	// Load the ABI
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
		oracleKeeper: oracleKeeper,
	}

	// Set the address of the precompile
	precompile.SetAddress(common.HexToAddress(OraclePrecompileAddress))

	// Return the precompile
	return precompile, nil
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
func (p Precompile) Run(evm *vm.EVM, contract *vm.Contract, readOnly bool) (bz []byte, err error) {
	// Initialize the context, db and chain data
	ctx, statedb, snapshot, method, initialGas, args, err := p.RunSetup(evm, contract, readOnly, p.IsTransaction)
	if err != nil {
		return nil, err
	}

	// This handles any out of gas errors
	defer cmn.HandleGasError(ctx, contract, initialGas, &err)()

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
	if err != nil {
		return nil, err
	}

	// Check the gas cost
	cost := ctx.GasMeter().GasConsumed() - initialGas
	if !contract.UseGas(cost) {
		return nil, vm.ErrOutOfGas
	}

	// Add the new journal entry to the stateDB
	if err := p.AddJournalEntries(statedb, snapshot); err != nil {
		return nil, err
	}

	return bz, nil
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
