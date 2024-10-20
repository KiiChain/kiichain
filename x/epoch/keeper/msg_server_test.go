package keeper_test

import (
	"testing"

	keepertest "github.com/kiichain/kiichain3/testutil/keeper"
	"github.com/kiichain/kiichain3/x/epoch/keeper"
	"github.com/stretchr/testify/require"
)

func TestSetupMsgServer(t *testing.T) {
	k, _ := keepertest.EpochKeeper(t)
	msg := keeper.NewMsgServerImpl(*k)
	require.NotNil(t, msg)
}
