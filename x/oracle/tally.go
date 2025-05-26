package oracle

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kiichain/kiichain/v1/x/oracle/keeper"
	"github.com/kiichain/kiichain/v1/x/oracle/types"
)

// pickReferenceDenom selects a denom with the highest vote power as reference denom.
// If the power of 2 denominations is the same, select the reference denom
// in alphabetical order
func pickReferenceDenom(ctx sdk.Context, k keeper.Keeper, voteTargets map[string]types.Denom, voteMap map[string]types.ExchangeRateBallot) (string, map[string]types.ExchangeRateBallot) {
	highestBallotPower := int64(0)
	referenceDenom := ""
	belowThresholdVoteMap := map[string]types.ExchangeRateBallot{}

	// Get total bonded power
	powerReductionFactor := k.StakingKeeper.PowerReduction(ctx)                             // get the power reduction
	totalBondedTokens, _ := k.StakingKeeper.TotalBondedTokens(ctx)                          // total of tokens in staking
	totalBondedPower := sdk.TokensToConsensusPower(totalBondedTokens, powerReductionFactor) // Get the blockchain vote power

	// Get threshold (minimum power necessary to considerate a successful ballot)
	voteThreshold := k.VoteThreshold(ctx)                                 // Get vote threshold from params
	thresholdVotes := voteThreshold.MulInt64(totalBondedPower).RoundInt() // Threshold to allow a ballot

	// Iterate the voting map
	for denom, ballot := range voteMap {

		// If a denom is not in the vote targets or the ballot for it has failed
		// that denom is removed from votemap (for efficiency)
		_, exists := voteTargets[denom]
		if !exists {
			delete(voteMap, denom)
			continue
		}

		// Get ballot power and check if is greater than the threshold
		ballotPower, ok := ballotIsPassing(ballot, thresholdVotes)

		// if the ballot power is lower than threshold, add denom in below
		// threshold map to separe for tally evaluation
		if !ok {
			belowThresholdVoteMap[denom] = voteMap[denom]
			delete(voteTargets, denom)
			delete(voteMap, denom)
			continue
		}

		if ballotPower.Int64() > highestBallotPower || highestBallotPower == 0 {
			referenceDenom = denom
			highestBallotPower = ballotPower.Int64()
		}

		// If the power is equal, select the the denom by alphabetical order
		if ballotPower.Int64() == highestBallotPower && referenceDenom > denom {
			referenceDenom = denom
		}
	}
	return referenceDenom, belowThresholdVoteMap
}

// ballotIsPassing calculate the sum of each vote power per denom, then check
// if the ballot power is greater than the threshold
func ballotIsPassing(ballot types.ExchangeRateBallot, thresholdVotes math.Int) (math.Int, bool) {
	ballotPower := math.NewInt(ballot.Power()) // Get the validator power

	// return ballot power and if the ballot is greater than the threshold
	return ballotPower, !ballotPower.IsZero() && ballotPower.GTE(thresholdVotes)
}

// Tally calculates the median and returns it. Sets the set of voters to be rewarded, i.e. voted within
// a reasonable spread from the weighted median to the store
// CONTRACT: ex must be sorted
func Tally(_ sdk.Context, ex types.ExchangeRateBallot, rewardBand math.LegacyDec, validatorClaimMap map[string]types.Claim) (weightedMedian math.LegacyDec) {
	weightedMedian = ex.WeightedMedianWithAssertion() // Get weighted median

	// Check if result is on the reward interval
	standardDeviation := ex.StandardDeviation(weightedMedian)
	rewardSpread := weightedMedian.Mul(rewardBand.QuoInt64(2)) // this is the interval that will be added around weightedMedian

	if standardDeviation.GT(rewardSpread) { // if rewardSpread > deviation means the data is disperse
		rewardSpread = standardDeviation
	}

	// Check each vote and reward
	for _, vote := range ex {
		// Filter ballot winners
		voter := vote.Voter.String()
		claim := validatorClaimMap[voter]

		// If exchange rate is in the interval reward the validator
		if vote.ExchangeRate.GTE(weightedMedian.Sub(rewardSpread)) && // lower limit
			vote.ExchangeRate.LTE(weightedMedian.Add(rewardSpread)) { // upper limit

			claim.Weight += vote.Power
			claim.WinCount++
		}
		claim.DidVote = true
		validatorClaimMap[voter] = claim
	}

	return
}
