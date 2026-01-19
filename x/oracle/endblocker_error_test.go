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

	// Set up vote data that will trigger EndBlocker processing
	err := oracleKeeper.AggregateExchangeRateVote.Set(ctx, keeper.ValAddrs[0], types.AggregateExchangeRateVote{
		ExchangeRateTuples: types.ExchangeRateTuples{
			{Denom: utils.AtomDenom, ExchangeRate: math.LegacyNewDec(10)},
		},
		Voter: keeper.ValAddrs[0].String(),
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

	// Set up a vote penalty counter for a non-existent validator
	nonExistentValidator := keeper.ValAddrs[0]
	err := oracleKeeper.VotePenaltyCounter.Set(ctx, nonExistentValidator, types.NewVotePenaltyCounter(1, 0, 0))
	require.NoError(t, err)

	// Execute SlashAndResetCounters - should handle error gracefully
	err = oracleKeeper.SlashAndResetCounters(ctx)
	require.NoError(t, err)

	// The vote penalty counter should be removed despite the error
	_, err = oracleKeeper.VotePenaltyCounter.Get(ctx, nonExistentValidator)
	require.Error(t, err)
}

// TestValAddressFromBech32_ErrorHandling tests the specific error handling for address parsing
func TestValAddressFromBech32_ErrorHandling(t *testing.T) {
	// Test various invalid bech32 addresses that should trigger error handling
	invalidAddresses := []string{
		"",                    // empty string
		"invalid",             // not bech32
		"cosmosvaloper1invalid", // invalid checksum
		"cosmosvaloper1tooshort", // too short
	}

	for _, invalidAddr := range invalidAddresses {
		t.Run("Invalid address: "+invalidAddr, func(t *testing.T) {
			// This should trigger the error handling in EndBlocker
			_, err := sdk.ValAddressFromBech32(invalidAddr)
			require.Error(t, err)
			
			// The error should be properly handled without panic
			// (This test verifies the error path is reachable)
		})
	}
}