package kiichain

import (
	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	evmmempool "github.com/cosmos/evm/mempool"
	"github.com/cosmos/evm/server"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

// configureEVMMempool sets up the Krakatoa mempool. See the Cosmos EVM
// v0.6.x to v0.7.0 migration guide, step 5a.
func (app *KiichainApp) configureEVMMempool(appOpts servertypes.AppOptions, logger log.Logger) error {
	if evmtypes.GetChainConfig() == nil {
		logger.Debug("evm chain config is not set, skipping mempool configuration")
		return nil
	}

	mpConfig := server.ResolveMempoolConfig(app.GetAnteHandler(), appOpts, logger)
	txEncoder := evmmempool.NewTxEncoder(app.txConfig)
	evmRechecker := evmmempool.NewTxRechecker(mpConfig.AnteHandler, txEncoder)
	cosmosRechecker := evmmempool.NewTxRechecker(mpConfig.AnteHandler, txEncoder)
	cosmosPoolMaxTx := server.GetCosmosPoolMaxTx(appOpts, logger)
	checkTxTimeout := server.GetMempoolCheckTxTimeout(appOpts, logger)

	if cosmosPoolMaxTx < 0 {
		logger.Debug("evm mempool is disabled, skipping configuration")
		return nil
	}

	if err := server.ValidateReapBounds(appOpts, mpConfig.BlockGasLimit); err != nil {
		return err
	}

	pool := evmmempool.NewMempool(
		app.CreateQueryContext,
		logger,
		app.EVMKeeper,
		app.FeeMarketKeeper,
		app.txConfig,
		evmRechecker,
		cosmosRechecker,
		mpConfig,
		cosmosPoolMaxTx,
	)

	app.EVMMempool = pool
	proposalHandler := baseapp.NewDefaultProposalHandler(pool, NewNoCheckProposalTxVerifier(app.BaseApp))
	app.SetPrepareProposal(proposalHandler.PrepareProposalHandler())
	app.SetProcessProposal(proposalHandler.ProcessProposalHandler())
	app.SetInsertTxHandler(pool.NewInsertTxHandler(app.TxDecode))
	app.SetReapTxsHandler(pool.NewReapTxsHandler())
	app.SetCheckTxHandler(pool.NewCheckTxHandler(app.TxDecode, checkTxTimeout))
	app.SetMempool(pool)
	app.SetPrepareCheckStater(func(_ sdk.Context) {
		if !pool.HasEventBus() {
			pool.NotifyNewBlock()
		}
	})

	return nil
}
