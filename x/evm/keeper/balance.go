package keeper

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain3/x/evm/state"
)

func (k *Keeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress) *big.Int {
	denom := k.GetBaseDenom(ctx)
	allUkii := k.BankKeeper().GetBalance(ctx, addr, denom).Amount
	lockedUkii := k.BankKeeper().LockedCoins(ctx, addr).AmountOf(denom) // LockedCoins doesn't use iterators
	ukii := allUkii.Sub(lockedUkii)
	wei := k.BankKeeper().GetWeiBalance(ctx, addr)
	return ukii.Mul(state.SdkUkiiToSweiMultiplier).Add(wei).BigInt()
}
