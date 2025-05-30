package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kiichain/kiichain/v1/x/oracle/types"
)

func TestSlashAndResetMissCounters(t *testing.T) {
	// initial setup
	input := CreateTestInput(t)
	bankKeeper := input.BankKeeper
	stakingKeeper := input.StakingKeeper
	oracleKeeper := input.OracleKeeper

	addr1, val1 := ValAddrs[0], ValPubKeys[0]
	addr2, val2 := ValAddrs[1], ValPubKeys[1]
	amount := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
	msgServer := stakingkeeper.NewMsgServerImpl(&stakingKeeper)
	ctx := input.Ctx

	// Validators created
	_, err := msgServer.CreateValidator(ctx, NewTestMsgCreateValidator(addr1, val1, amount))
	require.NoError(t, err)
	_, err = msgServer.CreateValidator(ctx, NewTestMsgCreateValidator(addr2, val2, amount))
	require.NoError(t, err)
	_, err = stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	balance1 := bankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr1))
	stakingParams, err := stakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	expectedBalance := sdk.NewCoins(sdk.NewCoin(stakingParams.BondDenom, InitTokens.Sub(amount)))
	getVal1, err := stakingKeeper.Validator(ctx, addr1)
	require.NoError(t, err)
	bondedTokens1 := getVal1.GetBondedTokens()
	require.Equal(t, balance1, expectedBalance)
	require.Equal(t, amount, bondedTokens1)

	balance2 := bankKeeper.GetAllBalances(ctx, sdk.AccAddress(addr2))
	getVal2, err := stakingKeeper.Validator(ctx, addr2)
	require.NoError(t, err)
	bondedTokens2 := getVal2.GetBondedTokens()
	require.Equal(t, balance2, expectedBalance)
	require.Equal(t, amount, bondedTokens2)

	// Define slash fraction
	params, err := oracleKeeper.Params.Get(ctx)
	require.NoError(t, err)

	votePeriodsPerWindow := math.LegacyNewDec(int64(params.SlashWindow)).QuoInt64(int64(params.VotePeriod)).TruncateInt64()
	slashFraction := params.SlashFraction
	minValidVotes := params.MinValidPerWindow.MulInt64(votePeriodsPerWindow).TruncateInt64()

	t.Run("no slash", func(t *testing.T) {
		// Set the vote penalty counter for the validator
		err := oracleKeeper.VotePenaltyCounter.Set(input.Ctx, ValAddrs[0], types.NewVotePenaltyCounter(
			uint64(votePeriodsPerWindow-minValidVotes),
			0,
			uint64(minValidVotes),
		))
		require.NoError(t, err)

		err = oracleKeeper.SlashAndResetCounters(input.Ctx)
		require.NoError(t, err)
		_, err = stakingKeeper.EndBlocker(ctx)
		require.NoError(t, err)

		validator, _ := stakingKeeper.GetValidator(input.Ctx, ValAddrs[0])
		require.Equal(t, amount, validator.GetBondedTokens())
	})

	t.Run("no slash - total votes is greater than votes per window", func(t *testing.T) {
		// Set the vote penalty counter for the validator
		err := oracleKeeper.VotePenaltyCounter.Set(input.Ctx, ValAddrs[0], types.VotePenaltyCounter{
			MissCount:    uint64(votePeriodsPerWindow),
			AbstainCount: 0,
			SuccessCount: uint64(votePeriodsPerWindow),
		})
		require.NoError(t, err)

		err = oracleKeeper.SlashAndResetCounters(input.Ctx)
		require.NoError(t, err)
		_, err = stakingKeeper.EndBlocker(ctx)
		require.NoError(t, err)

		validator, _ := stakingKeeper.GetValidator(input.Ctx, ValAddrs[0])
		require.Equal(t, amount, validator.GetBondedTokens())
	})

	t.Run("successfully slash", func(t *testing.T) {
		// Set the vote penalty counter for the validator
		err := oracleKeeper.VotePenaltyCounter.Set(input.Ctx, ValAddrs[0], types.NewVotePenaltyCounter(
			uint64(votePeriodsPerWindow-minValidVotes+1),
			0,
			uint64(minValidVotes-1),
		))
		require.NoError(t, err)

		err = oracleKeeper.SlashAndResetCounters(input.Ctx)
		require.NoError(t, err)
		validator, _ := stakingKeeper.GetValidator(input.Ctx, ValAddrs[0])
		require.Equal(t, amount.Sub(slashFraction.MulInt(amount).TruncateInt()), validator.GetBondedTokens())
		require.True(t, validator.IsJailed())
	})

	t.Run("slash and jail for abstaining too much along with misses", func(t *testing.T) {
		validator, _ := stakingKeeper.GetValidator(input.Ctx, ValAddrs[0])
		validator.Jailed = false
		validator.Tokens = amount
		err := stakingKeeper.SetValidator(input.Ctx, validator)
		require.NoError(t, err)
		require.Equal(t, amount, validator.GetBondedTokens())

		// Set the vote penalty counter for the validator
		err = oracleKeeper.VotePenaltyCounter.Set(input.Ctx, ValAddrs[0], types.NewVotePenaltyCounter(
			0,
			uint64(votePeriodsPerWindow-minValidVotes+1),
			0,
		))
		require.NoError(t, err)

		err = oracleKeeper.SlashAndResetCounters(input.Ctx)
		require.NoError(t, err)
		validator, _ = stakingKeeper.GetValidator(input.Ctx, ValAddrs[0])

		// slashing for not voting validly sufficiently
		require.Equal(t, amount.Sub(slashFraction.MulInt(amount).TruncateInt()), validator.GetBondedTokens())
		require.True(t, validator.IsJailed())
	})

	t.Run("slash unbonded validator", func(t *testing.T) {
		validator, _ := stakingKeeper.GetValidator(input.Ctx, ValAddrs[0])
		validator.Status = stakingtypes.Unbonded
		validator.Jailed = false
		validator.Tokens = amount
		err := stakingKeeper.SetValidator(input.Ctx, validator)
		require.NoError(t, err)

		// Set the vote penalty counter for the validator
		err = oracleKeeper.VotePenaltyCounter.Set(input.Ctx, ValAddrs[0], types.NewVotePenaltyCounter(
			uint64(votePeriodsPerWindow-minValidVotes+1),
			0,
			0,
		))
		require.NoError(t, err)

		err = oracleKeeper.SlashAndResetCounters(input.Ctx)
		require.NoError(t, err)
		validator, _ = stakingKeeper.GetValidator(input.Ctx, ValAddrs[0])
		require.Equal(t, amount, validator.Tokens)
		require.False(t, validator.IsJailed())
	})

	t.Run("slash jailed validator", func(t *testing.T) {
		validator, _ := stakingKeeper.GetValidator(input.Ctx, ValAddrs[0])
		validator.Status = stakingtypes.Bonded
		validator.Jailed = true
		validator.Tokens = amount
		err := stakingKeeper.SetValidator(input.Ctx, validator)
		require.NoError(t, err)

		// Set the vote penalty counter for the validator
		err = oracleKeeper.VotePenaltyCounter.Set(input.Ctx, ValAddrs[0], types.NewVotePenaltyCounter(
			uint64(votePeriodsPerWindow-minValidVotes+1),
			0,
			0,
		))
		require.NoError(t, err)

		err = oracleKeeper.SlashAndResetCounters(input.Ctx)
		require.NoError(t, err)
		validator, _ = stakingKeeper.GetValidator(input.Ctx, ValAddrs[0])
		require.Equal(t, amount, validator.Tokens)
	})
}
