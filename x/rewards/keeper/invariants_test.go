package keeper_test

import (
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

func (suite *KeeperTestSuite) TestValidateModuleAccounting() {
	denom := types.DefaultParams().TokenDenom

	// An empty community pool is trivially consistent
	suite.Require().NoError(suite.App.RewardsKeeper.RewardPool.Set(suite.Ctx, types.RewardPool{}))
	suite.Require().NoError(suite.App.RewardsKeeper.ValidateModuleAccounting(suite.Ctx))

	// Funding the pool moves coins into the module account, so balance matches accounting
	err := suite.App.RewardsKeeper.FundCommunityPool(suite.Ctx, sdk.NewCoin(denom, math.NewInt(1000)), suite.TestAccs[0])
	suite.Require().NoError(err)
	suite.Require().NoError(suite.App.RewardsKeeper.ValidateModuleAccounting(suite.Ctx))

	// Inflating the accounting beyond the module balance must break the invariant
	pool, err := suite.App.RewardsKeeper.RewardPool.Get(suite.Ctx)
	suite.Require().NoError(err)
	pool.CommunityPool = pool.CommunityPool.Add(sdk.NewDecCoin(denom, math.NewInt(1)))
	suite.Require().NoError(suite.App.RewardsKeeper.RewardPool.Set(suite.Ctx, pool))
	suite.Require().Error(suite.App.RewardsKeeper.ValidateModuleAccounting(suite.Ctx))
}
