package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v7/x/rewards/keeper"
	"github.com/kiichain/kiichain/v7/x/rewards/types"
	v1beta1 "github.com/kiichain/kiichain/v7/x/rewards/types/v1beta1"
)

func (suite *KeeperTestSuite) TestLegacyMsgServer() {
	srv := keeper.NewLegacyMsgServer(suite.App.RewardsKeeper)
	authority := suite.App.RewardsKeeper.GetAuthority()

	_, err := srv.UpdateParams(suite.Ctx, &v1beta1.MsgUpdateParams{
		Authority: suite.TestAccs[0].String(),
		Params:    v1beta1.Params{TokenDenom: "akii"},
	})
	suite.Require().Error(err)

	_, err = srv.UpdateParams(suite.Ctx, &v1beta1.MsgUpdateParams{
		Authority: authority,
		Params:    v1beta1.Params{TokenDenom: "1bad"},
	})
	suite.Require().Error(err)

	bad := types.DefaultParams()
	bad.GoalBonded = math.LegacyZeroDec()
	suite.Require().NoError(suite.App.RewardsKeeper.Params.Set(suite.Ctx, bad))
	_, err = srv.UpdateParams(suite.Ctx, &v1beta1.MsgUpdateParams{
		Authority: authority,
		Params:    v1beta1.Params{TokenDenom: "akii"},
	})
	suite.Require().Error(err)

	original := types.DefaultParams()
	original.SupplyBase = math.NewInt(42)
	suite.Require().NoError(suite.App.RewardsKeeper.Params.Set(suite.Ctx, original))
	_, err = srv.UpdateParams(suite.Ctx, &v1beta1.MsgUpdateParams{
		Authority: authority,
		Params:    v1beta1.Params{TokenDenom: "ukii"},
	})
	suite.Require().NoError(err)

	got, err := suite.App.RewardsKeeper.Params.Get(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal("ukii", got.TokenDenom)
	suite.Require().True(got.SupplyBase.Equal(math.NewInt(42)))
	suite.Require().True(got.GoalBonded.Equal(original.GoalBonded))

	_, err = srv.FundPool(suite.Ctx, &v1beta1.MsgFundPool{})
	suite.Require().Error(err)
	_, err = srv.ChangeSchedule(suite.Ctx, &v1beta1.MsgChangeSchedule{})
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestLegacyProposalUpdateParamsRoundTrip() {
	original := types.DefaultParams()
	original.SupplyBase = math.NewInt(42)
	suite.Require().NoError(suite.App.RewardsKeeper.Params.Set(suite.Ctx, original))

	legacyMsg := &v1beta1.MsgUpdateParams{
		Authority: suite.App.RewardsKeeper.GetAuthority(),
		Params:    v1beta1.Params{TokenDenom: "ukii"},
	}
	packed, err := codectypes.NewAnyWithValue(legacyMsg)
	suite.Require().NoError(err)
	suite.Require().Equal("/kiichain.rewards.v1beta1.MsgUpdateParams", packed.TypeUrl)

	var unpacked sdk.Msg
	suite.Require().NoError(suite.App.InterfaceRegistry().UnpackAny(packed, &unpacked))
	gotMsg, ok := unpacked.(*v1beta1.MsgUpdateParams)
	suite.Require().True(ok)

	srv := keeper.NewLegacyMsgServer(suite.App.RewardsKeeper)
	_, err = srv.UpdateParams(suite.Ctx, gotMsg)
	suite.Require().NoError(err)

	got, err := suite.App.RewardsKeeper.Params.Get(suite.Ctx)
	suite.Require().NoError(err)
	suite.Require().Equal("ukii", got.TokenDenom)
	suite.Require().True(got.SupplyBase.Equal(original.SupplyBase))
	suite.Require().True(got.GoalBonded.Equal(original.GoalBonded))
	suite.Require().True(got.InflationMin.Equal(original.InflationMin))
	suite.Require().True(got.InflationMax.Equal(original.InflationMax))
	suite.Require().True(got.InflationRateChange.Equal(original.InflationRateChange))
	suite.Require().Equal(original.BlocksPerYear, got.BlocksPerYear)
}

func TestLegacyUpdateParamsMissingParams(t *testing.T) {
	ctx, k := setupRewardsKeeper(t, mockBankKeeper{}, mockStakingKeeper{})
	srv := keeper.NewLegacyMsgServer(k)

	_, err := srv.UpdateParams(ctx, &v1beta1.MsgUpdateParams{
		Authority: k.GetAuthority(),
		Params:    v1beta1.Params{TokenDenom: "akii"},
	})
	require.Error(t, err)
}

func (suite *KeeperTestSuite) TestLegacyQuerier() {
	q := keeper.NewLegacyQuerier(suite.App.RewardsKeeper)

	paramsRes, err := q.Params(suite.Ctx, &v1beta1.QueryParamsRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(types.DefaultParams().TokenDenom, paramsRes.Params.TokenDenom)

	_, err = q.ReleaseSchedule(suite.Ctx, &v1beta1.QueryReleaseScheduleRequest{})
	suite.Require().Error(err)

	emptyDenom := types.DefaultParams()
	emptyDenom.TokenDenom = ""
	suite.Require().NoError(suite.App.RewardsKeeper.Params.Set(suite.Ctx, emptyDenom))
	poolRes, err := q.RewardPool(suite.Ctx, &v1beta1.QueryRewardPoolRequest{})
	suite.Require().NoError(err)
	suite.Require().True(poolRes.RewardPool.CommunityPool.IsZero())

	suite.Require().NoError(suite.App.RewardsKeeper.Params.Set(suite.Ctx, types.DefaultParams()))
	coin := sdk.NewCoin("akii", math.NewInt(1000))
	suite.Require().NoError(suite.App.BankKeeper.SendCoinsFromAccountToModule(
		suite.Ctx, suite.TestAccs[0], types.ModuleName, sdk.NewCoins(coin),
	))
	poolRes, err = q.RewardPool(suite.Ctx, &v1beta1.QueryRewardPoolRequest{})
	suite.Require().NoError(err)
	suite.Require().True(poolRes.RewardPool.CommunityPool.AmountOf("akii").Equal(math.LegacyNewDec(1000)))
}

func TestLegacyQuerierMissingParams(t *testing.T) {
	ctx, k := setupRewardsKeeper(t, mockBankKeeper{}, mockStakingKeeper{})
	q := keeper.NewLegacyQuerier(k)

	_, err := q.Params(ctx, &v1beta1.QueryParamsRequest{})
	require.Error(t, err)
	_, err = q.RewardPool(ctx, &v1beta1.QueryRewardPoolRequest{})
	require.Error(t, err)
}
