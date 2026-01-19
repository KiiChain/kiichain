package oracle

import (
        "testing"
        "github.com/stretchr/testify/require"
        sdk "github.com/cosmos/cosmos-sdk/types"
        "github.com/kiichain/kiichain/v7/x/oracle/types"
)

func TestSlashAndResetCounters_ValidatorNotFound(t *testing.T) {
        input, _ := SetUp(t)
        ctx := input.Ctx
        oracleKeeper := input.OracleKeeper

	nonExistentValidator := sdk.ValAddress(make([]byte, 20))
        err := oracleKeeper.VotePenaltyCounter.Set(ctx, nonExistentValidator, types.NewVotePenaltyCounter(1, 0, 0))
        require.NoError(t, err)

        err = oracleKeeper.SlashAndResetCounters(ctx)
        require.NoError(t, err)

        counter, err := oracleKeeper.VotePenaltyCounter.Get(ctx, nonExistentValidator)
        require.NoError(t, err)
        require.Equal(t, uint64(1), counter.MissCount)
}