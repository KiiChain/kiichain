package oracle

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v1/x/oracle/keeper"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
)

// InitGenesis initialize the module with the default parameters
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, data *types.GenesisState) error {
	// Start the genesis with the data input
	err := keeper.Params.Set(ctx, data.Params)
	if err != nil {
		return err
	}

	// Iterate over the feeder delegation list to set the feeder
	for _, feederDelegation := range data.FeederDelegations {
		// Get the validator address
		valAddress, err := sdk.ValAddressFromBech32(feederDelegation.ValidatorAddress)
		if err != nil {
			panic(err)
		}

		// Get the delegator address
		feederAddress, err := sdk.AccAddressFromBech32(feederDelegation.FeederAddress)
		if err != nil {
			panic(err)
		}

		// Assign the feeder delegator on the module
		keeper.SetFeederDelegation(ctx, valAddress, feederAddress)
	}

	// Assign on the KVStore the exchange rate
	for _, exchangeRate := range data.ExchangeRates {
		err := keeper.SetBaseExchangeRate(ctx, exchangeRate.Denom, exchangeRate.ExchangeRate)
		if err != nil {
			return err
		}
	}

	// Add the penaltyCounter array to the KVStore
	for _, penaltyCounter := range data.PenaltyCounters {
		operator, err := sdk.ValAddressFromBech32(penaltyCounter.ValidatorAddress)
		if err != nil {
			panic(err)
		}

		keeper.SetVotePenaltyCounter(ctx, operator, penaltyCounter.VotePenaltyCounter.MissCount,
			penaltyCounter.VotePenaltyCounter.AbstainCount, penaltyCounter.VotePenaltyCounter.SuccessCount)
	}

	// Add the AggregateExchangeRateVotes to the KVStore defined on the input object
	for _, aggregateExchange := range data.AggregateExchangeRateVotes {
		valAddress, err := sdk.ValAddressFromBech32(aggregateExchange.Voter)
		if err != nil {
			panic(err)
		}

		err = keeper.AggregateExchangeRateVote.Set(ctx, valAddress, aggregateExchange)
		if err != nil {
			return err
		}
	}

	// Add the price snapshots to the KVStore defined on the input object
	for _, priceSnapshot := range data.PriceSnapshots {
		err = keeper.AddPriceSnapshot(ctx, priceSnapshot)
		if err != nil {
			return err
		}
	}

	// Check if the module account exists
	moduleAccount := keeper.GetOracleAccount(ctx)
	if moduleAccount == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	return nil
}

// ExportGenesis collect and return the params of the blockchain
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) (*types.GenesisState, error) {
	// Current params of the module
	params, err := keeper.Params.Get(ctx)
	if err != nil {
		panic(err)
	}

	// Extract the FeederDelegation array
	feederDelegations := []types.FeederDelegation{}
	keeper.IterateFeederDelegations(ctx, func(valAddr sdk.ValAddress, delegatedFeeder string) (bool, error) {
		feederDelegations = append(feederDelegations, types.FeederDelegation{
			FeederAddress:    delegatedFeeder,
			ValidatorAddress: valAddr.String(),
		})
		return false, nil
	})

	// Extract the exchangeRatesTuple
	exchangeRates := []types.ExchangeRateTuple{}
	keeper.IterateBaseExchangeRates(ctx, func(denom string, exchangeRate types.OracleExchangeRate) (bool, error) {
		exRate := types.ExchangeRateTuple{Denom: denom, ExchangeRate: exchangeRate.ExchangeRate}
		exchangeRates = append(exchangeRates, exRate)
		return false, nil
	})

	// Extract penalty counters
	penaltyCounters := []types.PenaltyCounter{}
	err = keeper.VotePenaltyCounter.Walk(ctx, nil, func(operator sdk.ValAddress, votePenaltyCounter types.VotePenaltyCounter) (bool, error) {
		penalty := types.PenaltyCounter{ValidatorAddress: operator.String(), VotePenaltyCounter: &votePenaltyCounter}
		penaltyCounters = append(penaltyCounters, penalty)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	// Extract Aggregate exchange rate votes
	aggregateExchangeRateVotes := []types.AggregateExchangeRateVote{}
	err = keeper.AggregateExchangeRateVote.Walk(ctx, nil, func(voterAddr sdk.ValAddress, aggregateVote types.AggregateExchangeRateVote) (bool, error) {
		aggregateExchangeRateVotes = append(aggregateExchangeRateVotes, aggregateVote)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	// Extract priceSnapshots
	priceSnapshots := []types.PriceSnapshot{}
	err = keeper.PriceSnapshot.Walk(ctx, nil, func(_ int64, snapshot types.PriceSnapshot) (bool, error) {
		priceSnapshots = append(priceSnapshots, snapshot)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	// Extract votePenaltyCounters
	votePenaltyCounters := []types.VotePenaltyCounter{}
	err = keeper.VotePenaltyCounter.Walk(ctx, nil, func(operator sdk.ValAddress, votePenaltyCounter types.VotePenaltyCounter) (bool, error) {
		votePenaltyCounters = append(votePenaltyCounters, votePenaltyCounter)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	// Build the genesis state
	genesisState := types.NewGenesisState(
		params,
		exchangeRates,
		feederDelegations,
		penaltyCounters,
		aggregateExchangeRateVotes,
		priceSnapshots,
		votePenaltyCounters,
	)

	return genesisState, nil
}
