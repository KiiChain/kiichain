package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	txlistener "github.com/cosmos/evm/ante"

	kiievmante "github.com/kiichain/kiichain/v7/x/feeabstraction/ante/evm"
)

// newMonoEVMAnteHandler creates the sdk.AnteHandler implementation for the EVM transactions
func newMonoEVMAnteHandler(ctx sdk.Context, options HandlerOptions) sdk.AnteHandler {
	evmParams := options.EvmKeeper.GetParams(ctx)
	feemarketParams := options.FeeMarketKeeper.GetParams(ctx)
	decorators := []sdk.AnteDecorator{
		NewBlockedAddrDecorator(options.Cdc), // reject incident addrs before EVM execution
		kiievmante.NewEVMMonoDecorator(
			options.AccountKeeper,
			options.FeeMarketKeeper,
			options.EvmKeeper,
			options.FeeAbstractionKeeper,
			options.MaxTxGasWanted,
			&evmParams,
			&feemarketParams,
		),
		txlistener.NewTxListenerDecorator(options.PendingTxListener),
	}

	return sdk.ChainAnteDecorators(decorators...)
}
