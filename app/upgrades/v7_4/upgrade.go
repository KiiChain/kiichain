package v740

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/kiichain/kiichain/v7/app/blockedaddrs"
	"github.com/kiichain/kiichain/v7/app/keepers"
)

// denom is the chain's native, 18-decimal token denom.
const denom = "akii"

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

		ctx.Logger().Info("Enabling bank send restriction for incident addresses...")
		blockedaddrs.Enable(ctx, k.GetKey(banktypes.StoreKey))

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
// three run sequentially against the same, still-uncommitted block context.
//
// On any chain-id other than MainnetChainID, no money moves at all: this
// logs and returns immediately, so a testnet/devnet rehearsal can confirm
// the Plan got scheduled and the handler ran.
func recoverFunds(ctx sdk.Context, k *keepers.AppKeepers) error {
	if ctx.ChainID() != MainnetChainID {
		ctx.Logger().Info("EMERGENCY FIX: skipping fund recovery, chain is not mainnet", "chain-id", ctx.ChainID())
		return nil
	}

	staging, err := sdk.AccAddressFromBech32(stagingAddr)
	if err != nil {
		return fmt.Errorf("invalid staging address: %w", err)
	}

	sweptAKII, err := sweepAttackerFunds(ctx, k, staging)
	if err != nil {
		return err
	}

	payoutTotal, err := totalPayoutAmount()
	if err != nil {
		return err
	}

	if sweptAKII.LT(payoutTotal) {
		return fmt.Errorf("insufficient recovered akii: swept=%s payouts=%s", sweptAKII, payoutTotal)
	}

	if err := distributePayouts(ctx, k, staging); err != nil {
		return err
	}

	remainderAmount := sweptAKII.Sub(payoutTotal)
	return distributeRemainder(ctx, k, staging, remainderAmount)
}

// totalPayoutAmount sums payouts with math.Int, so recoverFunds can compare
// it against what was actually swept before moving anything out of staging.
func totalPayoutAmount() (math.Int, error) {
	total := math.ZeroInt()
	for _, p := range payouts {
		amount, ok := math.NewIntFromString(p.amount)
		if !ok {
			return math.ZeroInt(), fmt.Errorf("invalid payout amount %q for %s", p.amount, p.addr)
		}
		total = total.Add(amount)
	}
	return total, nil
}

// sweepAttackerFunds moves every attacker wallet's akii balance into staging
// and returns the total amount actually swept. These are plain
// accounts/contracts, not vesting accounts, so a direct bank transfer of
// just the akii balance (not GetAllBalances) is all that's needed.
//
// Iterates blockedaddrs.SortedAttackerAddresses(), not the map directly:
// Go's map iteration order is randomized per process, and every validator
// must run this in the same order for identical behavior.
func sweepAttackerFunds(ctx sdk.Context, k *keepers.AppKeepers, staging sdk.AccAddress) (math.Int, error) {
	stagingBefore := k.BankKeeper.GetBalance(ctx, staging, denom)
	sweptAKII := math.ZeroInt()

	for _, addrStr := range blockedaddrs.SortedAttackerAddresses() {
		attackerAddr, err := sdk.AccAddressFromBech32(addrStr)
		if err != nil {
			return math.ZeroInt(), fmt.Errorf("invalid attacker address %s: %w", addrStr, err)
		}

		coin := k.BankKeeper.GetBalance(ctx, attackerAddr, denom)
		if coin.IsZero() {
			continue
		}

		if err := k.BankKeeper.SendCoins(ctx, attackerAddr, staging, sdk.NewCoins(coin)); err != nil {
			return math.ZeroInt(), fmt.Errorf("sweep from %s: %w", addrStr, err)
		}

		ctx.Logger().Info("emergency-fix: swept to staging", "addr", addrStr, "amount", coin.String())
		sweptAKII = sweptAKII.Add(coin.Amount)
	}

	// Validate the amount swept matches the difference in staging's akii
	// balance before and after.
	stagingAfter := k.BankKeeper.GetBalance(ctx, staging, denom)
	if !stagingAfter.Amount.Equal(stagingBefore.Amount.Add(sweptAKII)) {
		return math.ZeroInt(), fmt.Errorf("swept akii %s does not match staging balance change %s",
			sweptAKII.String(), stagingAfter.Amount.Sub(stagingBefore.Amount).String())
	}

	return sweptAKII, nil
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

// distributeRemainder sends exactly amount (swept minus paid out, computed
// by the caller) from staging to remainderAddr.
func distributeRemainder(ctx sdk.Context, k *keepers.AppKeepers, staging sdk.AccAddress, amount math.Int) error {
	remainder, err := sdk.AccAddressFromBech32(remainderAddr)
	if err != nil {
		return fmt.Errorf("invalid remainder address: %w", err)
	}

	if !amount.IsPositive() {
		return nil
	}

	coins := sdk.NewCoins(sdk.NewCoin(denom, amount))
	if err := k.BankKeeper.SendCoins(ctx, staging, remainder, coins); err != nil {
		return fmt.Errorf("remainder distribution: %w", err)
	}
	ctx.Logger().Info("emergency-fix: remainder distributed", "amount", coins.String())

	return nil
}
