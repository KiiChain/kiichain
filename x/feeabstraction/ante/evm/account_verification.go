// This file is based on the Cosmos EVM package
// The original implementation can be found at: `https://github.com/cosmos/evm/blob/main/ante/evm/06_account_verification.go`
// The main changes here are related to the balance checks
// Differently from the original implementation, here we only check if the user has balance to pay for the transaction value
// fees are ignored at this point and considered paid

package evm

import (
	"github.com/ethereum/go-ethereum/common"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	anteinterfaces "github.com/cosmos/evm/ante/interfaces"
	"github.com/cosmos/evm/x/vm/statedb"
	evmtypes "github.com/cosmos/evm/x/vm/types"
)

// VerifyIfAccountExists checks if the account exists in the store and creates it if it doesn't
func VerifyIfAccountExists(
	ctx sdk.Context,
	accountKeeper anteinterfaces.AccountKeeper,
	account *statedb.Account,
	from common.Address,
) error {
	// Only EOA are allowed to send transactions.
	if account != nil && account.IsContract() {
		return errorsmod.Wrapf(
			errortypes.ErrInvalidType,
			"the sender is not EOA: address %s", from,
		)
	}

	// Check if the account is nil, if so, create a new account
	if account == nil {
		acc := accountKeeper.NewAccountWithAddress(ctx, from.Bytes())
		accountKeeper.SetAccount(ctx, acc)
	}

	return nil
}

// VerifyAccountBalance checks that the account balance is greater than the total transaction value.
// The account will be set to store if it doesn't exist, i.e. cannot be found on store.
// This method will fail if:
// - from address is NOT an EOA
// - account balance is lower than the transaction value
func VerifyAccountBalance(
	ctx sdk.Context,
	accountKeeper anteinterfaces.AccountKeeper,
	account *statedb.Account,
	txData evmtypes.TxData,
) error {
	// Check the sender balance against the TX data
	// This checks if the sender has enough funds to pay for the transaction value
	if err := checkSenderBalance(sdkmath.NewIntFromBigInt(account.Balance), txData); err != nil {
		return errorsmod.Wrap(err, "failed to check sender balance")
	}

	return nil
}

// checkSenderBalance validates that the tx cost value is positive and that the
// sender has enough funds to pay for the fees and value of the transaction.
func checkSenderBalance(
	balance sdkmath.Int,
	txData evmtypes.TxData,
) error {
	// Get the value, this is only the value of the transaction, not the fees
	value := txData.GetValue()

	// Check if the value is valid
	if value.Sign() < 0 {
		return errorsmod.Wrapf(
			errortypes.ErrInvalidCoins,
			"tx value (%s) is negative and invalid", value,
		)
	}

	// Check if the balance is negative or if the balance is less than the cost
	if balance.IsNegative() || balance.BigInt().Cmp(value) < 0 {
		return errorsmod.Wrapf(
			errortypes.ErrInsufficientFunds,
			"sender balance < tx value (%s < %s)", balance, value,
		)
	}
	return nil
}
