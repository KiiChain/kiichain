package integration

import (
	"os"

	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simutils "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/evm"

	kiichain "github.com/kiichain/kiichain/v5/app"
)

// CreateKiichain creates a kiichain app for regular integration tests (non-mempool)
// This version uses a noop mempool to avoid state issues during transaction processing
func CreateKiichain(chainID string, evmChaindID uint64, customBaseAppOptions ...func(*baseapp.BaseApp)) evm.EvmApp {
	appOptions := simutils.NewAppOptionsWithFlagHome(kiichain.DefaultNodeHome)

	customBaseAppOptions = append(customBaseAppOptions, baseapp.SetChainID(chainID))

	// Disable cache for integration tests to avoid state issues
	sdk.SetAddrCacheEnabled(false)

	// Set kiichain evm chain ID
	kiichain.KiichainID = evmChaindID

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
		customBaseAppOptions...,
	)
}
