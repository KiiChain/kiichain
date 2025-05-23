package keeper

import (
	"testing"

	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/kiichain/kiichain/v1/x/oracle/utils"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	params := oracleKeeper.GetParams(ctx)
	require.NotNil(t, params)
	require.True(t, len(params.Whitelist) > 0)
}

func TestSetParams(t *testing.T) {
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	// get current params
	params := oracleKeeper.GetParams(ctx)
	require.NotNil(t, params)

	// set params
	beforParam := params.VotePeriod
	params.VotePeriod = 123456
	oracleKeeper.SetParams(ctx, params) // update params

	newParams := oracleKeeper.GetParams(ctx)

	// validation
	require.True(t, beforParam != newParams.VotePeriod)
	require.Equal(t, uint64(123456), newParams.VotePeriod)
}

func TestVotePeriod(t *testing.T) {
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	params := oracleKeeper.GetParams(ctx)
	require.NotNil(t, params)
	require.Equal(t, types.DefaultVotePeriod, params.VotePeriod)
}

func TestVoteThreshold(t *testing.T) {
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	params := oracleKeeper.GetParams(ctx)
	require.NotNil(t, params)
	require.Equal(t, types.DefaultVoteThreshold, params.VoteThreshold)
}

func TestRewardBand(t *testing.T) {
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	params := oracleKeeper.GetParams(ctx)
	require.NotNil(t, params)
	require.Equal(t, types.DefaultRewardBand, params.RewardBand)
}

func TestSetWhitelist(t *testing.T) {
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	params := oracleKeeper.GetParams(ctx)
	require.NotNil(t, params)
	require.Equal(t, types.DefaultWhitelist, params.Whitelist)

	// update the param
	newDenomList := types.DenomList{
		types.Denom{Name: utils.MicroAtomDenom},
		types.Denom{Name: utils.MicroEthDenom},
		types.Denom{Name: utils.MicroKiiDenom},
	}
	oracleKeeper.SetWhitelist(ctx, newDenomList)

	// get new whiteList
	newWhiteList := oracleKeeper.Whitelist(ctx)

	// validation
	require.Equal(t, utils.MicroAtomDenom, newWhiteList[0].Name)
	require.Equal(t, utils.MicroEthDenom, newWhiteList[1].Name)
	require.True(t, len(newWhiteList) == 3)
}

func TestSlashFraction(t *testing.T) {
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	params := oracleKeeper.GetParams(ctx)
	require.NotNil(t, params)
	require.Equal(t, types.DefaultSlashFraction, params.SlashFraction)
}

func TestSlashWindow(t *testing.T) {
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	params := oracleKeeper.GetParams(ctx)
	require.NotNil(t, params)
	require.Equal(t, types.DefaultSlashWindow, params.SlashWindow)
}

func TestMinValidPerWindow(t *testing.T) {
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	params := oracleKeeper.GetParams(ctx)
	require.NotNil(t, params)
	require.Equal(t, types.DefaultMinValidPerWindow, params.MinValidPerWindow)
}

func LookbackDuration(t *testing.T) {
	init := CreateTestInput(t)
	oracleKeeper := init.OracleKeeper
	ctx := init.Ctx

	params := oracleKeeper.GetParams(ctx)
	require.NotNil(t, params)
	require.Equal(t, types.DefaultLookbackDuration, params.LookbackDuration)
}
