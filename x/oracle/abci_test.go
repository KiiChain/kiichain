package oracle

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	"github.com/kiichain/kiichain/v7/x/oracle/keeper"
	"github.com/kiichain/kiichain/v7/x/oracle/types"
	"github.com/kiichain/kiichain/v7/x/oracle/utils"
)

/* SetUp conditions:
voting target:
- atom
- eth
- usd
- kii

validators:
- val 1
- val 2
- val 3

Default Vote Threshold: 66.7%
bonded tokens: 30 kii
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
		err = oracleKeeper.VoteTarget.Set(ctx, utils.AtomDenom, types.Denom{Name: utils.AtomDenom})
		require.NoError(t, err)
		exchangeRate := randomAExchangeRate.String() + utils.AtomDenom

		ctx = input.Ctx.WithBlockHeight(1)

		// Multiple validators submit their votes
		for i := 0; i < 3; i++ {
			voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[i], keeper.ValAddrs[i])
			_, err := msgServer.AggregateExchangeRateVote(ctx, voteMsg)
			require.NoError(t, err)
		}

		err = EndBlocker(ctx, oracleKeeper)
		require.NoError(t, err)
		err = BeginBlocker(ctx, oracleKeeper)
		require.NoError(t, err)

		exchangeRateResponse, err := oracleKeeper.ExchangeRate.Get(ctx, utils.AtomDenom)
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
		err = oracleKeeper.VoteTarget.Set(ctx, utils.AtomDenom, types.Denom{Name: utils.AtomDenom})
		require.NoError(t, err)
		exchangeRate := randomAExchangeRate.String() + utils.AtomDenom

		ctx = input.Ctx.WithBlockHeight(1)

		// Multiple validators submit their votes
		for i := 0; i < 3; i++ {
			voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[i], keeper.ValAddrs[i])
			_, err := msgServer.AggregateExchangeRateVote(ctx, voteMsg)
			require.NoError(t, err)
		}

		err = EndBlocker(ctx, oracleKeeper)
		require.NoError(t, err)
		err = BeginBlocker(ctx, oracleKeeper)
		require.NoError(t, err)

		// validate snapshot
		err = oracleKeeper.PriceSnapshot.Walk(ctx, nil, func(_ int64, snapshot types.PriceSnapshot) (bool, error) {
			require.Equal(t, snapshot.PriceSnapshotItems[0].Denom, utils.AtomDenom)
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
		err = oracleKeeper.VoteTarget.Set(ctx, utils.AtomDenom, types.Denom{Name: utils.AtomDenom})
		require.NoError(t, err)
		exchangeRate := randomAExchangeRate.String() + utils.AtomDenom

		ctx = input.Ctx.WithBlockHeight(1)

		// Only one validator votes (insufficient power)
		voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[0], keeper.ValAddrs[0])
		_, err = msgServer.AggregateExchangeRateVote(ctx, voteMsg)
		require.NoError(t, err)

		err = EndBlocker(ctx, oracleKeeper) // rate did not storage on KVStore, ballot below ballot threshold
		require.NoError(t, err)
		err = BeginBlocker(ctx, oracleKeeper)
		require.NoError(t, err)

		_, err = oracleKeeper.ExchangeRate.Get(ctx, utils.AtomDenom)
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
		err = oracleKeeper.VoteTarget.Set(ctx, utils.AtomDenom, types.Denom{Name: utils.AtomDenom})
		require.NoError(t, err)
		exchangeRate := randomAExchangeRate.String() + utils.AtomDenom

		ctx = input.Ctx.WithBlockHeight(1)

		// Only two validators vote, one validator abstains
		for i := 0; i < 2; i++ {
			voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[i], keeper.ValAddrs[i])
			_, err := msgServer.AggregateExchangeRateVote(ctx, voteMsg)
			require.NoError(t, err)
		}

		err = EndBlocker(ctx, oracleKeeper)
		require.NoError(t, err)
		err = BeginBlocker(ctx, oracleKeeper)
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
		err = oracleKeeper.VoteTarget.Set(ctx, utils.AtomDenom, types.Denom{Name: utils.AtomDenom})
		require.NoError(t, err)
		exchangeRate := randomAExchangeRate.String() + utils.AtomDenom

		ctx = input.Ctx.WithBlockHeight(1)

		// Validator submits an incorrect exchange rate
		wrongRate := "100000000.0" + utils.AtomDenom
		voteMsg := types.NewMsgAggregateExchangeRateVote(wrongRate, keeper.Addrs[0], keeper.ValAddrs[0])
		_, err = msgServer.AggregateExchangeRateVote(ctx, voteMsg)
		require.NoError(t, err)

		// Other validators submit correct votes
		for i := 1; i < 3; i++ {
			voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[i], keeper.ValAddrs[i])
			_, err := msgServer.AggregateExchangeRateVote(ctx, voteMsg)
			require.NoError(t, err)
		}

		err = EndBlocker(ctx, oracleKeeper)
		require.NoError(t, err)
		err = BeginBlocker(ctx, oracleKeeper)
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

		// Modify the whitelist and apply it (kii and usdc will be 'new assets')
		err := oracleKeeper.VoteTarget.Clear(ctx, nil)
		require.NoError(t, err)
		newWhitelist := types.DenomList{
			{Name: utils.AtomDenom},
			{Name: utils.EthDenom},
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

		err = EndBlocker(ctx, oracleKeeper)
		require.NoError(t, err)

		voteTargetsAfter := make(map[string]types.Denom)
		err = oracleKeeper.VoteTarget.Walk(ctx, nil, func(denom string, denomInfo types.Denom) (bool, error) {
			voteTargetsAfter[denom] = denomInfo
			return false, nil
		})
		require.NoError(t, err)

		// validate the vote target
		require.NotEqual(t, voteTargetsBefore, voteTargetsAfter)
		require.Len(t, voteTargetsAfter, 2) // Only atom and eth must be on the vote target

		_, err = oracleKeeper.VoteTarget.Get(ctx, utils.KiiDenom)
		require.Error(t, err)
		_, err = oracleKeeper.VoteTarget.Get(ctx, utils.UsdcDenom)
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
	err = oracleKeeper.VoteTarget.Set(ctx, utils.AtomDenom, types.Denom{Name: utils.AtomDenom})
	require.NoError(t, err)
	err = input.OracleKeeper.SetBaseExchangeRateWithDefault(ctx, utils.AtomDenom, randomAExchangeRate)
	require.NoError(t, err)

	// Sample exchange rate for the test
	exchangeRate := randomAExchangeRate.String() + utils.AtomDenom

	// simulate val 0 votation
	voteMsg := types.NewMsgAggregateExchangeRateVote(exchangeRate, keeper.Addrs[0], keeper.ValAddrs[0])
	_, err = msgServer.AggregateExchangeRateVote(ctx.WithBlockHeight(9), voteMsg)
	require.NoError(t, err)

	// Immediately swap halt after an illiquid oracle vote
	err = EndBlocker(ctx, oracleKeeper)
	require.NoError(t, err)
	err = BeginBlocker(ctx, oracleKeeper)
	require.NoError(t, err)

	exchangeRateRes, err := oracleKeeper.ExchangeRate.Get(ctx, utils.AtomDenom)
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
		err = BeginBlocker(ctx, oracleKeeper)
		require.NoError(t, err)

		// Get the validator
		validator, err := oracleKeeper.StakingKeeper.Validator(ctx, operator)
		require.NoError(t, err)

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
		err = BeginBlocker(ctx, oracleKeeper)
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
		err = oracleKeeper.VoteTarget.Set(ctx, utils.AtomDenom, types.Denom{Name: utils.AtomDenom})
		require.NoError(t, err)

		err = oracleKeeper.VoteTarget.Set(ctx, utils.EthDenom, types.Denom{Name: utils.EthDenom})
		require.NoError(t, err)

		// Aggregate base exchange rate
		err = oracleKeeper.SetBaseExchangeRateWithDefault(ctx, utils.AtomDenom, math.LegacyNewDec(1))
		require.NoError(t, err)
		err = oracleKeeper.SetBaseExchangeRateWithDefault(ctx, utils.EthDenom, math.LegacyNewDec(2))
		require.NoError(t, err)
		err = oracleKeeper.SetBaseExchangeRateWithDefault(ctx, utils.KiiDenom, math.LegacyNewDec(3)) // extra denom
		require.NoError(t, err)

		// Execute EndBlocker on the last block of slash window
		params, err := oracleKeeper.Params.Get(ctx)
		require.NoError(t, err)
		slashWindow := params.SlashWindow
		ctx = ctx.WithBlockHeight(int64(slashWindow) - 1)
		err = BeginBlocker(ctx, oracleKeeper)
		require.NoError(t, err)

		// Validate the successful erased of the extra denoms
		err = oracleKeeper.ExchangeRate.Walk(ctx, nil, func(denom string, exchangeRate types.OracleExchangeRate) (bool, error) {
			require.True(t, denom != utils.KiiDenom)
			return false, nil
		})
		require.NoError(t, err)
	})

	t.Run("No votes submitted - Exchange rate should not be stored", func(t *testing.T) {
		// Reset blockchain state
		input, _ := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper

		// Set a single vote target
		err := oracleKeeper.VoteTarget.Clear(ctx, nil)
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.AtomDenom, types.Denom{Name: utils.AtomDenom})
		require.NoError(t, err)

		ctx = input.Ctx.WithBlockHeight(1)

		// No validators submit votes — skip directly to end/begin blocker

		err = EndBlocker(ctx, oracleKeeper)
		require.NoError(t, err)

		// No exchange rate should have been stored since nobody voted
		_, err = oracleKeeper.ExchangeRate.Get(ctx, utils.AtomDenom)
		require.Error(t, err)

		// Check if all three are getting abstain counts
		for i := 0; i < 3; i++ {
			counter, err := oracleKeeper.VotePenaltyCounter.Get(ctx, keeper.ValAddrs[i])
			require.NoError(t, err)
			require.EqualValues(t, uint64(1), counter.AbstainCount)
		}
	})

	t.Run("All validators miss one out of four denom votes - Miss count should increase for all", func(t *testing.T) {
		// Reset blockchain state
		input, msgServer := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper

		// Set four vote targets: atom, eth, kii and usdc
		err := oracleKeeper.VoteTarget.Clear(ctx, nil)
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.AtomDenom, types.Denom{Name: utils.AtomDenom})
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.EthDenom, types.Denom{Name: utils.EthDenom})
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.KiiDenom, types.Denom{Name: utils.KiiDenom})
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.UsdcDenom, types.Denom{Name: utils.UsdcDenom})
		require.NoError(t, err)

		ctx = input.Ctx.WithBlockHeight(1)

		// All validators vote for atom, eth and kii — skipping usdc entirely
		partialRate := randomAExchangeRate.String() + utils.AtomDenom +
			"," + randomAExchangeRate.String() + utils.EthDenom +
			"," + randomAExchangeRate.String() + utils.KiiDenom
		for i := 0; i < 3; i++ {
			voteMsg := types.NewMsgAggregateExchangeRateVote(partialRate, keeper.Addrs[i], keeper.ValAddrs[i])
			_, err := msgServer.AggregateExchangeRateVote(ctx, voteMsg)
			require.NoError(t, err)
		}

		err = EndBlocker(ctx, oracleKeeper)
		require.NoError(t, err)

		// Atom, eth and kii should be stored normally since all validators voted for them
		for _, denom := range []string{utils.AtomDenom, utils.EthDenom, utils.KiiDenom} {
			exchangeRateResponse, err := oracleKeeper.ExchangeRate.Get(ctx, denom)
			require.NoError(t, err)
			require.Equal(t, randomAExchangeRate, exchangeRateResponse.ExchangeRate)
		}

		// Usdc should not be stored since nobody voted for it
		_, err = oracleKeeper.ExchangeRate.Get(ctx, utils.UsdcDenom)
		require.Error(t, err)

		// All validators should have a miss count of 1 since none of them
		// achieved a WinCount equal to the total number of vote targets (4)
		for i := 0; i < 3; i++ {
			counter, err := oracleKeeper.VotePenaltyCounter.Get(ctx, keeper.ValAddrs[i])
			require.NoError(t, err)
			require.EqualValues(t, uint64(1), counter.MissCount)
			require.EqualValues(t, uint64(0), counter.AbstainCount)
		}
	})

	t.Run("One denom below threshold - validators who voted for it still miss", func(t *testing.T) {
		// Reset blockchain state
		input, msgServer := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper

		// Set four vote targets: atom, eth, kii and usdc
		err := oracleKeeper.VoteTarget.Clear(ctx, nil)
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.AtomDenom, types.Denom{Name: utils.AtomDenom})
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.EthDenom, types.Denom{Name: utils.EthDenom})
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.KiiDenom, types.Denom{Name: utils.KiiDenom})
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.UsdcDenom, types.Denom{Name: utils.UsdcDenom})
		require.NoError(t, err)

		ctx = input.Ctx.WithBlockHeight(1)

		// All 3 validators vote for atom, eth and kii (above threshold)
		// Only validator 0 votes for usdc — keeping it below the ballot threshold
		fullRate := randomAExchangeRate.String() + utils.AtomDenom +
			"," + randomAExchangeRate.String() + utils.EthDenom +
			"," + randomAExchangeRate.String() + utils.KiiDenom +
			"," + randomAExchangeRate.String() + utils.UsdcDenom
		partialRate := randomAExchangeRate.String() + utils.AtomDenom +
			"," + randomAExchangeRate.String() + utils.EthDenom +
			"," + randomAExchangeRate.String() + utils.KiiDenom

		// Validator 0 votes for all four denoms (including usdc)
		voteMsg := types.NewMsgAggregateExchangeRateVote(fullRate, keeper.Addrs[0], keeper.ValAddrs[0])
		_, err = msgServer.AggregateExchangeRateVote(ctx, voteMsg)
		require.NoError(t, err)

		// Validators 1 and 2 skip usdc, keeping it below threshold
		for i := 1; i < 3; i++ {
			voteMsg := types.NewMsgAggregateExchangeRateVote(partialRate, keeper.Addrs[i], keeper.ValAddrs[i])
			_, err := msgServer.AggregateExchangeRateVote(ctx, voteMsg)
			require.NoError(t, err)
		}

		err = EndBlocker(ctx, oracleKeeper)
		require.NoError(t, err)

		// Atom, eth and kii should be stored — they passed the ballot threshold
		for _, denom := range []string{utils.AtomDenom, utils.EthDenom, utils.KiiDenom} {
			exchangeRateResponse, err := oracleKeeper.ExchangeRate.Get(ctx, denom)
			require.NoError(t, err)
			require.Equal(t, randomAExchangeRate, exchangeRateResponse.ExchangeRate)
		}

		// Usdc should not be stored — it failed the ballot threshold
		_, err = oracleKeeper.ExchangeRate.Get(ctx, utils.UsdcDenom)
		require.Error(t, err)

		// Validators 1 and 2 voted for 3 out of 3 valid targets — their WinCount (3)
		// will match len(voteTargets) (3), so they get a success
		for i := 1; i < 3; i++ {
			counter, err := oracleKeeper.VotePenaltyCounter.Get(ctx, keeper.ValAddrs[i])
			require.NoError(t, err)
			require.EqualValues(t, uint64(0), counter.MissCount)
			require.EqualValues(t, uint64(0), counter.AbstainCount)
		}

		// Validator 0 voted for all 4 denoms. Usdc went to belowThresholdVoteMap
		// and still runs through Tally, so validator 0 gets WinCount=4 and
		// does not matches len(voteTargets) — earning a miss
		counter, err := oracleKeeper.VotePenaltyCounter.Get(ctx, keeper.ValAddrs[0])
		require.NoError(t, err)
		require.EqualValues(t, uint64(1), counter.MissCount)
		require.EqualValues(t, uint64(0), counter.AbstainCount)
	})

	t.Run("Validator hits below-threshold denom but deviates on an above-threshold one - should miss", func(t *testing.T) {
		// Reset blockchain state
		input, msgServer := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper

		// Set four vote targets: atom, eth, kii and usdc
		err := oracleKeeper.VoteTarget.Clear(ctx, nil)
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.AtomDenom, types.Denom{Name: utils.AtomDenom})
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.EthDenom, types.Denom{Name: utils.EthDenom})
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.KiiDenom, types.Denom{Name: utils.KiiDenom})
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.UsdcDenom, types.Denom{Name: utils.UsdcDenom})
		require.NoError(t, err)

		ctx = input.Ctx.WithBlockHeight(1)

		// Validators 1 and 2 vote correctly on atom, eth and kii — skipping usdc
		// keeping it below threshold
		partialRate := randomAExchangeRate.String() + utils.AtomDenom +
			"," + randomAExchangeRate.String() + utils.EthDenom +
			"," + randomAExchangeRate.String() + utils.KiiDenom
		for i := 1; i < 3; i++ {
			voteMsg := types.NewMsgAggregateExchangeRateVote(partialRate, keeper.Addrs[i], keeper.ValAddrs[i])
			_, err := msgServer.AggregateExchangeRateVote(ctx, voteMsg)
			require.NoError(t, err)
		}

		// Validator 0 deviates heavily on eth but votes correctly on usdc
		// (which will land in belowThresholdVoteMap since only val 0 voted for it)
		deviatedRate := randomAExchangeRate.String() + utils.AtomDenom +
			"," + aboveExchangeRate.String() + utils.EthDenom +
			"," + randomAExchangeRate.String() + utils.KiiDenom +
			"," + randomAExchangeRate.String() + utils.UsdcDenom
		voteMsg := types.NewMsgAggregateExchangeRateVote(deviatedRate, keeper.Addrs[0], keeper.ValAddrs[0])
		_, err = msgServer.AggregateExchangeRateVote(ctx, voteMsg)
		require.NoError(t, err)

		err = EndBlocker(ctx, oracleKeeper)
		require.NoError(t, err)

		// Atom, eth and kii should be stored based on validators 1 and 2 votes
		for _, denom := range []string{utils.AtomDenom, utils.EthDenom, utils.KiiDenom} {
			exchangeRateResponse, err := oracleKeeper.ExchangeRate.Get(ctx, denom)
			require.NoError(t, err)
			require.Equal(t, randomAExchangeRate, exchangeRateResponse.ExchangeRate)
		}

		// Usdc should not be stored — only validator 0 voted, below threshold
		_, err = oracleKeeper.ExchangeRate.Get(ctx, utils.UsdcDenom)
		require.Error(t, err)

		// Validators 1 and 2 voted correctly on all 3 above-threshold denoms
		// WinCount == 3 == len(voteTargets) → success
		// Validator 0 voted incorrectly on Eth, but correctly on under threshold
		// So WinCount == 3 as well
		for i := 0; i < 3; i++ {
			counter, err := oracleKeeper.VotePenaltyCounter.Get(ctx, keeper.ValAddrs[i])
			require.NoError(t, err)
			require.EqualValues(t, uint64(0), counter.MissCount)
			require.EqualValues(t, uint64(0), counter.AbstainCount)
		}
	})
}
