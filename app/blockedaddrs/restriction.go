package blockedaddrs

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// enabledKey is stored in the bank KV store. Chosen to sit outside the
// x/bank collections prefixes.
var enabledKey = []byte{0xF1, 'i', 'n', 'c', 'i', 'd', 'e', 'n', 't', '-', 'b', 'l', 'o', 'c', 'k'}

// Enable turns on the incident send restriction. Called from the v7.4.0
// upgrade handler after fund recovery.
func Enable(ctx sdk.Context, key storetypes.StoreKey) {
	ctx.KVStore(key).Set(enabledKey, []byte{1})
}

// IsEnabled reports whether the incident send restriction has been turned on.
func IsEnabled(ctx sdk.Context, key storetypes.StoreKey) bool {
	if key == nil {
		return false
	}
	return ctx.KVStore(key).Has(enabledKey)
}

// NewSendRestriction returns a bank send hook that no-ops until Enable is
// written, then rejects sends whose from or to address is on the incident list.
func NewSendRestriction(key storetypes.StoreKey) banktypes.SendRestrictionFn {
	return func(ctx context.Context, fromAddr, toAddr sdk.AccAddress, _ sdk.Coins) (sdk.AccAddress, error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		if !IsEnabled(sdkCtx, key) {
			return toAddr, nil
		}
		if IsBlockedAccAddress(fromAddr) {
			return nil, blockedSendErr(fromAddr.String())
		}
		if IsBlockedAccAddress(toAddr) {
			return nil, blockedSendErr(toAddr.String())
		}
		return toAddr, nil
	}
}

func blockedSendErr(addr string) error {
	return errorsmod.Wrapf(errortypes.ErrUnauthorized, "address is blocked: %s", addr)
}
