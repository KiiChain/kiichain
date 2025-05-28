package keeper

import (
	"sort"
	"strings"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/kiichain/kiichain/v1/x/oracle/utils"
	"github.com/stretchr/testify/require"
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

	// execute staking endblocker to start validators bon

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
	oracleKeeper.SetAggregateExchangeRateVote(ctx, ValAddrs[0], exchangeRateVote1)
	require.NoError(t, err)

	exchangeRateVote2, err := types.NewAggregateExchangeRateVote(exchangeRate2, ValAddrs[1]) // Aggregate rate tuples from Val1
	oracleKeeper.SetAggregateExchangeRateVote(ctx, ValAddrs[1], exchangeRateVote2)
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

	ukiiBallot := types.ExchangeRateBallot{
		{Denom: utils.MicroKiiDenom, ExchangeRate: math.LegacyNewDec(4), Power: int64(10), Voter: ValAddrs[0]},
		{Denom: utils.MicroKiiDenom, ExchangeRate: math.LegacyNewDec(4), Power: int64(10), Voter: ValAddrs[1]},
	}

	sort.Sort(uatomBallot)
	sort.Sort(uethBallot)
	sort.Sort(uusdcBallot)
	sort.Sort(ukiiBallot)

	// Call function
	denomBallot := oracleKeeper.OrganizeBallotByDenom(ctx, validatorClaimMap)

	// Validation
	require.ElementsMatch(t, uatomBallot, denomBallot[utils.MicroAtomDenom])
	require.ElementsMatch(t, uethBallot, denomBallot[utils.MicroEthDenom])
	require.ElementsMatch(t, uusdcBallot, denomBallot[utils.MicroUsdcDenom])
	require.ElementsMatch(t, ukiiBallot, denomBallot[utils.MicroKiiDenom])
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
	stakingKeeper.EndBlocker(ctx)

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
	oracleKeeper.SetAggregateExchangeRateVote(ctx, ValAddrs[0], exchangeRateVote1)
	require.NoError(t, err)

	exchangeRateVote2, err := types.NewAggregateExchangeRateVote(exchangeRate2, ValAddrs[1]) // Aggregate rate tuples from Val1
	oracleKeeper.SetAggregateExchangeRateVote(ctx, ValAddrs[1], exchangeRateVote2)
	require.NoError(t, err)

	// Delete the added exchange rate
	oracleKeeper.ClearBallots(ctx)

	// Validate process
	_, err = oracleKeeper.GetAggregateExchangeRateVote(ctx, ValAddrs[0])
	require.Error(t, err)
	_, err = oracleKeeper.GetAggregateExchangeRateVote(ctx, ValAddrs[1])
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
	oracleKeeper.ClearVoteTargets(ctx) // Delete voting target to start test from scrath

	// Define new whitelist (adds uusdc)
	whiteList := types.DenomList{
		{Name: utils.MicroAtomDenom},
		{Name: utils.MicroEthDenom},
		{Name: utils.MicroKiiDenom},
		{Name: utils.MicroUsdcDenom}, // New Denom
	}
	oracleParams.Whitelist = whiteList
	oracleKeeper.Params.Set(ctx, oracleParams)

	// Set vote targets manually before applying the new whitelist
	oracleKeeper.SetVoteTarget(ctx, utils.MicroAtomDenom)
	oracleKeeper.SetVoteTarget(ctx, utils.MicroEthDenom)
	oracleKeeper.SetVoteTarget(ctx, utils.MicroKiiDenom)

	// Ensure that uusdc is NOT present before applying the whitelist
	_, err = oracleKeeper.GetVoteTarget(ctx, utils.MicroUsdcDenom)
	require.Error(t, err)

	// Apply whitelist
	oracleKeeper.ApplyWhitelist(ctx, whiteList, map[string]types.Denom{})

	// Check that all elements in whitelist are now in voteTargets
	for _, item := range whiteList {
		_, err := oracleKeeper.GetVoteTarget(ctx, item.Name)
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
