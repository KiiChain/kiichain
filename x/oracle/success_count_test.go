package oracle

import (
	"testing"
	"github.com/stretchr/testify/require"
	"cosmossdk.io/math"
	"github.com/kiichain/kiichain/v7/x/oracle/keeper"
	"github.com/kiichain/kiichain/v7/x/oracle/types"
)

func TestValidatorSuccessCountWithBelowThresholdDenoms(t *testing.T) {
	input, _ := SetUp(t)
	ctx := input.Ctx.WithBlockHeight(1)
	k := input.OracleKeeper
	
	k.VoteTarget.Set(ctx, "atom", types.Denom{Name: "atom"})
	k.VoteTarget.Set(ctx, "eth", types.Denom{Name: "eth"})
	k.VoteTarget.Set(ctx, "usdt", types.Denom{Name: "usdt"})
	
	valAddr := keeper.ValAddrs[0]
	vote, _ := types.NewAggregateExchangeRateVote(
		types.ExchangeRateTuples{
			{Denom: "atom", ExchangeRate: math.LegacyNewDec(100)},
			{Denom: "eth", ExchangeRate: math.LegacyNewDec(100)},
			{Denom: "usdt", ExchangeRate: math.LegacyNewDec(100)},
		},
		valAddr,
	)
	k.AggregateExchangeRateVote.Set(ctx, valAddr, vote)
	
	initialCounter := types.NewVotePenaltyCounter(0, 0, 0)
	k.VotePenaltyCounter.Set(ctx, valAddr, initialCounter)
	
	params, _ := k.Params.Get(ctx)
	params.VoteThreshold = math.LegacyNewDecWithPrec(99, 2)
	params.VotePeriod = 1
	k.Params.Set(ctx, params)
	
	initial, _ := k.VotePenaltyCounter.Get(ctx, valAddr)
	initialSuccess := initial.SuccessCount
	
	EndBlocker(ctx, k)
	
	final, _ := k.VotePenaltyCounter.Get(ctx, valAddr)
	
	require.Greater(t, final.SuccessCount, initialSuccess)
}
