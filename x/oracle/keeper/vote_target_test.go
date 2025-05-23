package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetVoteTargets(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper

	// clear vote target
	oracleKeeper.ClearVoteTargets(input.Ctx)

	// set new expected targets
	expectedTargets := []string{"ukii", "ubtc", "ueth"}
	for _, target := range expectedTargets {
		oracleKeeper.SetVoteTarget(input.Ctx, target)
	}

	// get voting target
	targets := oracleKeeper.GetVoteTargets(input.Ctx)

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
	oracleKeeper.ClearVoteTargets(input.Ctx)

	// set new expected targets and validate
	validTargets := []string{"ukii", "ubtc", "ueth"}
	for _, target := range validTargets {
		oracleKeeper.SetVoteTarget(input.Ctx, target)
		require.True(t, oracleKeeper.IsVoteTarget(input.Ctx, target))
	}
}
