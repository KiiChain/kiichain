package types

import (
	"fmt"
	"math"
	"sort"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Claim represents a claim action ticket from the validator, it will store the information
// about voting and who will receive the reward (or slashing)
type Claim struct {
	Power     int64
	Weight    int64
	WinCount  int64
	DidVote   bool
	Recipient sdk.ValAddress
}

// NewClaim creates a new instance of Claim with the input parameters
func NewClaim(power, weight, winCount int64, didVote bool, recipient sdk.ValAddress) Claim {
	return Claim{
		Power:     power,
		Weight:    power,
		WinCount:  winCount,
		DidVote:   didVote,
		Recipient: recipient,
	}
}

// VoteForTally is the struct that represents the validator's vote
type VoteForTally struct {
	Denom        string         // What denom validator is voting
	ExchangeRate sdk.Dec        // Rate for that specific denom
	Voter        sdk.ValAddress // Voter (the validator)
	Power        int64          // Validator's power
}

// NewVoteForTally creates a new instance of VoteForTally with the input parameters
func NewVoteForTally(rate sdk.Dec, denom string, voter sdk.ValAddress, power int64) VoteForTally {
	return VoteForTally{
		Denom:        denom,
		ExchangeRate: rate,
		Voter:        voter,
		Power:        power,
	}
}

// ExchangeRateBallot is a wrapper that means an arrya of VoteForTally
type ExchangeRateBallot []VoteForTally

// ToMap returns organized exchange rate map by validator
func (ex ExchangeRateBallot) ToMap() map[string]sdk.Dec {
	exchangeRateMap := make(map[string]sdk.Dec) // recipient to return

	for _, vote := range ex { // Iterate the ExchangeRateBallot (remember is []VoteForTally)
		if vote.ExchangeRate.IsPositive() {
			exchangeRateMap[string(vote.Voter)] = vote.ExchangeRate // extract voter and exchange rate
		}
	}

	return exchangeRateMap
}

// Len implements sort.Interface, necessary to use sort.Sort()
// Len returns the length of the []VoteForTally
func (ex ExchangeRateBallot) Len() int {
	return len(ex)
}

// Less implements sort.Interface, necessary to use sort.Sort()
// Less compare the exchangeRate whether i position is less than j position
func (ex ExchangeRateBallot) Less(i, j int) bool {
	return ex[i].ExchangeRate.LT(ex[j].ExchangeRate)
}

// Swap implements sort.Interface, necessary to use sort.Sort()
// Swap switch the exchangeRate elements from position i to j
func (ex ExchangeRateBallot) Swap(i, j int) {
	ex[i], ex[j] = ex[j], ex[i]
}

// Power returns the total amount of voting power in the ballot
func (ex ExchangeRateBallot) Power() int64 {
	totalPower := int64(0)

	for _, vote := range ex {
		totalPower += vote.Power
	}

	return totalPower
}

// WeightedMedianWithAssertion returns the median weighted by the power
// of the exchange rate vote. Must be sorted because I selected
// the exchange rate that the accomulated power is equal or major to 50% of total power
func (ex ExchangeRateBallot) WeightedMedianWithAssertion() sdk.Dec {
	// Validate if the exchange rate is sorted
	if !sort.IsSorted(ex) {
		panic("ballot must be sorted")
	}

	totalPower := ex.Power() // get the ballot power

	if ex.Len() > 0 {
		pivot := int64(0)

		// Iterate the votes
		for _, vote := range ex {
			pivot += vote.Power // accomulate the vote's power
			if pivot >= (totalPower / 2) {
				return vote.ExchangeRate
			}
		}
	}

	return sdk.ZeroDec() // Return zero if the ballot doesn't have exchange rates
}

func (ex ExchangeRateBallot) ToCrossRateWithSort(bases map[string]sdk.Dec) ExchangeRateBallot {
	ballot := ex.ToCrossRate(bases)
	sort.Sort(ballot)
	return ballot
}

// ToCrossRate return cross_rate(base/exchange_rate) ballot
func (ex ExchangeRateBallot) ToCrossRate(bases map[string]sdk.Dec) ExchangeRateBallot {
	// Iterate over the exchange rates
	crossRateBallot := make(ExchangeRateBallot, 0, len(ex))

	for i := range ex {
		vote := ex[i]

		exchangeRateBase, ok := bases[string(vote.Voter)]

		if ok && vote.ExchangeRate.IsPositive() {
			// Quo will panic on overflow, so we wrap it in a defer/recover
			func() {
				defer func() {
					if r := recover(); r != nil {
						// if overflow, set exchange rate to 0 and power to 0
						vote.ExchangeRate = sdk.ZeroDec()
						vote.Power = 0
					}
				}()
				vote.ExchangeRate = exchangeRateBase.Quo(vote.ExchangeRate) // get cross = base / ex
			}()
		} else {
			// If we can't get exchange rate, convert the vote as abstain vote
			vote.ExchangeRate = sdk.ZeroDec()
			vote.Power = 0
		}

		crossRateBallot = append(crossRateBallot, vote)
	}

	return crossRateBallot
}

// StandardDeviation calculates the standard deviation by the power
func (ex ExchangeRateBallot) StandardDeviation(median sdk.Dec) (standardDeviation sdk.Dec) {
	// Validate the Ballot has votes
	if len(ex) == 0 {
		return sdk.ZeroDec()
	}

	// Panic handler (returns zero)
	defer func() {
		e := recover()
		if e != nil {
			standardDeviation = sdk.ZeroDec()
		}
	}()

	sum := sdk.ZeroDec()
	for _, votes := range ex {
		deviation := votes.ExchangeRate.Sub(median) // calculate the ex - median
		sum = sum.Add(deviation.Mul(deviation))     // Calculate sum += (ex - median)^2
	}

	variance := sum.QuoInt64(int64(len(ex))) // Divide the result by the number of ex

	floatNum, _ := strconv.ParseFloat(variance.String(), 64)
	floatNum = math.Sqrt(floatNum)
	standardDeviation, _ = sdk.NewDecFromStr(fmt.Sprintf("%f", floatNum))

	return

}
