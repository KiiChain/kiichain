//go:build test

package ante_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kiichain/kiichain/v7/ante"
	"github.com/kiichain/kiichain/v7/app/helpers"
)

func TestGovVoteDecoratorWeightedAndNestedAuthz(t *testing.T) {
	kiiApp := helpers.Setup(t)
	ctx := kiiApp.NewUncachedContext(true, tmproto.Header{})
	decorator := ante.NewGovVoteDecorator(kiiApp.AppCodec(), kiiApp.StakingKeeper)
	stakingKeeper := kiiApp.StakingKeeper

	ante.SetMinStakedTokens(math.LegacyNewDec(1000000))

	validators, err := stakingKeeper.GetAllValidators(ctx)
	require.NoError(t, err)
	valAddr, err := stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[0].GetOperator())
	require.NoError(t, err)
	valAddr = sdk.ValAddress(valAddr)

	addr, err := kiiApp.AccountKeeper.Accounts.Indexes.Number.MatchExact(ctx, 0)
	require.NoError(t, err)
	delegator, err := sdk.AccAddressFromBech32(addr.String())
	require.NoError(t, err)

	unbondAll := func() {
		delegations, err := stakingKeeper.GetAllDelegatorDelegations(ctx, delegator)
		require.NoError(t, err)
		for _, del := range delegations {
			vAddr, err := sdk.ValAddressFromBech32(del.GetValidatorAddr())
			require.NoError(t, err)
			_, _, err = stakingKeeper.Undelegate(ctx, delegator, vAddr, del.GetShares())
			require.NoError(t, err)
		}
	}

	delegate := func(amount math.Int) {
		val, err := stakingKeeper.GetValidator(ctx, valAddr)
		require.NoError(t, err)
		_, err = stakingKeeper.Delegate(ctx, delegator, amount, stakingtypes.Unbonded, val, true)
		require.NoError(t, err)
	}

	exec := func(msgs ...sdk.Msg) sdk.Msg {
		m := authz.NewMsgExec(delegator, msgs)
		return &m
	}

	vote := govv1.NewMsgVote(delegator, 0, govv1.VoteOption_VOTE_OPTION_YES, "")
	voteBeta := govv1beta1.NewMsgVote(delegator, 0, govv1beta1.OptionYes)
	weighted := govv1.NewMsgVoteWeighted(delegator, 0, govv1.NewNonSplitVoteOption(govv1.OptionYes), "")
	weightedBeta := govv1beta1.NewMsgVoteWeighted(delegator, 0, govv1beta1.NewNonSplitVoteOption(govv1beta1.OptionYes))
	send := banktypes.NewMsgSend(delegator, delegator, sdk.NewCoins())

	t.Run("zero stake rejects votes but allows non-votes", func(t *testing.T) {
		unbondAll()

		require.NoError(t, decorator.ValidateVoteMsgs(ctx, []sdk.Msg{send}))
		require.NoError(t, decorator.ValidateVoteMsgs(ctx, []sdk.Msg{exec(send)}))
		require.NoError(t, decorator.ValidateVoteMsgs(ctx, []sdk.Msg{exec(exec(send))}))

		require.Error(t, decorator.ValidateVoteMsgs(ctx, []sdk.Msg{weighted}))
		require.Error(t, decorator.ValidateVoteMsgs(ctx, []sdk.Msg{weightedBeta}))

		require.Error(t, decorator.ValidateVoteMsgs(ctx, []sdk.Msg{exec(vote)}))
		require.Error(t, decorator.ValidateVoteMsgs(ctx, []sdk.Msg{exec(voteBeta)}))
		require.Error(t, decorator.ValidateVoteMsgs(ctx, []sdk.Msg{exec(weighted)}))
		require.Error(t, decorator.ValidateVoteMsgs(ctx, []sdk.Msg{exec(weightedBeta)}))

		require.Error(t, decorator.ValidateVoteMsgs(ctx, []sdk.Msg{exec(exec(vote))}))
		require.Error(t, decorator.ValidateVoteMsgs(ctx, []sdk.Msg{exec(exec(exec(weighted)))}))
	})

	t.Run("sufficient stake allows votes through every path", func(t *testing.T) {
		unbondAll()
		delegate(math.NewInt(10000000))

		require.NoError(t, decorator.ValidateVoteMsgs(ctx, []sdk.Msg{vote}))
		require.NoError(t, decorator.ValidateVoteMsgs(ctx, []sdk.Msg{weighted}))
		require.NoError(t, decorator.ValidateVoteMsgs(ctx, []sdk.Msg{weightedBeta}))
		require.NoError(t, decorator.ValidateVoteMsgs(ctx, []sdk.Msg{exec(vote)}))
		require.NoError(t, decorator.ValidateVoteMsgs(ctx, []sdk.Msg{exec(weighted)}))
		require.NoError(t, decorator.ValidateVoteMsgs(ctx, []sdk.Msg{exec(exec(vote))}))
	})
}
