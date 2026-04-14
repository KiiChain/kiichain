package v2_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/app/apptesting"
	v2 "github.com/kiichain/kiichain/v7/x/tokenfactory/migrations/v2"
	"github.com/kiichain/kiichain/v7/x/tokenfactory/types"
)

type MigrateTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(MigrateTestSuite))
}

// TestMigrateStore verifies that MigrateStore correctly backfills the admin
// secondary index for all existing tokenfactory denoms
func (suite *MigrateTestSuite) TestMigrateStore() {
	suite.Setup()

	fundAmount := sdk.NewCoins(
		sdk.NewCoin(
			types.DefaultParams().DenomCreationFee[0].Denom,
			types.DefaultParams().DenomCreationFee[0].Amount.MulRaw(100),
		),
	)
	for _, acc := range suite.TestAccs {
		suite.FundAcc(acc, fundAmount)
	}

	k := suite.App.TokenFactoryKeeper

	// acc0 owns two denoms; acc1 owns one
	acc0 := suite.TestAccs[0].String()
	acc1 := suite.TestAccs[1].String()

	denom0, err := k.CreateDenom(suite.Ctx, acc0, "akii")
	suite.Require().NoError(err)

	denom1, err := k.CreateDenom(suite.Ctx, acc0, "uusdc")
	suite.Require().NoError(err)

	denom2, err := k.CreateDenom(suite.Ctx, acc1, "test")
	suite.Require().NoError(err)

	// Simulate pre-migration state: remove all admin index entries so the
	// index looks as if it was never populated
	adminDenoms := map[string][]string{
		acc0: {denom0, denom1},
		acc1: {denom2},
	}
	for admin, denoms := range adminDenoms {
		for _, denom := range denoms {
			k.GetAdminPrefixStore(suite.Ctx, admin).Delete([]byte(denom))
		}
	}

	// Confirm the index is empty before the migration
	for _, acc := range []string{acc0, acc1} {
		denoms, _, err := k.GetDenomsFromAdmin(suite.Ctx, acc, nil)
		suite.Require().NoError(err)
		suite.Require().Empty(denoms, "admin index should be empty before migration for %s", acc)
	}

	// Run the migration
	suite.Require().NoError(v2.MigrateStore(suite.Ctx, k))

	// After migration, each admin must have its denoms back in the index
	denoms0, _, err := k.GetDenomsFromAdmin(suite.Ctx, acc0, nil)
	suite.Require().NoError(err)
	suite.Require().ElementsMatch([]string{denom0, denom1}, denoms0)

	denoms1, _, err := k.GetDenomsFromAdmin(suite.Ctx, acc1, nil)
	suite.Require().NoError(err)
	suite.Require().ElementsMatch([]string{denom2}, denoms1)
}

// TestMigrateStore_EmptyState verifies that migrating with no denoms succeeds
func (suite *MigrateTestSuite) TestMigrateStore_EmptyState() {
	suite.Setup()
	suite.Require().NoError(v2.MigrateStore(suite.Ctx, suite.App.TokenFactoryKeeper))
}
