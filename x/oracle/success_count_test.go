package oracle

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/kiichain/kiichain/v7/x/oracle/keeper"
	"github.com/kiichain/kiichain/v7/x/oracle/types"
	"github.com/kiichain/kiichain/v7/x/oracle/utils"
)

func TestValidatorSuccessCountWithBelowThresholdDenoms(t *testing.T) {
	t.Run("Validator votes all denoms", func(t *testing.T) {
		input, _ := SetUp(t)
		ctx := input.Ctx.WithBlockHeight(1)
		oracleKeeper := input.OracleKeeper

		params, err := oracleKeeper.Params.Get(ctx)
		require.NoError(t, err)

		params.VotePeriod = 1
		err = oracleKeeper.Params.Set(ctx, params)
		require.NoError(t, err)

		voteTargets := []string{utils.AtomDenom, utils.EthDenom, utils.KiiDenom}
		for _, denom := range voteTargets {
			err := oracleKeeper.VoteTarget.Set(ctx, denom, types.Denom{Name: denom})
			require.NoError(t, err)
		}

		valOperator := keeper.ValAddrs[0]

		vote, err := types.NewAggregateExchangeRateVote(
			types.ExchangeRateTuples{
				{Denom: utils.AtomDenom, ExchangeRate: math.LegacyNewDecWithPrec(100, 2)},
				{Denom: utils.EthDenom, ExchangeRate: math.LegacyNewDecWithPrec(100, 2)},
				{Denom: utils.KiiDenom, ExchangeRate: math.LegacyNewDecWithPrec(100, 2)},
			},
			valOperator,
		)
		require.NoError(t, err)

		err = oracleKeeper.AggregateExchangeRateVote.Set(ctx, valOperator, vote)
		require.NoError(t, err)

		initialCounter, err := oracleKeeper.VotePenaltyCounter.Get(ctx, valOperator)
		if err != nil {
			initialCounter = types.NewVotePenaltyCounter(0, 0, 0)
		}
		initialSuccessCount := initialCounter.SuccessCount

		err = EndBlocker(ctx, oracleKeeper)
		require.NoError(t, err)

		finalCounter, err := oracleKeeper.VotePenaltyCounter.Get(ctx, valOperator)
		require.NoError(t, err)

		require.Greater(t, finalCounter.SuccessCount, initialSuccessCount, "should get success count")
	})
}
