package ante

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkvesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestVestingAccountCreationDecorator(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	authz.RegisterInterfaces(registry)
	sdkvesting.RegisterInterfaces(registry)
	banktypes.RegisterInterfaces(registry)
	decorator := NewVestingAccountCreationDecorator(codec.NewProtoCodec(registry))

	from := sdk.AccAddress("from________________")
	to := sdk.AccAddress("to__________________")
	coins := sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1)))

	exec := func(msgs ...sdk.Msg) sdk.Msg {
		m := authz.NewMsgExec(from, msgs)
		return &m
	}

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
			msgs:      []sdk.Msg{exec(sdkvesting.NewMsgCreateVestingAccount(from, to, coins, 1, false))},
			expectErr: true,
		},
		{
			name: "block MsgCreatePeriodicVestingAccount inside nested authz.MsgExec",
			msgs: []sdk.Msg{exec(exec(sdkvesting.NewMsgCreatePeriodicVestingAccount(from, to, 1, []sdkvesting.Period{{
				Length: 1,
				Amount: coins,
			}})))},
			expectErr: true,
		},
		{
			name:      "block MsgCreatePermanentLockedAccount inside authz.MsgExec",
			msgs:      []sdk.Msg{exec(sdkvesting.NewMsgCreatePermanentLockedAccount(from, to, coins))},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := decorator.validateMsgs(tc.msgs)
			if tc.expectErr {
				require.Error(t, err)
				require.ErrorContains(t, err, "vesting account creation is disabled")
			} else {
				require.NoError(t, err)
			}
		})
	}
}
