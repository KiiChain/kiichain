package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMsgAggregateExchangeRateVote(t *testing.T) {

	type test struct {
		voter         sdk.AccAddress
		exchangeRates string
		expectPass    bool
	}

	addrs := []sdk.AccAddress{
		sdk.AccAddress([]byte("addr1___________")),
	}

	invalidExchangeRates := "a,b"
	exchangeRates := "12.00atom,1234.12eth"
	abstainExchangeRates := "0.0atom,123.12eth"
	overFlowExchangeRates := "1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000.0atom,123.13eth"

	tests := []test{
		{addrs[0], exchangeRates, true},
		{addrs[0], invalidExchangeRates, false},
		{addrs[0], abstainExchangeRates, true},
		{addrs[0], overFlowExchangeRates, false},
		{sdk.AccAddress{}, exchangeRates, false},
	}

	// validation
	for i, test := range tests {
		msg := NewMsgAggregateExchangeRateVote(test.exchangeRates, test.voter, sdk.ValAddress(test.voter))
		if test.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", i)

			// last test must panic because there is no signer or feeder address
			if i == len(tests)-1 {
				require.Panics(t, func() { msg.GetSigners() })
			}
			continue
		}

		require.NotNil(t, msg.ValidateBasic(), "test: %v", i)
	}
}

func TestMsgDelegateFeedConsent(t *testing.T) {
	type test struct {
		delegator  sdk.ValAddress
		delegated  sdk.AccAddress
		expectPass bool
	}

	addrs := []sdk.AccAddress{
		sdk.AccAddress([]byte("addr1___________")),
		sdk.AccAddress([]byte("addr2___________")),
	}

	tests := []test{
		{sdk.ValAddress(addrs[0]), addrs[1], true},
		{sdk.ValAddress{}, addrs[1], false},
		{sdk.ValAddress(addrs[0]), sdk.AccAddress{}, false},
		{sdk.ValAddress(addrs[0]), addrs[0], true},
	}

	// validation
	for i, test := range tests {
		msg := NewMsgDelegateFeedConsent(test.delegator, test.delegated)
		if test.expectPass {
			require.Nil(t, msg.ValidateBasic(), "test: %v", i)
			continue
		}

		require.NotNil(t, msg.ValidateBasic(), "test: %v", i)

		// must panic because there is not delegator address
		if i == 1 {
			require.Panics(t, func() { msg.GetSigners() })
		}
	}

}
