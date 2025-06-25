package v300_test

import (
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	tmtypes "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/kiichain/kiichain/v2/app/helpers"
	utils "github.com/kiichain/kiichain/v2/app/upgrades/utils"
	"github.com/kiichain/kiichain/v2/precompiles/oracle"
)

// TestUpgrade tests the upgrade handler for v3.0.0
func TestUpgrade(t *testing.T) {
	// Create the app and the context
	app := helpers.Setup(t)
	ctx := app.BaseApp.NewUncachedContext(true, tmtypes.Header{Height: 1, ChainID: "test_1010-1", Time: time.Now().UTC()})

	// Create a pre-populated list of pre-compiles
	precompiles := []string{
		"0x0000000000000000000000000000000000000001",
		"0x0000000000000000000000000000000000000002",
	}

	// Install the precompiles
	evmParams := app.EVMKeeper.GetParams(ctx)
	evmParams.ActiveStaticPrecompiles = precompiles
	err := app.EVMKeeper.SetParams(ctx, evmParams)
	require.NoError(t, err)

	// Now run add wasmd upgrade
	err = utils.InstallNewPrecompiles(
		ctx,
		&app.AppKeepers,
		[]common.Address{
			common.HexToAddress(oracle.OraclePrecompileAddress),
		},
	)
	require.NoError(t, err)

	// Get the params again
	evmParams = app.EVMKeeper.GetParams(ctx)

	// Check that the precompiles was added
	require.Len(t, evmParams.ActiveStaticPrecompiles, 3)
	require.Contains(t, evmParams.ActiveStaticPrecompiles, "0x0000000000000000000000000000000000000001")
	require.Contains(t, evmParams.ActiveStaticPrecompiles, "0x0000000000000000000000000000000000000002")
	require.Contains(t, evmParams.ActiveStaticPrecompiles, oracle.OraclePrecompileAddress)
}
