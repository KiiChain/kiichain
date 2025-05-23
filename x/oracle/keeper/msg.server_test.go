package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/kiichain/kiichain/v1/x/oracle/utils"
	"github.com/stretchr/testify/require"
)

func TestAggregateExchangeRateVote(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	stakingKeeper := input.StakingKeeper
	ctx := input.Ctx
	stakingHandler := staking.NewHandler(stakingKeeper)

	// create msg server
	msgServer := NewMsgServer(oracleKeeper)

	// Create validators
	stakingAmount := sdk.TokensFromConsensusPower(50, sdk.DefaultPowerReduction)
	val := NewTestMsgCreateValidator(ValAddrs[0], ValPubKeys[0], stakingAmount)

	// Register validators
	_, err := stakingHandler(ctx, val)
	require.NoError(t, err)

	// execute staking endblocker to start validators bonding
	staking.EndBlocker(ctx, stakingKeeper)

	// send messages
	exchangeRate := sdk.NewDec(12).String() + utils.MicroUsdcDenom
	context := sdk.WrapSDKContext(ctx)
	_, err = msgServer.AggregateExchangeRateVote(context, types.NewMsgAggregateExchangeRateVote(exchangeRate, Addrs[0], ValAddrs[0]))

	// validation
	require.NoError(t, err)
}

func TestDelegateFeedConsent(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	stakingKeeper := input.StakingKeeper
	ctx := input.Ctx
	stakingHandler := staking.NewHandler(stakingKeeper)

	// create msg server
	msgServer := NewMsgServer(oracleKeeper)

	// Create validators
	stakingAmount := sdk.TokensFromConsensusPower(50, sdk.DefaultPowerReduction)
	val := NewTestMsgCreateValidator(ValAddrs[0], ValPubKeys[0], stakingAmount)

	// Register validators
	_, err := stakingHandler(ctx, val)
	require.NoError(t, err)

	// execute staking endblocker to start validators bonding
	staking.EndBlocker(ctx, stakingKeeper)

	// send messages
	context := sdk.WrapSDKContext(ctx)
	_, err = msgServer.DelegateFeedConsent(context, types.NewMsgDelegateFeedConsent(ValAddrs[0], Addrs[0]))
	require.NoError(t, err)

	// create query server
	querier := NewQueryServer(oracleKeeper)
	res, err := querier.FeederDelegation(context, &types.QueryFeederDelegationRequest{ValidatorAddr: ValAddrs[0].String()})
	require.NoError(t, err)

	// validation
	require.Equal(t, Addrs[0].String(), res.FeedAddr)
}
