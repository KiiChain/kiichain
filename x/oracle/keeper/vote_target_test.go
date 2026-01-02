package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kiichain/kiichain/v7/x/oracle/types"
	"github.com/kiichain/kiichain/v7/x/oracle/utils"
)

func TestGetVoteTargets(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper

	// clear vote target
	err := oracleKeeper.VoteTarget.Clear(input.Ctx, nil)
	require.NoError(t, err)

	// set new expected targets
	expectedTargets := []string{utils.KiiDenom, utils.BtcDenom, utils.EthDenom}
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
	validTargets := []string{utils.KiiDenom, utils.BtcDenom, utils.EthDenom}
	for _, target := range validTargets {
		err = oracleKeeper.VoteTarget.Set(input.Ctx, target, types.Denom{Name: target})
		require.NoError(t, err)

		// check if the target exist
		found, err := oracleKeeper.VoteTarget.Has(input.Ctx, target)
		require.NoError(t, err)
		require.True(t, found)
	}
}
