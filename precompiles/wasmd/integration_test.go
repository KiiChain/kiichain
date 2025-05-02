package wasmd_test

import (
	_ "embed"
	"testing"
	"time"

	wasmdkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	app "github.com/kiichain/kiichain/v1/app"
	helpers "github.com/kiichain/kiichain/v1/app/helpers"
	wasmdprecompile "github.com/kiichain/kiichain/v1/precompiles/wasmd"
	"github.com/stretchr/testify/suite"
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
	App *app.KiichainApp
	Ctx sdk.Context

	// Precompile
	Precompile *wasmdprecompile.Precompile

	// Contract for testing
	WasmdMsgServer wasmtypes.MsgServer
	CounterCodeID  uint64
}

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
