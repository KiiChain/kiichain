package integration

import (
	"encoding/json"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/evm"
	"github.com/cosmos/evm/testutil/config"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	ibctesting "github.com/cosmos/ibc-go/v10/testing"
	kiichain "github.com/kiichain/kiichain/v5/app"

	clienthelpers "cosmossdk.io/client/v2/helpers"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simutils "github.com/cosmos/cosmos-sdk/testutil/sims"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// CreateKiichain creates a kiichain app for regular integration tests (non-mempool)
// This version uses a noop mempool to avoid state issues during transaction processing
func CreateKiichain(chainID string, customBaseAppOptions ...func(*baseapp.BaseApp)) evm.EvmApp {
	defaultNodeHome, err := clienthelpers.GetNodeHomeDirectory(".kiichain")
	if err != nil {
		panic(err)
	}

	appOptions := simutils.NewAppOptionsWithFlagHome(defaultNodeHome)

	baseAppOptions := append(customBaseAppOptions, baseapp.SetChainID(chainID))

	return kiichain.NewKiichainApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		defaultNodeHome,
		appOptions,
		kiichain.EmptyWasmOptions,
		kiichain.EVMAppOptions,
		baseAppOptions...,
	)
}

// SetupKiichain initializes a new kiichain app with default genesis state.
// It is used in IBC integration tests to create a new evmd app instance.
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
