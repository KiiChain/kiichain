package types

import (
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
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
	denom := "ukii"
	rate := sdk.NewDec(1)
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
	denom := "ukii"
	voter1 := sdk.ValAddress([]byte("validator1"))
	voter2 := sdk.ValAddress([]byte("validator2"))
	voter3 := sdk.ValAddress([]byte("validator3"))
	voter4 := sdk.ValAddress([]byte("validator4"))
	power := int64(10)

	ballot := ExchangeRateBallot{
		NewVoteForTally(sdk.NewDec(1), denom, voter1, power),
		NewVoteForTally(sdk.NewDec(2), denom, voter2, power),
		NewVoteForTally(sdk.NewDec(3), denom, voter3, power),
		NewVoteForTally(sdk.NewDec(4), denom, voter4, power),
	}

	reference := map[string]sdk.Dec{
		"validator1": sdk.NewDec(1),
		"validator2": sdk.NewDec(2),
		"validator3": sdk.NewDec(3),
		"validator4": sdk.NewDec(4),
	}

	require.Equal(t, reference, ballot.ToMap())
}

func TestSortInterfaceExchangeRateBallot(t *testing.T) {
	// Create exchangeRate ballot
	denom := "ukii"
	voter1 := sdk.ValAddress([]byte("validator1"))
	voter2 := sdk.ValAddress([]byte("validator2"))
	voter3 := sdk.ValAddress([]byte("validator3"))
	voter4 := sdk.ValAddress([]byte("validator4"))
	power := int64(10)

	ballot := ExchangeRateBallot{
		NewVoteForTally(sdk.NewDec(3), denom, voter3, power),
		NewVoteForTally(sdk.NewDec(4), denom, voter4, power),
		NewVoteForTally(sdk.NewDec(1), denom, voter1, power),
		NewVoteForTally(sdk.NewDec(2), denom, voter2, power),
	}

	sortedBallot := ExchangeRateBallot{
		NewVoteForTally(sdk.NewDec(1), denom, voter1, power),
		NewVoteForTally(sdk.NewDec(2), denom, voter2, power),
		NewVoteForTally(sdk.NewDec(3), denom, voter3, power),
		NewVoteForTally(sdk.NewDec(4), denom, voter4, power),
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
	require.Equal(t, sdk.NewDec(4), ballot[0].ExchangeRate)
	require.Equal(t, sdk.NewDec(3), ballot[1].ExchangeRate)

	// Validate sort process (sort by exchangeRate value)
	sort.Sort(ballot)
	require.Equal(t, sortedBallot, ballot)
	require.True(t, sort.IsSorted(ballot)) // Validate response
}

func TestWeightedMedianWithAssertion(t *testing.T) {
	// Create exchangeRate ballot
	denom := "ukii"
	voter1 := sdk.ValAddress([]byte("validator1"))
	voter2 := sdk.ValAddress([]byte("validator2"))
	voter3 := sdk.ValAddress([]byte("validator3"))
	voter4 := sdk.ValAddress([]byte("validator4"))

	ballot := ExchangeRateBallot{
		NewVoteForTally(sdk.NewDec(3), denom, voter3, 30),
		NewVoteForTally(sdk.NewDec(4), denom, voter4, 40),
		NewVoteForTally(sdk.NewDec(1), denom, voter1, 10),
		NewVoteForTally(sdk.NewDec(2), denom, voter2, 20),
	}

	// This must returns panic because ballot is not sorted
	require.Panics(t, func() { ballot.WeightedMedianWithAssertion() })

	sortedBallot := ExchangeRateBallot{
		NewVoteForTally(sdk.NewDec(1), denom, voter1, 10),
		NewVoteForTally(sdk.NewDec(2), denom, voter2, 20),
		NewVoteForTally(sdk.NewDec(3), denom, voter3, 30), // 3 is the median rate
		NewVoteForTally(sdk.NewDec(4), denom, voter4, 40),
	}

	require.Equal(t, sdk.NewDec(3), sortedBallot.WeightedMedianWithAssertion())

	// This must returns zero because there is no votes
	emptyBallot := ExchangeRateBallot{}
	require.Equal(t, sdk.ZeroDec(), emptyBallot.WeightedMedianWithAssertion())
}

func TestStandardDeviation(t *testing.T) {
	// Create exchangeRate ballot
	denom := "ukii"
	voter1 := sdk.ValAddress([]byte("validator1"))
	voter2 := sdk.ValAddress([]byte("validator2"))
	voter3 := sdk.ValAddress([]byte("validator3"))
	voter4 := sdk.ValAddress([]byte("validator4"))

	ballot := ExchangeRateBallot{
		NewVoteForTally(sdk.NewDec(1), denom, voter1, 10),
		NewVoteForTally(sdk.NewDec(2), denom, voter2, 20),
		NewVoteForTally(sdk.NewDec(3), denom, voter3, 30), // 3 is the median rate
		NewVoteForTally(sdk.NewDec(4), denom, voter4, 40),
	}

	// Must return zero (no votes)
	emptyBallot := ExchangeRateBallot{}
	require.Equal(t, sdk.ZeroDec(), emptyBallot.StandardDeviation(sdk.ZeroDec()))

	// Calculate the standard deviation
	median := ballot.WeightedMedianWithAssertion()
	deviation := ballot.StandardDeviation(median)
	require.Equal(t, sdk.NewDecWithPrec(1224745, 6), deviation)
}

func TestToCrossRate(t *testing.T) {
	// Create exchangeRate ballot (reference and other)
	denom := "ukii"
	denomRefernce := "uusdc"
	voter1 := sdk.ValAddress([]byte("validator1"))
	voter2 := sdk.ValAddress([]byte("validator2"))
	voter3 := sdk.ValAddress([]byte("validator3"))
	voter4 := sdk.ValAddress([]byte("validator4"))

	referenceBallot := ExchangeRateBallot{
		NewVoteForTally(sdk.NewDec(6), denom, voter3, 30),
		NewVoteForTally(sdk.NewDec(2), denom, voter1, 10),
		NewVoteForTally(sdk.NewDec(8), denom, voter4, 40),
		NewVoteForTally(sdk.NewDec(4), denom, voter2, 20),
	}

	ballot := ExchangeRateBallot{
		NewVoteForTally(sdk.NewDec(2), denomRefernce, voter3, 30),
		NewVoteForTally(sdk.NewDec(2), denomRefernce, voter1, 10),
		NewVoteForTally(sdk.NewDec(2), denomRefernce, voter4, 40),
		NewVoteForTally(sdk.NewDec(2), denomRefernce, voter2, 20),
	}

	expectedCrossRate := ExchangeRateBallot{
		NewVoteForTally(sdk.NewDec(3), denomRefernce, voter3, 30),
		NewVoteForTally(sdk.NewDec(1), denomRefernce, voter1, 10),
		NewVoteForTally(sdk.NewDec(4), denomRefernce, voter4, 40),
		NewVoteForTally(sdk.NewDec(2), denomRefernce, voter2, 20),
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
	denom := "ukii"
	denomRefernce := "uusdc"
	voter1 := sdk.ValAddress([]byte("validator1"))
	voter2 := sdk.ValAddress([]byte("validator2"))
	voter3 := sdk.ValAddress([]byte("validator3"))
	voter4 := sdk.ValAddress([]byte("validator4"))
	voter5 := sdk.ValAddress([]byte("validator5"))
	voter6 := sdk.ValAddress([]byte("validator6"))

	referenceBallot := ExchangeRateBallot{
		NewVoteForTally(sdk.NewDec(6), denom, voter3, 30),
		NewVoteForTally(sdk.NewDec(2), denom, voter1, 10),
		NewVoteForTally(sdk.NewDec(8), denom, voter4, 40),
		NewVoteForTally(sdk.NewDec(4), denom, voter2, 20),
	}

	ballot := ExchangeRateBallot{
		NewVoteForTally(sdk.NewDec(2), denomRefernce, voter5, 10),
		NewVoteForTally(sdk.NewDec(2), denomRefernce, voter6, 30),
		NewVoteForTally(sdk.NewDec(2), denomRefernce, voter4, 40),
		NewVoteForTally(sdk.NewDec(2), denomRefernce, voter2, 20),
	}

	// must returns zero because val6 is not on referenceBallot
	// must returns zero because val5 is not on referenceBallot
	expectedCrossRate := ExchangeRateBallot{
		NewVoteForTally(sdk.ZeroDec(), denomRefernce, voter5, 0),
		NewVoteForTally(sdk.ZeroDec(), denomRefernce, voter6, 0),
		NewVoteForTally(sdk.NewDec(4), denomRefernce, voter4, 40),
		NewVoteForTally(sdk.NewDec(2), denomRefernce, voter2, 20),
	}

	crossRate := ballot.ToCrossRate(referenceBallot.ToMap())
	require.Equal(t, expectedCrossRate, crossRate)
}

func TestToCrossRateWithSort(t *testing.T) {
	// Create exchangeRate ballot (reference and other)
	denom := "ukii"
	denomRefernce := "uusdc"
	voter1 := sdk.ValAddress([]byte("validator1"))
	voter2 := sdk.ValAddress([]byte("validator2"))
	voter3 := sdk.ValAddress([]byte("validator3"))
	voter4 := sdk.ValAddress([]byte("validator4"))

	referenceBallot := ExchangeRateBallot{
		NewVoteForTally(sdk.NewDec(6), denom, voter3, 30),
		NewVoteForTally(sdk.NewDec(2), denom, voter1, 10),
		NewVoteForTally(sdk.NewDec(8), denom, voter4, 40),
		NewVoteForTally(sdk.NewDec(4), denom, voter2, 20),
	}

	ballot := ExchangeRateBallot{
		NewVoteForTally(sdk.NewDec(2), denomRefernce, voter3, 30),
		NewVoteForTally(sdk.NewDec(2), denomRefernce, voter1, 10),
		NewVoteForTally(sdk.NewDec(2), denomRefernce, voter4, 40),
		NewVoteForTally(sdk.NewDec(2), denomRefernce, voter2, 20),
	}

	// expected cross rate and SORTED
	expectedCrossRate := ExchangeRateBallot{
		NewVoteForTally(sdk.NewDec(1), denomRefernce, voter1, 10),
		NewVoteForTally(sdk.NewDec(2), denomRefernce, voter2, 20),
		NewVoteForTally(sdk.NewDec(3), denomRefernce, voter3, 30),
		NewVoteForTally(sdk.NewDec(4), denomRefernce, voter4, 40),
	}

	// must calcualte the cross rate and sort the response
	crossRate := ballot.ToCrossRateWithSort(referenceBallot.ToMap())
	require.Equal(t, expectedCrossRate, crossRate)
}
