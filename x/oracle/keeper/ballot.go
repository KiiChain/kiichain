package keeper

import (
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/kiichain/kiichain/v1/x/oracle/types"
)

// OrganizeBallotByDenom iterates over the map with validators and create its voting tally.
// returns a map with the denom and its ballot (denom alphabetical ordered)
func (k Keeper) OrganizeBallotByDenom(ctx sdk.Context, validatorClaimMap map[string]types.Claim) map[string]types.ExchangeRateBallot {
	votes := map[string]types.ExchangeRateBallot{} // Here I will collect the array of votes by denom

	// Aggregate votes by denom
	aggregateHandler := func(voterAddr sdk.ValAddress, aggregateVote types.AggregateExchangeRateVote) (bool, error) {
		// Aggregate only for validators who have registered on the map
		claim, ok := validatorClaimMap[aggregateVote.Voter]

		if ok {
			power := claim.Power
			for _, tuple := range aggregateVote.ExchangeRateTuples {
				tmpPower := power

				// Validate invalids exchange rates
				if !tuple.ExchangeRate.IsPositive() {
					tmpPower = 0
				}

				vote := types.NewVoteForTally(tuple.ExchangeRate, tuple.Denom, voterAddr, tmpPower) // Create validator vote
				votes[tuple.Denom] = append(votes[tuple.Denom], vote)                               // Append vote on that specific denom
			}
		}
		return false, nil
	}

	k.IterateAggregateExchangeRateVotes(ctx, aggregateHandler)

	// sort created ballot
	for denom, ballot := range votes {
		sort.Sort(ballot) // sort by denom
		votes[denom] = ballot
	}

	return votes
}

// ClearBallots clears all votes from the KV Store
func (k Keeper) ClearBallots(ctx sdk.Context) {
	// Clear all aggregate votes
	k.IterateAggregateExchangeRateVotes(ctx, func(voterAddr sdk.ValAddress, aggregateVote types.AggregateExchangeRateVote) (bool, error) {
		k.DeleteAggregateExchangeRateVote(ctx, voterAddr)
		return false, nil
	})
}

// ApplyWhitelist update the vote target on the KVStore if there are new desired denoms on the parameters
// for the new denoms on the whitelist creaste its mili and micro version
func (k Keeper) ApplyWhitelist(ctx sdk.Context, whitelist types.DenomList, voteTargets map[string]types.Denom) {
	// Check if there is an update in whitelist
	updateRequire := false
	if len(voteTargets) != len(whitelist) {
		updateRequire = true
	}

	// iterate whitelist and check for an item on the whitelist but no on the vote target list
	for _, item := range whitelist {
		if _, ok := voteTargets[item.Name]; !ok {
			updateRequire = true
			break
		}
	}

	if updateRequire {
		k.ClearVoteTargets(ctx) // Delete the current targets on the KVStore

		// Iterate the new whitelist
		for _, item := range whitelist {
			k.SetVoteTarget(ctx, item.Name)

			// Register meta data to bank module
			_, ok := k.bankKeeper.GetDenomMetaData(ctx, item.Name)
			if !ok {
				base := item.Name
				display := base[1:] // remove the first character. i.e: uKII -> display = KII
				nameSymbol := strings.ToUpper(display)

				// define meta data of the param and its mili and micro
				// i.e: 1 KII = 1000 mKII = 1000000 uKII
				bankMetadata := bankTypes.Metadata{
					Description: display,
					DenomUnits: []*bankTypes.DenomUnit{
						{Denom: "u" + display, Exponent: uint32(0), Aliases: []string{"micro" + display}},
						{Denom: "m" + display, Exponent: uint32(3), Aliases: []string{"mili" + display}},
						{Denom: display, Exponent: uint32(6), Aliases: []string{}},
					},
					Base:    base,
					Display: display,
					Name:    nameSymbol,
					Symbol:  nameSymbol,
				}

				k.bankKeeper.SetDenomMetaData(ctx, bankMetadata)
			}
		}

	}
}
