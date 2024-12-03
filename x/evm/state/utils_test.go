package state_test

import (
	"math/big"
	"testing"

	"github.com/kiichain/kiichain3/x/evm/state"
	"github.com/stretchr/testify/require"
)

func TestGetCoinbaseAddress(t *testing.T) {
	coinbaseAddr := state.GetCoinbaseAddress(1).String()
	require.Equal(t, coinbaseAddr, "kii1v4mx6hmrda5kucnpwdjsqqqqqqqqqqqpkaxwqq")
}

func TestSplitUkiiWeiAmount(t *testing.T) {
	for _, test := range []struct {
		amt         *big.Int
		expectedSei *big.Int
		expectedWei *big.Int
	}{
		{
			amt:         big.NewInt(0),
			expectedSei: big.NewInt(0),
			expectedWei: big.NewInt(0),
		}, {
			amt:         big.NewInt(1),
			expectedSei: big.NewInt(0),
			expectedWei: big.NewInt(1),
		}, {
			amt:         big.NewInt(999_999_999_999),
			expectedSei: big.NewInt(0),
			expectedWei: big.NewInt(999_999_999_999),
		}, {
			amt:         big.NewInt(1_000_000_000_000),
			expectedSei: big.NewInt(1),
			expectedWei: big.NewInt(0),
		}, {
			amt:         big.NewInt(1_000_000_000_001),
			expectedSei: big.NewInt(1),
			expectedWei: big.NewInt(1),
		}, {
			amt:         big.NewInt(123_456_789_123_456_789),
			expectedSei: big.NewInt(123456),
			expectedWei: big.NewInt(789_123_456_789),
		},
	} {
		ukii, wei := state.SplitUkiiWeiAmount(test.amt)
		require.Equal(t, test.expectedSei, ukii.BigInt())
		require.Equal(t, test.expectedWei, wei.BigInt())
	}
}
