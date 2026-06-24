//go:build test

package keepers_test

import (
	"math/big"
	"testing"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	evmtypes "github.com/cosmos/evm/x/vm/types"

	kiichain "github.com/kiichain/kiichain/v7/app"
	"github.com/kiichain/kiichain/v7/app/apptesting"
	apphelpers "github.com/kiichain/kiichain/v7/app/helpers"
	kiiparams "github.com/kiichain/kiichain/v7/app/params"
	tokenfactorytypes "github.com/kiichain/kiichain/v7/x/tokenfactory/types"
)

func newUnsignedEVMMsg(from sdk.AccAddress) *evmtypes.MsgEthereumTx {
	to := gethcommon.BytesToAddress(apptesting.RandomAccountAddress().Bytes())
	msg := evmtypes.NewTx(&evmtypes.EvmTxArgs{
		ChainID:  big.NewInt(int64(kiichain.KiichainID)),
		Nonce:    0,
		To:       &to,
		GasLimit: 21_000,
		GasPrice: big.NewInt(0),
		Amount:   big.NewInt(1),
	})
	msg.From = from.Bytes()
	return msg
}

func TestAuthzRouterRejectsDirectMsgEthereumTx(t *testing.T) {
	app, ctx := apphelpers.SetupWithContext(t)
	grantee := apptesting.RandomAccountAddress()

	_, err := app.AuthzKeeper.DispatchActions(ctx, grantee, []sdk.Msg{newUnsignedEVMMsg(grantee)})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not allowed to be executed through authz")
}

func TestAuthzRouterRejectsNestedMsgEthereumTx(t *testing.T) {
	app, ctx := apphelpers.SetupWithContext(t)
	grantee := apptesting.RandomAccountAddress()

	inner := authz.NewMsgExec(grantee, []sdk.Msg{newUnsignedEVMMsg(grantee)})

	_, err := app.AuthzKeeper.DispatchActions(ctx, grantee, []sdk.Msg{&inner})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not allowed to be executed through authz")
}

func TestAuthzRouterAllowsNonEVMMsg(t *testing.T) {
	app, ctx := apphelpers.SetupWithContext(t)
	grantee := apptesting.RandomAccountAddress()
	recipient := apptesting.RandomAccountAddress()

	fundAmount := sdkmath.NewInt(1_000)
	sendAmount := sdkmath.NewInt(100)
	coins := sdk.NewCoins(sdk.NewCoin(kiiparams.BaseDenom, fundAmount))
	require.NoError(t, app.BankKeeper.MintCoins(ctx, tokenfactorytypes.ModuleName, coins))
	require.NoError(t, app.BankKeeper.SendCoinsFromModuleToAccount(ctx, tokenfactorytypes.ModuleName, grantee, coins))

	sendMsg := banktypes.NewMsgSend(grantee, recipient, sdk.NewCoins(sdk.NewCoin(kiiparams.BaseDenom, sendAmount)))

	_, err := app.AuthzKeeper.DispatchActions(ctx, grantee, []sdk.Msg{sendMsg})
	require.NoError(t, err)
	require.Equal(t, sendAmount, app.BankKeeper.GetBalance(ctx, recipient, kiiparams.BaseDenom).Amount)
}
