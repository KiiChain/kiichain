package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/kiichain/kiichain/v1/x/oracle/utils"
)

func TestQueryParams(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// query params
	res, err := querier.Params(ctx, &types.QueryParamsRequest{})

	// validation
	require.NoError(t, err)
	params, err := oracleKeeper.Params.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, params, *res.Params)
}

func TestQueryExchangeRate(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// insert data on the module
	rate := math.LegacyNewDec(12)
	err := oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroAtomDenom, rate)
	require.NoError(t, err)
	// query params
	res, err := querier.ExchangeRate(ctx, &types.QueryExchangeRateRequest{Denom: utils.MicroAtomDenom})

	// validation
	require.NoError(t, err)
	require.Equal(t, rate, res.OracleExchangeRate.ExchangeRate)
}

func TestQueryExchangeRates(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// insert data on the module
	rate := math.LegacyNewDec(12)
	err := oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroAtomDenom, rate)
	require.NoError(t, err)
	err = oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroEthDenom, rate)
	require.NoError(t, err)

	// query params
	res, err := querier.ExchangeRates(ctx, &types.QueryExchangeRatesRequest{})

	// validation
	require.NoError(t, err)
	require.Equal(t, 2, len(res.DenomOracleExchangeRate))
}

func TestQueryActives(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// insert data on the module
	rate := math.LegacyNewDec(12)
	err := oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroAtomDenom, rate)
	require.NoError(t, err)
	err = oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroEthDenom, rate)
	require.NoError(t, err)

	// query params
	res, err := querier.Actives(ctx, &types.QueryActivesRequest{})

	// validation
	require.NoError(t, err)
	require.Equal(t, 2, len(res.Actives))
	require.Equal(t, utils.MicroAtomDenom, res.Actives[0])
	require.Equal(t, utils.MicroEthDenom, res.Actives[1])
}

func TestQueryVoteTargets(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// insert data on the module
	err := oracleKeeper.VoteTarget.Clear(ctx, nil)
	require.NoError(t, err)
	err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroAtomDenom, types.Denom{Name: utils.MicroAtomDenom})
	require.NoError(t, err)
	err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroEthDenom, types.Denom{Name: utils.MicroEthDenom})
	require.NoError(t, err)

	// query params
	res, err := querier.VoteTargets(ctx, &types.QueryVoteTargetsRequest{})

	// validation
	require.NoError(t, err)
	require.Equal(t, 2, len(res.VoteTargets))
	require.Equal(t, utils.MicroAtomDenom, res.VoteTargets[0])
	require.Equal(t, utils.MicroEthDenom, res.VoteTargets[1])
}

func TestQueryPriceSnapshotHistory(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// insert data on the module
	snapShot1 := types.NewPriceSnapshot(1, types.PriceSnapshotItems{
		types.NewPriceSnapshotItem(utils.MicroEthDenom, types.OracleExchangeRate{
			ExchangeRate: math.LegacyNewDec(11),
			LastUpdate:   math.NewInt(20),
		}),
		types.NewPriceSnapshotItem(utils.MicroAtomDenom, types.OracleExchangeRate{
			ExchangeRate: math.LegacyNewDec(12),
			LastUpdate:   math.NewInt(20),
		}),
	})

	snapShot2 := types.NewPriceSnapshot(2, types.PriceSnapshotItems{
		types.NewPriceSnapshotItem(utils.MicroEthDenom, types.OracleExchangeRate{
			ExchangeRate: math.LegacyNewDec(21),
			LastUpdate:   math.NewInt(30),
		}),
		types.NewPriceSnapshotItem(utils.MicroAtomDenom, types.OracleExchangeRate{
			ExchangeRate: math.LegacyNewDec(22),
			LastUpdate:   math.NewInt(30),
		}),
	})

	priceSnapshots := types.PriceSnapshots{snapShot1, snapShot2}

	err := oracleKeeper.PriceSnapshot.Set(ctx, priceSnapshots[0].SnapshotTimestamp, priceSnapshots[0])
	require.NoError(t, err)
	err = oracleKeeper.PriceSnapshot.Set(ctx, priceSnapshots[1].SnapshotTimestamp, priceSnapshots[1])
	require.NoError(t, err)

	// query params
	res, err := querier.PriceSnapshotHistory(ctx, &types.QueryPriceSnapshotHistoryRequest{})

	// validation
	require.NoError(t, err)
	require.Equal(t, priceSnapshots, res.PriceSnapshot)
}

func TestQueryTwaps(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// insert data on the module
	exchangeRate1 := types.OracleExchangeRate{
		ExchangeRate:        math.LegacyNewDec(1),
		LastUpdate:          math.NewInt(1),
		LastUpdateTimestamp: 1,
	}
	exchangeRate2 := types.OracleExchangeRate{
		ExchangeRate:        math.LegacyNewDec(2),
		LastUpdate:          math.NewInt(2),
		LastUpdateTimestamp: 2,
	}
	snapshotItem1 := types.NewPriceSnapshotItem(utils.MicroKiiDenom, exchangeRate1)
	snapshotItem2 := types.NewPriceSnapshotItem(utils.MicroEthDenom, exchangeRate2)
	snapshot1 := types.NewPriceSnapshot(1, types.PriceSnapshotItems{snapshotItem1, snapshotItem1})
	snapshot2 := types.NewPriceSnapshot(2, types.PriceSnapshotItems{snapshotItem2, snapshotItem2})

	err := oracleKeeper.PriceSnapshot.Set(ctx, snapshot1.SnapshotTimestamp, snapshot1)
	require.NoError(t, err)
	err = oracleKeeper.PriceSnapshot.Set(ctx, snapshot2.SnapshotTimestamp, snapshot2)
	require.NoError(t, err)

	// set vote target on params
	params := types.DefaultParams()
	err = oracleKeeper.Params.Set(ctx, params)
	require.NoError(t, err)
	for _, denom := range params.Whitelist {
		err = oracleKeeper.VoteTarget.Set(ctx, denom.Name, types.Denom{Name: denom.Name})
		require.NoError(t, err)
	}

	// query params
	res, err := querier.Twaps(ctx, &types.QueryTwapsRequest{LookbackSeconds: 3600})

	// validation
	require.NoError(t, err)
	require.Equal(t, utils.MicroEthDenom, res.OracleTwap[0].Denom)
	require.Equal(t, math.LegacyNewDec(2), res.OracleTwap[0].Twap)
}

func TestQueryFeederDelegation(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// delegate voting power
	oracleKeeper.SetFeederDelegation(ctx, ValAddrs[0], Addrs[0])

	// query params
	res, err := querier.FeederDelegation(ctx, &types.QueryFeederDelegationRequest{ValidatorAddr: ValAddrs[0].String()})

	// validation
	require.NoError(t, err)
	require.Equal(t, Addrs[0].String(), res.FeedAddr)
}

func TestQueryVotePenaltyCounter(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// calculate the expected slashwindow
	voteCounter := types.NewVotePenaltyCounter(30, 20, 10)
	err := oracleKeeper.VotePenaltyCounter.Set(ctx, ValAddrs[0], voteCounter) // Set the voting info
	require.NoError(t, err)

	// query params
	res, err := querier.VotePenaltyCounter(ctx, &types.QueryVotePenaltyCounterRequest{ValidatorAddr: ValAddrs[0].String()})

	// validation
	require.NoError(t, err)
	require.Equal(t, res.VotePenaltyCounter.AbstainCount, voteCounter.AbstainCount)
	require.Equal(t, res.VotePenaltyCounter.MissCount, voteCounter.MissCount)
	require.Equal(t, res.VotePenaltyCounter.SuccessCount, voteCounter.SuccessCount)
}

func TestQuerySlashWindow(t *testing.T) {
	// prepare env
	input := CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// create query server
	querier := NewQueryServer(oracleKeeper)

	// calculate the expected slashwindow
	params := types.DefaultParams()
	expectedWindowProgress := (uint64(ctx.BlockHeight()) % params.SlashWindow) / params.VotePeriod

	// query params
	res, err := querier.SlashWindow(ctx, &types.QuerySlashWindowRequest{})

	// validation
	require.NoError(t, err)
	require.Equal(t, expectedWindowProgress, res.WindowProgress)
}
