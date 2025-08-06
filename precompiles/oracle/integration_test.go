package oracle_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	testkeyring "github.com/cosmos/evm/testutil/integration/os/keyring"

	app "github.com/kiichain/kiichain/v4/app"
	"github.com/kiichain/kiichain/v4/app/helpers"
	oracleprecompile "github.com/kiichain/kiichain/v4/precompiles/oracle"
)

// OraclePrecompileTestSuite is a test suite for the oracle precompile
type OraclePrecompileTestSuite struct {
	suite.Suite

	// App and context
	App     *app.KiichainApp
	Ctx     sdk.Context
	keyring testkeyring.Keyring

	// Precompile
	Precompile *oracleprecompile.Precompile
}

// TestOraclePrecompileTestSuite runs all the tests under the oracle pre-compile test suite
func TestOraclePrecompileTestSuite(t *testing.T) {
	suite.Run(t, new(OraclePrecompileTestSuite))
}

// SetupSuite sets up the test suite
func (s *OraclePrecompileTestSuite) SetupSuite() {
	// Get the test context
	t := s.T()

	// Create the app and the context
	s.App = helpers.Setup(t)
	s.Ctx = s.App.BaseApp.NewUncachedContext(true, tmtypes.Header{Height: 1, ChainID: "test_1010-1", Time: time.Now().UTC()})

	// Start a new keyring
	keyring := testkeyring.New(2)
	s.keyring = keyring

	// Start the precompile
	pc, err := oracleprecompile.NewPrecompile(s.App.OracleKeeper, s.App.AuthzKeeper)
	s.Require().NoError(err)
	s.Precompile = pc
}
