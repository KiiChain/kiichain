package ante

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

func TestValidateVoteMsgsTypesAndAuthz(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	authz.RegisterInterfaces(registry)
	govv1.RegisterInterfaces(registry)
	govv1beta1.RegisterInterfaces(registry)
	banktypes.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	decorator := NewGovVoteDecorator(cdc, nil)

	original := minStakedTokens
	minStakedTokens = math.LegacyZeroDec()
	defer func() { minStakedTokens = original }()

	addr := sdk.AccAddress([]byte("voteraddress00000001"))

	voteV1 := govv1.NewMsgVote(addr, 1, govv1.OptionYes, "")
	weightedV1 := govv1.NewMsgVoteWeighted(addr, 1, govv1.NewNonSplitVoteOption(govv1.OptionYes), "")
	weightedBeta := govv1beta1.NewMsgVoteWeighted(addr, 1, govv1beta1.NewNonSplitVoteOption(govv1beta1.OptionYes))
	send := banktypes.NewMsgSend(addr, addr, sdk.NewCoins())

	badVoteV1 := &govv1.MsgVote{Voter: "not-bech32", ProposalId: 1, Option: govv1.OptionYes}
	badWeightedV1 := &govv1.MsgVoteWeighted{Voter: "not-bech32", ProposalId: 1, Options: govv1.NewNonSplitVoteOption(govv1.OptionYes)}
	badWeightedBeta := &govv1beta1.MsgVoteWeighted{Voter: "not-bech32", ProposalId: 1, Options: govv1beta1.NewNonSplitVoteOption(govv1beta1.OptionYes)}

	exec := func(msgs ...sdk.Msg) sdk.Msg {
		m := authz.NewMsgExec(addr, msgs)
		return &m
	}

	ctx := sdk.Context{}

	testCases := []struct {
		name      string
		msgs      []sdk.Msg
		expectErr bool
	}{
		{"weighted vote v1 routes through switch", []sdk.Msg{weightedV1}, false},
		{"weighted vote v1beta1 routes through switch", []sdk.Msg{weightedBeta}, false},
		{"weighted vote v1 invalid voter returns error", []sdk.Msg{badWeightedV1}, true},
		{"weighted vote v1beta1 invalid voter returns error", []sdk.Msg{badWeightedBeta}, true},
		{"non-vote message is ignored", []sdk.Msg{send}, false},
		{"authz exec wrapping a weighted vote", []sdk.Msg{exec(weightedV1)}, false},
		{"authz exec wrapping a non-vote", []sdk.Msg{exec(send)}, false},
		{"nested authz exec wrapping a vote", []sdk.Msg{exec(exec(voteV1))}, false},
		{"nested authz exec propagates inner error", []sdk.Msg{exec(exec(badVoteV1))}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := decorator.ValidateVoteMsgs(ctx, tc.msgs)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
