package ante

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type blockedMsgsTx struct {
	msgs []sdk.Msg
}

func (t blockedMsgsTx) GetMsgs() []sdk.Msg                    { return t.msgs }
func (t blockedMsgsTx) GetMsgsV2() ([]protov2.Message, error) { return nil, nil }

func TestBlockedAddrPairs(t *testing.T) {
	require.Len(t, blockedAddrPairs, 40)
	for _, pair := range blockedAddrPairs {
		require.True(t, IsBlockedAddr(pair[0]), pair[0])
		require.True(t, IsBlockedAddr(pair[1]), pair[1])
		require.True(t, IsBlockedAddr("0x"+pair[0][2:]), pair[0])

		raw, err := hex.DecodeString(pair[0][2:])
		require.NoError(t, err)
		require.True(t, IsBlockedAccAddress(sdk.AccAddress(raw)), pair[0])
	}
	require.False(t, IsBlockedAddr("0x0000000000000000000000000000000000000001"))
}

func TestBlockedAddrDecorator(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	authz.RegisterInterfaces(registry)
	banktypes.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)
	decorator := NewBlockedAddrDecorator(cdc)

	blocked := sdk.AccAddress(mustDecodeHex("0e7a96227fcf09f53d644ba6462d8c73993ef246"))
	allowed := sdk.AccAddress(mustDecodeHex("0000000000000000000000000000000000000001"))
	coins := sdk.NewCoins(sdk.NewCoin("akii", math.NewInt(1)))

	exec := func(msgs ...sdk.Msg) sdk.Msg {
		m := authz.NewMsgExec(allowed, msgs)
		return &m
	}

	testCases := []struct {
		name      string
		msgs      []sdk.Msg
		expectErr bool
	}{
		{
			name: "allow bank send between unlisted addrs",
			msgs: []sdk.Msg{banktypes.NewMsgSend(allowed, allowed, coins)},
		},
		{
			name:      "block bank send from listed addr",
			msgs:      []sdk.Msg{banktypes.NewMsgSend(blocked, allowed, coins)},
			expectErr: true,
		},
		{
			name:      "block bank send to listed addr",
			msgs:      []sdk.Msg{banktypes.NewMsgSend(allowed, blocked, coins)},
			expectErr: true,
		},
		{
			name:      "block bank send from listed addr inside authz.MsgExec",
			msgs:      []sdk.Msg{exec(banktypes.NewMsgSend(blocked, allowed, coins))},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decorator.AnteHandle(sdk.Context{}, blockedMsgsTx{msgs: tc.msgs}, false,
				func(ctx sdk.Context, _ sdk.Tx, _ bool) (sdk.Context, error) { return ctx, nil })
			if tc.expectErr {
				require.Error(t, err)
				require.ErrorContains(t, err, "address is blocked")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func mustDecodeHex(s string) []byte {
	bz, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return bz
}
