package keeper

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/kiichain/kiichain/v1/x/oracle/utils"
)

func TestOrganizeBallotByDenom(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	stakingKeeper := init.StakingKeeper
	ctx := init.Ctx

	// Create handlers
	msgServer := stakingkeeper.NewMsgServerImpl(&stakingKeeper)

	// Create validators
	stakingAmount := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	val0 := NewTestMsgCreateValidator(ValAddrs[0], ValPubKeys[0], stakingAmount)
	val1 := NewTestMsgCreateValidator(ValAddrs[1], ValPubKeys[1], stakingAmount)

	// Register validators
	_, err := msgServer.CreateValidator(ctx, val0)
	require.NoError(t, err)
	_, err = msgServer.CreateValidator(ctx, val1)
	require.NoError(t, err)

	// execute staking endblocker to start validators bonded
	_, err = stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	// Simulate aggregation exchange rate process
	exchangeRate1 := types.ExchangeRateTuples{
		{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(1)},
		{Denom: utils.MicroEthDenom, ExchangeRate: math.LegacyNewDec(2)},
		{Denom: utils.MicroUsdcDenom, ExchangeRate: math.LegacyNewDec(3)},
		{Denom: utils.MicroKiiDenom, ExchangeRate: math.LegacyNewDec(4)},
	}

	exchangeRate2 := types.ExchangeRateTuples{
		{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(1)},
		{Denom: utils.MicroEthDenom, ExchangeRate: math.LegacyNewDec(2)},
		{Denom: utils.MicroUsdcDenom, ExchangeRate: math.LegacyNewDec(3)},
		{Denom: utils.MicroKiiDenom, ExchangeRate: math.LegacyNewDec(4)},
	}

	exchangeRateVote1, err := types.NewAggregateExchangeRateVote(exchangeRate1, ValAddrs[0]) // Aggregate rate tuples from Val0
	require.NoError(t, err)
	err = oracleKeeper.AggregateExchangeRateVote.Set(ctx, ValAddrs[0], exchangeRateVote1)
	require.NoError(t, err)

	exchangeRateVote2, err := types.NewAggregateExchangeRateVote(exchangeRate2, ValAddrs[1]) // Aggregate rate tuples from Val1
	require.NoError(t, err)
	err = oracleKeeper.AggregateExchangeRateVote.Set(ctx, ValAddrs[1], exchangeRateVote2)
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

		// Set the validator as bonded for calculations
		valPower := validator.GetConsensusPower(powerReduction)
		operator := validator.GetOperator()

		// Get the operator as a valaddress
		operatorAddr, err := sdk.ValAddressFromBech32(operator)
		require.NoError(t, err)

		claim := types.NewClaim(valPower, 0, 0, false, operatorAddr)

		validatorClaimMap[operator] = claim // Assign the validator on the list to receive
	}

	// Create expected result (with denom organized alphabetically)
	uatomBallot := types.ExchangeRateBallot{
		{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(1), Power: int64(10), Voter: ValAddrs[0]},
		{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(1), Power: int64(10), Voter: ValAddrs[1]},
	}

	uethBallot := types.ExchangeRateBallot{
		{Denom: utils.MicroEthDenom, ExchangeRate: math.LegacyNewDec(2), Power: int64(10), Voter: ValAddrs[0]},
		{Denom: utils.MicroEthDenom, ExchangeRate: math.LegacyNewDec(2), Power: int64(10), Voter: ValAddrs[1]},
	}

	uusdcBallot := types.ExchangeRateBallot{
		{Denom: utils.MicroUsdcDenom, ExchangeRate: math.LegacyNewDec(3), Power: int64(10), Voter: ValAddrs[0]},
		{Denom: utils.MicroUsdcDenom, ExchangeRate: math.LegacyNewDec(3), Power: int64(10), Voter: ValAddrs[1]},
	}

	akiiBallot := types.ExchangeRateBallot{
		{Denom: utils.MicroKiiDenom, ExchangeRate: math.LegacyNewDec(4), Power: int64(10), Voter: ValAddrs[0]},
		{Denom: utils.MicroKiiDenom, ExchangeRate: math.LegacyNewDec(4), Power: int64(10), Voter: ValAddrs[1]},
	}

	sort.Sort(uatomBallot)
	sort.Sort(uethBallot)
	sort.Sort(uusdcBallot)
	sort.Sort(akiiBallot)

	// Call function
	denomBallot, err := oracleKeeper.OrganizeBallotByDenom(ctx, validatorClaimMap)
	require.NoError(t, err)

	// Validation
	microAtomDenomBallot := denomBallot[utils.MicroAtomDenom]
	require.ElementsMatch(t, uatomBallot, microAtomDenomBallot)
	require.ElementsMatch(t, uethBallot, denomBallot[utils.MicroEthDenom])
	require.ElementsMatch(t, uusdcBallot, denomBallot[utils.MicroUsdcDenom])
	require.ElementsMatch(t, akiiBallot, denomBallot[utils.MicroKiiDenom])
}

func TestClearBallots(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	stakingKeeper := init.StakingKeeper
	ctx := init.Ctx

	// Create handlers
	msgServer := stakingkeeper.NewMsgServerImpl(&stakingKeeper)

	// Create validators
	stakingAmount := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	val0 := NewTestMsgCreateValidator(ValAddrs[0], ValPubKeys[0], stakingAmount)
	val1 := NewTestMsgCreateValidator(ValAddrs[1], ValPubKeys[1], stakingAmount)

	// Register validators
	_, err := msgServer.CreateValidator(ctx, val0)
	require.NoError(t, err)
	_, err = msgServer.CreateValidator(ctx, val1)
	require.NoError(t, err)

	// execute staking endblocker to start validators bonding
	_, err = stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	// Simulate aggregation exchange rate process
	exchangeRate1 := types.ExchangeRateTuples{
		{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(1)},
		{Denom: utils.MicroEthDenom, ExchangeRate: math.LegacyNewDec(2)},
		{Denom: utils.MicroUsdcDenom, ExchangeRate: math.LegacyNewDec(3)},
		{Denom: utils.MicroKiiDenom, ExchangeRate: math.LegacyNewDec(4)},
	}

	exchangeRate2 := types.ExchangeRateTuples{
		{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(1)},
		{Denom: utils.MicroEthDenom, ExchangeRate: math.LegacyNewDec(2)},
		{Denom: utils.MicroUsdcDenom, ExchangeRate: math.LegacyNewDec(3)},
		{Denom: utils.MicroKiiDenom, ExchangeRate: math.LegacyNewDec(4)},
	}

	exchangeRateVote1, err := types.NewAggregateExchangeRateVote(exchangeRate1, ValAddrs[0]) // Aggregate rate tuples from Val0
	require.NoError(t, err)
	err = oracleKeeper.AggregateExchangeRateVote.Set(ctx, ValAddrs[0], exchangeRateVote1)
	require.NoError(t, err)

	exchangeRateVote2, err := types.NewAggregateExchangeRateVote(exchangeRate2, ValAddrs[1]) // Aggregate rate tuples from Val1
	require.NoError(t, err)
	err = oracleKeeper.AggregateExchangeRateVote.Set(ctx, ValAddrs[1], exchangeRateVote2)
	require.NoError(t, err)

	// Clear all votes
	err = oracleKeeper.AggregateExchangeRateVote.Clear(ctx, nil)
	require.NoError(t, err)

	// Validate process
	_, err = oracleKeeper.AggregateExchangeRateVote.Get(ctx, ValAddrs[0])
	require.Error(t, err)
	_, err = oracleKeeper.AggregateExchangeRateVote.Get(ctx, ValAddrs[1])
	require.Error(t, err)
}

func TestApplyWhitelist(t *testing.T) {
	// Prepare the test environment
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	bankKeeper := init.BankKeeper
	ctx := init.Ctx
	oracleParams, err := oracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	err = oracleKeeper.VoteTarget.Clear(ctx, nil) // Delete voting target to start test from scrath
	require.NoError(t, err)

	// Define new whitelist (adds uusdc)
	whiteList := types.DenomList{
		{Name: utils.MicroAtomDenom},
		{Name: utils.MicroEthDenom},
		{Name: utils.MicroKiiDenom},
		{Name: utils.MicroUsdcDenom}, // New Denom
	}
	oracleParams.Whitelist = whiteList
	err = oracleKeeper.Params.Set(ctx, oracleParams)
	require.NoError(t, err)

	// Set vote targets manually before applying the new whitelist
	err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroAtomDenom, types.Denom{Name: utils.MicroAtomDenom})
	require.NoError(t, err)
	err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroEthDenom, types.Denom{Name: utils.MicroEthDenom})
	require.NoError(t, err)
	err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroKiiDenom, types.Denom{Name: utils.MicroKiiDenom})
	require.NoError(t, err)

	// Ensure that uusdc is NOT present before applying the whitelist
	_, err = oracleKeeper.VoteTarget.Get(ctx, utils.MicroUsdcDenom)
	require.Error(t, err)

	// Apply whitelist
	err = oracleKeeper.ApplyWhitelist(ctx, whiteList, map[string]types.Denom{})
	require.NoError(t, err)

	// Check that all elements in whitelist are now in voteTargets
	for _, item := range whiteList {
		_, err := oracleKeeper.VoteTarget.Get(ctx, item.Name)
		require.NoError(t, err)
	}

	// Verify metadata was created in the bank module
	for _, item := range whiteList {
		metadata, found := bankKeeper.GetDenomMetaData(ctx, item.Name)
		require.True(t, found)

		// Validate metadata fields
		require.Equal(t, item.Name, metadata.Base)
		require.Equal(t, strings.ToUpper(item.Name[1:]), metadata.Name)
		require.Equal(t, "u"+item.Name[1:], metadata.DenomUnits[0].Denom)
		require.Equal(t, "m"+item.Name[1:], metadata.DenomUnits[1].Denom)
		require.Equal(t, item.Name[1:], metadata.DenomUnits[2].Denom)
	}
}
