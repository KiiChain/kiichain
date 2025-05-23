package keeper

import (
	"testing"

	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/stretchr/testify/require"
)

func TestGetOracleAccount(t *testing.T) {
	init := CreateTestInput(t)
	accountKeeper := init.AccountKeeper
	ctx := init.Ctx

	// must create the account
	oracleAccount := accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	require.NotNil(t, oracleAccount)
}
