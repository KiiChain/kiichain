package oracle

import (
	"testing"

	"github.com/stretchr/testify/require"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/kiichain/kiichain/v7/x/oracle/keeper"
	"github.com/kiichain/kiichain/v7/x/oracle/types"
	"github.com/kiichain/kiichain/v7/x/oracle/utils"
)

// TestEndBlocker_InvalidValidatorAddress tests that EndBlocker properly handles
// invalid validator addresses by logging errors and continuing execution
func TestEndBlocker_InvalidValidatorAddress(t *testing.T) {
	input, _ := keeper.SetUp(t)
	ctx := input.Ctx
	oracleKeeper := input.OracleKeeper
	stakingKeeper := input.StakingKeeper

	// Create a truly non-existent validator address
	nonExistentValidatorAddr := sdk.ValAddress("non_existent_validator_addr")

	// Set up vote data for a non-existent validator that will trigger error handling
	err := oracleKeeper.AggregateExchangeRateVote.Set(ctx, nonExistentValidatorAddr, types.AggregateExchangeRateVote{
		ExchangeRateTuples: types.ExchangeRateTuples{
			{Denom: utils.AtomDenom, ExchangeRate: math.LegacyNewDec(10)},
		},
		Voter: nonExistentValidatorAddr.String(),
	})
	require.NoError(t, err)

	// Execute EndBlocker - this should not panic and should handle any errors gracefully
	err = EndBlocker(ctx, oracleKeeper)
	require.NoError(t, err)

	// The execution should continue normally
	// Verify that other processing completed successfully
	_, err = oracleKeeper.ExchangeRate.Get(ctx, utils.AtomDenom)
	require.NoError(t, err)
}

// TestSlashAndResetCounters_ValidatorNotFound tests error handling when validator not found
func TestSlashAndResetCounters_ValidatorNotFound(t *testing.T) {
	input, _ := keeper.SetUp(t)
	ctx := input.Ctx
	oracleKeeper := input.OracleKeeper

	// Create a truly non-existent validator address (not registered with staking keeper)
	nonExistentValidator := sdk.ValAddress("truly_non_existent_validator")
	err := oracleKeeper.VotePenaltyCounter.Set(ctx, nonExistentValidator, types.NewVotePenaltyCounter(1, 0, 0))
	require.NoError(t, err)

	// Execute SlashAndResetCounters - should handle error gracefully
	err = oracleKeeper.SlashAndResetCounters(ctx)
	require.NoError(t, err)

	// The vote penalty counter should be removed despite the error
	_, err = oracleKeeper.VotePenaltyCounter.Get(ctx, nonExistentValidator)
	require.Error(t, err)
}

// TestEndBlocker_InvalidOperatorAddress tests EndBlocker's error handling for invalid operator addresses
func TestEndBlocker_InvalidOperatorAddress(t *testing.T) {
	input, _ := keeper.SetUp(t)
	ctx := input.Ctx
	oracleKeeper := input.OracleKeeper
	stakingKeeper := input.StakingKeeper

	// Create a validator with an invalid operator address that will cause ValAddressFromBech32 to fail
	invalidValidator := stakingtypes.Validator{
		OperatorAddress: "invalid_bech32_address_that_will_fail_parsing",
		Status:          stakingtypes.Bonded,
		Tokens:          sdk.NewInt(1000000),
		DelegatorShares: sdk.NewDec(1000000),
	}

	// Mock the staking keeper to return our invalid validator
	// This simulates the scenario where a validator exists but has malformed operator address
	stakingKeeper.Validator = func(ctx sdk.Context, addr sdk.ValAddress) (stakingtypes.ValidatorI, error) {
		return &invalidValidator, nil
	}

	// Set up vote data for the invalid validator to trigger EndBlocker error handling
	err := oracleKeeper.AggregateExchangeRateVote.Set(ctx, sdk.ValAddress("invalid_validator_addr"), types.AggregateExchangeRateVote{
		ExchangeRateTuples: types.ExchangeRateTuples{
			{Denom: utils.AtomDenom, ExchangeRate: math.LegacyNewDec(10)},
		},
		Voter: "invalid_validator_addr",
	})
	require.NoError(t, err)

	// Execute EndBlocker - this should not panic and should handle the invalid address gracefully
	err = EndBlocker(ctx, oracleKeeper)
	require.NoError(t, err)

	// Verify that processing continued normally for other valid operations
	_, err = oracleKeeper.ExchangeRate.Get(ctx, utils.AtomDenom)
	require.NoError(t, err)
}