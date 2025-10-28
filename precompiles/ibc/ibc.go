package ibc

import (
	"embed"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	clientkeeper "github.com/cosmos/ibc-go/v10/modules/core/02-client/keeper"
	connectionkeeper "github.com/cosmos/ibc-go/v10/modules/core/03-connection/keeper"
	channelkeeper "github.com/cosmos/ibc-go/v10/modules/core/04-channel/keeper"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cmn "github.com/cosmos/evm/precompiles/common"
	ibctransferkeeper "github.com/cosmos/evm/x/ibc/transfer/keeper"
)

const (
	TransferMethod                   = "transfer"
	TransferWithDefaultTimeoutMethod = "transferWithDefaultTimeout"
)

const (
	IBCPrecompileAddress = "0x0000000000000000000000000000000000001002"
)

var (
	// Embed abi json file to the executable binary. Needed when importing as dependency.
	//
	//go:embed abi.json
	f   embed.FS
	ABI abi.ABI
)

func init() {
	var err error
	ABI, err = cmn.LoadABI(f, "abi.json")
	if err != nil {
		panic(err)
	}
}

// Precompile is a struct that implements the PrecompiledContract interface
var _ vm.PrecompiledContract = &Precompile{}

// Precompile defines the struct for the ibc precompile
type Precompile struct {
	cmn.Precompile

	abi.ABI
	transferKeeper   ibctransferkeeper.Keeper
	clientKeeper     clientkeeper.Keeper
	connectionKeeper connectionkeeper.Keeper
	channelKeeper    channelkeeper.Keeper
}

// NewPrecompile defines creates a new instance of ibc precompile
func NewPrecompile(
	transferKeeper ibctransferkeeper.Keeper,
	clientKeeper clientkeeper.Keeper,
	connectionKeeper connectionkeeper.Keeper,
	channelKeeper channelkeeper.Keeper,
) *Precompile {
	// Setup keepers
	return &Precompile{
		Precompile: cmn.Precompile{
			KvGasConfig:          storetypes.KVGasConfig(),
			TransientKVGasConfig: storetypes.TransientGasConfig(),
			ContractAddress:      common.HexToAddress(IBCPrecompileAddress),
		},
		ABI:              ABI,
		transferKeeper:   transferKeeper,
		clientKeeper:     clientKeeper,
		connectionKeeper: connectionKeeper,
		channelKeeper:    channelKeeper,
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

// Run executes the ibc precompile
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
	case TransferMethod:
		bz, err = p.Transfer(ctx, method, stateDB, args, caller)
	case TransferWithDefaultTimeoutMethod:
		bz, err = p.TransferWithDefaultTimeout(ctx, method, stateDB, args, caller)
	default:
		// If default error out
		return nil, fmt.Errorf(cmn.ErrUnknownMethod, method.Name)
	}

	return bz, err
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
