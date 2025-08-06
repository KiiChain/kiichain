package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper is used to send and receive coins into module account
type BankKeeper interface {
	// Methods imported from bank should be defined here
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
