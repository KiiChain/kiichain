package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	kiievmante "github.com/kiichain/kiichain/v5/x/feeabstraction/ante/evm"
)

// newMonoEVMAnteHandler creates the sdk.AnteHandler implementation for the EVM transactions
func newMonoEVMAnteHandler(ctx sdk.Context, options HandlerOptions) sdk.AnteHandler {
	evmParams := options.EvmKeeper.GetParams(ctx)
	feemarketParams := options.FeeMarketKeeper.GetParams(ctx)
	return sdk.ChainAnteDecorators(
		kiievmante.NewEVMMonoDecorator(
			options.AccountKeeper,
			options.FeeMarketKeeper,
			options.EvmKeeper,
			options.FeeAbstractionKeeper,
			options.MaxTxGasWanted,
			&evmParams,
			&feemarketParams,
		),
		NewTxListenerDecorator(options.PendingTxListener),
	)
}
