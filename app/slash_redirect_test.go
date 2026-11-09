package kiichain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/comet"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	kiichain "github.com/kiichain/kiichain/v7/app"
	kiihelpers "github.com/kiichain/kiichain/v7/app/helpers"
	"github.com/kiichain/kiichain/v7/app/keepers"
)

// slashBalances is the bond denom state that a slash moves between.
type slashBalances struct {
	supply        math.Int
	bondedPool    math.Int
	notBondedPool math.Int
	distrModule   math.Int
	communityPool math.LegacyDec
}

func getSlashBalances(t *testing.T, app *kiichain.KiichainApp, ctx sdk.Context, denom string) slashBalances {
	t.Helper()
	feePool, err := app.DistrKeeper.FeePool.Get(ctx)
	require.NoError(t, err)
	balance := func(moduleName string) math.Int {
		return app.BankKeeper.GetBalance(ctx, authtypes.NewModuleAddress(moduleName), denom).Amount
	}
	return slashBalances{
		supply:        app.BankKeeper.GetSupply(ctx, denom).Amount,
		bondedPool:    balance(stakingtypes.BondedPoolName),
		notBondedPool: balance(stakingtypes.NotBondedPoolName),
		distrModule:   balance(distrtypes.ModuleName),
		communityPool: feePool.CommunityPool.AmountOf(denom),
	}
}

// requireSlashedToCommunityPool checks that everything taken from the staking
// pools went to the community pool and that total supply did not change.
func requireSlashedToCommunityPool(t *testing.T, before, after slashBalances) math.Int {
	t.Helper()
	slashed := before.bondedPool.Add(before.notBondedPool).Sub(after.bondedPool).Sub(after.notBondedPool)
	require.True(t, slashed.IsPositive(), "expected tokens to be slashed from the staking pools")
	require.Equal(t, before.supply, after.supply, "slashing must not change total supply")
	require.Equal(t, before.distrModule.Add(slashed), after.distrModule)
	require.Equal(t, before.communityPool.Add(math.LegacyNewDecFromInt(slashed)), after.communityPool)
	return slashed
}

// requireSlashEvents checks that the slash_to_community_pool events add up to
// the slashed amount and returns the amount reported for each staking pool.
func requireSlashEvents(t *testing.T, ctx sdk.Context, denom string, slashed math.Int) map[string]math.Int {
	t.Helper()
	byPool := map[string]math.Int{}
	total := math.ZeroInt()
	for _, event := range ctx.EventManager().Events() {
		if event.Type != keepers.EventTypeSlashToCommunityPool {
			continue
		}
		pool, ok := event.GetAttribute(keepers.AttributeKeyPool)
		require.True(t, ok, "missing pool attribute")
		require.Contains(t, []string{stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName}, pool.Value)
		amountAttr, ok := event.GetAttribute(sdk.AttributeKeyAmount)
		require.True(t, ok, "missing amount attribute")
		coins, err := sdk.ParseCoinsNormalized(amountAttr.Value)
		require.NoError(t, err)
		amount := coins.AmountOf(denom)
		if prev, found := byPool[pool.Value]; found {
			amount = amount.Add(prev)
		}
		byPool[pool.Value] = amount
		total = total.Add(coins.AmountOf(denom))
	}
	require.Equal(t, slashed, total, "slash events must report the slashed amount")
	return byPool
}

func TestSlashBondedValidatorFundsCommunityPool(t *testing.T) {
	app, ctx := kiihelpers.SetupWithContext(t)
	denom, err := app.StakingKeeper.BondDenom(ctx)
	require.NoError(t, err)

	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	validator := validators[0]
	require.Equal(t, stakingtypes.Bonded, validator.GetStatus())
	consAddr, err := validator.GetConsAddr()
	require.NoError(t, err)
	power := validator.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx))

	ctx = ctx.WithEventManager(sdk.NewEventManager())
	before := getSlashBalances(t, app, ctx, denom)
	burned, err := app.StakingKeeper.Slash(ctx, consAddr, ctx.BlockHeight(), power, math.LegacyNewDecWithPrec(1, 1))
	require.NoError(t, err)
	after := getSlashBalances(t, app, ctx, denom)

	slashed := requireSlashedToCommunityPool(t, before, after)
	require.Equal(t, burned, slashed)
	require.Equal(t, before.bondedPool.Sub(burned), after.bondedPool)
	events := requireSlashEvents(t, ctx, denom, slashed)
	require.Equal(t, slashed, events[stakingtypes.BondedPoolName])
}

func TestSlashUnbondingDelegationFundsCommunityPool(t *testing.T) {
	app, ctx := kiihelpers.SetupWithContext(t)
	denom, err := app.StakingKeeper.BondDenom(ctx)
	require.NoError(t, err)

	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	validator := validators[0]
	valAddr, err := sdk.ValAddressFromBech32(validator.GetOperator())
	require.NoError(t, err)
	consAddr, err := validator.GetConsAddr()
	require.NoError(t, err)
	power := validator.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx))

	delegations, err := app.StakingKeeper.GetValidatorDelegations(ctx, valAddr)
	require.NoError(t, err)
	require.NotEmpty(t, delegations)
	delAddr, err := sdk.AccAddressFromBech32(delegations[0].DelegatorAddress)
	require.NoError(t, err)

	// Unbond half of the stake at the infraction height so the slash reaches
	// the unbonding entry in the not-bonded pool as well as the validator.
	infractionHeight := ctx.BlockHeight()
	_, unbonded, err := app.StakingKeeper.Undelegate(ctx, delAddr, valAddr, delegations[0].Shares.QuoInt64(2))
	require.NoError(t, err)
	require.True(t, unbonded.IsPositive())
	ctx = ctx.WithBlockHeight(infractionHeight + 1)

	ctx = ctx.WithEventManager(sdk.NewEventManager())
	before := getSlashBalances(t, app, ctx, denom)
	_, err = app.StakingKeeper.Slash(ctx, consAddr, infractionHeight, power, math.LegacyNewDecWithPrec(1, 1))
	require.NoError(t, err)
	after := getSlashBalances(t, app, ctx, denom)

	slashed := requireSlashedToCommunityPool(t, before, after)
	require.True(t, after.notBondedPool.LT(before.notBondedPool), "expected the unbonding delegation to be slashed")
	require.True(t, after.bondedPool.LT(before.bondedPool), "expected the bonded validator to be slashed")
	events := requireSlashEvents(t, ctx, denom, slashed)
	require.Equal(t, before.bondedPool.Sub(after.bondedPool), events[stakingtypes.BondedPoolName])
	require.Equal(t, before.notBondedPool.Sub(after.notBondedPool), events[stakingtypes.NotBondedPoolName])
}

func TestDowntimeSlashFundsCommunityPool(t *testing.T) {
	app, ctx := kiihelpers.SetupWithContext(t)
	denom, err := app.StakingKeeper.BondDenom(ctx)
	require.NoError(t, err)

	params, err := app.SlashingKeeper.GetParams(ctx)
	require.NoError(t, err)
	params.SignedBlocksWindow = 10
	params.MinSignedPerWindow = math.LegacyNewDecWithPrec(5, 1)
	params.DowntimeJailDuration = 60 * time.Second
	params.SlashFractionDowntime = math.LegacyNewDecWithPrec(1, 2)
	require.NoError(t, app.SlashingKeeper.SetParams(ctx, params))

	validators, err := app.StakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	validator := validators[0]
	require.Equal(t, stakingtypes.Bonded, validator.GetStatus())
	consAddr, err := validator.GetConsAddr()
	require.NoError(t, err)
	power := validator.GetConsensusPower(app.StakingKeeper.PowerReduction(ctx))

	startHeight := ctx.BlockHeight()
	if !app.SlashingKeeper.HasValidatorSigningInfo(ctx, consAddr) {
		info := slashingtypes.NewValidatorSigningInfo(consAddr, startHeight, 0, time.Unix(0, 0), false, 0)
		require.NoError(t, app.SlashingKeeper.SetValidatorSigningInfo(ctx, consAddr, info))
	}

	ctx = ctx.WithEventManager(sdk.NewEventManager())
	before := getSlashBalances(t, app, ctx, denom)
	// Downtime only fires once the height is past StartHeight + SignedBlocksWindow;
	// once the validator is jailed further calls are no-ops.
	for height := startHeight + 1; height <= startHeight+2*params.SignedBlocksWindow; height++ {
		ctx = ctx.WithBlockHeight(height)
		require.NoError(t, app.SlashingKeeper.HandleValidatorSignature(ctx, consAddr, power, comet.BlockIDFlagAbsent))
	}
	after := getSlashBalances(t, app, ctx, denom)

	jailed, err := app.StakingKeeper.GetValidatorByConsAddr(ctx, consAddr)
	require.NoError(t, err)
	require.True(t, jailed.IsJailed(), "expected the validator to be jailed for downtime")

	slashed := requireSlashedToCommunityPool(t, before, after)
	expected := math.LegacyNewDecFromInt(app.StakingKeeper.TokensFromConsensusPower(ctx, power)).Mul(params.SlashFractionDowntime).TruncateInt()
	require.Equal(t, expected, slashed)
	require.Equal(t, before.bondedPool.Sub(slashed), after.bondedPool)
	events := requireSlashEvents(t, ctx, denom, slashed)
	require.Equal(t, slashed, events[stakingtypes.BondedPoolName])
}
