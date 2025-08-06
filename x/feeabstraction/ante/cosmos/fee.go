// This ante handler is a exact copy of the Cosmos SDK's ante handler
// The original implementation can be found at: `x/auth/ante/fee.go`
// These are the main changes to the original implementation:
// - The fee abstraction module is used to convert the fees from the native coin to a available coin
package cosmos

import (
	"bytes"
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	antetypes "github.com/kiichain/kiichain/v3/ante/types"
)

// DeductFeeDecorator deducts fees from the fee payer. The fee payer is the fee granter (if specified) or first signer of the tx.
// If the fee payer does not have the funds to pay for the fees, return an InsufficientFunds error.
// Call next AnteHandler if fees successfully deducted.
// CONTRACT: Tx must implement FeeTx interface to use DeductFeeDecorator
type DeductFeeDecorator struct {
	accountKeeper        ante.AccountKeeper
	bankKeeper           types.BankKeeper
	feegrantKeeper       ante.FeegrantKeeper
	feeAbstractionKeeper antetypes.FeeAbstractionKeeper
	txFeeChecker         ante.TxFeeChecker
}

// NewDeductFeeDecorator creates a new DeductFeeDecorator instance
func NewDeductFeeDecorator(ak ante.AccountKeeper, bk types.BankKeeper, fk ante.FeegrantKeeper, fak antetypes.FeeAbstractionKeeper, tfc ante.TxFeeChecker) DeductFeeDecorator {
	if tfc == nil {
		// This is different from the original implementation
		// Originally, we set as checkTxFeeWithValidatorMinGasPrices
		// But due to our EVM implementation, a custom param is passed to use the fee market module
		panic("txFeeChecker cannot be nil")
	}

	// Return the DeductFeeDecorator
	return DeductFeeDecorator{
		accountKeeper:        ak,
		bankKeeper:           bk,
		feegrantKeeper:       fk,
		feeAbstractionKeeper: fak,
		txFeeChecker:         tfc,
	}
}

// AnteHandle handles the deduction of fees from the fee payer account
func (dfd DeductFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// Parse the tx as a feeTx interface
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// Check if gas is provided for non-simulated transactions
	if !simulate && ctx.BlockHeight() > 0 && feeTx.GetGas() == 0 {
		return ctx, errorsmod.Wrap(sdkerrors.ErrInvalidGasLimit, "must provide positive gas")
	}

	var (
		priority int64
		err      error
	)

	// Extract the fee from the feeTx
	fee := feeTx.GetFee()
	if !simulate {
		fee, priority, err = dfd.txFeeChecker(ctx, tx)
		if err != nil {
			return ctx, err
		}
	}
	// Check and deduct the fees from the fee payer account
	if err := dfd.checkDeductFee(ctx, tx, fee); err != nil {
		return ctx, err
	}

	// Set the TX priority
	newCtx := ctx.WithPriority(priority)

	return next(newCtx, tx, simulate)
}

// checkDeductFee checks if the fee payer has enough funds to pay for the fees and deducts the fees from the fee payer account
func (dfd DeductFeeDecorator) checkDeductFee(ctx sdk.Context, sdkTx sdk.Tx, fee sdk.Coins) error {
	// Parse the tx as a feeTx interface
	feeTx, ok := sdkTx.(sdk.FeeTx)
	if !ok {
		return errorsmod.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	// Check if the fee collector module account is set
	if addr := dfd.accountKeeper.GetModuleAddress(types.FeeCollectorName); addr == nil {
		return fmt.Errorf("fee collector module account (%s) has not been set", types.FeeCollectorName)
	}

	// Get the fee payer and the fee granter from the feeTx
	feePayer := feeTx.FeePayer()
	feeGranter := feeTx.FeeGranter()
	deductFeesFrom := feePayer

	// if feegranter set deduct fee from feegranter account.
	// this works with only when feegrant enabled.
	if feeGranter != nil {
		feeGranterAddr := sdk.AccAddress(feeGranter)

		// If feegranter is set, we need to check if the feegrant module is enabled
		if dfd.feegrantKeeper == nil {
			return sdkerrors.ErrInvalidRequest.Wrap("fee grants are not enabled")
		} else if !bytes.Equal(feeGranterAddr, feePayer) {
			err := dfd.feegrantKeeper.UseGrantedFees(ctx, feeGranterAddr, feePayer, fee, sdkTx.GetMsgs())
			if err != nil {
				return errorsmod.Wrapf(err, "%s does not allow to pay fees for %s", feeGranter, feePayer)
			}
		}

		// If feegranter is set, we deduct the fees from the feegranter account
		deductFeesFrom = feeGranterAddr
	}

	// Get the account of the fee payer
	deductFeesFromAcc := dfd.accountKeeper.GetAccount(ctx, deductFeesFrom)
	if deductFeesFromAcc == nil {
		return sdkerrors.ErrUnknownAddress.Wrapf("fee payer address: %s does not exist", deductFeesFrom)
	}

	// Deduct the fees
	var convertedFee sdk.Coins
	if !fee.IsZero() {
		// Apply the fee conversion from the fee abstraction module
		// This is the only change from the original implementation
		var err error
		convertedFee, err = dfd.feeAbstractionKeeper.ConvertNativeFee(ctx, deductFeesFromAcc.GetAddress(), fee)
		if err != nil {
			return err
		}

		// Deduct the fees from the fee payer account
		err = ante.DeductFees(dfd.bankKeeper, ctx, deductFeesFromAcc, convertedFee)
		if err != nil {
			return err
		}
	}

	// Emit the events
	events := sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeTx,
			sdk.NewAttribute(sdk.AttributeKeyFee, convertedFee.String()),
			sdk.NewAttribute(sdk.AttributeKeyFeePayer, sdk.AccAddress(deductFeesFrom).String()),
		),
	}
	ctx.EventManager().EmitEvents(events)

	return nil
}
