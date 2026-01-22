package types

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	sdkMath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ChainDenom = "akii"
	USDC       = "usdc"
)

func TestNewClaim(t *testing.T) {
	power := int64(10)
	weight := int64(10)
	winCount := int64(0)
	didVote := false
	recipient := sdk.ValAddress([]byte("validator1"))

	reference := Claim{
		Power:     power,
		Weight:    weight,
		WinCount:  winCount,
		DidVote:   didVote,
		Recipient: recipient,
	}

	claim := NewClaim(power, weight, winCount, didVote, recipient)

	require.Equal(t, reference, claim)
}

func TestNewVoteForTally(t *testing.T) {
	denom := ChainDenom
	rate := sdkMath.LegacyNewDec(1)
	voter := sdk.ValAddress([]byte("validator1"))
	power := int64(10)

	reference := VoteForTally{
		Denom:        denom,
		ExchangeRate: rate,
		Voter:        voter,
		Power:        power,
	}

	vote := NewVoteForTally(rate, denom, voter, power)

	require.Equal(t, reference, vote)
}

func TestToMapExchangeRateBallot(t *testing.T) {
	// Create exchangeRate ballot
	denom := ChainDenom
	voter1 := sdk.ValAddress([]byte("validator1"))
	voter2 := sdk.ValAddress([]byte("validator2"))
	voter3 := sdk.ValAddress([]byte("validator3"))
	voter4 := sdk.ValAddress([]byte("validator4"))
	power := int64(10)

	ballot := ExchangeRateBallot{
		NewVoteForTally(sdkMath.LegacyNewDec(1), denom, voter1, power),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denom, voter2, power),
		NewVoteForTally(sdkMath.LegacyNewDec(3), denom, voter3, power),
		NewVoteForTally(sdkMath.LegacyNewDec(4), denom, voter4, power),
	}

	reference := map[string]sdkMath.LegacyDec{
		"validator1": sdkMath.LegacyNewDec(1),
		"validator2": sdkMath.LegacyNewDec(2),
		"validator3": sdkMath.LegacyNewDec(3),
		"validator4": sdkMath.LegacyNewDec(4),
	}

	require.Equal(t, reference, ballot.ToMap())
}

func TestSortInterfaceExchangeRateBallot(t *testing.T) {
	// Create exchangeRate ballot
	denom := ChainDenom
	voter1 := sdk.ValAddress([]byte("validator1"))
	voter2 := sdk.ValAddress([]byte("validator2"))
	voter3 := sdk.ValAddress([]byte("validator3"))
	voter4 := sdk.ValAddress([]byte("validator4"))
	power := int64(10)

	ballot := ExchangeRateBallot{
		NewVoteForTally(sdkMath.LegacyNewDec(3), denom, voter3, power),
		NewVoteForTally(sdkMath.LegacyNewDec(4), denom, voter4, power),
		NewVoteForTally(sdkMath.LegacyNewDec(1), denom, voter1, power),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denom, voter2, power),
	}

	sortedBallot := ExchangeRateBallot{
		NewVoteForTally(sdkMath.LegacyNewDec(1), denom, voter1, power),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denom, voter2, power),
		NewVoteForTally(sdkMath.LegacyNewDec(3), denom, voter3, power),
		NewVoteForTally(sdkMath.LegacyNewDec(4), denom, voter4, power),
	}

	// Validate ExchangeRateBallot implements sort.Interface
	var _ sort.Interface = ballot

	// Validate the len method
	require.Equal(t, len(ballot), ballot.Len())

	// Validate the less method
	require.True(t, ballot.Less(2, 1))
	require.False(t, ballot.Less(1, 3))
	require.True(t, ballot.Less(3, 1))

	// Validate the swap method
	ballot.Swap(0, 1)
	require.Equal(t, sdkMath.LegacyNewDec(4), ballot[0].ExchangeRate)
	require.Equal(t, sdkMath.LegacyNewDec(3), ballot[1].ExchangeRate)

	// Validate sort process (sort by exchangeRate value)
	sort.Sort(ballot)
	require.Equal(t, sortedBallot, ballot)
	require.True(t, sort.IsSorted(ballot)) // Validate response
}

func TestWeightedMedianWithAssertion(t *testing.T) {
	// Create exchangeRate ballot
	denom := ChainDenom
	voter1 := sdk.ValAddress([]byte("validator1"))
	voter2 := sdk.ValAddress([]byte("validator2"))
	voter3 := sdk.ValAddress([]byte("validator3"))
	voter4 := sdk.ValAddress([]byte("validator4"))

	ballot := ExchangeRateBallot{
		NewVoteForTally(sdkMath.LegacyNewDec(3), denom, voter3, 30),
		NewVoteForTally(sdkMath.LegacyNewDec(4), denom, voter4, 40),
		NewVoteForTally(sdkMath.LegacyNewDec(1), denom, voter1, 10),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denom, voter2, 20),
	}

	// This must returns panic because ballot is not sorted
	require.Panics(t, func() { ballot.WeightedMedianWithAssertion() })

	sortedBallot := ExchangeRateBallot{
		NewVoteForTally(sdkMath.LegacyNewDec(1), denom, voter1, 10),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denom, voter2, 20),
		NewVoteForTally(sdkMath.LegacyNewDec(3), denom, voter3, 30), // 3 is the median rate
		NewVoteForTally(sdkMath.LegacyNewDec(4), denom, voter4, 40),
	}

	require.Equal(t, sdkMath.LegacyNewDec(3), sortedBallot.WeightedMedianWithAssertion())

	// This must returns zero because there is no votes
	emptyBallot := ExchangeRateBallot{}
	require.Equal(t, sdkMath.LegacyZeroDec(), emptyBallot.WeightedMedianWithAssertion())
}

func TestStandardDeviation(t *testing.T) {
	// Create exchangeRate ballot
	denom := ChainDenom
	voter1 := sdk.ValAddress([]byte("validator1"))
	voter2 := sdk.ValAddress([]byte("validator2"))
	voter3 := sdk.ValAddress([]byte("validator3"))
	voter4 := sdk.ValAddress([]byte("validator4"))

	ballot := ExchangeRateBallot{
		NewVoteForTally(sdkMath.LegacyNewDec(1), denom, voter1, 10),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denom, voter2, 20),
		NewVoteForTally(sdkMath.LegacyNewDec(3), denom, voter3, 30), // 3 is the median rate
		NewVoteForTally(sdkMath.LegacyNewDec(4), denom, voter4, 40),
	}

	// Must return zero (no votes)
	emptyBallot := ExchangeRateBallot{}
	require.Equal(t, sdkMath.LegacyZeroDec(), emptyBallot.StandardDeviation(sdkMath.LegacyZeroDec()))

	// Calculate the standard deviation
	median := ballot.WeightedMedianWithAssertion()
	deviation := ballot.StandardDeviation(median)

	expected := sdkMath.LegacyNewDecWithPrec(1224745, 6)
	// 2e-7 tolerance
	delta := sdkMath.LegacyNewDecWithPrec(2, 7)

	require.True(t, deviation.Sub(expected).Abs().LTE(delta))
}

func TestToCrossRate(t *testing.T) {
	// Create exchangeRate ballot (reference and other)
	denom := ChainDenom
	denomRefernce := USDC
	voter1 := sdk.ValAddress([]byte("validator1"))
	voter2 := sdk.ValAddress([]byte("validator2"))
	voter3 := sdk.ValAddress([]byte("validator3"))
	voter4 := sdk.ValAddress([]byte("validator4"))

	referenceBallot := ExchangeRateBallot{
		NewVoteForTally(sdkMath.LegacyNewDec(6), denom, voter3, 30),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denom, voter1, 10),
		NewVoteForTally(sdkMath.LegacyNewDec(8), denom, voter4, 40),
		NewVoteForTally(sdkMath.LegacyNewDec(4), denom, voter2, 20),
	}

	ballot := ExchangeRateBallot{
		NewVoteForTally(sdkMath.LegacyNewDec(2), denomRefernce, voter3, 30),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denomRefernce, voter1, 10),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denomRefernce, voter4, 40),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denomRefernce, voter2, 20),
	}

	expectedCrossRate := ExchangeRateBallot{
		NewVoteForTally(sdkMath.LegacyNewDec(3), denomRefernce, voter3, 30),
		NewVoteForTally(sdkMath.LegacyNewDec(1), denomRefernce, voter1, 10),
		NewVoteForTally(sdkMath.LegacyNewDec(4), denomRefernce, voter4, 40),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denomRefernce, voter2, 20),
	}

	// Calculate the cross rate as:
	//						reference
	//		cross rate = ---------------
	//				      exchange rate

	// must calculates the cross rate base on the reference ballot
	crossRate := ballot.ToCrossRate(referenceBallot.ToMap())
	require.Equal(t, expectedCrossRate, crossRate)
}

func TestToCrossRateNotFound(t *testing.T) {
	// Create exchangeRate ballot (reference and other)
	denom := ChainDenom
	denomRefernce := USDC
	voter1 := sdk.ValAddress([]byte("validator1"))
	voter2 := sdk.ValAddress([]byte("validator2"))
	voter3 := sdk.ValAddress([]byte("validator3"))
	voter4 := sdk.ValAddress([]byte("validator4"))
	voter5 := sdk.ValAddress([]byte("validator5"))
	voter6 := sdk.ValAddress([]byte("validator6"))

	referenceBallot := ExchangeRateBallot{
		NewVoteForTally(sdkMath.LegacyNewDec(6), denom, voter3, 30),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denom, voter1, 10),
		NewVoteForTally(sdkMath.LegacyNewDec(8), denom, voter4, 40),
		NewVoteForTally(sdkMath.LegacyNewDec(4), denom, voter2, 20),
	}

	ballot := ExchangeRateBallot{
		NewVoteForTally(sdkMath.LegacyNewDec(2), denomRefernce, voter5, 10),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denomRefernce, voter6, 30),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denomRefernce, voter4, 40),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denomRefernce, voter2, 20),
	}

	// must returns zero because val6 is not on referenceBallot
	// must returns zero because val5 is not on referenceBallot
	expectedCrossRate := ExchangeRateBallot{
		NewVoteForTally(sdkMath.LegacyZeroDec(), denomRefernce, voter5, 0),
		NewVoteForTally(sdkMath.LegacyZeroDec(), denomRefernce, voter6, 0),
		NewVoteForTally(sdkMath.LegacyNewDec(4), denomRefernce, voter4, 40),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denomRefernce, voter2, 20),
	}

	crossRate := ballot.ToCrossRate(referenceBallot.ToMap())
	require.Equal(t, expectedCrossRate, crossRate)
}

func TestToCrossRateWithSort(t *testing.T) {
	// Create exchangeRate ballot (reference and other)
	denom := ChainDenom
	denomRefernce := USDC
	voter1 := sdk.ValAddress([]byte("validator1"))
	voter2 := sdk.ValAddress([]byte("validator2"))
	voter3 := sdk.ValAddress([]byte("validator3"))
	voter4 := sdk.ValAddress([]byte("validator4"))

	referenceBallot := ExchangeRateBallot{
		NewVoteForTally(sdkMath.LegacyNewDec(6), denom, voter3, 30),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denom, voter1, 10),
		NewVoteForTally(sdkMath.LegacyNewDec(8), denom, voter4, 40),
		NewVoteForTally(sdkMath.LegacyNewDec(4), denom, voter2, 20),
	}

	ballot := ExchangeRateBallot{
		NewVoteForTally(sdkMath.LegacyNewDec(2), denomRefernce, voter3, 30),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denomRefernce, voter1, 10),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denomRefernce, voter4, 40),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denomRefernce, voter2, 20),
	}

	// expected cross rate and SORTED
	expectedCrossRate := ExchangeRateBallot{
		NewVoteForTally(sdkMath.LegacyNewDec(1), denomRefernce, voter1, 10),
		NewVoteForTally(sdkMath.LegacyNewDec(2), denomRefernce, voter2, 20),
		NewVoteForTally(sdkMath.LegacyNewDec(3), denomRefernce, voter3, 30),
		NewVoteForTally(sdkMath.LegacyNewDec(4), denomRefernce, voter4, 40),
	}

	// must calculate the cross rate and sort the response
	crossRate := ballot.ToCrossRateWithSort(referenceBallot.ToMap())
	require.Equal(t, expectedCrossRate, crossRate)
}

// TestWeightedMedianOddTotalPower tests that weighted median requires strict majority (>50%)
// This is a regression test for issue #236 where >= was incorrectly used instead of >
func TestWeightedMedianOddTotalPower(t *testing.T) {
	// Create ballot with odd total power (101)
	// Validator A: 50 power (49.5%), votes 1.00
	// Validator B: 51 power (50.5%), votes 2.00
	ballot := ExchangeRateBallot{
		NewVoteForTally(sdkMath.LegacyNewDec(1), "test", sdk.ValAddress([]byte("val1")), 50),
		NewVoteForTally(sdkMath.LegacyNewDec(2), "test", sdk.ValAddress([]byte("val2")), 51),
	}

	sort.Sort(ballot)
	require.Equal(t, int64(101), ballot.Power())

	// Median should be 2 (from validator with 51 power = 50.5%)
	// NOT 1 (from validator with 50 power = 49.5%)
	median := ballot.WeightedMedianWithAssertion()
	require.Equal(t, sdkMath.LegacyNewDec(2), median, "weighted median should require strict majority")
}

// TestWeightedMedianExactHalf tests edge case where validator has exactly half power
func TestWeightedMedianExactHalf(t *testing.T) {
	// Total power: 100 (even)
	// Validator A: 50 power, votes 1.00
	// Validator B: 50 power, votes 2.00
	ballot := ExchangeRateBallot{
		NewVoteForTally(sdkMath.LegacyNewDec(1), "test", sdk.ValAddress([]byte("val1")), 50),
		NewVoteForTally(sdkMath.LegacyNewDec(2), "test", sdk.ValAddress([]byte("val2")), 50),
	}

	sort.Sort(ballot)
	require.Equal(t, int64(100), ballot.Power())

	// With > check: first 50 is not > 50, so continue to second
	median := ballot.WeightedMedianWithAssertion()
	require.Equal(t, sdkMath.LegacyNewDec(2), median, "with exactly 50 percent should continue to next")
}
