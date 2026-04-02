package kiichain_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"

	db "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	kiichain "github.com/kiichain/kiichain/v7/app"
	kiihelpers "github.com/kiichain/kiichain/v7/app/helpers"
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

func TestInitChainerReturnsErrorOnInvalidJSON(t *testing.T) {
	app := kiichain.NewKiichainApp(
		log.NewNopLogger(),
		db.NewMemDB(),
		nil,
		true,
		map[int64]bool{},
		kiichain.DefaultNodeHome,
		EmptyAppOptions{},
		emptyWasmOption,
	)

	ctx := app.NewContext(true)

	var initErr error
	require.NotPanics(t, func() {
		_, initErr = app.InitChainer(ctx, &abci.RequestInitChain{
			AppStateBytes: []byte("invalid json"),
		})
	})
	require.Error(t, initErr, "InitChainer should return an error for invalid JSON")
	require.Contains(t, initErr.Error(), "failed to unmarshal genesis state")
}
