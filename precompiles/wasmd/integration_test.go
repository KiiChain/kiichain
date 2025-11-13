package wasmd_test

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/stretchr/testify/suite"

	_ "embed"

	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmdkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	testkeyring "github.com/cosmos/evm/testutil/keyring"
	"github.com/cosmos/evm/x/vm/statedb"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	app "github.com/kiichain/kiichain/v5/app"
	helpers "github.com/kiichain/kiichain/v5/app/helpers"
	wasmdprecompile "github.com/kiichain/kiichain/v5/precompiles/wasmd"
)

// CounterWasmCode is the bytecode of the counter smart contract
// Schema can be found at: precompiles/wasmd/testdata/counter_schema.json
//
//go:embed testdata/counter.wasm
var CounterWasmCode []byte

// WasmdPrecompileTestSuite is a test suite for the wasmd precompile
type WasmdPrecompileTestSuite struct {
	suite.Suite

	// App and context
	App     *app.KiichainApp
	Ctx     sdk.Context
	keyring testkeyring.Keyring

	// Evm
	stateDB *statedb.StateDB

	// Precompile
	Precompile *wasmdprecompile.Precompile

	// Contract for testing
	WasmdMsgServer wasmtypes.MsgServer
	CounterCodeID  uint64
}

// TestWasmdPrecompileTestSuite runs all the tests under the wasmd pre-compile test suite
func TestWasmdPrecompileTestSuite(t *testing.T) {
	suite.Run(t, new(WasmdPrecompileTestSuite))
}

// SetupTest sets up the test case
func (s *WasmdPrecompileTestSuite) SetupTest() {
	// Get the test context
	t := s.T()

	// Create the app and the context
	s.App = helpers.Setup(t)
	s.Ctx = s.App.BaseApp.NewUncachedContext(true, tmtypes.Header{Height: 1, ChainID: "test_1010-1", Time: time.Now().UTC()})

	// Start a new keyring
	keyring := testkeyring.New(2)
	s.keyring = keyring

	// Store a counter smart contract
	s.WasmdMsgServer = wasmdkeeper.NewMsgServerImpl(&s.App.WasmKeeper)
	res, err := s.WasmdMsgServer.StoreCode(s.Ctx, &wasmtypes.MsgStoreCode{
		Sender:       sdk.AccAddress([]byte("wasm")).String(),
		WASMByteCode: CounterWasmCode,
		InstantiatePermission: &wasmtypes.AccessConfig{
			Permission: wasmtypes.AccessTypeEverybody,
		},
	})
	s.Require().NoError(err)
	s.CounterCodeID = res.CodeID

	// Start the precompile
	pc, err := wasmdprecompile.NewPrecompile(s.App.WasmKeeper)
	s.Require().NoError(err)
	s.Precompile = pc

	// Start the state DB
	// Get the header hash
	headerHash := s.Ctx.HeaderHash()

	// Return the statedb
	stateDB := statedb.New(
		s.Ctx,
		s.App.EVMKeeper,
		statedb.NewEmptyTxConfig(common.BytesToHash(headerHash)),
	)

	s.stateDB = stateDB
}

// NewVmInstance creates a new EVM instance for the test suite
func (s *WasmdPrecompileTestSuite) NewVMInstance(ctx sdk.Context) *vm.EVM {
	return vm.NewEVM(
		vm.BlockContext{},
		s.stateDB,
		evmtypes.GetEthChainConfig(),
		vm.Config{},
	)
}

// GetStateDB returns the state database for the precompile
func (s *WasmdPrecompileTestSuite) GetStateDB() *statedb.StateDB {
	return s.stateDB
}

// PrepareInputData prepares the input data for the precompile method call
func (s *WasmdPrecompileTestSuite) PrepareInputData(methodName string, args []interface{}) []byte {
	s.T().Helper()

	// Get the method from the precompile ABI
	_, exists := s.Precompile.ABI.Methods[methodName]
	s.Require().True(exists, "method %s not found in precompile ABI", methodName)

	// Pack the arguments
	inputData, err := s.Precompile.Pack(methodName, args...)
	s.Require().NoError(err, "failed to pack arguments for method %s", methodName)

	return inputData
}
