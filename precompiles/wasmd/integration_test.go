package wasmd_test

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"

	_ "embed"

	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmdkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	testkeyring "github.com/cosmos/evm/testutil/integration/os/keyring"
	"github.com/cosmos/evm/x/vm/statedb"

	app "github.com/kiichain/kiichain/v2/app"
	helpers "github.com/kiichain/kiichain/v2/app/helpers"
	wasmdprecompile "github.com/kiichain/kiichain/v2/precompiles/wasmd"
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

// SetupSuite sets up the test suite
func (s *WasmdPrecompileTestSuite) SetupSuite() {
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
	pc, err := wasmdprecompile.NewPrecompile(s.App.WasmKeeper, s.App.AuthzKeeper)
	s.Require().NoError(err)
	s.Precompile = pc
}

// GetStateDB returns the state database for the precompile
func (s *WasmdPrecompileTestSuite) GetStateDB() *statedb.StateDB {
	// Get the header hash
	headerHash := s.Ctx.HeaderHash()

	// Return the statedb
	return statedb.New(
		s.Ctx,
		s.App.EVMKeeper,
		statedb.NewEmptyTxConfig(common.BytesToHash(headerHash)),
	)
}
