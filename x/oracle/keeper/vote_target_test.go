package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kiichain/kiichain/v2/x/oracle/types"
)

func TestGetVoteTargets(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper

	// clear vote target
	err := oracleKeeper.VoteTarget.Clear(input.Ctx, nil)
	require.NoError(t, err)

	// set new expected targets
	expectedTargets := []string{"akii", "ubtc", "ueth"}
	for _, target := range expectedTargets {
		err = oracleKeeper.VoteTarget.Set(input.Ctx, target, types.Denom{Name: target})
		require.NoError(t, err)
	}

	// get voting target
	targets, err := oracleKeeper.GetVoteTargets(input.Ctx)
	require.NoError(t, err)

	// validation
	elements := make(map[string]bool)
	for _, target := range targets {
		elements[target] = true
	}

	for _, expectedTarget := range expectedTargets {
		require.True(t, elements[expectedTarget])
	}
}

func TestIsVoteTarget(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper

	// clear vote target
	err := oracleKeeper.VoteTarget.Clear(input.Ctx, nil)
	require.NoError(t, err)

	// set new expected targets and validate
	validTargets := []string{"akii", "ubtc", "ueth"}
	for _, target := range validTargets {
		err = oracleKeeper.VoteTarget.Set(input.Ctx, target, types.Denom{Name: target})
		require.NoError(t, err)

		// check if the target exist
		found, err := oracleKeeper.VoteTarget.Has(input.Ctx, target)
		require.NoError(t, err)
		require.True(t, found)
	}
}
