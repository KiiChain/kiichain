package keeper

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/kiichain/kiichain/v3/app/params"
	"github.com/kiichain/kiichain/v3/x/feeabstraction/types"

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
	// Get the module params
	params, err := k.Params.Get(ctx)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Check if the module is enabled
	if !params.Enabled {
		return fees, nil // If the module is disabled, we return the fees as is
	}

	// Validate the input fees
	// We only support a single asset coin for now
	// This is ensured on both Cosmos and EVM sides:
	// - On Cosmos, when the TX goes though the fee market fee ante handler, it returns the only supported asset as the staking coin
	// - On EVM we always use the staking coin as the fee coin
	if fees.IsZero() {
		return fees, nil
	}
	if len(fees) != 1 {
		// We don't support multi tokens
		return fees, nil
	}
	fee := fees[0]

	// Check if the fee is under the native denom
	if fee.Denom != params.NativeDenom {
		return fees, nil
	}

	// Check for the native fees
	ok, err := k.hasSufficientNativeBalance(ctx, account, fee)
	if err != nil {
		return sdk.Coins{}, err
	}
	if ok {
		return fees, nil
	}

	// convert ERC20 tokens to fees
	return k.convertERC20ForFees(ctx, account, fee)
}

// hasSufficientNativeBalance checks if the user has enough balance to pay using the native coin
func (k Keeper) hasSufficientNativeBalance(ctx sdk.Context, account sdk.AccAddress, fee sdk.Coin) (bool, error) {
	// Then we check if the user has enough balance for the fee
	balance := k.bankKeeper.GetBalance(ctx, account, fee.Denom)
	if balance.Amount.GTE(fee.Amount) {
		return true, nil
	}

	// If we reach here, the coin is good for conversion but the user does not have enough balance
	return false, nil
}

// convertERC20ForFees prepares the user balance for fees by converting the native coin to the fee token
// It checks if the user has enough balance in the native token, if not it tries to
// convert the ERC20 token to the native token
func (k Keeper) convertERC20ForFees(ctx sdk.Context, account sdk.AccAddress, fee sdk.Coin) (sdk.Coins, error) {
	// Get the fee prices
	feePrices, err := k.FeeTokens.Get(ctx)
	if err != nil {
		return sdk.Coins{}, err
	}

	// Iterate over the fee prices and try to convert the native fee
	for _, feePrice := range feePrices.Items {
		// Check if the token is enabled
		if !feePrice.Enabled {
			continue
		}

		// Convert the amount using the price
		amountEquivalent, err := types.CalculateTokenAmountWithDecimals(
			feePrice.Price,
			fee.Amount,
			params.BaseDenomUnit,
			uint64(feePrice.Decimals),
		)
		if err != nil {
			return sdk.Coins{}, err
		}
		// Truncate the decimals
		amountEquivalentInt := amountEquivalent.TruncateInt()

		// Prepare the user balance for fees
		ok, err := k.convertERC20ToNative(ctx, account, fee.Denom, amountEquivalentInt)
		if err != nil {
			return sdk.Coins{}, err
		}

		// If all went well we return the selected fee
		if ok {
			return sdk.Coins{sdk.NewCoin(fee.Denom, amountEquivalentInt)}, nil
		}
	}

	// If no suitable pair was found we return an error
	return sdk.Coins{}, errorsmod.Wrapf(
		errortypes.ErrInsufficientFunds,
		"insufficient funds for fee, no suitable pair found for amount %s",
		fee.String(),
	)
}

// convertERC20ToNative converts the ERC20 token to the native token
// It checks if the user has enough balance in the native token, if not it tries to
// convert the ERC20 token to the native token
func (k Keeper) convertERC20ToNative(ctx sdk.Context, account sdk.AccAddress, denom string, amount math.Int) (bool, error) {
	// Get the balance for the pair native token on cosmos
	// and check if the user has enough balance, if so we return true
	balance := k.bankKeeper.GetBalance(ctx, account, denom)
	if balance.Amount.GTE(amount) {
		return true, nil
	}

	// Get the pair ID and check if it exists
	pairID := k.erc20Keeper.GetTokenPairID(ctx, denom)
	pair, found := k.erc20Keeper.GetTokenPair(ctx, pairID)
	if !found {
		return false, nil
	}

	// Take the ABI
	erc20 := contracts.ERC20MinterBurnerDecimalsContract.ABI

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
