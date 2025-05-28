package keeper

import (
	"testing"

	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/stretchr/testify/require"
)

func TestCreateModuleAccount(t *testing.T) {
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// Check the module account was created
	account := oracleKeeper.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
	require.NotNil(t, account)

}
