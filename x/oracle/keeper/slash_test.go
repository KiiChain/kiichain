package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kiichain/kiichain/v7/x/oracle/types"
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

// TestSlashAndResetCounters_MultipleValidatorsNotFound ensures that SlashAndResetCounters does not panic
func TestSlashAndResetCounters_MultipleValidatorsNotFound(t *testing.T) {
	// initial setup
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	stakingKeeper := input.StakingKeeper
	ctx := input.Ctx

	// Create a staking message server
	amount := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
	msgServer := stakingkeeper.NewMsgServerImpl(&stakingKeeper)

	// Set the vote penalty counter for multiple non existing validators
	for i := range 5 {
		addr, val := ValAddrs[i], ValPubKeys[i]

		// Skip creating validator for ValAddrs[2] to simulate not found scenario
		if i != 2 {
			_, err := msgServer.CreateValidator(ctx, NewTestMsgCreateValidator(addr, val, amount))
			require.NoError(t, err)
		}

		// Add vote penalty counter
		err := oracleKeeper.VotePenaltyCounter.Set(input.Ctx, addr, types.NewVotePenaltyCounter(
			100000,
			0,
			0,
		))
		require.NoError(t, err)
	}

	// This should not panic even if some validators are not found
	err := oracleKeeper.SlashAndResetCounters(ctx)
	require.NoError(t, err)

	// Ensure that existing validators are slashed
	for i := range 5 {
		if i == 2 {
			// skip the non existing validator
			continue
		}

		// Check that the validator is slashed
		validator, err := stakingKeeper.GetValidator(ctx, ValAddrs[i])
		require.NoError(t, err)

		// Check that the validator has been slashed
		oracleParams, err := input.OracleKeeper.Params.Get(ctx)
		require.NoError(t, err)

		// Calculate expected bonded tokens after slashing
		expectedTokens := amount.Sub(oracleParams.SlashFraction.MulInt(amount).TruncateInt())
		require.Equal(t, expectedTokens, validator.Tokens)
	}
}

// TestSlashToPool_SupplyPreserved is the core regression test for the supply-burn bug.
//
// Previous behavior: oracle slash called StakingKeeper.Slash which burned tokens from the
// bonded pool, permanently reducing total supply. 
//
// Current behavior: slashToPool redirects those tokens to the community pool instead.
// Total supply is unchanged; only ownership shifts (bonded pool → community pool).
func TestSlashToPool_SupplyPreserved(t *testing.T) {
	input := CreateTestInput(t)
	ctx := input.Ctx
	bankKeeper := input.BankKeeper
	stakingKeeper := input.StakingKeeper
	oracleKeeper := input.OracleKeeper
	distKeeper := input.DistKeeper

	// Create a validator with 100 units of staked tokens
	stakedAmount := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
	msgServer := stakingkeeper.NewMsgServerImpl(&stakingKeeper)
	_, err := msgServer.CreateValidator(ctx, NewTestMsgCreateValidator(ValAddrs[0], ValPubKeys[0], stakedAmount))
	require.NoError(t, err)
	_, err = stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	// Fetch oracle params and set a non-zero slash fraction so the slash fires
	params, err := oracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.SlashFraction = math.LegacyNewDecWithPrec(1, 2) // 1%
	require.NoError(t, oracleKeeper.Params.Set(ctx, params))

	bondDenom := utils.KiiDenom
	expectedSlash := params.SlashFraction.MulInt(stakedAmount).TruncateInt()

	// Snapshot state before the slash
	totalSupplyBefore := bankKeeper.GetSupply(ctx, bondDenom).Amount
	feePoolBefore, err := distKeeper.FeePool.Get(ctx)
	require.NoError(t, err)
	communityPoolBefore := feePoolBefore.CommunityPool.AmountOf(bondDenom)
	bondedPoolBefore := bankKeeper.GetBalance(ctx, authtypes.NewModuleAddress(stakingtypes.BondedPoolName), bondDenom).Amount

	// Trigger the oracle slash: miss enough votes to fall below MinValidPerWindow
	votePeriodsPerWindow := math.LegacyNewDec(int64(params.SlashWindow)).QuoInt64(int64(params.VotePeriod)).TruncateInt64()
	minValidVotes := params.MinValidPerWindow.MulInt64(votePeriodsPerWindow).TruncateInt64()
	require.NoError(t, oracleKeeper.VotePenaltyCounter.Set(ctx, ValAddrs[0], types.NewVotePenaltyCounter(
		uint64(votePeriodsPerWindow-minValidVotes+1), // misses
		0,
		uint64(minValidVotes-1), // successes (below threshold)
	)))
	require.NoError(t, oracleKeeper.SlashAndResetCounters(ctx))

	// Snapshot state after the slash
	totalSupplyAfter := bankKeeper.GetSupply(ctx, bondDenom).Amount
	feePoolAfter, err := distKeeper.FeePool.Get(ctx)
	require.NoError(t, err)
	communityPoolAfter := feePoolAfter.CommunityPool.AmountOf(bondDenom)
	bondedPoolAfter := bankKeeper.GetBalance(ctx, authtypes.NewModuleAddress(stakingtypes.BondedPoolName), bondDenom).Amount

	// Previous behavior would have reduced supply by expectedSlash:
	//   require.Equal(t, totalSupplyBefore.Sub(expectedSlash), totalSupplyAfter) // BURNS
	//
	// Current behavior: supply is unchanged
	require.Equal(t, totalSupplyBefore, totalSupplyAfter, "total supply must not change after oracle slash")

	// The slashed tokens moved from the bonded pool to the community pool
	require.Equal(t, bondedPoolBefore.Sub(expectedSlash), bondedPoolAfter, "bonded pool must decrease by slash amount")
	require.Equal(t, communityPoolBefore.Add(math.LegacyNewDecFromInt(expectedSlash)), communityPoolAfter, "community pool must increase by slash amount")

	// The validator's bonded token count is reduced
	validator, err := stakingKeeper.GetValidator(ctx, ValAddrs[0])
	require.NoError(t, err)
	require.Equal(t, stakedAmount.Sub(expectedSlash), validator.GetBondedTokens())
}
