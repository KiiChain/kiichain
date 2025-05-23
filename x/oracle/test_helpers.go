package oracle

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/kiichain/kiichain/v1/x/oracle/keeper"
	"github.com/stretchr/testify/require"
)

var (
	stakingAmount       = sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	randomAExchangeRate = sdk.NewDec(1700)
	randomBExchangeRate = sdk.NewDecWithPrec(4882, 2)
)

func SetUp(t *testing.T) (keeper.TestInput, sdk.Handler) {
	input := keeper.CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	stakingKeeper := input.StakingKeeper
	ctx := input.Ctx

	// Update params to test easier and faster
	params := oracleKeeper.GetParams(ctx)
	params.VotePeriod = 1
	params.SlashWindow = 100
	oracleKeeper.SetParams(ctx, params)

	stakingParams := stakingKeeper.GetParams(ctx)
	stakingParams.MinCommissionRate = sdk.NewDecWithPrec(0, 2) // 0.00
	stakingKeeper.SetParams(ctx, stakingParams)

	// Create handlers
	oracleHandler := NewHandler(oracleKeeper)
	stakingHandler := staking.NewHandler(stakingKeeper)

	// Create validators
	val0 := keeper.NewTestMsgCreateValidator(keeper.ValAddrs[0], keeper.ValPubKeys[0], stakingAmount)
	val1 := keeper.NewTestMsgCreateValidator(keeper.ValAddrs[1], keeper.ValPubKeys[1], stakingAmount)
	val2 := keeper.NewTestMsgCreateValidator(keeper.ValAddrs[2], keeper.ValPubKeys[2], stakingAmount)

	// Register validators
	_, err := stakingHandler(ctx, val0)
	require.NoError(t, err)
	_, err = stakingHandler(ctx, val1)
	require.NoError(t, err)
	_, err = stakingHandler(ctx, val2)
	require.NoError(t, err)

	// execute staking endblocker to start validators bonding
	staking.EndBlocker(ctx, stakingKeeper)

	return input, oracleHandler
}
