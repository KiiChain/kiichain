package keepers

import (
	"context"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	tokenfactorytypes "github.com/kiichain/kiichain/v7/x/tokenfactory/types"
)

// slashStakingKeeper is the staking-keeper subset needed by SlashHooks
type slashStakingKeeper interface {
	GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error)
	BondDenom(ctx context.Context) (string, error)
}

// slashBankKeeper is the bank-keeper subset needed by SlashHooks
type slashBankKeeper interface {
	MintCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error
}

// slashDistrKeeper is the distribution-keeper subset needed by SlashHooks
type slashDistrKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

// SlashHooks implements stakingtypes.StakingHooks
type SlashHooks struct {
	stakingKeeper slashStakingKeeper
	bankKeeper    slashBankKeeper
	distrKeeper   slashDistrKeeper
}

// NewSlashHooks returns a SlashHooks
func NewSlashHooks(sk slashStakingKeeper, bk slashBankKeeper, dk slashDistrKeeper) SlashHooks {
	return SlashHooks{stakingKeeper: sk, bankKeeper: bk, distrKeeper: dk}
}

var _ stakingtypes.StakingHooks = SlashHooks{}

// BeforeValidatorSlashed mints the equivalent of the slash amount into the
// community pool before the tokens are burned
func (h SlashHooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction math.LegacyDec) error {
	validator, err := h.stakingKeeper.GetValidator(ctx, valAddr)
	if err != nil || !validator.Tokens.IsPositive() {
		return err
	}

	// The effectiveFraction passed by slashToPool is tokensToBurn / validator.Tokens,
	// so multiplying back recovers tokensToBurn
	mintAmount := math.LegacyNewDecFromInt(validator.Tokens).Mul(fraction).TruncateInt()
	if mintAmount.IsZero() {
		return nil
	}

	bondDenom, err := h.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return err
	}

	coins := sdk.NewCoins(sdk.NewCoin(bondDenom, mintAmount))

	// Mint into the tokenfactory module account (needs mint permission)
	// and immediately forward to the distribution community pool
	if err := h.bankKeeper.MintCoins(ctx, tokenfactorytypes.ModuleName, coins); err != nil {
		return err
	}

	tokenfactoryAddr := authtypes.NewModuleAddress(tokenfactorytypes.ModuleName)
	return h.distrKeeper.FundCommunityPool(ctx, coins, tokenfactoryAddr)
}

// The remaining hooks are not in use; only BeforeValidatorSlashed is used

func (h SlashHooks) AfterValidatorCreated(_ context.Context, _ sdk.ValAddress) error {
	return nil
}

func (h SlashHooks) BeforeValidatorModified(_ context.Context, _ sdk.ValAddress) error {
	return nil
}

func (h SlashHooks) AfterValidatorRemoved(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h SlashHooks) AfterValidatorBonded(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h SlashHooks) AfterValidatorBeginUnbonding(_ context.Context, _ sdk.ConsAddress, _ sdk.ValAddress) error {
	return nil
}

func (h SlashHooks) BeforeDelegationCreated(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h SlashHooks) BeforeDelegationSharesModified(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h SlashHooks) BeforeDelegationRemoved(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h SlashHooks) AfterDelegationModified(_ context.Context, _ sdk.AccAddress, _ sdk.ValAddress) error {
	return nil
}

func (h SlashHooks) AfterUnbondingInitiated(_ context.Context, _ uint64) error {
	return nil
}
