package oracle

import (
	"testing"

	"github.com/stretchr/testify/require"
	"cosmossdk.io/math"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	
	"github.com/kiichain/kiichain/v7/x/oracle/types"
	"github.com/kiichain/kiichain/v7/x/oracle/utils"
)

func TestEndBlockerInvalidValidatorAddress(t *testing.T) {
	t.Run("EndBlocker handles invalid validator operator address gracefully", func(t *testing.T) {
		// Setup test environment
		input, _ := SetUp(t)
		ctx := input.Ctx
		oracleKeeper := input.OracleKeeper
		stakingKeeper := input.StakingKeeper

		// Set up vote target
		err := oracleKeeper.VoteTarget.Clear(ctx, nil)
		require.NoError(t, err)
		err = oracleKeeper.VoteTarget.Set(ctx, utils.AtomDenom, types.Denom{Name: utils.AtomDenom})
		require.NoError(t, err)

		// Create a validator with invalid operator address
		// We'll create a validator manually with corrupted operator address
		invalidValidator := stakingtypes.Validator{
			OperatorAddress: "invalid_bech32_address_that_will_fail_parsing",
			Status:          stakingtypes.Bonded,
			Tokens:          math.NewInt(1000000),
			DelegatorShares: math.LegacyNewDec(1000000),
		}

		// Set the invalid validator in staking keeper
		stakingKeeper.SetValidator(ctx, invalidValidator)

		// Execute EndBlocker - should handle invalid address gracefully
		err = EndBlocker(ctx, oracleKeeper)
		require.NoError(t, err)

		// Test passes if EndBlocker completes without panic
		// The invalid validator should be skipped and logged
	})
}
