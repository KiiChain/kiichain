package keeper

import (
	"github.com/ethereum/go-ethereum/common"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/evm/contracts"
	erc20types "github.com/cosmos/evm/x/erc20/types"
)

// ConvertNativeFee prepares the user balance for fees though the registered pairs
// this function considers that the amount passed is the staking denom
func (k Keeper) ConvertNativeFee(ctx sdk.Context, account sdk.AccAddress, fees sdk.Coins) (sdk.Coins, error) {
	// First we check if the amount is zero, if zero we return the zero coin
	if fees.IsZero() {
		return fees, nil
	}

	// We only support a single asset coin for now
	// This is ensure on both Cosmos and EVM side
	// On Cosmos, when the TX goes though the fee market fee ante handler, it returns the only supported asset as the staking coin
	// On EVM we always use the staking coin as the fee coin
	if len(fees) != 1 {
		return sdk.Coins{}, nil
	}

	// Check for the native fees
	ok, err := k.checkNativeFees(ctx, account, fees)
	if err != nil {
		return sdk.Coins{}, err
	}
	if ok {
		return fees, nil
	}

	// Here we know that we have a single token token non zero
	fee := fees[0]

	// Get the fee prices
	feePrices, err := k.GetFeePrices(ctx)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Iterate over the fee prices and try to convert the native fee
	for _, feePrice := range feePrices {
		pairID := k.erc20Keeper.GetTokenPairID(ctx, feePrice.Denom)
		pair, found := k.erc20Keeper.GetTokenPair(ctx, pairID)
		if !found {
			continue // Skip if the pair is not found
		}
		// Take the kii amount
		amount := fee.Amount

		// Calculate the amount the coin is worth in the native token
		amountEquivalent := math.LegacyNewDecFromInt(amount).Mul(feePrice.Price)
		amount = amountEquivalent.TruncateInt()

		// Prepare the user balance for fees
		ok, err := k.PrepareUserBalanceForFees(ctx, account, pair, amount)
		if err != nil {
			return sdk.Coins{}, err
		}

		// If all went well we return the selected fee
		if ok {
			return sdk.Coins{sdk.NewCoin(pair.Denom, amount)}, nil
		}
	}

	// If no suitable pair was found we return an error
	return sdk.Coins{}, errorsmod.Wrapf(
		errortypes.ErrInsufficientFunds,
		"insufficient funds for fee, no suitable pair found for amount %s",
		fees.String(),
	)
}

// checkNativeFees checks if the user has enough native fees
func (k Keeper) checkNativeFees(ctx sdk.Context, account sdk.AccAddress, fees sdk.Coins) (bool, error) {
	// First we check if the we have a single asset coin
	if len(fees) != 1 {
		return false, errorsmod.Wrapf(
			errortypes.ErrInvalidCoins,
			"expected a single asset coin, got %d assets",
			len(fees),
		)
	}
	fee := fees[0]

	// Then we check if the coin is the native coin
	if fee.Denom != k.GetParams(ctx).DefaultDenom {
		return false, errorsmod.Wrapf(
			errortypes.ErrInvalidCoins,
			"expected the native coin %s, got %s",
			k.GetParams(ctx).DefaultDenom,
			fee.Denom,
		)
	}

	// Then we check if the user has enough balance for the fee
	balance := k.bankKeeper.GetBalance(ctx, account, fee.Denom)
	if balance.Amount.GTE(fee.Amount) {
		return true, nil
	}

	// If we reach here, the coin is good for conversion but the user does not have enough balance
	return false, nil
}

// PrepareUserBalanceForFees prepare the user balance for fees
// this checks if the user has enough native balance for the fee
// if not tries to pay for the fee using the erc20 token
func (k Keeper) PrepareUserBalanceForFees(ctx sdk.Context, account sdk.AccAddress, pair erc20types.TokenPair, amount math.Int) (bool, error) {
	// Take the ABI
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI

	// Get the balance for the pair native token on cosmos
	// and check if the user has enough balance, if so we return true
	balance := k.bankKeeper.GetBalance(ctx, account, pair.Denom)
	if balance.Amount.GTE(amount) {
		return true, nil
	}

	// Get the balance for the erc20 token
	erc20Balance := k.erc20Keeper.BalanceOf(ctx, erc20, pair.GetERC20Contract(), common.BytesToAddress(account.Bytes()))
	erc20BalanceInt := math.NewIntFromBigInt(erc20Balance)
	// If the user has enough erc20 balance, we convert the erc20 token to the native token
	// and return true
	if erc20Balance != nil && erc20BalanceInt.GTE(amount) {
		// Prepare the convert message
		msg := erc20types.NewMsgConvertERC20(
			amount,
			account,
			pair.GetERC20Contract(),
			common.BytesToAddress(account.Bytes()),
		)
		if _, err := k.erc20Keeper.ConvertERC20(ctx, msg); err != nil {
			return false, err
		}
		return true, nil
	}

	// If the user does not have enough balance, we return false
	return false, nil
}
