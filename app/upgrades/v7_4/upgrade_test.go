package v740_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	evmtypes "github.com/cosmos/evm/x/vm/types"

	kiichain "github.com/kiichain/kiichain/v7/app"
	"github.com/kiichain/kiichain/v7/app/blockedaddrs"
	kiihelpers "github.com/kiichain/kiichain/v7/app/helpers"
	v740 "github.com/kiichain/kiichain/v7/app/upgrades/v7_4"
	tokenfactorytypes "github.com/kiichain/kiichain/v7/x/tokenfactory/types"
)

const denom = "akii"

// attacker1 and attacker2 are two of the real, hardcoded attackerAddrs from
// v7_4's upgrade.go. stagingAddr and remainderAddr mirror its hardcoded
// constants of the same name.
const (
	attacker1     = "kii1peafvgnleuyl20tyfwnyvtvvwwvnaujxmqe5qe"
	attacker2     = "kii1a5v3eaeaugdh3vk57nlh8q8xcu7z46w0ttlrw9"
	stagingAddr   = "kii1vqu8rska6swzdmnhf90zuv0xmelej4lq5el7zh"
	remainderAddr = "kii1c6cgjmsx0ewl6j552sp06musutmfcvxcaq4n9h"
)

// payouts mirrors v7_4's hardcoded redistribution list exactly, so the test
// can fund staging with precisely enough to cover it and verify every leg.
var payouts = []struct {
	addr   string
	amount string
}{
	{"kii19c6q309u7c9atnvefqajdjzzjhn82cfcakx4cc", "1023953000000000000000"},
	{"kii14r6lynfqtl6cznllajhms8xy9ce8udnqw73zwz", "366930000000000000000000"},
	{"kii1af3ecamzq3zcllmdahx2sc0gaqxsp0r72h6x6j", "959472000000000000000000"},
	{"kii10jtnkmhlkqnprvng9yenqgr0jtujp0guk3pysp", "982062000000000000000000"},
	{"kii1meq2vy7rlnnceurju0uz9qeshfaaun0x5xsg06", "992286000000000000000000"},
	{"kii196ceqskhhyj93hejczj6py8f0am7vs3fykark7", "1004832000000000000000000"},
	{"kii1rt7arm6ckp0lcfuq9r0fmyl22urdcyfgfer35g", "1022706000000000000000000"},
	{"kii18fqfr2j7v96xy4lggvaca5cef586jp7pry27v0", "1038420000000000000000000"},
	{"kii1ehfns3qwnuhlunkhk5l2d0la8d8erenjn0482a", "1115262000000000000000000"},
	{"kii18ufxrsncyegu9qactah4hzrn0xmqqlkxr6p3z5", "1116000000000000000000000"},
	{"kii1qqyn3zg7g648pwc46y0depq8f82rj9400ulj4g", "1116000000000000000000000"},
	{"kii13u7hu5lscvdc2yqg0x5t2qj27m8fj8jw4ez046", "3247115883451428571420551"},
	{"kii1fc80es03yhle3xjpqp8e8pezl7an65h50fl2pm", "4496107929952857142857255"},
	{"kii1syezrzevu6ycshvgtm4sxtreh3pxvk0mtfe6rd", "5139906382822857142856976"},
	{"kii106tcwjead6wdj9xegyes80vfxd4da6sr4f5npu", "9000001000000000000000000"},
	{"kii1n4mskp6c83rzvl9eraddqdgwuqt6zec46qv06q", "36000001000000000000000000"},
}

// fund mints coins via tokenfactory (the same source app/apptesting's
// FundAcc helper uses) and sends them to addr.
func fund(t *testing.T, app *kiichain.KiichainApp, ctx sdk.Context, addr string, amount math.Int) {
	t.Helper()
	coins := sdk.NewCoins(sdk.NewCoin(denom, amount))
	require.NoError(t, app.BankKeeper.MintCoins(ctx, tokenfactorytypes.ModuleName, coins))
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToAccount(ctx, tokenfactorytypes.ModuleName, sdk.MustAccAddressFromBech32(addr), coins))
}

// fundStaging mints coins via tokenfactory and sends them into the "evm"
// module account by name. stagingAddr is a real module account, and bank's
// SendCoinsFromModuleToAccount rejects it as a recipient (blocked-address
// check), so module-to-module is the only way to fund it directly in a test.
func fundStaging(t *testing.T, app *kiichain.KiichainApp, ctx sdk.Context, amount math.Int) {
	t.Helper()
	coins := sdk.NewCoins(sdk.NewCoin(denom, amount))
	require.NoError(t, app.BankKeeper.MintCoins(ctx, tokenfactorytypes.ModuleName, coins))
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToModule(ctx, tokenfactorytypes.ModuleName, evmtypes.ModuleName, coins))
}

// totalPayouts sums payouts using math.Int so the huge amounts aren't
// hand-added (and hand-added wrong).
func totalPayouts(t *testing.T) math.Int {
	t.Helper()
	total := math.ZeroInt()
	for _, p := range payouts {
		amt, ok := math.NewIntFromString(p.amount)
		require.True(t, ok, "bad test fixture amount %q", p.amount)
		total = total.Add(amt)
	}
	return total
}

// TestCreateUpgradeHandler_RecoversAndRedistributesFunds runs the full
// three-stage recovery: sweep two attacker wallets into the evm module
// account (stagingAddr), pay out the fixed redistribution list, then sweep
// whatever remains to remainderAddr.
func TestCreateUpgradeHandler_RecoversAndRedistributesFunds(t *testing.T) {
	app, ctx := kiihelpers.SetupWithContext(t)
	ctx = ctx.WithChainID(v740.MainnetChainID).WithBlockHeight(v740.UpgradeHeight)

	remainder := math.NewIntWithDecimal(500, 18) // 500 KII left over, on purpose
	total := totalPayouts(t).Add(remainder)

	// Split the funding across two real attacker addresses to exercise the
	// sweep loop over more than one account.
	attacker2Amount := math.NewIntWithDecimal(1, 18) // 1 KII
	attacker1Amount := total.Sub(attacker2Amount)

	fund(t, app, ctx, attacker1, attacker1Amount)
	fund(t, app, ctx, attacker2, attacker2Amount)

	mm := app.GetModuleManager()
	handler := v740.CreateUpgradeHandler(mm, app.GetConfigurator(), &app.AppKeepers)
	vm, err := handler(ctx, upgradetypes.Plan{Name: v740.UpgradeName}, mm.GetVersionMap())
	require.NoError(t, err)
	require.NotNil(t, vm)

	// Attacker wallets end up empty.
	require.True(t, app.BankKeeper.GetAllBalances(ctx, sdk.MustAccAddressFromBech32(attacker1)).IsZero())
	require.True(t, app.BankKeeper.GetAllBalances(ctx, sdk.MustAccAddressFromBech32(attacker2)).IsZero())

	// Staging (the evm module account) passes everything through.
	require.True(t, app.BankKeeper.GetAllBalances(ctx, sdk.MustAccAddressFromBech32(stagingAddr)).IsZero())

	// Every payout landed exactly as specified.
	for _, p := range payouts {
		expected, ok := math.NewIntFromString(p.amount)
		require.True(t, ok)
		got := app.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(p.addr), denom)
		require.Equal(t, expected.String(), got.Amount.String(), "payout mismatch for %s", p.addr)
	}

	// Whatever was left over went to the remainder address.
	gotRemainder := app.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(remainderAddr), denom)
	require.Equal(t, remainder.String(), gotRemainder.Amount.String())

	// Freeze turns on only after recoverFunds, so the sweep above can succeed.
	require.True(t, blockedaddrs.IsEnabled(ctx, app.GetKey(banktypes.StoreKey)))
}

// TestCreateUpgradeHandler_RemainderExcludesPreexistingStagingBalance is a
// regression test: stagingAddr is the real "evm" module account, so it can
// hold akii that has nothing to do with this recovery. The remainder step
// must send only (swept - payouts), never staging's full balance.
func TestCreateUpgradeHandler_RemainderExcludesPreexistingStagingBalance(t *testing.T) {
	app, ctx := kiihelpers.SetupWithContext(t)
	ctx = ctx.WithChainID(v740.MainnetChainID).WithBlockHeight(v740.UpgradeHeight)

	// staging already holds akii unrelated to the incident before the
	// handler ever runs.
	preexisting := math.NewIntWithDecimal(777, 18)
	fundStaging(t, app, ctx, preexisting)

	extraRemainder := math.NewIntWithDecimal(500, 18)
	fund(t, app, ctx, attacker1, totalPayouts(t).Add(extraRemainder))

	mm := app.GetModuleManager()
	handler := v740.CreateUpgradeHandler(mm, app.GetConfigurator(), &app.AppKeepers)
	_, err := handler(ctx, upgradetypes.Plan{Name: v740.UpgradeName}, mm.GetVersionMap())
	require.NoError(t, err)

	// Only the swept-minus-payouts delta reached remainderAddr...
	gotRemainder := app.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(remainderAddr), denom)
	require.Equal(t, extraRemainder.String(), gotRemainder.Amount.String())

	// ...and staging's pre-existing balance was left exactly where it was.
	stagingBal := app.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(stagingAddr), denom)
	require.Equal(t, preexisting.String(), stagingBal.Amount.String())
}

// TestCreateUpgradeHandler_PanicsWhenStagingCannotCoverPayouts verifies the
// handler fails closed: if the swept balance can't cover the fixed payout
// list, it panics instead of sending partial or incorrect amounts.
func TestCreateUpgradeHandler_PanicsWhenStagingCannotCoverPayouts(t *testing.T) {
	app, ctx := kiihelpers.SetupWithContext(t)
	ctx = ctx.WithChainID(v740.MainnetChainID).WithBlockHeight(v740.UpgradeHeight)

	// Fund the attacker with far less than the fixed payout list requires.
	fund(t, app, ctx, attacker1, math.OneInt())

	mm := app.GetModuleManager()
	handler := v740.CreateUpgradeHandler(mm, app.GetConfigurator(), &app.AppKeepers)

	require.Panics(t, func() {
		_, _ = handler(ctx, upgradetypes.Plan{Name: v740.UpgradeName}, mm.GetVersionMap())
	})
}

// TestCreateUpgradeHandler_SkipsFundMovementWhenNotMainnet verifies that on
// any chain-id other than MainnetChainID (e.g. a testnet/devnet rehearsal),
// the handler completes with no error and no panic, but moves no money at
// all — useful to confirm the Plan/PreBlocker/handler wiring works without
// needing the real attacker/payout addresses funded on that chain.
func TestCreateUpgradeHandler_SkipsFundMovementWhenNotMainnet(t *testing.T) {
	app, ctx := kiihelpers.SetupWithContext(t) // default test chain-id, not v740.MainnetChainID
	ctx = ctx.WithBlockHeight(v740.UpgradeHeight)

	funded := math.NewIntWithDecimal(1, 18) // 1 KII
	fund(t, app, ctx, attacker1, funded)

	mm := app.GetModuleManager()
	handler := v740.CreateUpgradeHandler(mm, app.GetConfigurator(), &app.AppKeepers)

	vm, err := handler(ctx, upgradetypes.Plan{Name: v740.UpgradeName}, mm.GetVersionMap())
	require.NoError(t, err)
	require.NotNil(t, vm)

	// Nothing moved: the attacker still holds exactly what it was funded
	// with, and staging/remainder never received anything.
	got := app.BankKeeper.GetBalance(ctx, sdk.MustAccAddressFromBech32(attacker1), denom)
	require.Equal(t, funded.String(), got.Amount.String())
	require.True(t, app.BankKeeper.GetAllBalances(ctx, sdk.MustAccAddressFromBech32(stagingAddr)).IsZero())
	require.True(t, app.BankKeeper.GetAllBalances(ctx, sdk.MustAccAddressFromBech32(remainderAddr)).IsZero())
}
