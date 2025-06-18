package kiichain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	db "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	kiichain "github.com/kiichain/kiichain/v2/app"
	kiihelpers "github.com/kiichain/kiichain/v2/app/helpers"
)

type EmptyAppOptions struct{}

var emptyWasmOption []wasmkeeper.Option

func (ao EmptyAppOptions) Get(_ string) interface{} {
	return nil
}

func TestKiichainApp_BlockedModuleAccountAddrs(t *testing.T) {
	app := kiichain.NewKiichainApp(
		log.NewNopLogger(),
		db.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		kiichain.DefaultNodeHome,
		EmptyAppOptions{},
		emptyWasmOption,
		kiichain.NoOpEVMOptions,
	)

	moduleAccountAddresses := app.ModuleAccountAddrs()
	blockedAddrs := app.BlockedModuleAccountAddrs(moduleAccountAddresses)

	require.NotContains(t, blockedAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())
}

func TestKiichainApp_Export(t *testing.T) {
	app := kiihelpers.Setup(t)
	_, err := app.ExportAppStateAndValidators(true, []string{}, []string{})
	require.NoError(t, err, "ExportAppStateAndValidators should not have an error")
}
