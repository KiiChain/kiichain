package kiichain

import (
	erc20keeper "github.com/cosmos/evm/x/erc20/keeper"
	feemarketkeeper "github.com/cosmos/evm/x/feemarket/keeper"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v10/modules/core/keeper"

	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmempool "github.com/cosmos/cosmos-sdk/types/mempool"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	ibccallbackskeeper "github.com/cosmos/evm/x/ibc/callbacks/keeper"
	transferkeeper "github.com/cosmos/evm/x/ibc/transfer/keeper"
	precisebankkeeper "github.com/cosmos/evm/x/precisebank/keeper"
)

// GetStakingKeeper implements the TestingApp interface. Needed for ICS.
func (app *KiichainApp) GetStakingKeeper() *stakingkeeper.Keeper { //nolint:nolintlint
	return app.StakingKeeper
}

// GetIBCKeeper implements the TestingApp interface.
func (app *KiichainApp) GetIBCKeeper() *ibckeeper.Keeper { //nolint:nolintlint
	return app.IBCKeeper
}

func (app *KiichainApp) GetEVMKeeper() *evmkeeper.Keeper {
	return app.EVMKeeper
}

func (app *KiichainApp) GetErc20Keeper() *erc20keeper.Keeper {
	return &app.Erc20Keeper
}

func (app *KiichainApp) SetErc20Keeper(erc20Keeper erc20keeper.Keeper) {
	app.Erc20Keeper = erc20Keeper
}

func (app *KiichainApp) GetGovKeeper() govkeeper.Keeper {
	return *app.GovKeeper
}

func (app *KiichainApp) GetEvidenceKeeper() *evidencekeeper.Keeper {
	return &app.EvidenceKeeper
}

func (app *KiichainApp) GetSlashingKeeper() slashingkeeper.Keeper {
	return app.SlashingKeeper
}

func (app *KiichainApp) GetBankKeeper() bankkeeper.Keeper {
	return app.BankKeeper
}

func (app *KiichainApp) GetFeeMarketKeeper() *feemarketkeeper.Keeper {
	return &app.FeeMarketKeeper
}

func (app *KiichainApp) GetFeeGrantKeeper() feegrantkeeper.Keeper {
	return app.FeeGrantKeeper
}

func (app *KiichainApp) GetConsensusParamsKeeper() consensusparamkeeper.Keeper {
	return app.ConsensusParamsKeeper
}

func (app *KiichainApp) GetAccountKeeper() authkeeper.AccountKeeper {
	return app.AccountKeeper
}

func (app *KiichainApp) GetAuthzKeeper() authzkeeper.Keeper {
	return app.AuthzKeeper
}

func (app *KiichainApp) GetDistrKeeper() distrkeeper.Keeper {
	return app.DistrKeeper
}

func (app *KiichainApp) GetMintKeeper() mintkeeper.Keeper {
	return mintkeeper.Keeper{}
}

func (app *KiichainApp) GetPreciseBankKeeper() *precisebankkeeper.Keeper {
	return &precisebankkeeper.Keeper{}
}

func (app *KiichainApp) GetCallbackKeeper() ibccallbackskeeper.ContractKeeper {
	return ibccallbackskeeper.ContractKeeper{}
}

func (app *KiichainApp) GetTransferKeeper() transferkeeper.Keeper {
	return app.TransferKeeper
}

func (app *KiichainApp) SetTransferKeeper(transferKeeper transferkeeper.Keeper) {
	app.TransferKeeper = transferKeeper
}

func (app *KiichainApp) GetMempool() sdkmempool.ExtMempool {
	return app.EVMKeeper.GetEvmMempool()
}

func (app *KiichainApp) GetAnteHandler() sdk.AnteHandler {
	return app.BaseApp.AnteHandler()
}
