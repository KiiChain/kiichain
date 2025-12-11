package kiichain

import (
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"

	evmmempool "github.com/cosmos/evm/mempool"
)

// SetupEVMMempool creates and sets the EVM mempool
func (app *KiichainApp) SetupEVMMempool(appOpts servertypes.AppOptions, logger log.Logger) {
	mempoolConfig := &evmmempool.EVMMempoolConfig{
		AnteHandler:   app.GetAnteHandler(),
		BlockGasLimit: 100_000_000,
	}

	evmMempool := evmmempool.NewExperimentalEVMMempool(app.CreateQueryContext, logger, app.EVMKeeper, app.FeeMarketKeeper, app.txConfig, app.clientCtx, mempoolConfig)
	app.EVMMempool = evmMempool

	// Set the global mempool for RPC access
	if evmmempool.GetGlobalEVMMempool() == nil {
		if err := evmmempool.SetGlobalEVMMempool(evmMempool); err != nil {
			panic(err)
		}
	}
	app.SetMempool(evmMempool)
	checkTxHandler := evmmempool.NewCheckTxHandler(evmMempool)
	app.SetCheckTxHandler(checkTxHandler)

	abciProposalHandler := baseapp.NewDefaultProposalHandler(evmMempool, app)
	abciProposalHandler.SetSignerExtractionAdapter(evmmempool.NewEthSignerExtractionAdapter(sdkmempool.NewDefaultSignerExtractionAdapter()))
	app.SetPrepareProposal(abciProposalHandler.PrepareProposalHandler())
}
