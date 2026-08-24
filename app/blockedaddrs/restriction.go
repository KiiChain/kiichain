package blockedaddrs

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// SendRestriction rejects bank sends whose from or to address is on the
// incident list. It is registered on the bank keeper so Cosmos sends,
// precompile sends, and EVM native commits (mint/burn via SendCoins) are
// all covered.
func SendRestriction(_ context.Context, fromAddr, toAddr sdk.AccAddress, _ sdk.Coins) (sdk.AccAddress, error) {
	if IsBlockedAccAddress(fromAddr) {
		return nil, blockedSendErr(fromAddr.String())
	}
	if IsBlockedAccAddress(toAddr) {
		return nil, blockedSendErr(toAddr.String())
	}
	return toAddr, nil
}

func blockedSendErr(addr string) error {
	return errorsmod.Wrapf(errortypes.ErrUnauthorized, "address is blocked: %s", normalizeAddr(addr))
}

var _ banktypes.SendRestrictionFn = SendRestriction
