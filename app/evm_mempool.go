package kiichain

import (
	"fmt"

	"cosmossdk.io/log"

	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"

	evmmempool "github.com/cosmos/evm/mempool"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

// SetupEVMMempool creates and sets the EVM mempool
func (app *KiichainApp) SetupEVMMempool(appOpts servertypes.AppOptions, logger log.Logger) {
	mempoolConfig := &evmmempool.EVMMempoolConfig{
		AnteHandler:   app.GetAnteHandler(),
		BroadCastTxFn: app.broadcastEVMTransactions,
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

// broadcastEVMTransactions converts Ethereum transactions to Cosmos SDK format and broadcasts them.
// This function wraps EVM transactions in MsgEthereumTx messages and submits them to the network
// using the provided client context. It handles encoding and error reporting for each transaction.
// TODO: Remove once EVM pushes a fix
func (app *KiichainApp) broadcastEVMTransactions(ethTxs []*ethtypes.Transaction) error {
	for _, ethTx := range ethTxs {
		msg := &evmtypes.MsgEthereumTx{}
		msg.FromEthereumTx(ethTx)

		txBuilder := app.txConfig.NewTxBuilder()
		if err := txBuilder.SetMsgs(msg); err != nil {
			return fmt.Errorf("failed to set msg in tx builder: %w", err)
		}

		txBytes, err := app.txConfig.TxEncoder()(txBuilder.GetTx())
		if err != nil {
			return fmt.Errorf("failed to encode transaction: %w", err)
		}

		res, err := app.clientCtx.BroadcastTxSync(txBytes)
		if err != nil {
			return fmt.Errorf("failed to broadcast transaction %s: %w", ethTx.Hash().Hex(), err)
		}
		if res.Code != 0 {
			return fmt.Errorf("transaction %s rejected by mempool: code=%d, log=%s", ethTx.Hash().Hex(), res.Code, res.RawLog)
		}
	}
	return nil
}
