package v740

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/kiichain/kiichain/v7/app/keepers"
)

// denom is the chain's native, 18-decimal token denom.
const denom = "akii"

// attackerAddrs holds the exploited accounts whose full balance must be
// clawed back into stagingAddr before redistribution.
var attackerAddrs = []string{
	"kii1peafvgnleuyl20tyfwnyvtvvwwvnaujxmqe5qe",
	"kii1vvwu93nya4ku9yds3v6ns2uq0fsmrnf4cf4yht",
	"kii1p3zmn7m6xq82jna6me04p8awt7k4u4k2alwu99",
	"kii1zamzjyjcwl0dejjvr90rtrwttxx2zhspqx4sm5",
	"kii1zlqdn7706xym7q3k2mdleag0uqjnhv8wu4sfsj",
	"kii1rehngnge8qn3ngszw4a8xxf2kqwmact602wtm8",
	"kii1y8m0qyc4n3m0rw4rcd7qnqahjh3r7p9uu3ert8",
	"kii19p9h2nw2y4fs85sgwgj2qrhhx7jmz6zujldh3n",
	"kii183h7rz9p4r8a7j8q2ardnrc7pgwnjp9jvhc8kq",
	"kii1gp7ar4hdlqntl5qkerm5n8mfxhqkegm76zqskr",
	"kii1gf9a9jjnnv8q3zcr8kczx0r5425zcfgpdw72tt",
	"kii1t7gzjh4gsrcuyfx3xdsem05chluqfsa43j9g54",
	"kii1wucgj4wxe0zvmmew2000cltc5qrl99eedtrzv4",
	"kii1syetlh585kl6yv5hmflhfehla5re7ay4um2skh",
	"kii1s7jw5ffqgjfn4ywxtgtq3nhpgcn05z28fsmkhm",
	"kii13ndtp734ntzx0jqvr80rlmj62slztqm9agzwce",
	"kii13umhqxg56cxwa9wv4gu6l9v4vyz9e70g4hupvn",
	"kii156expaxlymu5uhepe2dh647c9lu4slxpyml28q",
	"kii1k8vyx8d9ru2hk3k207p3az84xedjxz2gkdyle0",
	"kii16tr429kvneexqf4jttueuecm75ptc5l3gtj34q",
	"kii1mkhdmdgklsskgcgzz699nzhafav2hkea4qp2dj",
	"kii1a5v3eaeaugdh3vk57nlh8q8xcu7z46w0ttlrw9",
}

// stagingAddr is the chain's "evm" module account. BankKeeper.SendCoins moves balances
// by address and using this address as a plain intermediate
const stagingAddr = "kii1vqu8rska6swzdmnhf90zuv0xmelej4lq5el7zh"

// remainderAddr receives whatever is left in stagingAddr once every payout
// below has been sent
const remainderAddr = "kii1c6cgjmsx0ewl6j552sp06musutmfcvxcaq4n9h"

// payout is one redistribution leg out of stagingAddr.
type payout struct {
	addr   string
	amount string
}

// payouts lists the redistribution from stagingAddr, already in send order
// (the recovery plan requires paying out the last computed amount first and
// the first computed amount last) — distributePayouts just walks it forward.
var payouts = []payout{
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

// CreateUpgradeHandler creates the handler for the emergency fund recovery.
// It runs when the Plan scheduled from app.PreBlocker (see app.go) is applied
// by x/upgrade's own PreBlocker, in that same block
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	k *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(c context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		ctx := sdk.UnwrapSDKContext(c)
		ctx.Logger().Info("EMERGENCY FIX: starting funds recovery", "height", ctx.BlockHeight())

		if err := recoverFunds(ctx, k); err != nil {
			panic(fmt.Errorf("emergency fix failed: %w", err))
		}

		ctx.Logger().Info("EMERGENCY FIX: funds recovery completed successfully", "height", ctx.BlockHeight())

		vm, err := mm.RunMigrations(ctx, configurator, vm)
		if err != nil {
			return vm, err
		}

		// Log the upgrade completion
		ctx.Logger().Info("Upgrade v7.2.0 complete")
		return vm, nil
	}
}

// recoverFunds runs the three-stage recovery: sweep the exploited wallets
// into stagingAddr, pay out the computed amounts from stagingAddr, then send
// whatever remains in stagingAddr to remainderAddr. Each stage must finish
// before the next reads stagingAddr's balance, which holds here because all
// three run sequentially against the same, still-uncommitted block context
func recoverFunds(ctx sdk.Context, k *keepers.AppKeepers) error {
	staging, err := sdk.AccAddressFromBech32(stagingAddr)
	if err != nil {
		return fmt.Errorf("invalid staging address: %w", err)
	}

	if err := sweepAttackerFunds(ctx, k, staging); err != nil {
		return err
	}

	if err := distributePayouts(ctx, k, staging); err != nil {
		return err
	}

	return sweepRemainder(ctx, k, staging)
}

// sweepAttackerFunds moves every attacker wallet's full balance into staging
// These are plain accounts/contracts, not vesting accounts, so a direct
// bank transfer is all that's needed
func sweepAttackerFunds(ctx sdk.Context, k *keepers.AppKeepers, staging sdk.AccAddress) error {
	for _, addrStr := range attackerAddrs {
		attackerAddr, err := sdk.AccAddressFromBech32(addrStr)
		if err != nil {
			return fmt.Errorf("invalid attacker address %s: %w", addrStr, err)
		}

		balance := k.BankKeeper.GetAllBalances(ctx, attackerAddr)
		if balance.IsZero() {
			continue
		}
		if err := k.BankKeeper.SendCoins(ctx, attackerAddr, staging, balance); err != nil {
			return fmt.Errorf("sweep from %s: %w", addrStr, err)
		}
		ctx.Logger().Info("emergency-fix: swept to staging", "addr", addrStr, "amount", balance.String())
	}

	return nil
}

// distributePayouts sends each payout out of staging, in the order listed
func distributePayouts(ctx sdk.Context, k *keepers.AppKeepers, staging sdk.AccAddress) error {
	for _, p := range payouts {
		recipient, err := sdk.AccAddressFromBech32(p.addr)
		if err != nil {
			return fmt.Errorf("invalid payout address %s: %w", p.addr, err)
		}

		amount, ok := math.NewIntFromString(p.amount)
		if !ok {
			return fmt.Errorf("invalid payout amount %q for %s", p.amount, p.addr)
		}

		coins := sdk.NewCoins(sdk.NewCoin(denom, amount))
		if err := k.BankKeeper.SendCoins(ctx, staging, recipient, coins); err != nil {
			return fmt.Errorf("payout to %s: %w", p.addr, err)
		}
		ctx.Logger().Info("emergency-fix: payout sent", "addr", p.addr, "amount", coins.String())
	}

	return nil
}

// sweepRemainder sends whatever is left in staging to remainderAddr
func sweepRemainder(ctx sdk.Context, k *keepers.AppKeepers, staging sdk.AccAddress) error {
	remainder, err := sdk.AccAddressFromBech32(remainderAddr)
	if err != nil {
		return fmt.Errorf("invalid remainder address: %w", err)
	}

	balance := k.BankKeeper.GetAllBalances(ctx, staging)
	if balance.IsZero() {
		return nil
	}
	if err := k.BankKeeper.SendCoins(ctx, staging, remainder, balance); err != nil {
		return fmt.Errorf("remainder sweep: %w", err)
	}
	ctx.Logger().Info("emergency-fix: remainder swept", "amount", balance.String())

	return nil
}
