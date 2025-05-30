package oracle_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/kiichain/kiichain/v1/x/oracle"
	"github.com/kiichain/kiichain/v1/x/oracle/keeper"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/kiichain/kiichain/v1/x/oracle/utils"
)

func TestExportInitGenesis(t *testing.T) {
	// Prepare env
	input, _ := oracle.SetUp(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// Prepare genesis to be exported
	exchangeRateVote, err := types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{{Denom: utils.MicroAtomDenom, ExchangeRate: math.LegacyNewDec(123)}}, keeper.ValAddrs[0])
	require.NoError(t, err)

	snapshot1 := types.NewPriceSnapshot(int64(3600),
		types.PriceSnapshotItems{
			{
				Denom: utils.MicroAtomDenom,
				OracleExchangeRate: types.OracleExchangeRate{
					ExchangeRate: math.LegacyNewDec(12),
					LastUpdate:   math.NewInt(3600),
				},
			},
			{
				Denom: utils.MicroEthDenom,
				OracleExchangeRate: types.OracleExchangeRate{
					ExchangeRate: math.LegacyNewDec(10),
					LastUpdate:   math.NewInt(3600),
				},
			},
		},
	)

	snapshot2 := types.NewPriceSnapshot(int64(3700),
		types.PriceSnapshotItems{
			{
				Denom: utils.MicroAtomDenom,
				OracleExchangeRate: types.OracleExchangeRate{
					ExchangeRate: math.LegacyNewDec(15),
					LastUpdate:   math.NewInt(3700),
				},
			},
			{
				Denom: utils.MicroEthDenom,
				OracleExchangeRate: types.OracleExchangeRate{
					ExchangeRate: math.LegacyNewDec(13),
					LastUpdate:   math.NewInt(3700),
				},
			},
		},
	)

	oracleKeeper.SetFeederDelegation(ctx, keeper.ValAddrs[0], keeper.Addrs[1])
	err = oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroAtomDenom, math.LegacyNewDec(123))
	require.NoError(t, err)
	err = oracleKeeper.AggregateExchangeRateVote.Set(ctx, keeper.ValAddrs[0], exchangeRateVote)
	require.NoError(t, err)

	err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroAtomDenom, types.Denom{Name: utils.MicroAtomDenom})
	require.NoError(t, err)
	err = oracleKeeper.VoteTarget.Set(ctx, utils.MicroEthDenom, types.Denom{Name: utils.MicroEthDenom})
	require.NoError(t, err)

	err = oracleKeeper.VotePenaltyCounter.Set(ctx, keeper.ValAddrs[0], types.NewVotePenaltyCounter(2, 3, 0))
	require.NoError(t, err)
	err = oracleKeeper.VotePenaltyCounter.Set(ctx, keeper.ValAddrs[1], types.NewVotePenaltyCounter(4, 5, 0))
	require.NoError(t, err)
	err = oracleKeeper.AddPriceSnapshot(ctx, snapshot1)
	require.NoError(t, err)
	err = oracleKeeper.AddPriceSnapshot(ctx, snapshot2)
	require.NoError(t, err)

	// Export genesis
	genesis, err := oracle.ExportGenesis(ctx, oracleKeeper)
	require.NoError(t, err)

	// Create new test env
	newInput := keeper.CreateTestInput(t)
	neworacleKeeper := newInput.OracleKeeper
	newctx := newInput.Ctx

	// use the exported genesis on the new env
	err = oracle.InitGenesis(newctx, neworacleKeeper, genesis)
	require.NoError(t, err)
	newGenesis, err := oracle.ExportGenesis(newctx, neworacleKeeper)
	require.NoError(t, err)

	// validation
	require.Equal(t, genesis, newGenesis)
}
