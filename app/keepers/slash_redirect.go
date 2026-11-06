package keepers

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// communityPoolFunder is the part of the distribution keeper used to move
// slashed tokens into the community pool.
type communityPoolFunder interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

// slashRedirectBankKeeper is the bank keeper handed to x/staking. Staking only
// burns coins from its bonded and not-bonded pools when a validator is
// slashed, so those burns are sent to the community pool instead and total
// supply is left unchanged. Every other call goes to the wrapped bank keeper.
//
// distrKeeper is set after the distribution keeper is built, because the
// distribution keeper itself depends on the staking keeper.
type slashRedirectBankKeeper struct {
	bankkeeper.Keeper
	distrKeeper communityPoolFunder
}

var _ stakingtypes.BankKeeper = (*slashRedirectBankKeeper)(nil)

// BurnCoins sends coins burned from the staking pools to the community pool,
// and burns coins from any other module account as usual.
func (k *slashRedirectBankKeeper) BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error {
	if moduleName != stakingtypes.BondedPoolName && moduleName != stakingtypes.NotBondedPoolName {
		return k.Keeper.BurnCoins(ctx, moduleName, amt)
	}
	return k.distrKeeper.FundCommunityPool(ctx, amt, authtypes.NewModuleAddress(moduleName))
}
