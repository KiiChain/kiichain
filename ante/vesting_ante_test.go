//go:build test

package ante_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/kiichain/kiichain/v7/ante"
	"github.com/kiichain/kiichain/v7/app/helpers"
)

func TestVestingAccountCreationDecorator(t *testing.T) {
	kiiApp := helpers.Setup(t)
	from := sdk.AccAddress("from________________")
	to := sdk.AccAddress("to__________________")
	coins := sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1)))

	testCases := []struct {
		name      string
		msgs      []sdk.Msg
		expectErr bool
	}{
		{
			name: "allow bank send",
			msgs: []sdk.Msg{&banktypes.MsgSend{
				FromAddress: from.String(),
				ToAddress:   to.String(),
				Amount:      coins,
			}},
			expectErr: false,
		},
		{
			name:      "block MsgCreateVestingAccount",
			msgs:      []sdk.Msg{sdkvesting.NewMsgCreateVestingAccount(from, to, coins, 1, false)},
			expectErr: true,
		},
		{
			name:      "block delayed MsgCreateVestingAccount",
			msgs:      []sdk.Msg{sdkvesting.NewMsgCreateVestingAccount(from, to, coins, 1, true)},
			expectErr: true,
		},
		{
			name: "block MsgCreatePeriodicVestingAccount",
			msgs: []sdk.Msg{sdkvesting.NewMsgCreatePeriodicVestingAccount(from, to, 1, []sdkvesting.Period{{
				Length: 1,
				Amount: coins,
			}})},
			expectErr: true,
		},
		{
			name:      "block MsgCreatePermanentLockedAccount",
			msgs:      []sdk.Msg{sdkvesting.NewMsgCreatePermanentLockedAccount(from, to, coins)},
			expectErr: true,
		},
		{
			name:      "block MsgCreateVestingAccount inside authz.MsgExec",
			msgs:      []sdk.Msg{newAuthzExec(sdkvesting.NewMsgCreateVestingAccount(from, to, coins, 1, false))},
			expectErr: true,
		},
		{
			name:      "block MsgCreatePeriodicVestingAccount inside nested authz.MsgExec",
			msgs:      []sdk.Msg{newAuthzExec(newAuthzExec(sdkvesting.NewMsgCreatePeriodicVestingAccount(from, to, 1, []sdkvesting.Period{{Length: 1, Amount: coins}})))},
			expectErr: true,
		},
		{
			name:      "block MsgCreatePermanentLockedAccount inside authz.MsgExec",
			msgs:      []sdk.Msg{newAuthzExec(sdkvesting.NewMsgCreatePermanentLockedAccount(from, to, coins))},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			txCfg := kiiApp.GetTxConfig()
			decorator := ante.NewVestingAccountCreationDecorator(kiiApp.AppCodec())

			txBuilder := txCfg.NewTxBuilder()
			require.NoError(t, txBuilder.SetMsgs(tc.msgs...))

			_, err := decorator.AnteHandle(sdk.Context{}, txBuilder.GetTx(), false,
				func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) { return ctx, nil })
			if tc.expectErr {
				require.Error(t, err)
				require.ErrorContains(t, err, "vesting account creation is disabled")
			} else {
				require.NoError(t, err)
			}
		})
	}
}
