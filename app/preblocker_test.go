package kiichain_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	kiichain "github.com/kiichain/kiichain/v7/app"
	"github.com/kiichain/kiichain/v7/app/blockedaddrs"
	kiihelpers "github.com/kiichain/kiichain/v7/app/helpers"
	v740 "github.com/kiichain/kiichain/v7/app/upgrades/v7_4"
	tokenfactorytypes "github.com/kiichain/kiichain/v7/x/tokenfactory/types"
)

// fundAnAttacker mints a generous akii balance (comfortably more than the
// fixed payout list could ever total) into one real attacker address, so a
// PreBlocker call that actually applies the upgrade Plan can complete the
// fund-recovery step instead of failing on insufficient recovered funds.
// The recovery math itself is exercised in detail by
// app/upgrades/v7_4/upgrade_test.go; here we only need it to succeed.
func fundAnAttacker(t *testing.T, app *kiichain.KiichainApp, ctx sdk.Context) {
	t.Helper()
	attacker := blockedaddrs.SortedAttackerAddresses()[0]
	coins := sdk.NewCoins(sdk.NewCoin("akii", math.NewIntWithDecimal(1, 30)))
	require.NoError(t, app.BankKeeper.MintCoins(ctx, tokenfactorytypes.ModuleName, coins))
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToAccount(ctx, tokenfactorytypes.ModuleName, sdk.MustAccAddressFromBech32(attacker), coins))
}

// expectedPlan mirrors what app.PreBlocker constructs for the emergency
// upgrade at v740.UpgradeHeight.
func expectedPlan() upgradetypes.Plan {
	return upgradetypes.Plan{
		Name:   v740.UpgradeName,
		Height: v740.UpgradeHeight,
		Info:   "emergency fund recovery post-exploit",
	}
}

// withHeight sets height on both places sdk.Context tracks it. WithBlockHeight
// only updates the legacy BlockHeight()/BlockHeader() pair; x/upgrade's own
// PreBlocker reads HeaderInfo().Height instead, which a real node always
// keeps in sync with the block being processed but a bare test context does
// not unless told to.
func withHeight(ctx sdk.Context, height int64) sdk.Context {
	ctx = ctx.WithBlockHeight(height)
	info := ctx.HeaderInfo()
	info.Height = height
	return ctx.WithHeaderInfo(info)
}

// TestPreBlocker_SchedulesUpgradeAtHeight verifies the happy path: no plan
// exists yet, so PreBlocker schedules the expected one, and — because
// x/upgrade's own PreBlocker runs right after ours in the same call — it
// gets applied immediately, in this same block. By the time PreBlocker
// returns, ApplyUpgrade has already cleared the plan (that's normal x/upgrade
// behavior), so completion is checked via GetDoneHeight, not GetUpgradePlan.
func TestPreBlocker_SchedulesUpgradeAtHeight(t *testing.T) {
	app, ctx := kiihelpers.SetupWithContext(t)
	ctx = withHeight(ctx.WithChainID(v740.MainnetChainID), v740.UpgradeHeight)
	fundAnAttacker(t, app, ctx)

	_, err := app.PreBlocker(ctx, nil)
	require.NoError(t, err)

	_, err = app.UpgradeKeeper.GetUpgradePlan(ctx)
	require.ErrorIs(t, err, upgradetypes.ErrNoUpgradePlanFound, "plan should be cleared once applied")

	doneHeight, err := app.UpgradeKeeper.GetDoneHeight(ctx, v740.UpgradeName)
	require.NoError(t, err)
	require.Equal(t, v740.UpgradeHeight, doneHeight)
}

// TestPreBlocker_NoOpWhenChainIDDoesNotMatch verifies the mainnet-only gate:
// at the right height but the wrong chain-id, nothing gets scheduled.
func TestPreBlocker_NoOpWhenChainIDDoesNotMatch(t *testing.T) {
	app, ctx := kiihelpers.SetupWithContext(t) // default test chain-id, not v740.MainnetChainID
	ctx = withHeight(ctx, v740.UpgradeHeight)

	_, err := app.PreBlocker(ctx, nil)
	require.NoError(t, err)

	_, err = app.UpgradeKeeper.GetUpgradePlan(ctx)
	require.ErrorIs(t, err, upgradetypes.ErrNoUpgradePlanFound)
}

// TestPreBlocker_NoOpWhenHeightDoesNotMatch verifies the same for the height
// half of the gate, on the right chain-id.
func TestPreBlocker_NoOpWhenHeightDoesNotMatch(t *testing.T) {
	app, ctx := kiihelpers.SetupWithContext(t)
	ctx = withHeight(ctx.WithChainID(v740.MainnetChainID), v740.UpgradeHeight-1)

	_, err := app.PreBlocker(ctx, nil)
	require.NoError(t, err)

	_, err = app.UpgradeKeeper.GetUpgradePlan(ctx)
	require.ErrorIs(t, err, upgradetypes.ErrNoUpgradePlanFound)
}

// TestPreBlocker_NoPanicWhenExistingPlanMatchesExactly covers the replay
// safety net: if the exact expected plan is somehow already scheduled when
// PreBlocker runs, it must not panic or try to reschedule.
func TestPreBlocker_NoPanicWhenExistingPlanMatchesExactly(t *testing.T) {
	app, ctx := kiihelpers.SetupWithContext(t)
	ctx = withHeight(ctx.WithChainID(v740.MainnetChainID), v740.UpgradeHeight)
	fundAnAttacker(t, app, ctx)

	require.NoError(t, app.UpgradeKeeper.ScheduleUpgrade(ctx, expectedPlan()))

	require.NotPanics(t, func() {
		_, err := app.PreBlocker(ctx, nil)
		require.NoError(t, err)
	})
}

// TestPreBlocker_PanicsWhenExistingPlanDoesNotMatch verifies the fail-closed
// guard: an unrelated plan already occupying the slot at the same height
// must not be silently ignored.
func TestPreBlocker_PanicsWhenExistingPlanDoesNotMatch(t *testing.T) {
	app, ctx := kiihelpers.SetupWithContext(t)
	ctx = withHeight(ctx.WithChainID(v740.MainnetChainID), v740.UpgradeHeight)

	conflicting := upgradetypes.Plan{
		Name:   "some-other-upgrade",
		Height: v740.UpgradeHeight,
		Info:   "unrelated",
	}
	require.NoError(t, app.UpgradeKeeper.ScheduleUpgrade(ctx, conflicting))

	defer func() {
		r := recover()
		require.NotNil(t, r, "expected a panic")
		err, ok := r.(error)
		require.True(t, ok, "panic value should be an error, got %T", r)
		require.ErrorContains(t, err, "unexpected active upgrade plan")
		require.ErrorContains(t, err, "some-other-upgrade")
	}()

	_, _ = app.PreBlocker(ctx, nil)
	t.Fatal("expected PreBlocker to panic")
}

// farFutureOffset stands in for "long after the incident" — e.g. a node that
// joins via state-sync, whose first PreBlocker call ever is at whatever
// height its snapshot targets, not UpgradeHeight+1. The guard must not have
// a blind spot that only covers the immediately-following block.
const farFutureOffset = 1_000_000

// TestPreBlocker_ContinuesPastUpgradeHeightOnceRecoveryVerified verifies the
// guard added for blocks after UpgradeHeight: once the recovery has genuinely
// run (done-height recorded, incident restriction enabled), every later
// block on mainnet processes normally — not just the one right after it.
func TestPreBlocker_ContinuesPastUpgradeHeightOnceRecoveryVerified(t *testing.T) {
	for _, offset := range []int64{1, farFutureOffset} {
		t.Run(fmt.Sprintf("height+%d", offset), func(t *testing.T) {
			app, ctx := kiihelpers.SetupWithContext(t)
			ctx = withHeight(ctx.WithChainID(v740.MainnetChainID), v740.UpgradeHeight)
			fundAnAttacker(t, app, ctx)

			_, err := app.PreBlocker(ctx, nil)
			require.NoError(t, err)

			ctx = withHeight(ctx, v740.UpgradeHeight+offset)
			require.NotPanics(t, func() {
				_, err := app.PreBlocker(ctx, nil)
				require.NoError(t, err)
			})
		})
	}
}

// TestPreBlocker_PanicsWhenRecoveryNeverApplied is the regression test for
// this guard: a node past UpgradeHeight on mainnet that never actually ran
// the recovery (e.g. it skipped the upgrade, or state-synced from a snapshot
// that predates the fix) must refuse to keep processing blocks — at any
// height past the target, not just the very next one — rather than silently
// running on with the exploited funds unfrozen.
func TestPreBlocker_PanicsWhenRecoveryNeverApplied(t *testing.T) {
	for _, offset := range []int64{1, farFutureOffset} {
		t.Run(fmt.Sprintf("height+%d", offset), func(t *testing.T) {
			app, ctx := kiihelpers.SetupWithContext(t)
			ctx = withHeight(ctx.WithChainID(v740.MainnetChainID), v740.UpgradeHeight+offset)

			defer func() {
				r := recover()
				require.NotNil(t, r, "expected a panic")
				err, ok := r.(error)
				require.True(t, ok, "panic value should be an error, got %T", r)
				require.ErrorContains(t, err, "emergency recovery missing")
			}()

			_, _ = app.PreBlocker(ctx, nil)
			t.Fatal("expected PreBlocker to panic")
		})
	}
}
