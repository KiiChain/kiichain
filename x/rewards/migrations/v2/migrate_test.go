package v2_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/app/apptesting"
	v2 "github.com/kiichain/kiichain/v7/x/rewards/migrations/v2"
	"github.com/kiichain/kiichain/v7/x/rewards/types"
)

type MigrateTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(MigrateTestSuite))
}

func (suite *MigrateTestSuite) TestMigrateStore() {
	suite.Setup()

	k := suite.App.RewardsKeeper
	denom := types.DefaultParams().TokenDenom

	// Simulate v1 params (denom only meaningful; other fields empty/zero).
	suite.Require().NoError(k.Params.Set(suite.Ctx, types.Params{
		TokenDenom: denom,
	}))
	suite.Require().NoError(k.RewardPool.Set(suite.Ctx, types.RewardPool{
		CommunityPool: sdk.NewDecCoins(sdk.NewDecCoin(denom, math.NewInt(100))),
	}))

	// Write a dummy legacy ReleaseSchedule key under prefix 2.
	legacyPrefix := collections.NewPrefix(2)
	suite.Require().NoError(k.StoreService().OpenKVStore(suite.Ctx).Set(legacyPrefix, []byte("legacy")))

	suite.Require().NoError(v2.MigrateStore(suite.Ctx, k))

	got, err := k.StoreService().OpenKVStore(suite.Ctx).Get(legacyPrefix)
	suite.Require().NoError(err)
	suite.Require().Nil(got)

	params, err := k.Params.Get(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(denom, params.TokenDenom)
	defaults := types.DefaultParams()
	suite.Require().True(params.GoalBonded.Equal(defaults.GoalBonded), "got GoalBonded=%s want %s", params.GoalBonded, defaults.GoalBonded)
	suite.Require().True(params.InflationMax.Equal(defaults.InflationMax), "got InflationMax=%s want %s", params.InflationMax, defaults.InflationMax)
	suite.Require().True(params.InflationRateChange.Equal(defaults.InflationRateChange), "got InflationRateChange=%s want %s", params.InflationRateChange, defaults.InflationRateChange)
	suite.Require().True(params.SupplyBase.IsZero())
}
