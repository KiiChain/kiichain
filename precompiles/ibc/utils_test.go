package ibc_test

import (
	"encoding/json"
	"fmt"
	"os"

	dbm "github.com/cosmos/cosmos-db"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	"cosmossdk.io/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"

	"github.com/kiichain/kiichain/v4/ante"
	kiichain "github.com/kiichain/kiichain/v4/app"
	"github.com/kiichain/kiichain/v4/app/params"
)

var app *kiichain.KiichainApp

// Some tests require a random directory to be created when running IBC testing suite with kiichain.
// This is due to how CosmWasmVM initializes the VM - all IBC testing apps must have different dirs so they don't conflict.
func KiichainAppIniterTempDir() (ibctesting.TestingApp, map[string]json.RawMessage) {
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}

	// Set the base options
	baseAppOptions := bam.SetChainID(
		fmt.Sprintf("%s-1", params.LocalChainID),
	)

	// Disable the fee market
	ante.UseFeeMarketDecorator = false

	// Initialize the app
	app = kiichain.NewKiichainApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		tmpDir,
		kiichain.EmptyAppOptions{},
		kiichain.EmptyWasmOptions,
		kiichain.EVMAppOptions,
		baseAppOptions,
	)

	testApp := ibctesting.TestingApp(app)

	return testApp, app.ModuleBasics.DefaultGenesis(app.AppCodec())
}

// KiichainAppIniter implements ibctesting.AppIniter for the kiichain app
func KiichainAppIniter() (ibctesting.TestingApp, map[string]json.RawMessage) {
	app = kiichain.NewKiichainApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		kiichain.DefaultNodeHome,
		kiichain.EmptyAppOptions{},
		kiichain.EmptyWasmOptions,
		kiichain.NoOpEVMOptions)

	testApp := ibctesting.TestingApp(app)

	return testApp, app.ModuleBasics.DefaultGenesis(app.AppCodec())
}

func GetApp(chain *ibctesting.TestChain) *kiichain.KiichainApp {
	return chain.App.(*kiichain.KiichainApp)
}
