package antedecorators

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkacltypes "github.com/cosmos/cosmos-sdk/types/accesscontrol"
	evmkeeper "github.com/kiichain/kiichain3/x/evm/keeper"
	evmtypes "github.com/kiichain/kiichain3/x/evm/types"
)

type GaslessDecorator struct {
	wrapped []sdk.AnteFullDecorator
	// oracleKeeper oraclekeeper.Keeper
	evmKeeper *evmkeeper.Keeper
}

func NewGaslessDecorator(wrapped []sdk.AnteFullDecorator, evmKeeper *evmkeeper.Keeper) GaslessDecorator {
	return GaslessDecorator{wrapped: wrapped, evmKeeper: evmKeeper}
}

func (gd GaslessDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	originalGasMeter := ctx.GasMeter()
	// eagerly set infinite gas meter so that queries performed by IsTxGasless will not incur gas cost
	ctx = ctx.WithGasMeter(storetypes.NewNoConsumptionInfiniteGasMeter())

	isGasless, err := IsTxGasless(tx, ctx, gd.evmKeeper)
	if err != nil {
		return ctx, err
	}
	if !isGasless {
		ctx = ctx.WithGasMeter(originalGasMeter)
	}
	isDeliverTx := !ctx.IsCheckTx() && !ctx.IsReCheckTx() && !simulate
	if isDeliverTx || !isGasless {
		// In the case of deliverTx, we want to deduct fees regardless of whether the tx is considered gasless or not, since
		// gasless txs will be subject to application-specific fee requirements in later stage of ante, for which the payment
		// of those app-specific fees happens here. Note that the minimum fee check in the wrapped deduct fee handler is only
		// performed if the context is for CheckTx, so the check will be skipped for deliverTx and the deduct fee handler will
		// only deduct fee without checking.
		// Otherwise (i.e. in the case of checkTx), we only want to perform fee checks and fee deduction if the tx is not considered
		// gasless, or if it specifies a non-zero gas limit even if it is considered gasless, so that the wrapped deduct fee
		// handler will assign an appropriate priority to it.
		return gd.handleWrapped(ctx, tx, simulate, next)
	}

	return next(ctx, tx, simulate)
}

func (gd GaslessDecorator) handleWrapped(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// AnteHandle always takes a `next` so we need a no-op to execute only one handler at a time
	terminatorHandler := func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) {
		return ctx, nil
	}
	// iterating instead of recursing the handler for readability
	for _, handler := range gd.wrapped {
		ctx, err := handler.AnteHandle(ctx, tx, simulate, terminatorHandler)
		if err != nil {
			return ctx, err
		}
	}
	return next(ctx, tx, simulate)
}

func (gd GaslessDecorator) AnteDeps(txDeps []sdkacltypes.AccessOperation, tx sdk.Tx, txIndex int, next sdk.AnteDepGenerator) (newTxDeps []sdkacltypes.AccessOperation, err error) {
	deps := []sdkacltypes.AccessOperation{}
	terminatorDeps := func(txDeps []sdkacltypes.AccessOperation, _ sdk.Tx, _ int) ([]sdkacltypes.AccessOperation, error) {
		return txDeps, nil
	}
	for _, depGen := range gd.wrapped {
		deps, _ = depGen.AnteDeps(deps, tx, txIndex, terminatorDeps)
	}

	return next(append(txDeps, deps...), tx, txIndex)
}

func IsTxGasless(tx sdk.Tx, ctx sdk.Context, evmKeeper *evmkeeper.Keeper) (bool, error) {
	if len(tx.GetMsgs()) == 0 {
		// empty TX shouldn't be gasless
		return false, nil
	}
	for _, msg := range tx.GetMsgs() {
		switch m := msg.(type) {
		case *evmtypes.MsgAssociate:
			if !evmAssociateIsGasless(m, ctx, evmKeeper) {
				return false, nil
			}
			// ddos prevention
			return len(tx.GetMsgs()) == 1, nil
		default:
			return false, nil
		}
	}
	return true, nil
}

func evmAssociateIsGasless(msg *evmtypes.MsgAssociate, ctx sdk.Context, keeper *evmkeeper.Keeper) bool {
	// not gasless if already associated
	kiiAddr := sdk.MustAccAddressFromBech32(msg.Sender)
	_, associated := keeper.GetEVMAddress(ctx, kiiAddr)
	return !associated
}
