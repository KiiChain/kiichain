package e2e

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) testBankTokenTransfer() {
	s.Run("send_tokens_between_accounts", func() {
		var (
			err           error
			valIdx        = 0
			c             = s.chainA
			chainEndpoint = fmt.Sprintf("http://%s", s.valResources[c.id][valIdx].GetHostPort("1317/tcp"))
		)

		// define one sender and two recipient accounts
		alice, _ := c.genesisAccounts[1].keyInfo.GetAddress()
		bob, _ := c.genesisAccounts[2].keyInfo.GetAddress()
		charlie, _ := c.genesisAccounts[3].keyInfo.GetAddress()

		var beforeAliceAKiiBalance,
			beforeBobAkiiBalance,
			beforeCharlieAKiiBalance,
			afterAliceAKiiBalance,
			afterBobUAKiiBalance,
			afterCharlieAKiiBalance sdk.Coin

		// get balances of sender and recipient accounts
		s.Require().Eventually(
			func() bool {
				beforeAliceAKiiBalance, err = getSpecificBalance(chainEndpoint, alice.String(), akiiDenom)
				s.Require().NoError(err)

				beforeBobAkiiBalance, err = getSpecificBalance(chainEndpoint, bob.String(), akiiDenom)
				s.Require().NoError(err)

				beforeCharlieAKiiBalance, err = getSpecificBalance(chainEndpoint, charlie.String(), akiiDenom)
				s.Require().NoError(err)

				return beforeAliceAKiiBalance.IsValid() && beforeBobAkiiBalance.IsValid() && beforeCharlieAKiiBalance.IsValid()
			},
			10*time.Second,
			5*time.Second,
		)

		// alice sends tokens to bob
		s.execBankSend(s.chainA, valIdx, alice.String(), bob.String(), tokenAmount.String(), standardFees.String(), false)

		// check that the transfer was successful
		s.Require().Eventually(
			func() bool {
				afterAliceAKiiBalance, err = getSpecificBalance(chainEndpoint, alice.String(), akiiDenom)
				s.Require().NoError(err)

				afterBobUAKiiBalance, err = getSpecificBalance(chainEndpoint, bob.String(), akiiDenom)
				s.Require().NoError(err)

				decremented := beforeAliceAKiiBalance.Sub(tokenAmount).Sub(standardFees).IsEqual(afterAliceAKiiBalance)
				incremented := beforeBobAkiiBalance.Add(tokenAmount).IsEqual(afterBobUAKiiBalance)

				return decremented && incremented
			},
			10*time.Second,
			5*time.Second,
		)

		// save the updated account balances of alice and bob
		beforeAliceAKiiBalance, beforeBobAkiiBalance = afterAliceAKiiBalance, afterBobUAKiiBalance

		// alice sends tokens to bob and charlie, at once
		s.execBankMultiSend(s.chainA, valIdx, alice.String(), []string{bob.String(), charlie.String()}, tokenAmount.String(), standardFees.String(), false)

		s.Require().Eventually(
			func() bool {
				afterAliceAKiiBalance, err = getSpecificBalance(chainEndpoint, alice.String(), akiiDenom)
				s.Require().NoError(err)

				afterBobUAKiiBalance, err = getSpecificBalance(chainEndpoint, bob.String(), akiiDenom)
				s.Require().NoError(err)

				afterCharlieAKiiBalance, err = getSpecificBalance(chainEndpoint, charlie.String(), akiiDenom)
				s.Require().NoError(err)

				decremented := beforeAliceAKiiBalance.Sub(tokenAmount).Sub(tokenAmount).Sub(standardFees).IsEqual(afterAliceAKiiBalance)
				incremented := beforeBobAkiiBalance.Add(tokenAmount).IsEqual(afterBobUAKiiBalance) &&
					beforeCharlieAKiiBalance.Add(tokenAmount).IsEqual(afterCharlieAKiiBalance)

				return decremented && incremented
			},
			10*time.Second,
			5*time.Second,
		)
	})
}
