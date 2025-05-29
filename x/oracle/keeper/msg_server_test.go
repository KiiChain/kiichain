package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/kiichain/kiichain/v1/x/oracle/utils"
)

func TestAggregateExchangeRateVote(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	stakingKeeper := input.StakingKeeper
	ctx := input.Ctx
	msgServerStaking := stakingkeeper.NewMsgServerImpl(&stakingKeeper)

	// create msg server
	msgServer := NewMsgServer(oracleKeeper)

	// Create validators
	stakingAmount := sdk.TokensFromConsensusPower(50, sdk.DefaultPowerReduction)
	val := NewTestMsgCreateValidator(ValAddrs[0], ValPubKeys[0], stakingAmount)

	// Register validators
	_, err := msgServerStaking.CreateValidator(ctx, val)
	require.NoError(t, err)

	// execute staking endblocker to start validators bonding
	_, err = stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	// send messages
	exchangeRate := math.LegacyNewDec(12).String() + utils.MicroUsdcDenom
	_, err = msgServer.AggregateExchangeRateVote(ctx, types.NewMsgAggregateExchangeRateVote(exchangeRate, Addrs[0], ValAddrs[0]))

	// validation
	require.NoError(t, err)
}

func TestDelegateFeedConsent(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	stakingKeeper := input.StakingKeeper
	ctx := input.Ctx
	msgServerStaking := stakingkeeper.NewMsgServerImpl(&stakingKeeper)

	// create msg server
	msgServer := NewMsgServer(oracleKeeper)

	// Create validators
	stakingAmount := sdk.TokensFromConsensusPower(50, sdk.DefaultPowerReduction)
	val := NewTestMsgCreateValidator(ValAddrs[0], ValPubKeys[0], stakingAmount)

	// Register validators
	_, err := msgServerStaking.CreateValidator(ctx, val)
	require.NoError(t, err)

	// execute staking endblocker to start validators bonding
	_, err = stakingKeeper.EndBlocker(ctx)
	require.NoError(t, err)

	// send messages
	_, err = msgServer.DelegateFeedConsent(ctx, types.NewMsgDelegateFeedConsent(ValAddrs[0], Addrs[0]))
	require.NoError(t, err)

	// create query server
	querier := NewQueryServer(oracleKeeper)
	res, err := querier.FeederDelegation(ctx, &types.QueryFeederDelegationRequest{ValidatorAddr: ValAddrs[0].String()})
	require.NoError(t, err)

	// validation
	require.Equal(t, Addrs[0].String(), res.FeedAddr)
}
