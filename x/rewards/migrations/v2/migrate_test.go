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
	suite.Require().NoError(k.RewardPool.Set(suite.Ctx, types.RewardPool{}))

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
	suite.Require().Equal(defaults.BlocksPerYear, params.BlocksPerYear)
	suite.Require().True(params.SupplyBase.IsZero())
}

func (suite *MigrateTestSuite) TestMigrateStoreBackfillsPartialParams() {
	suite.Setup()
	k := suite.App.RewardsKeeper
	defaults := types.DefaultParams()

	suite.Require().NoError(k.Params.Set(suite.Ctx, types.Params{
		TokenDenom: "",
		GoalBonded: math.LegacyZeroDec(),
	}))
	suite.Require().NoError(v2.MigrateStore(suite.Ctx, k))
	params, err := k.Params.Get(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(defaults.TokenDenom, params.TokenDenom)
	suite.Require().True(params.GoalBonded.Equal(defaults.GoalBonded))

	suite.Require().NoError(k.Params.Set(suite.Ctx, types.Params{
		TokenDenom:   defaults.TokenDenom,
		GoalBonded:   defaults.GoalBonded,
		InflationMax: math.LegacyNewDec(-1),
	}))
	suite.Require().NoError(k.RewardPool.Set(suite.Ctx, types.RewardPool{
		TotalReleased: sdk.NewCoin(defaults.TokenDenom, math.NewInt(9)),
	}))
	suite.Require().NoError(v2.MigrateStore(suite.Ctx, k))
	params, err = k.Params.Get(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().True(params.InflationMin.Equal(defaults.InflationMin))
	suite.Require().True(params.InflationMax.Equal(defaults.InflationMax))
	suite.Require().True(params.SupplyBase.IsZero())
	suite.Require().True(params.InflationRateChange.Equal(defaults.InflationRateChange))
	suite.Require().Equal(defaults.BlocksPerYear, params.BlocksPerYear)

	pool, err := k.RewardPool.Get(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().True(pool.TotalReleased.Amount.Equal(math.NewInt(9)))

	suite.Require().NoError(k.Params.Set(suite.Ctx, types.Params{
		TokenDenom:          defaults.TokenDenom,
		GoalBonded:          defaults.GoalBonded,
		InflationMin:        defaults.InflationMin,
		InflationMax:        math.LegacyNewDec(-1),
		SupplyBase:          math.NewInt(5),
		InflationRateChange: math.LegacyZeroDec(),
		BlocksPerYear:       defaults.BlocksPerYear,
	}))
	suite.Require().NoError(v2.MigrateStore(suite.Ctx, k))
	params, err = k.Params.Get(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().True(params.InflationMax.Equal(defaults.InflationMax))
	suite.Require().True(params.SupplyBase.Equal(math.NewInt(5)))
	suite.Require().True(params.InflationRateChange.Equal(defaults.InflationRateChange))
}

func (suite *MigrateTestSuite) TestMigrateStoreMissingState() {
	suite.Setup()
	k := suite.App.RewardsKeeper
	suite.Require().NoError(k.Params.Remove(suite.Ctx))
	suite.Require().NoError(k.RewardPool.Remove(suite.Ctx))
	suite.Require().NoError(v2.MigrateStore(suite.Ctx, k))
}
