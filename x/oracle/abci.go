package oracle

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v1/x/oracle/keeper"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
	"github.com/kiichain/kiichain/v1/x/oracle/utils"
)

// MidBlocker get the votes from the validators, calculate the exchange rate using
// weighted median logic when the vote period is almost finished
func EndBlocker(ctx sdk.Context, k keeper.Keeper) error {
	// Get the params
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// Check if the current block is the last one to finish the voting period
	if utils.IsPeriodLastBlock(ctx, params.VotePeriod) {
		validatorClaimMap := make(map[string]types.Claim) // here I will store the claim per validator

		iterator, err := k.StakingKeeper.ValidatorsPowerStoreIterator(ctx)
		if err != nil {
			return err
		}

		defer iterator.Close()

		powerReduction := k.StakingKeeper.PowerReduction(ctx) // Get the power reduction factor

		// Iterate over validators and register only the bonded ones
		for ; iterator.Valid(); iterator.Next() {
			valAddr := sdk.ValAddress(iterator.Value()) // Get validator address
			validator, err := k.StakingKeeper.Validator(ctx, valAddr)
			if err != nil {
				return err
			}

			// add bonded validators
			if validator.IsBonded() {
				valPower := validator.GetConsensusPower(powerReduction) // Get validator's power
				operator := validator.GetOperator()                     // Get address to receive coins

				// Parse the operator address
				operatorAddr, _ := sdk.ValAddressFromBech32(operator)

				claim := types.NewClaim(valPower, 0, 0, false, operatorAddr) // Create claim object
				validatorClaimMap[operator] = claim                          // Assign the validator on the list to receive
			}
		}

		// Get the voting targets from the KVStore
		voteTargets := make(map[string]types.Denom)
		err = k.VoteTarget.Walk(ctx, nil, func(denom string, denomInfo types.Denom) (bool, error) {
			voteTargets[denom] = denomInfo
			return false, nil
		})
		if err != nil {
			return err
		}

		// Create a reference denom (RD) based on the voting power
		voteMap, err := k.OrganizeBallotByDenom(ctx, validatorClaimMap) // Create a map (denom sorted) with the votes by denom
		if err != nil {
			return err
		}
		referenceDenom, belowThresholdVoteMap := pickReferenceDenom(ctx, k, voteTargets, voteMap)

		if referenceDenom != "" {
			ballotRD := voteMap[referenceDenom] // get the ballot of the RD
			votingMapRD := ballotRD.ToMap()     // Conver the ballot into a map by voting tally

			// calculate the weighted median of the reference denom ballot
			exchangeRateRD := ballotRD.WeightedMedianWithAssertion()

			// Get the denoms from the ballot
			denoms := make([]string, 0, len(voteMap))
			for denom := range voteMap {
				denoms = append(denoms, denom)
			}
			sort.Strings(denoms)

			// Iterate the denoms on the voting map to calculate the final exchange rate
			for _, denom := range denoms {
				votingTally := voteMap[denom] // get the voting tally per denom

				// Convert the voting tally to cross exchange rate
				if denom != referenceDenom {
					votingTally = votingTally.ToCrossRateWithSort(votingMapRD)
				}

				// Get weighted median of cross exchange rates
				exchangeRate := Tally(ctx, votingTally, params.RewardBand, validatorClaimMap)

				// Validate invalid exchangeRate
				if exchangeRate.IsZero() {
					continue // skip this denom
				}

				// transform into the original form base/quote
				if denom != referenceDenom {
					exchangeRate = exchangeRateRD.Quo(exchangeRate)
				}

				// set the exchange rate with event
				err = k.SetBaseExchangeRateWithEvent(ctx, denom, exchangeRate)
				if err != nil {
					return err
				}
			}
		}

		// Extract the denoms stored on belowThresholdVote map
		belowThresholdDenoms := make([]string, 0, len(belowThresholdVoteMap))
		for denom := range belowThresholdVoteMap {
			belowThresholdDenoms = append(belowThresholdDenoms, denom)
		}
		sort.Strings(belowThresholdDenoms) // sort by denom name

		// Calculate tally for below threshold assets lists
		for _, denom := range belowThresholdDenoms {
			ballot := belowThresholdVoteMap[denom]
			Tally(ctx, ballot, params.RewardBand, validatorClaimMap)
		}

		// Validate miss voting process
		for _, claim := range validatorClaimMap {
			if int(claim.WinCount) == len(voteTargets) {
				err = k.IncrementSuccessCount(ctx, claim.Recipient)
				if err != nil {
					return err
				}
				continue
			}

			if !claim.DidVote {
				err = k.IncrementAbstainCount(ctx, claim.Recipient)
				if err != nil {
					return err
				}
				continue
			}

			err = k.IncrementMissCount(ctx, claim.Recipient)
			if err != nil {
				return err
			}
		}

		// Clear the ballot
		err = k.AggregateExchangeRateVote.Clear(ctx, nil)
		if err != nil {
			return err
		}

		// Update vote target
		err = k.ApplyWhitelist(ctx, params.Whitelist, voteTargets)
		if err != nil {
			return err
		}

		// take an snapshot for each price
		priceSnapshotItems := []types.PriceSnapshotItem{}
		err = k.ExchangeRate.Walk(ctx, nil, func(denom string, exchangeRate types.OracleExchangeRate) (bool, error) {
			priceSnapshotItem := types.PriceSnapshotItem{
				Denom:              denom,
				OracleExchangeRate: exchangeRate,
			}

			priceSnapshotItems = append(priceSnapshotItems, priceSnapshotItem)
			return false, nil
		})
		if err != nil {
			return err
		}

		// create and save general snapshot
		if len(priceSnapshotItems) > 0 {
			priceSnapshot := types.PriceSnapshot{
				SnapshotTimestamp:  ctx.BlockTime().Unix(),
				PriceSnapshotItems: priceSnapshotItems,
			}
			err := k.AddPriceSnapshot(ctx, priceSnapshot)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// BeginBlocker is the function that slashes the validators and resets the miss counters
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) error {
	// Get the params
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// Slash who did miss voting over threshold
	// reset miss counter of all validators at the last block of slash window
	if utils.IsPeriodLastBlock(ctx, params.SlashWindow) {
		err = k.SlashAndResetCounters(ctx) // slash validator and reset voting counter
		if err != nil {
			return err
		}
		err = k.RemoveExcessFeeds(ctx) // remove aditional rates added on the votes
		if err != nil {
			return err
		}
	}

	return nil
}
