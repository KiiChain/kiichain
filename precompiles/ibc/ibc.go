package ibc

import (
	"embed"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	clientkeeper "github.com/cosmos/ibc-go/v8/modules/core/02-client/keeper"
	connectionkeeper "github.com/cosmos/ibc-go/v8/modules/core/03-connection/keeper"
	channelkeeper "github.com/cosmos/ibc-go/v8/modules/core/04-channel/keeper"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"

	cmn "github.com/cosmos/evm/precompiles/common"
	ibctransferkeeper "github.com/cosmos/evm/x/ibc/transfer/keeper"
	"github.com/cosmos/evm/x/vm/core/vm"
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

// Precompile is a struct that implements the PrecompiledContract interface
var _ vm.PrecompiledContract = &Precompile{}

type Precompile struct {
	cmn.Precompile
	transferKeeper   ibctransferkeeper.Keeper
	clientKeeper     clientkeeper.Keeper
	connectionKeeper connectionkeeper.Keeper
	channelKeeper    channelkeeper.Keeper

	TransferID                   []byte
	TransferWithDefaultTimeoutID []byte
}

func NewPrecompile(
	transferKeeper ibctransferkeeper.Keeper,
	clientKeeper clientkeeper.Keeper,
	connectionKeeper connectionkeeper.Keeper,
	channelKeeper channelkeeper.Keeper,
	authzKeeper authzkeeper.Keeper,
) (*Precompile, error) {
	// Load abi
	abi, err := cmn.LoadABI(f, "abi.json")
	if err != nil {
		return nil, err
	}

	// Setup keepers
	p := &Precompile{
		Precompile: cmn.Precompile{
			ABI:                  abi,
			AuthzKeeper:          authzKeeper,
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
			ApprovalExpiration:   cmn.DefaultExpirationDuration,
		},
		transferKeeper:   transferKeeper,
		clientKeeper:     clientKeeper,
		connectionKeeper: connectionKeeper,
		channelKeeper:    channelKeeper,
	}

	// Set method IDs
	for name, m := range abi.Methods {
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

// Run executes the ibc precompile
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
		bz, err = p.Transfer(ctx, method, stateDB, args, evm.Origin)
	case TransferWithDefaultTimeoutMethod:
		bz, err = p.TransferWithDefaultTimeout(ctx, method, stateDB, args, evm.Origin)
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
func (Precompile) IsTransaction(method *abi.Method) bool {
	switch method.Name {
	case TransferMethod, TransferWithDefaultTimeoutMethod:
		return true
	default:
		return false
	}
}

// Logger returns the logger for the precompile
func (p Precompile) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("evm extension", "ibc")
}
