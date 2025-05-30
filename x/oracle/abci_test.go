package oracle

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	"github.com/kiichain/kiichain/v1/x/oracle/keeper"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/kiichain/kiichain/v1/x/oracle/utils"
)

/* SetUp conditions:
voting target:
- uatom
- ueth
- uusd
- akii

validators:
- val 1
- val 2
- val 3

Default Vote Threshold: 66.7%
bonded tokens: 30 akii
ballot threshold: 20 power units

*/

func TestMidBlocker(t *testing.T) {
	t.Run("Success case - Exchange rate created on KVStore", func(t *testing.T) {
		// Reset blockchain state
		input, msgServer := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper

		// Sample exchange rate for the test
		err := oracleKeeper.VoteTarget.Clear(ctx, nil)
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroAtomDenom, types.Denom{Name: utils.MicroAtomDenom})
		require.NoError(t, err)
		exchangeRate := randomAExchangeRate.String() + utils.MicroAtomDenom

		ctx = input.Ctx.WithBlockHeight(1)

		// Multiple validators submit their votes
		for i := 0; i < 3; i++ {
			voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[i], keeper.ValAddrs[i])
			_, err := msgServer.AggregateExchangeRateVote(ctx, voteMsg)
			require.NoError(t, err)
		}

		err = MidBlocker(ctx, oracleKeeper)
		require.NoError(t, err)
		err = Endblocker(ctx, oracleKeeper)
		require.NoError(t, err)

		exchangeRateResponse, err := oracleKeeper.GetBaseExchangeRate(ctx, utils.MicroAtomDenom)
		require.NoError(t, err)
		require.Equal(t, randomAExchangeRate, exchangeRateResponse.ExchangeRate)
		require.Equal(t, int64(1), exchangeRateResponse.LastUpdate.Int64()) // Last update block should be 1
	})

	t.Run("Success case - snapshot created", func(t *testing.T) {
		// Reset blockchain state
		input, msgServer := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper

		// Sample exchange rate for the test
		err := oracleKeeper.VoteTarget.Clear(ctx, nil)
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroAtomDenom, types.Denom{Name: utils.MicroAtomDenom})
		require.NoError(t, err)
		exchangeRate := randomAExchangeRate.String() + utils.MicroAtomDenom

		ctx = input.Ctx.WithBlockHeight(1)

		// Multiple validators submit their votes
		for i := 0; i < 3; i++ {
			voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[i], keeper.ValAddrs[i])
			_, err := msgServer.AggregateExchangeRateVote(ctx, voteMsg)
			require.NoError(t, err)
		}

		err = MidBlocker(ctx, oracleKeeper)
		require.NoError(t, err)
		err = Endblocker(ctx, oracleKeeper)
		require.NoError(t, err)

		// validate snapshot
		err = oracleKeeper.PriceSnapshot.Walk(ctx, nil, func(_ int64, snapshot types.PriceSnapshot) (bool, error) {
			require.Equal(t, snapshot.PriceSnapshotItems[0].Denom, utils.MicroAtomDenom)
			return false, nil
		})
		require.NoError(t, err)
	})

	t.Run("Error case - Ballot power less than threshold", func(t *testing.T) {
		// Reset blockchain state
		input, msgServer := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper

		// Sample exchange rate for the test
		err := oracleKeeper.VoteTarget.Clear(ctx, nil)
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroAtomDenom, types.Denom{Name: utils.MicroAtomDenom})
		require.NoError(t, err)
		exchangeRate := randomAExchangeRate.String() + utils.MicroAtomDenom

		ctx = input.Ctx.WithBlockHeight(1)

		// Only one validator votes (insufficient power)
		voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[0], keeper.ValAddrs[0])
		_, err = msgServer.AggregateExchangeRateVote(ctx, voteMsg)
		require.NoError(t, err)

		err = MidBlocker(ctx, oracleKeeper) // rate did not storage on KVStore, ballot below ballot threshold
		require.NoError(t, err)
		err = Endblocker(ctx, oracleKeeper)
		require.NoError(t, err)

		_, err = oracleKeeper.GetBaseExchangeRate(ctx, utils.MicroAtomDenom)
		require.Error(t, err)
	})

	t.Run("Validator does not vote - AbstainCount should increase", func(t *testing.T) {
		// Reset blockchain state
		input, msgServer := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper

		// Sample exchange rate for the test
		err := oracleKeeper.VoteTarget.Clear(ctx, nil)
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroAtomDenom, types.Denom{Name: utils.MicroAtomDenom})
		require.NoError(t, err)
		exchangeRate := randomAExchangeRate.String() + utils.MicroAtomDenom

		ctx = input.Ctx.WithBlockHeight(1)

		// Only two validators vote, one validator abstains
		for i := 0; i < 2; i++ {
			voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[i], keeper.ValAddrs[i])
			_, err := msgServer.AggregateExchangeRateVote(ctx, voteMsg)
			require.NoError(t, err)
		}

		err = MidBlocker(ctx, oracleKeeper)
		require.NoError(t, err)
		err = Endblocker(ctx, oracleKeeper)
		require.NoError(t, err)

		// Get the Vote Penalty Counter for the abstaining validator
		votePenaltyCounter, err := oracleKeeper.VotePenaltyCounter.Get(ctx, keeper.ValAddrs[2])
		require.NoError(t, err)

		require.EqualValues(t, uint64(1), votePenaltyCounter.AbstainCount) // Validator 2 has 1 abstained
	})

	t.Run("Validator votes out of acceptable range - Should count as Miss", func(t *testing.T) {
		// Reset blockchain state
		input, msgServer := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper

		// Sample exchange rate for the test
		err := oracleKeeper.VoteTarget.Clear(ctx, nil)
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroAtomDenom, types.Denom{Name: utils.MicroAtomDenom})
		require.NoError(t, err)
		exchangeRate := randomAExchangeRate.String() + utils.MicroAtomDenom

		ctx = input.Ctx.WithBlockHeight(1)

		// Validator submits an incorrect exchange rate
		wrongRate := "100000000.0" + utils.MicroAtomDenom
		voteMsg := types.NewMsgAggregateExchangeRateVote(wrongRate, keeper.Addrs[0], keeper.ValAddrs[0])
		_, err = msgServer.AggregateExchangeRateVote(ctx, voteMsg)
		require.NoError(t, err)

		// Other validators submit correct votes
		for i := 1; i < 3; i++ {
			voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[i], keeper.ValAddrs[i])
			_, err := msgServer.AggregateExchangeRateVote(ctx, voteMsg)
			require.NoError(t, err)
		}

		err = MidBlocker(ctx, oracleKeeper)
		require.NoError(t, err)
		err = Endblocker(ctx, oracleKeeper)
		require.NoError(t, err)

		// Get the Vote Penalty Counter for the abstaining validator
		votePenaltyCounter, err := oracleKeeper.VotePenaltyCounter.Get(ctx, keeper.ValAddrs[0])
		require.NoError(t, err)

		require.Equal(t, uint64(1), votePenaltyCounter.MissCount) // Validator 0 has 1 Miss
	})

	t.Run("Verify upgrading the vote targets", func(t *testing.T) {
		// Reset blockchain state
		input, _ := SetUp(t)
		oracleKeeper := input.OracleKeeper

		ctx := input.Ctx.WithBlockHeight(1)

		// Modify the whitelist and apply it (akii and uusdc will be 'new assets')
		err := oracleKeeper.VoteTarget.Clear(ctx, nil)
		require.NoError(t, err)
		newWhitelist := types.DenomList{
			{Name: utils.MicroAtomDenom},
			{Name: utils.MicroEthDenom},
		}
		params, err := oracleKeeper.Params.Get(ctx)
		require.NoError(t, err)
		params.Whitelist = newWhitelist
		err = oracleKeeper.Params.Set(ctx, params)
		require.NoError(t, err)

		voteTargetsBefore := make(map[string]types.Denom)
		err = oracleKeeper.VoteTarget.Walk(ctx, nil, func(denom string, denomInfo types.Denom) (bool, error) {
			voteTargetsBefore[denom] = denomInfo
			return false, nil
		})
		require.NoError(t, err)

		err = MidBlocker(ctx, oracleKeeper)
		require.NoError(t, err)

		voteTargetsAfter := make(map[string]types.Denom)
		err = oracleKeeper.VoteTarget.Walk(ctx, nil, func(denom string, denomInfo types.Denom) (bool, error) {
			voteTargetsAfter[denom] = denomInfo
			return false, nil
		})
		require.NoError(t, err)

		// validate the vote target
		require.NotEqual(t, voteTargetsBefore, voteTargetsAfter)
		require.Len(t, voteTargetsAfter, 2) // Only uatom and ueth must be on the vote target

		_, err = oracleKeeper.VoteTarget.Get(ctx, utils.MicroKiiDenom)
		require.Error(t, err)
		_, err = oracleKeeper.VoteTarget.Get(ctx, utils.MicroUsdcDenom)
		require.Error(t, err)
	})
}

func TestOracleDrop(t *testing.T) {
	// Reset blockchain state
	input, msgServer := SetUp(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx.WithBlockHeight(1)

	err := oracleKeeper.VoteTarget.Clear(ctx, nil)
	require.NoError(t, err)
	err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroAtomDenom, types.Denom{Name: utils.MicroAtomDenom})
	require.NoError(t, err)
	err = input.OracleKeeper.SetBaseExchangeRate(ctx, utils.MicroAtomDenom, randomAExchangeRate)
	require.NoError(t, err)

	// Sample exchange rate for the test
	exchangeRate := randomAExchangeRate.String() + utils.MicroAtomDenom

	// simulate val 0 votation
	voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[0], keeper.ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(ctx.WithBlockHeight(9), voteMsg)
	require.NoError(t, err)

	// Immediately swap halt after an illiquid oracle vote
	err = MidBlocker(ctx, oracleKeeper)
	require.NoError(t, err)
	err = Endblocker(ctx, oracleKeeper)
	require.NoError(t, err)

	exchangeRateRes, err := oracleKeeper.GetBaseExchangeRate(ctx, utils.MicroAtomDenom)
	require.NoError(t, err)
	require.True(t, randomAExchangeRate.Equal(exchangeRateRes.ExchangeRate))
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

		// Set the vote penalty counter for the validator
		err := oracleKeeper.VotePenaltyCounter.Set(input.Ctx, operator, types.NewVotePenaltyCounter(15, 1, 5))
		require.NoError(t, err)

		// update MinValidPerWindow
		params, err := oracleKeeper.Params.Get(ctx)
		require.NoError(t, err)
		params.MinValidPerWindow = math.LegacyNewDecWithPrec(50, 2) // 50%
		params.SlashFraction = math.LegacyNewDecWithPrec(50, 2)     // 50%
		err = oracleKeeper.Params.Set(ctx, params)
		require.NoError(t, err)

		// Execute EndBlocker on the last block of slash window
		slashWindow := params.SlashWindow
		ctx = ctx.WithBlockHeight(int64(slashWindow) - 1)
		err = Endblocker(ctx, oracleKeeper)
		require.NoError(t, err)

		// Check if validator was jailed
		validator, err := oracleKeeper.StakingKeeper.Validator(ctx, operator)
		require.NoError(t, err)
		require.True(t, validator.IsJailed())

		// Check if validator was slashed (power reduced)
		slashedPower := validator.GetConsensusPower(stakingKeeper.PowerReduction(ctx))
		require.True(t, slashedPower < 10)

		// Check voting info deleted
		result, err := oracleKeeper.VotePenaltyCounter.Get(ctx, operator)
		require.Empty(t, result)
		require.ErrorIs(t, err, collections.ErrNotFound)
	})

	t.Run("Validator not jailed", func(t *testing.T) {
		// SetUp blockchain state
		input, _ := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper
		stakingKeeper := input.StakingKeeper

		// Simulate a validator with too many misses
		operator := keeper.ValAddrs[0]

		// Set the vote penalty counter for the validator
		err := oracleKeeper.VotePenaltyCounter.Set(input.Ctx, operator, types.NewVotePenaltyCounter(4, 5, 10))
		require.NoError(t, err)

		// update MinValidPerWindow
		params, err := oracleKeeper.Params.Get(ctx)
		require.NoError(t, err)
		params.MinValidPerWindow = math.LegacyNewDecWithPrec(50, 2) // 50%
		params.SlashFraction = math.LegacyNewDecWithPrec(50, 2)     // 50%
		err = oracleKeeper.Params.Set(ctx, params)
		require.NoError(t, err)

		// Execute EndBlocker on the last block of slash window
		slashWindow := params.SlashWindow
		ctx = ctx.WithBlockHeight(int64(slashWindow) - 1)
		err = Endblocker(ctx, oracleKeeper)
		require.NoError(t, err)

		// Check if validator was jailed
		validator, err := oracleKeeper.StakingKeeper.Validator(ctx, operator)
		require.NoError(t, err)
		require.False(t, validator.IsJailed())

		// vaidator must keep its voting power
		slashedPower := validator.GetConsensusPower(stakingKeeper.PowerReduction(ctx))
		require.True(t, slashedPower == 10) // voting power does not change

		// Check voting info deleted
		result, err := oracleKeeper.VotePenaltyCounter.Get(ctx, operator)
		require.Empty(t, result)
		require.ErrorIs(t, err, collections.ErrNotFound)
	})

	t.Run("Success remove excess feeds", func(t *testing.T) {
		// SetUp blockchain state
		input, _ := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper

		// Simulate a validator with too many misses
		operator := keeper.ValAddrs[0]

		// Set the vote penalty counter for the validator
		err := oracleKeeper.VotePenaltyCounter.Set(input.Ctx, operator, types.NewVotePenaltyCounter(4, 5, 10))
		require.NoError(t, err)

		// Aggregate voting targets
		err = oracleKeeper.VoteTarget.Clear(ctx, nil) // clean voting target list
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroAtomDenom, types.Denom{Name: utils.MicroAtomDenom})
		require.NoError(t, err)

		err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroEthDenom, types.Denom{Name: utils.MicroEthDenom})
		require.NoError(t, err)

		// Aggregate base exchange rate
		err = oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroAtomDenom, math.LegacyNewDec(1))
		require.NoError(t, err)
		err = oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroEthDenom, math.LegacyNewDec(2))
		require.NoError(t, err)
		err = oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroKiiDenom, math.LegacyNewDec(3)) // extra denom
		require.NoError(t, err)

		// Execute EndBlocker on the last block of slash window
		params, err := oracleKeeper.Params.Get(ctx)
		require.NoError(t, err)
		slashWindow := params.SlashWindow
		ctx = ctx.WithBlockHeight(int64(slashWindow) - 1)
		err = Endblocker(ctx, oracleKeeper)
		require.NoError(t, err)

		// Validate the successful erased of the extra denoms
		oracleKeeper.IterateBaseExchangeRates(ctx, func(denom string, exchangeRate types.OracleExchangeRate) (bool, error) {
			require.True(t, denom != utils.MicroKiiDenom)
			return false, nil
		})
	})
}
