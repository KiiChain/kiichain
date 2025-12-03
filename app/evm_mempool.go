package kiichain

import (
	"path/filepath"

	"github.com/spf13/cast"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	evmmempool "github.com/cosmos/evm/mempool"
)

// SetupEVMMempool creates and sets the EVM mempool
func (app *KiichainApp) SetupEVMMempool(appOpts servertypes.AppOptions, logger log.Logger) {
	mempoolConfig := &evmmempool.EVMMempoolConfig{
		AnteHandler:   app.GetAnteHandler(),
		BlockGasLimit: GetBlockGasLimit(appOpts, logger),
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

// GetBlockGasLimit reads the genesis json file using AppGenesisFromFile
// to extract the consensus block gas limit before InitChain is called.
func GetBlockGasLimit(appOpts servertypes.AppOptions, logger log.Logger) uint64 {
	if appOpts == nil {
		logger.Error("app options is nil, using zero block gas limit")
		return 0
	}

	homeDir := cast.ToString(appOpts.Get(flags.FlagHome))
	if homeDir == "" {
		logger.Error("home directory not found in app options, using zero block gas limit")
		return 0
	}
	genesisPath := filepath.Join(homeDir, "config", "genesis.json")

	appGenesis, err := genutiltypes.AppGenesisFromFile(genesisPath)
	if err != nil {
		logger.Error("failed to load genesis using SDK AppGenesisFromFile, using zero block gas limit", "path", genesisPath, "error", err)
		return 0
	}
	genDoc, err := appGenesis.ToGenesisDoc()
	if err != nil {
		logger.Error("failed to convert AppGenesis to GenesisDoc, using zero block gas limit", "path", genesisPath, "error", err)
		return 0
	}

	if genDoc.ConsensusParams == nil {
		logger.Error("consensus parameters not found in genesis (nil), using zero block gas limit")
		return 0
	}

	maxGas := genDoc.ConsensusParams.Block.MaxGas
	if maxGas == -1 {
		logger.Warn("genesis max_gas is unlimited (-1), using max uint64")
		return 0
	}
	if maxGas < -1 {
		logger.Error("invalid max_gas value in genesis, using zero block gas limit")
		return 0
	}
	blockGasLimit := uint64(maxGas) // #nosec G115 -- maxGas >= 0 checked above

	logger.Debug(
		"extracted block gas limit from genesis using SDK AppGenesisFromFile",
		"genesis_path", genesisPath,
		"max_gas", maxGas,
		"block_gas_limit", blockGasLimit,
	)

	return blockGasLimit
}
