package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/kiichain/kiichain3/x/evm/types"
	"github.com/stretchr/testify/require"
)

func TestMessageSendValidate(t *testing.T) {
	fromAddr, err := sdk.AccAddressFromBech32("kii1v4mx6hmrda5kucnpwdjsqqqqqqqqqqqpkaxwqq")
	require.Nil(t, err)
	msg := types.NewMsgSend(fromAddr, common.HexToAddress("to"), sdk.Coins{sdk.Coin{
		Denom:  "kii",
		Amount: sdk.NewInt(1),
	}})
	require.Nil(t, msg.ValidateBasic())

	// No coins
	msg = types.NewMsgSend(fromAddr, common.HexToAddress("to"), sdk.Coins{})
	require.Error(t, msg.ValidateBasic())

	// Negative coins
	msg = types.NewMsgSend(fromAddr, common.HexToAddress("to"), sdk.Coins{sdk.Coin{
		Denom:  "kii",
		Amount: sdk.NewInt(-1),
	}})
	require.Error(t, msg.ValidateBasic())
}
