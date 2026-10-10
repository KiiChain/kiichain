// This file is based on the Cosmos EVM v0.7 mono decorator:
// https://github.com/cosmos/evm/blob/v0.7.3/ante/evm/mono_decorator.go
//
// Kii changes, required by the v0.6.x to v0.7.0 migration (custom ante):
//   - fees are converted through the fee abstraction module before they are charged
//   - the balance check covers the transaction value only, because fees may be paid in another denom
//   - converted fees are stored on the context under evmkeeper.ContextPaidFeesKey so unused gas is refunded in the paid denom
package evm

import (
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/txpool"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	evmante "github.com/cosmos/evm/ante/evm"
	anteinterfaces "github.com/cosmos/evm/ante/interfaces"
	feemarkettypes "github.com/cosmos/evm/x/feemarket/types"
	evmkeeper "github.com/cosmos/evm/x/vm/keeper"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	antetypes "github.com/kiichain/kiichain/v7/ante/types"
)

const AcceptedTxType = 0 |
	1<<ethtypes.LegacyTxType |
	1<<ethtypes.AccessListTxType |
	1<<ethtypes.DynamicFeeTxType |
	1<<ethtypes.SetCodeTxType

// MonoDecorator is a single decorator that handles all the prechecks for
// ethereum transactions.
type MonoDecorator struct {
	accountKeeper        anteinterfaces.AccountKeeper
	feeMarketKeeper      anteinterfaces.FeeMarketKeeper
	evmKeeper            anteinterfaces.EVMKeeper
	feeAbstractionKeeper antetypes.FeeAbstractionKeeper
	maxGasWanted         uint64
	evmParams            *evmtypes.Params
	feemarketParams      *feemarkettypes.Params
}

// NewEVMMonoDecorator creates the mono decorator used for EVM transactions.
func NewEVMMonoDecorator(
	accountKeeper anteinterfaces.AccountKeeper,
	feeMarketKeeper anteinterfaces.FeeMarketKeeper,
	evmKeeper anteinterfaces.EVMKeeper,
	feeAbstractionKeeper antetypes.FeeAbstractionKeeper,
	maxGasWanted uint64,
	evmParams *evmtypes.Params,
	feemarketParams *feemarkettypes.Params,
) MonoDecorator {
	return MonoDecorator{
		accountKeeper:        accountKeeper,
		feeMarketKeeper:      feeMarketKeeper,
		evmKeeper:            evmKeeper,
		feeAbstractionKeeper: feeAbstractionKeeper,
		maxGasWanted:         maxGasWanted,
		evmParams:            evmParams,
		feemarketParams:      feemarketParams,
	}
}

// AnteHandle handles the entire decorator chain using a mono decorator.
func (md MonoDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	// EVM fee deduction sends to this module. The SDK sets it when the Cosmos
	// deduct-fee decorator is constructed; keep the default if that has not run.
	if authante.FeeRecipientModule == "" {
		authante.FeeRecipientModule = authtypes.FeeCollectorName
	}

	var txFeeInfo *txtypes.Fee
	if !ctx.IsReCheckTx() {
		txFeeInfo, err = evmante.ValidateTx(tx)
		if err != nil {
			return ctx, err
		}
	}

	evmDenom := evmtypes.GetEVMCoinDenom()

	ctx, err = evmante.SetupContextAndResetTransientGas(ctx, tx)
	if err != nil {
		return ctx, err
	}

	decUtils, err := evmante.NewMonoDecoratorUtils(ctx, md.evmKeeper, md.evmParams, md.feemarketParams)
	if err != nil {
		return ctx, err
	}

	msgs := tx.GetMsgs()
	if len(msgs) != 1 {
		return ctx, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "expected 1 message, got %d", len(msgs))
	}

	ethMsg, ethTx, err := evmtypes.UnpackEthMsg(msgs[0])
	if err != nil {
		return ctx, err
	}

	header := ethtypes.Header{
		GasLimit:   ethTx.Gas(),
		BaseFee:    decUtils.BaseFee,
		Number:     big.NewInt(ctx.BlockHeight()),
		Time:       uint64(ctx.BlockTime().Unix()),
		Difficulty: big.NewInt(0),
	}

	chainConfig := evmtypes.GetEthChainConfig()
	if err := txpool.ValidateTransaction(ethTx, &header, decUtils.Signer, &txpool.ValidationOptions{
		Config:  chainConfig,
		Accept:  AcceptedTxType,
		MaxSize: math.MaxUint64,
		MinTip:  new(big.Int),
	}); err != nil {
		return ctx, err
	}

	feeAmt := ethMsg.GetFee()
	gas := ethTx.Gas()
	fee := sdkmath.LegacyNewDecFromBigInt(feeAmt)
	gasLimit := sdkmath.LegacyNewDecFromBigInt(new(big.Int).SetUint64(gas))

	if ctx.IsCheckTx() && !simulate {
		if err := evmante.CheckMempoolFee(fee, decUtils.MempoolMinGasPrice, gasLimit, decUtils.Rules.IsLondon); err != nil {
			return ctx, err
		}
	}

	if ethTx.Type() >= ethtypes.DynamicFeeTxType && decUtils.BaseFee != nil {
		feeAmt = ethMsg.GetEffectiveFee(decUtils.BaseFee)
		fee = sdkmath.LegacyNewDecFromBigInt(feeAmt)
	}

	if err := evmante.CheckGlobalFee(fee, decUtils.GlobalMinGasPrice, gasLimit); err != nil {
		return ctx, err
	}

	if err := evmante.ValidateMsg(decUtils.EvmParams, ethTx); err != nil {
		return ctx, err
	}

	if v, ok := ctx.GetIncarnationCache(evmante.EthSigVerificationResultCacheKey); ok {
		if v != nil {
			cachedErr, ok := v.(error)
			if !ok {
				return ctx, fmt.Errorf("unexpected type %T cached under %s, want error", v, evmante.EthSigVerificationResultCacheKey)
			}
			return ctx, cachedErr
		}
	} else {
		err = evmante.SignatureVerification(ethMsg, ethTx, decUtils.Signer)
		ctx.SetIncarnationCache(evmante.EthSigVerificationResultCacheKey, err)
		if err != nil {
			return ctx, err
		}
	}

	from := ethMsg.GetFrom()
	fromAddr := common.BytesToAddress(from)

	account := md.evmKeeper.GetAccount(ctx, fromAddr)
	if err := VerifyIfAccountExists(ctx, md.accountKeeper, md.evmKeeper, account, fromAddr); err != nil {
		return ctx, err
	}

	coreMsg := ethMsg.AsMessage(decUtils.BaseFee)
	if err := evmante.CanTransfer(
		ctx,
		md.evmKeeper,
		*coreMsg,
		decUtils.BaseFee,
		decUtils.EvmParams,
		decUtils.Rules.IsLondon,
	); err != nil {
		return ctx, err
	}

	msgFees, err := evmkeeper.VerifyFee(
		ethTx,
		evmDenom,
		decUtils.BaseFee,
		decUtils.Rules.IsHomestead,
		decUtils.Rules.IsIstanbul,
		decUtils.Rules.IsShanghai,
		ctx.IsCheckTx(),
	)
	if err != nil {
		return ctx, err
	}

	convertedMsgFees, err := md.feeAbstractionKeeper.ConvertNativeFee(ctx, from, msgFees)
	if err != nil {
		return ctx, err
	}

	if err := evmante.ConsumeFeesAndEmitEvent(ctx, md.evmKeeper, convertedMsgFees, from); err != nil {
		return ctx, err
	}

	// Fees are already charged, possibly in a non-native denom. The remaining
	// check is that the sender can cover the transaction value.
	account = md.evmKeeper.GetAccount(ctx, fromAddr)
	if err := VerifyAccountBalance(ctx, md.accountKeeper, account, ethTx); err != nil {
		return ctx, err
	}

	decUtils.GasWanted = evmante.UpdateCumulativeGasWanted(ctx, gas, md.maxGasWanted, decUtils.GasWanted)
	decUtils.MinPriority = evmante.GetMsgPriority(ethTx, decUtils.MinPriority, decUtils.BaseFee)
	decUtils.TxFee.Add(decUtils.TxFee, ethMsg.GetFee())
	decUtils.TxGasLimit += gas

	acc := md.accountKeeper.GetAccount(ctx, from)
	if acc == nil {
		return ctx, errorsmod.Wrapf(errortypes.ErrUnknownAddress, "account %s does not exist", from)
	}
	if err := evmante.IncrementNonce(ctx, md.accountKeeper, acc, ethTx.Nonce()); err != nil {
		return ctx, err
	}

	evmante.EmitTxHashEvent(ctx, ethMsg, uint64(ctx.TxIndex()))

	if err := evmante.CheckTxFee(txFeeInfo, decUtils.TxFee, decUtils.TxGasLimit); err != nil {
		return ctx, err
	}

	ctx, err = evmante.CheckBlockGasLimit(ctx, decUtils.GasWanted, decUtils.MinPriority)
	if err != nil {
		return ctx, err
	}

	ctx = ctx.WithValue(evmkeeper.ContextPaidFeesKey{}, convertedMsgFees)
	return next(ctx, tx, simulate)
}
