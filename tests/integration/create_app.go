package integration

import (
	"encoding/json"
	"os"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/evm"
	"github.com/cosmos/evm/testutil/config"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	ibctesting "github.com/cosmos/ibc-go/v10/testing"
	kiichain "github.com/kiichain/kiichain/v5/app"

	clienthelpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simutils "github.com/cosmos/cosmos-sdk/testutil/sims"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// CreateKiichain creates a kiichain app for regular integration tests (non-mempool)
// This version uses a noop mempool to avoid state issues during transaction processing
func CreateKiichain(chainID string, _ uint64, customBaseAppOptions ...func(*baseapp.BaseApp)) evm.EvmApp {
	appOptions := simutils.NewAppOptionsWithFlagHome(kiichain.DefaultNodeHome)

	baseAppOptions := append(customBaseAppOptions, baseapp.SetChainID(chainID))

	// Disable cache for integration tests to avoid state issues
	sdk.SetAddrCacheEnabled(false)

	// Create a temporary path
	dir, err := os.MkdirTemp("", "kiichain-integration-test")
	if err != nil {
		panic(err)
	}

	return kiichain.NewKiichainApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		dir,
		appOptions,
		kiichain.EmptyWasmOptions,
		kiichain.EVMAppOptions,
		baseAppOptions...,
	)
}

// SetupKiichain initializes a new kiichain app with default genesis state.
// It is used in IBC integration tests to create a new kiichain app instance.
func SetupKiichain() (ibctesting.TestingApp, map[string]json.RawMessage) {
	defaultNodeHome, err := clienthelpers.GetNodeHomeDirectory(".kiichain")
	if err != nil {
		panic(err)
	}

	app := kiichain.NewKiichainApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		defaultNodeHome,
		kiichain.EmptyAppOptions{},
		kiichain.EmptyWasmOptions,
		kiichain.EVMAppOptions,
	)
	// disable base fee for testing
	genesisState := app.ModuleBasics.DefaultGenesis(app.AppCodec())
	fmGen := feemarkettypes.DefaultGenesisState()
	fmGen.Params.NoBaseFee = true
	genesisState[feemarkettypes.ModuleName] = app.AppCodec().MustMarshalJSON(fmGen)
	stakingGen := stakingtypes.DefaultGenesisState()
	stakingGen.Params.BondDenom = config.ExampleChainDenom
	genesisState[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGen)
	mintGen := minttypes.DefaultGenesisState()
	mintGen.Params.MintDenom = config.ExampleChainDenom
	genesisState[minttypes.ModuleName] = app.AppCodec().MustMarshalJSON(mintGen)

	return app, genesisState
}
