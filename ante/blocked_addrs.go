package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/app/blockedaddrs"
)

// IsBlockedAddr reports whether addr (hex or bech32) is on the incident deny list.
func IsBlockedAddr(addr string) bool {
	return blockedaddrs.IsBlockedAddr(addr)
}

// IsBlockedAccAddress reports whether addr's hex or bech32 form is denied.
func IsBlockedAccAddress(addr sdk.AccAddress) bool {
	return blockedaddrs.IsBlockedAccAddress(addr)
}
