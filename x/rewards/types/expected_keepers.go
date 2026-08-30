package types

import (
	context "context"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper is used to send and receive coins into module account
type BankKeeper interface {
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// StakingKeeper provides the bonded ratio used to drive the emission curve.
type StakingKeeper interface {
	BondedRatio(ctx context.Context) (math.LegacyDec, error)
}
