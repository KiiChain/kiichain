package oracle

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/kiichain/kiichain/v1/x/oracle/keeper"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/kiichain/kiichain/v1/x/oracle/utils"
)

func TestPickReferenceDenom(t *testing.T) {
	input := keeper.CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	stakingKeeper := input.StakingKeeper
	ctx := input.Ctx

	// **** Prepare staking environment (set total bonded power as 100 )
	// Create handlers
	msgServer := stakingkeeper.NewMsgServerImpl(&stakingKeeper)

	// Create validators
	stakingAmount := sdk.TokensFromConsensusPower(50, sdk.DefaultPowerReduction)
	val0 := keeper.NewTestMsgCreateValidator(keeper.ValAddrs[0], keeper.ValPubKeys[0], stakingAmount)
	val1 := keeper.NewTestMsgCreateValidator(keeper.ValAddrs[1], keeper.ValPubKeys[1], stakingAmount)

	// Register validators
	_, err := msgServer.CreateValidator(ctx, val0)
	require.NoError(t, err)
	_, err = msgServer.CreateValidator(ctx, val1)
	require.NoError(t, err)

	// execute staking endblocker to start validators bonding
	_, err = stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)
	// ********

	// Modify the oracle param vote threshold
	params, err := oracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	params.VoteThreshold = math.LegacyNewDecWithPrec(66, 2) // 0.66
	err = oracleKeeper.Params.Set(ctx, params)
	require.NoError(t, err)

	// Create voting targets
	votingTarget := map[string]types.Denom{
		utils.MicroAtomDenom: {Name: utils.MicroAtomDenom},
		utils.MicroEthDenom:  {Name: utils.MicroEthDenom},
		utils.MicroUsdcDenom: {Name: utils.MicroUsdcDenom},
		utils.MicroKiiDenom:  {Name: utils.MicroKiiDenom},
	}

	// Create vote map (the voting (ballot) per denom)
	uatomBallot := types.ExchangeRateBallot{
		{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(4000), Power: int64(20), Voter: keeper.ValAddrs[0]},
		{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(4100), Power: int64(10), Voter: keeper.ValAddrs[1]},
		{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(4200), Power: int64(30), Voter: keeper.ValAddrs[3]},
		{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(5000), Power: int64(30), Voter: keeper.ValAddrs[4]},
	}

	uethBallot := types.ExchangeRateBallot{
		{Denom: utils.MicroEthDenom, ExchangeRate: math.LegacyNewDec(10000), Power: int64(20), Voter: keeper.ValAddrs[0]},
		{Denom: utils.MicroEthDenom, ExchangeRate: math.LegacyNewDec(9580), Power: int64(30), Voter: keeper.ValAddrs[3]},
		{Denom: utils.MicroEthDenom, ExchangeRate: math.LegacyNewDec(10300), Power: int64(30), Voter: keeper.ValAddrs[4]},
	}

	uusdcBallot := types.ExchangeRateBallot{
		{Denom: utils.MicroUsdcDenom, ExchangeRate: math.LegacyNewDec(20000), Power: int64(20), Voter: keeper.ValAddrs[0]},
		{Denom: utils.MicroUsdcDenom, ExchangeRate: math.LegacyNewDec(20100), Power: int64(10), Voter: keeper.ValAddrs[1]},
		{Denom: utils.MicroUsdcDenom, ExchangeRate: math.LegacyNewDec(19580), Power: int64(30), Voter: keeper.ValAddrs[3]},
		{Denom: utils.MicroUsdcDenom, ExchangeRate: math.LegacyNewDec(20300), Power: int64(30), Voter: keeper.ValAddrs[4]},
	}

	akiiBallot := types.ExchangeRateBallot{
		{Denom: utils.MicroKiiDenom, ExchangeRate: math.LegacyNewDec(30000), Power: int64(20), Voter: keeper.ValAddrs[0]},
		{Denom: utils.MicroKiiDenom, ExchangeRate: math.LegacyNewDec(30100), Power: int64(10), Voter: keeper.ValAddrs[1]},
		{Denom: utils.MicroKiiDenom, ExchangeRate: math.LegacyNewDec(29580), Power: int64(30), Voter: keeper.ValAddrs[3]},
	}

	voteMap := map[string]types.ExchangeRateBallot{
		utils.MicroAtomDenom: uatomBallot,
		utils.MicroEthDenom:  uethBallot,
		utils.MicroUsdcDenom: uusdcBallot,
		utils.MicroKiiDenom:  akiiBallot,
		"extraDenom":         uatomBallot, // This denom will be removed because is not on the voting targets
	}

	// Expected below threshold vote map
	expectedBelowThreshold := map[string]types.ExchangeRateBallot{
		utils.MicroKiiDenom: akiiBallot,
	}

	// Must return denom MicroAtomDenom and akiiBallot as below threshold map
	referenceDenom, belowThresholdVoteMap := pickReferenceDenom(ctx, oracleKeeper, votingTarget, voteMap)
	require.Equal(t, utils.MicroAtomDenom, referenceDenom)
	require.Equal(t, expectedBelowThreshold, belowThresholdVoteMap)
}

func TestBallotIsPassing(t *testing.T) {
	uatomBallot := types.ExchangeRateBallot{
		{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(4000), Power: int64(20), Voter: keeper.ValAddrs[0]},
		{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(4100), Power: int64(10), Voter: keeper.ValAddrs[1]},
		{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(4200), Power: int64(30), Voter: keeper.ValAddrs[3]},
		{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(5000), Power: int64(30), Voter: keeper.ValAddrs[4]},
	}

	// must return true because the threshold is lower than the ballot power
	power, ispassing := ballotIsPassing(uatomBallot, math.NewInt(80))
	require.Equal(t, math.NewInt(90), power)
	require.True(t, ispassing)

	// must return false because the threshold is higher than the ballot power
	power, ispassing = ballotIsPassing(uatomBallot, math.NewInt(100))
	require.Equal(t, math.NewInt(90), power)
	require.False(t, ispassing)
}

func TestTally(t *testing.T) {
	input := keeper.CreateTestInput(t)
	stakingKeeper := input.StakingKeeper
	ctx := input.Ctx

	// Create handlers
	msgServer := stakingkeeper.NewMsgServerImpl(&stakingKeeper)

	// Create validators
	stakingAmount := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	val0 := keeper.NewTestMsgCreateValidator(keeper.ValAddrs[0], keeper.ValPubKeys[0], stakingAmount)
	val1 := keeper.NewTestMsgCreateValidator(keeper.ValAddrs[1], keeper.ValPubKeys[1], stakingAmount)

	// Register validators
	_, err := msgServer.CreateValidator(ctx, val0)
	require.NoError(t, err)
	_, err = msgServer.CreateValidator(ctx, val1)
	require.NoError(t, err)

	// execute staking endblocker to start validators bonding
	_, err = stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	// Get claim map
	validatorClaimMap := make(map[string]types.Claim)
	powerReduction := stakingKeeper.PowerReduction(ctx)

	iterator, err := stakingKeeper.ValidatorsPowerStoreIterator(ctx)
	require.NoError(t, err)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		valAddr := sdk.ValAddress(iterator.Value())             // Get validator address
		validator, err := stakingKeeper.Validator(ctx, valAddr) // get validator by address
		require.NoError(t, err)

		valPower := validator.GetConsensusPower(powerReduction)
		operator := validator.GetOperator()

		// Get the operator as a valaddress
		operatorValAddr := sdk.ValAddress(operator)

		claim := types.NewClaim(valPower, 0, 0, false, operatorValAddr)

		validatorClaimMap[operator] = claim // Assign the validator on the list to receive
	}

	uatomBallot := types.ExchangeRateBallot{
		{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(4160), Power: int64(10), Voter: keeper.ValAddrs[0]},
		{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(4180), Power: int64(20), Voter: keeper.ValAddrs[1]},
		{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(4200), Power: int64(30), Voter: keeper.ValAddrs[2]}, // weighted median
		{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(5000), Power: int64(40), Voter: keeper.ValAddrs[3]},
	}

	// median = 4200
	// deviation = 415.33
	// rewardBand = 0.02
	// reward spread = 42
	// upper limit = 4242
	// lower limit = 4158

	weightedMedian := Tally(ctx, uatomBallot, math.LegacyNewDecWithPrec(2, 2), validatorClaimMap)
	require.Equal(t, math.LegacyNewDec(4200), weightedMedian)

	// validate validators who voted
	for validator, claim := range validatorClaimMap {
		if validator == keeper.ValAddrs[3].String() {
			require.Zero(t, claim.Weight)
			continue
		}

		require.NotZero(t, claim.Weight) // val 0, 1 and 2 voted
	}
}
