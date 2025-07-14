package types

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"

	erc20types "github.com/cosmos/evm/x/erc20/types"
)

// Erc20Keeper defines the expected interface for the Erc20 keeper
type Erc20Keeper interface {
	GetTokenPairID(ctx sdk.Context, token string) []byte
	GetTokenPair(ctx sdk.Context, id []byte) (erc20types.TokenPair, bool)
	ConvertERC20(
		goCtx context.Context,
		msg *erc20types.MsgConvertERC20,
	) (*erc20types.MsgConvertERC20Response, error)
	BalanceOf(
		ctx sdk.Context,
		abi abi.ABI,
		contract, account common.Address,
	) *big.Int
}

// BankKeeper defines the expected interface for the Bank keeper
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
