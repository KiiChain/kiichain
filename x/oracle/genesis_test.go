package oracle_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain/v1/x/oracle"
	"github.com/kiichain/kiichain/v1/x/oracle/keeper"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/kiichain/kiichain/v1/x/oracle/utils"
	"github.com/stretchr/testify/require"
)

func TestExportInitGenesis(t *testing.T) {
	// Prepare env
	input, _ := oracle.SetUp(t)
	oracleKeeper := input.OracleKeeper
	ctx := input.Ctx

	// Prepare genesis to be exported
	exchangeRateVote, err := types.NewAggregateExchangeRateVote(types.ExchangeRateTuples{{Denom: utils.MicroAtomDenom, ExchangeRate: sdk.NewDec(123)}}, keeper.ValAddrs[0])
	require.NoError(t, err)

	snapshot1 := types.NewPriceSnapshot(int64(3600),
		types.PriceSnapshotItems{
			{
				Denom: utils.MicroAtomDenom,
				OracleExchangeRate: types.OracleExchangeRate{
					ExchangeRate: sdk.NewDec(12),
					LastUpdate:   sdk.NewInt(3600),
				},
			},
			{
				Denom: utils.MicroEthDenom,
				OracleExchangeRate: types.OracleExchangeRate{
					ExchangeRate: sdk.NewDec(10),
					LastUpdate:   sdk.NewInt(3600),
				},
			},
		},
	)

	snapshot2 := types.NewPriceSnapshot(int64(3700),
		types.PriceSnapshotItems{
			{
				Denom: utils.MicroAtomDenom,
				OracleExchangeRate: types.OracleExchangeRate{
					ExchangeRate: sdk.NewDec(15),
					LastUpdate:   sdk.NewInt(3700),
				},
			},
			{
				Denom: utils.MicroEthDenom,
				OracleExchangeRate: types.OracleExchangeRate{
					ExchangeRate: sdk.NewDec(13),
					LastUpdate:   sdk.NewInt(3700),
				},
			},
		},
	)

	oracleKeeper.SetFeederDelegation(ctx, keeper.ValAddrs[0], keeper.Addrs[1])
	oracleKeeper.SetBaseExchangeRate(ctx, utils.MicroAtomDenom, sdk.NewDec(123))
	oracleKeeper.SetAggregateExchangeRateVote(ctx, keeper.ValAddrs[0], exchangeRateVote)
	oracleKeeper.SetVoteTarget(ctx, utils.MicroAtomDenom)
	oracleKeeper.SetVoteTarget(ctx, utils.MicroEthDenom)
	oracleKeeper.SetVotePenaltyCounter(ctx, keeper.ValAddrs[0], 2, 3, 0)
	oracleKeeper.SetVotePenaltyCounter(ctx, keeper.ValAddrs[1], 4, 5, 0)
	oracleKeeper.AddPriceSnapshot(ctx, snapshot1)
	oracleKeeper.AddPriceSnapshot(ctx, snapshot2)

	// Export genesis
	genesis := oracle.ExportGenesis(ctx, oracleKeeper)

	// Create new test env
	newInput := keeper.CreateTestInput(t)
	neworacleKeeper := newInput.OracleKeeper
	newctx := newInput.Ctx

	// use the exported genesis on the new env
	oracle.InitGenesis(newctx, neworacleKeeper, &genesis)
	newGenesis := oracle.ExportGenesis(newctx, neworacleKeeper)

	// validation
	require.Equal(t, genesis, newGenesis)
}
