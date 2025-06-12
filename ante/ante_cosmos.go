package ante

import (
	ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	evmcosmosante "github.com/cosmos/evm/ante/cosmos"
	evmante "github.com/cosmos/evm/ante/evm"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

// UseFeeMarketDecorator to make the integration testing easier: we can switch off its ante and post decorators with this flag
var UseFeeMarketDecorator = true

// newCosmosAnteHandler creates the default ante handler for Cosmos transactions
func NewCosmosAnteHandler(options HandlerOptions) sdk.AnteHandler {
	anteDecorators := []sdk.AnteDecorator{
		evmcosmosante.NewRejectMessagesDecorator(), // reject MsgEthereumTxs
		evmcosmosante.NewAuthzLimiterDecorator( // disable the Msg types that cannot be included on an authz.MsgExec msgs field
			sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{}),
			sdk.MsgTypeURL(&sdkvesting.MsgCreateVestingAccount{}),
		),

		ante.NewSetUpContextDecorator(),
		wasmkeeper.NewLimitSimulationGasDecorator(options.WasmConfig.SimulationGasLimit), // after setup context to enforce limits early
		wasmkeeper.NewCountTXDecorator(options.TXCounterStoreService),
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		NewGovVoteDecorator(options.Cdc, options.StakingKeeper),
		NewGovExpeditedProposalsDecorator(options.Cdc),
		evmcosmosante.NewMinGasPriceDecorator(options.FeeMarketKeeper, options.EvmKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
		ibcante.NewRedundantRelayDecorator(options.IBCKeeper),
		evmante.NewGasWantedDecorator(options.EvmKeeper, options.FeeMarketKeeper),
	}

	// Skip the feemarket decorator is needed
	if UseFeeMarketDecorator {
		// This wraps using the gasless decorator
		gasLessFeeDecorator := NewGaslessDecorator(
			ante.NewDeductFeeDecorator(
				options.AccountKeeper,
				options.BankKeeper,
				options.FeegrantKeeper,
				options.TxFeeChecker,
			),
			options.OracleKeeper,
		)

		// Add to the ante decorators
		anteDecorators = append(anteDecorators, gasLessFeeDecorator)
	}

	return sdk.ChainAnteDecorators(anteDecorators...)
}
