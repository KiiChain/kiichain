package oracle

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/kiichain/kiichain/v1/x/oracle/keeper"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/kiichain/kiichain/v1/x/oracle/utils"
	"github.com/stretchr/testify/require"
)

/* SetUp conditions:
voting target:
- uatom
- ueth
- uusd
- ukii

validators:
- val 1
- val 2
- val 3

Default Vote Threshold: 66.7%
bonded tokens: 30 ukii
ballot threshold: 20 power units

*/

func TestMidBlocker(t *testing.T) {
	t.Run("Success case - Exchange rate created on KVStore", func(t *testing.T) {
		// Reset blockchain state
		input, handler := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper

		// Sample exchange rate for the test
		oracleKeeper.DeleteVoteTargets(ctx)
		oracleKeeper.SetVoteTarget(ctx, utils.MicroAtomDenom)
		exchangeRate := randomAExchangeRate.String() + utils.MicroAtomDenom

		ctx = input.Ctx.WithBlockHeight(1)

		// Multiple validators submit their votes
		for i := 0; i < 3; i++ {
			voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[i], keeper.ValAddrs[i])
			_, err := handler(ctx, voteMsg)
			require.NoError(t, err)
		}

		MidBlocker(ctx, oracleKeeper)
		Endblocker(ctx, oracleKeeper)

		rate, lastUpdate, _, err := oracleKeeper.GetBaseExchangeRate(ctx, utils.MicroAtomDenom)
		require.NoError(t, err)
		require.Equal(t, randomAExchangeRate, rate)
		require.Equal(t, int64(1), lastUpdate.Int64()) // Last update block should be 1
	})

	t.Run("Success case - snapshot created", func(t *testing.T) {
		// Reset blockchain state
		input, handler := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper

		// Sample exchange rate for the test
		oracleKeeper.DeleteVoteTargets(ctx)
		oracleKeeper.SetVoteTarget(ctx, utils.MicroAtomDenom)
		exchangeRate := randomAExchangeRate.String() + utils.MicroAtomDenom

		ctx = input.Ctx.WithBlockHeight(1)

		// Multiple validators submit their votes
		for i := 0; i < 3; i++ {
			voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[i], keeper.ValAddrs[i])
			_, err := handler(ctx, voteMsg)
			require.NoError(t, err)
		}

		MidBlocker(ctx, oracleKeeper)
		Endblocker(ctx, oracleKeeper)

		// validate snapshot
		oracleKeeper.IteratePriceSnapshots(ctx, func(snapshot types.PriceSnapshot) bool {
			require.Equal(t, snapshot.PriceSnapshotItems[0].Denom, utils.MicroAtomDenom)
			return false
		})
	})

	t.Run("Error case - Ballot power less than threshold", func(t *testing.T) {
		// Reset blockchain state
		input, handler := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper

		// Sample exchange rate for the test
		oracleKeeper.DeleteVoteTargets(ctx)
		oracleKeeper.SetVoteTarget(ctx, utils.MicroAtomDenom)
		exchangeRate := randomAExchangeRate.String() + utils.MicroAtomDenom

		ctx = input.Ctx.WithBlockHeight(1)

		// Only one validator votes (insufficient power)
		voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[0], keeper.ValAddrs[0])
		_, err := handler(ctx, voteMsg)
		require.NoError(t, err)

		MidBlocker(ctx, oracleKeeper) // rate did not storage on KVStore, ballot below ballot threshold
		Endblocker(ctx, oracleKeeper)

		_, _, _, err = oracleKeeper.GetBaseExchangeRate(ctx, utils.MicroAtomDenom)
		require.Error(t, err)
	})

	t.Run("Validator does not vote - AbstainCount should increase", func(t *testing.T) {
		// Reset blockchain state
		input, handler := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper

		// Sample exchange rate for the test
		oracleKeeper.DeleteVoteTargets(ctx)
		oracleKeeper.SetVoteTarget(ctx, utils.MicroAtomDenom)
		exchangeRate := randomAExchangeRate.String() + utils.MicroAtomDenom

		ctx = input.Ctx.WithBlockHeight(1)

		// Only two validators vote, one validator abstains
		for i := 0; i < 2; i++ {
			voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[i], keeper.ValAddrs[i])
			_, err := handler(ctx, voteMsg)
			require.NoError(t, err)
		}

		MidBlocker(ctx, oracleKeeper)
		Endblocker(ctx, oracleKeeper)

		abstainCount := oracleKeeper.GetAbstainCount(ctx, keeper.ValAddrs[2])
		require.Equal(t, uint64(1), abstainCount) // Validator 2 has 1 abstained
	})

	t.Run("Validator votes out of acceptable range - Should count as Miss", func(t *testing.T) {
		// Reset blockchain state
		input, handler := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper

		// Sample exchange rate for the test
		oracleKeeper.DeleteVoteTargets(ctx)
		oracleKeeper.SetVoteTarget(ctx, utils.MicroAtomDenom)
		exchangeRate := randomAExchangeRate.String() + utils.MicroAtomDenom

		ctx = input.Ctx.WithBlockHeight(1)

		// Validator submits an incorrect exchange rate
		wrongRate := "100000000.0" + utils.MicroAtomDenom
		voteMsg := types.NewMsgAggregateExchangeRateVote(wrongRate, keeper.Addrs[0], keeper.ValAddrs[0])
		_, err := handler(ctx, voteMsg)
		require.NoError(t, err)

		// Other validators submit correct votes
		for i := 1; i < 3; i++ {
			voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[i], keeper.ValAddrs[i])
			_, err := handler(ctx, voteMsg)
			require.NoError(t, err)
		}

		MidBlocker(ctx, oracleKeeper)
		Endblocker(ctx, oracleKeeper)

		missCount := oracleKeeper.GetMissCount(ctx, keeper.ValAddrs[0])
		require.Equal(t, uint64(1), missCount) // Validator 0 has 1 Miss
	})

	t.Run("Verify upgrading the vote targets", func(t *testing.T) {
		// Reset blockchain state
		input, _ := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper

		ctx = input.Ctx.WithBlockHeight(1)

		// Modify the whitelist and apply it (ukii and uusdc will be 'new assets')
		oracleKeeper.DeleteVoteTargets(ctx)
		newWhitelist := types.DenomList{
			{Name: utils.MicroAtomDenom},
			{Name: utils.MicroEthDenom},
		}
		params := oracleKeeper.GetParams(ctx)
		params.Whitelist = newWhitelist
		oracleKeeper.SetParams(ctx, params)

		voteTargetsBefore := make(map[string]types.Denom)
		oracleKeeper.IterateVoteTargets(ctx, func(denom string, denomInfo types.Denom) bool {
			voteTargetsBefore[denom] = denomInfo
			return false
		})

		MidBlocker(ctx, oracleKeeper)

		voteTargetsAfter := make(map[string]types.Denom)
		oracleKeeper.IterateVoteTargets(ctx, func(denom string, denomInfo types.Denom) bool {
			voteTargetsAfter[denom] = denomInfo
			return false
		})

		// validate the vote target
		require.NotEqual(t, voteTargetsBefore, voteTargetsAfter)
		require.Len(t, voteTargetsAfter, 2) // Only uatom and ueth must be on the vote target

		_, err := oracleKeeper.GetVoteTarget(ctx, utils.MicroKiiDenom)
		require.Error(t, err)
		_, err = oracleKeeper.GetVoteTarget(ctx, utils.MicroUsdcDenom)
		require.Error(t, err)

	})

}

func TestOracleDrop(t *testing.T) {
	// Reset blockchain state
	input, handler := SetUp(t)
	ctx := input.Ctx
	oracleKeeper := input.OracleKeeper
	ctx = input.Ctx.WithBlockHeight(1)

	oracleKeeper.DeleteVoteTargets(ctx)
	oracleKeeper.SetVoteTarget(ctx, utils.MicroAtomDenom)
	input.OracleKeeper.SetBaseExchangeRate(ctx, utils.MicroAtomDenom, randomAExchangeRate)

	// Sample exchange rate for the test
	exchangeRate := randomAExchangeRate.String() + utils.MicroAtomDenom

	// simulate val 0 votation
	voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[0], keeper.ValAddrs[0])
	_, err := handler(ctx.WithBlockHeight(9), voteMsg)
	require.NoError(t, err)

	// Immediately swap halt after an illiquid oracle vote
	MidBlocker(ctx, oracleKeeper)
	Endblocker(ctx, oracleKeeper)

	rate, _, _, err := oracleKeeper.GetBaseExchangeRate(ctx, utils.MicroAtomDenom)
	require.NoError(t, err)
	require.True(t, randomAExchangeRate.Equal(rate))
}

func TestEndblocker(t *testing.T) {
	t.Run("Validator Jailed - success voting below min valid per window", func(t *testing.T) {
		// SetUp blockchain state
		input, _ := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper
		stakingKeeper := input.StakingKeeper

		// Simulate a validator with too many misses
		operator := keeper.ValAddrs[0]
		oracleKeeper.SetVotePenaltyCounter(ctx, operator, 15, 1, 5)

		// update MinValidPerWindow
		params := oracleKeeper.GetParams(ctx)
		params.MinValidPerWindow = math.LegacyNewDecWithPrec(50, 2) // 50%
		params.SlashFraction = math.LegacyNewDecWithPrec(50, 2)     // 50%
		oracleKeeper.SetParams(ctx, params)

		// Execute EndBlocker on the last block of slash window
		slashWindow := params.SlashWindow
		ctx = ctx.WithBlockHeight(int64(slashWindow) - 1)
		Endblocker(ctx, oracleKeeper)

		// Check if validator was jailed
		validator := oracleKeeper.StakingKeeper.Validator(ctx, operator)
		require.True(t, validator.IsJailed())

		// Check if validator was slashed (power reduced)
		slashedPower := validator.GetConsensusPower(stakingKeeper.PowerReduction(ctx))
		require.True(t, slashedPower < 10)

		// Check voting info deleted
		result := oracleKeeper.GetVotePenaltyCounter(ctx, operator)
		require.Empty(t, result)
	})

	t.Run("Validator not jailed", func(t *testing.T) {
		// SetUp blockchain state
		input, _ := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper
		stakingKeeper := input.StakingKeeper

		// Simulate a validator with too many misses
		operator := keeper.ValAddrs[0]
		oracleKeeper.SetVotePenaltyCounter(ctx, operator, 4, 5, 10)

		// update MinValidPerWindow
		params := oracleKeeper.GetParams(ctx)
		params.MinValidPerWindow = math.LegacyNewDecWithPrec(50, 2) // 50%
		params.SlashFraction = math.LegacyNewDecWithPrec(50, 2)     // 50%
		oracleKeeper.SetParams(ctx, params)

		// Execute EndBlocker on the last block of slash window
		slashWindow := params.SlashWindow
		ctx = ctx.WithBlockHeight(int64(slashWindow) - 1)
		Endblocker(ctx, oracleKeeper)

		// Check if validator was jailed
		validator := oracleKeeper.StakingKeeper.Validator(ctx, operator)
		require.False(t, validator.IsJailed())

		// vaidator must keep its voting power
		slashedPower := validator.GetConsensusPower(stakingKeeper.PowerReduction(ctx))
		require.True(t, slashedPower == 10) // voting power does not change

		// Check voting info deleted
		result := oracleKeeper.GetVotePenaltyCounter(ctx, operator)
		require.Empty(t, result)
	})

	t.Run("Success remove excess feeds", func(t *testing.T) {
		// SetUp blockchain state
		input, _ := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper

		// Simulate a validator with too many misses
		operator := keeper.ValAddrs[0]
		oracleKeeper.SetVotePenaltyCounter(ctx, operator, 4, 5, 10)

		// Agregate voting targets
		oracleKeeper.DeleteVoteTargets(ctx) // clean voting target list
		oracleKeeper.SetVoteTarget(ctx, utils.MicroAtomDenom)
		oracleKeeper.SetVoteTarget(ctx, utils.MicroEthDenom)

		// Aggregate base exchange rate
		oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroAtomDenom, math.LegacyNewDec(1))
		oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroEthDenom, math.LegacyNewDec(2))
		oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroKiiDenom, math.LegacyNewDec(3)) // extra denom

		// Execute EndBlocker on the last block of slash window
		params := oracleKeeper.GetParams(ctx)
		slashWindow := params.SlashWindow
		ctx = ctx.WithBlockHeight(int64(slashWindow) - 1)
		Endblocker(ctx, oracleKeeper)

		// Validate the successfull erased of the extra denoms
		oracleKeeper.IterateBaseExchangeRates(ctx, func(denom string, exchangeRate types.OracleExchangeRate) bool {
			require.True(t, denom != utils.MicroKiiDenom)
			return false
		})

	})
}
