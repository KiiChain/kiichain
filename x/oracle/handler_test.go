package oracle

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/kiichain/kiichain/v1/x/oracle/keeper"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/kiichain/kiichain/v1/x/oracle/utils"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func TestOracleFilters(t *testing.T) {
	// Prepare env
	input, handler := SetUp(t)
	ctx := input.Ctx
	oracleKeeper := input.OracleKeeper

	// set voting targets
	oracleKeeper.DeleteVoteTargets(ctx)
	oracleKeeper.SetVoteTarget(ctx, utils.MicroAtomDenom)

	t.Run("Non-oracle message received", func(t *testing.T) {
		bankMsg := &banktypes.MsgSend{}
		_, err := handler(ctx, bankMsg)
		require.Error(t, err)
	})

	t.Run("oracle message received", func(t *testing.T) {
		exchangeRate := randomAExchangeRate.String() + utils.MicroAtomDenom
		msg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[0], keeper.ValAddrs[0])
		_, err := handler(ctx, msg)
		require.NoError(t, err)
	})

	t.Run("non'validator sent an oracle message", func(t *testing.T) {
		nonValidatorPub := secp256k1.GenPrivKey().PubKey()
		nonValidatorAddr := nonValidatorPub.Address()
		exchangeRate := randomAExchangeRate.String() + utils.MicroAtomDenom
		voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, sdk.AccAddress(nonValidatorAddr), sdk.ValAddress(nonValidatorAddr))
		require.Panics(t, func() { handler(ctx, voteMsg) })
	})
}

func TestMsgDelegateFeedConsent(t *testing.T) {
	// Prepare env
	input, handler := SetUp(t)
	ctx := input.Ctx
	oracleKeeper := input.OracleKeeper

	// set voting targets
	oracleKeeper.DeleteVoteTargets(ctx)
	oracleKeeper.SetVoteTarget(ctx, utils.MicroAtomDenom)

	t.Run("empty message", func(t *testing.T) {
		msg := &types.MsgDelegateFeedConsent{}
		_, err := handler(ctx, msg)
		require.Error(t, err)
	})

	t.Run("success vote without delegation", func(t *testing.T) {
		exchangeRate := randomAExchangeRate.String() + utils.MicroAtomDenom
		msg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[0], keeper.ValAddrs[0])
		_, err := handler(ctx, msg)
		require.NoError(t, err)
	})

	t.Run("fail vote with failed delegation ", func(t *testing.T) {
		exchangeRate := randomAExchangeRate.String() + utils.MicroAtomDenom
		msg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[1], keeper.ValAddrs[0])
		_, err := handler(ctx, msg)
		require.Error(t, err)
	})

	t.Run("success vote with success delegation", func(t *testing.T) {
		// delegate vote
		msgDelegate := types.NewMsgDelegateFeedConsent(keeper.ValAddrs[1], keeper.Addrs[0])
		_, err := handler(ctx, msgDelegate)
		require.NoError(t, err)

		// vote
		exchangeRate := randomAExchangeRate.String() + utils.MicroAtomDenom
		msgVote := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[0], keeper.ValAddrs[1])
		_, err = handler(ctx, msgVote)
		require.NoError(t, err)
	})
}

func TestMsgAggregateExchangeRateVote(t *testing.T) {
	// Prepare env
	input, handler := SetUp(t)
	ctx := input.Ctx

	t.Run("success delegation", func(t *testing.T) {
		msgDelegate := types.NewMsgDelegateFeedConsent(keeper.ValAddrs[1], keeper.Addrs[0])
		_, err := handler(ctx, msgDelegate)
		require.NoError(t, err)
	})

	t.Run("fail delegation - empty addr", func(t *testing.T) {
		msgDelegate := types.NewMsgDelegateFeedConsent(keeper.ValAddrs[0], sdk.AccAddress{})
		_, err := handler(ctx, msgDelegate)
		require.Error(t, err)
	})

	t.Run("fail delegation - validator not registered", func(t *testing.T) {
		nonValidatorPub := secp256k1.GenPrivKey().PubKey()
		nonValidatorAddr := nonValidatorPub.Address()
		msgDelegate := types.NewMsgDelegateFeedConsent(sdk.ValAddress(nonValidatorAddr), sdk.AccAddress{})
		_, err := handler(ctx, msgDelegate)
		require.Error(t, err)
	})
}
