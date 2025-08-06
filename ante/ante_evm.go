package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	kiievmante "github.com/kiichain/kiichain/v4/x/feeabstraction/ante/evm"
)

// newMonoEVMAnteHandler creates the sdk.AnteHandler implementation for the EVM transactions
func newMonoEVMAnteHandler(options HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		kiievmante.NewEVMMonoDecorator(
			options.AccountKeeper,
			options.FeeMarketKeeper,
			options.EvmKeeper,
			options.FeeAbstractionKeeper,
			options.MaxTxGasWanted,
		),
	)
}
