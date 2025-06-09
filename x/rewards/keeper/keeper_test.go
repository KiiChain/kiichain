package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/kiichain/kiichain/v1/app/apptesting"
	"github.com/kiichain/kiichain/v1/x/rewards/keeper"
	"github.com/kiichain/kiichain/v1/x/rewards/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient     types.QueryClient
	bankQueryClient banktypes.QueryClient
	msgServer       types.MsgServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()

	// Fund every TestAcc
	amount, ok := math.NewIntFromString("1000000000000000000000") // 1000 kii
	if !ok {
		suite.Error(fmt.Errorf("Could not create int to fund accs "))
	}
	fundAccsAmount := sdk.NewCoins(sdk.NewCoin(types.DefaultParams().TokenDenom, amount))
	for _, acc := range suite.TestAccs {
		suite.FundAcc(acc, fundAccsAmount)
	}

	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
	suite.bankQueryClient = banktypes.NewQueryClient(suite.QueryHelper)
	suite.msgServer = keeper.NewMsgServerImpl(suite.App.RewardsKeeper)
}

func (suite *KeeperTestSuite) OverrideMsgServer(newKeeper keeper.Keeper) {
	suite.msgServer = keeper.NewMsgServerImpl(newKeeper)
}
