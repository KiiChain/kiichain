package oracle_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/kiichain/kiichain/v7/x/oracle"
	"github.com/kiichain/kiichain/v7/x/oracle/keeper"
	"github.com/kiichain/kiichain/v7/x/oracle/types"
	"github.com/kiichain/kiichain/v7/x/oracle/utils"
)

// TestInitGenesisPopulatesVoteTarget is a regression test for the first-period
// lack of votes gap
func TestInitGenesisPopulatesVoteTarget(t *testing.T) {
	input := keeper.CreateTestInput(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// Clear VoteTarget to reproduce the pre-fix state (empty at genesis).
	err := oracleKeeper.VoteTarget.Clear(ctx, nil)
	require.NoError(t, err)
	targets, err := oracleKeeper.GetVoteTargets(ctx)
	require.NoError(t, err)
	require.Empty(t, targets, "precondition: VoteTarget must be empty before InitGenesis")

	// Build a genesis state with a known whitelist.
	whitelist := types.DenomList{
		{Name: utils.BtcDenom},
		{Name: utils.EthDenom},
		{Name: utils.SolDenom},
	}
	params := types.DefaultParams()
	params.Whitelist = whitelist
	genesis := types.NewGenesisState(
		params,
		[]types.ExchangeRateTuple{},
		[]types.FeederDelegation{},
		[]types.PenaltyCounter{},
		[]types.AggregateExchangeRateVote{},
		types.PriceSnapshots{},
		[]types.VotePenaltyCounter{},
	)

	err = oracle.InitGenesis(ctx, oracleKeeper, genesis)
	require.NoError(t, err)

	// Every whitelisted denom must be present in VoteTarget after InitGenesis.
	for _, denom := range whitelist {
		found, err := oracleKeeper.VoteTarget.Has(ctx, denom.Name)
		require.NoError(t, err)
		require.True(t, found, "VoteTarget missing denom %q after InitGenesis", denom.Name)
	}

	// VoteTarget must not contain extra entries.
	targets, err = oracleKeeper.GetVoteTargets(ctx)
	require.NoError(t, err)
	require.Len(t, targets, len(whitelist))
}

func TestExportInitGenesis(t *testing.T) {
	// Prepare env
	input, _ := oracle.SetUp(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// Prepare genesis to be exported
	exchangeRateVote, err := types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{{Denom: utils.AtomDenom, ExchangeRate: math.LegacyNewDec(123)}}, keeper.ValAddrs[0])
	require.NoError(t, err)

	snapshot1 := types.NewPriceSnapshot(int64(3600),
		types.PriceSnapshotItems{
			{
				Denom: utils.AtomDenom,
				OracleExchangeRate: types.OracleExchangeRate{
					ExchangeRate: math.LegacyNewDec(12),
					LastUpdate:   math.NewInt(3600),
				},
			},
			{
				Denom: utils.EthDenom,
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
				Denom: utils.AtomDenom,
				OracleExchangeRate: types.OracleExchangeRate{
					ExchangeRate: math.LegacyNewDec(15),
					LastUpdate:   math.NewInt(3700),
				},
			},
			{
				Denom: utils.EthDenom,
				OracleExchangeRate: types.OracleExchangeRate{
					ExchangeRate: math.LegacyNewDec(13),
					LastUpdate:   math.NewInt(3700),
				},
			},
		},
	)

	err = oracleKeeper.FeederDelegation.Set(ctx, keeper.ValAddrs[0], keeper.Addrs[1].String())
	require.NoError(t, err)
	err = oracleKeeper.SetBaseExchangeRateWithDefault(ctx, utils.AtomDenom, math.LegacyNewDec(123))
	require.NoError(t, err)
	err = oracleKeeper.AggregateExchangeRateVote.Set(ctx, keeper.ValAddrs[0], exchangeRateVote)
	require.NoError(t, err)

	err = oracleKeeper.VoteTarget.Set(ctx, utils.AtomDenom, types.Denom{Name: utils.AtomDenom})
	require.NoError(t, err)
	err = oracleKeeper.VoteTarget.Set(ctx, utils.EthDenom, types.Denom{Name: utils.EthDenom})
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
