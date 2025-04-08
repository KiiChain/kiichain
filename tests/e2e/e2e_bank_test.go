package e2e

import (
	"fmt"
	"time"

	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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

// tests the bank send command with invalid non_critical_extension_options field
// the tx should always fail to decode the extension options since no concrete type is registered for the provided extension field
func (s *IntegrationTestSuite) failedBankSendWithNonCriticalExtensionOptions() {
	s.Run("fail_encoding_invalid_non_critical_extension_options", func() {
		c := s.chainB

		submitterAccount := c.genesisAccounts[1]
		submitterAddress, err := submitterAccount.keyInfo.GetAddress()
		s.Require().NoError(err)
		sendMsg := banktypes.NewMsgSend(submitterAddress, submitterAddress, sdk.NewCoins(sdk.NewCoin(akiiDenom, math.NewInt(100))))

		// the message does not matter, as long as it is in the interface registry
		ext := &banktypes.MsgMultiSend{}

		extAny, err := codectypes.NewAnyWithValue(ext)
		s.Require().NoError(err)
		s.Require().NotNil(extAny)

		txBuilder := encodingConfig.TxConfig.NewTxBuilder()

		s.Require().NoError(txBuilder.SetMsgs(sendMsg))

		txBuilder.SetMemo("fail-non-critical-ext-message")
		txBuilder.SetFeeAmount(sdk.NewCoins(standardFees))
		txBuilder.SetGasLimit(200000)

		// add extension options
		tx := txBuilder.GetTx()
		if etx, ok := tx.(authTx.ExtensionOptionsTxBuilder); ok {
			etx.SetNonCriticalExtensionOptions(extAny)
		}

		bz, err := encodingConfig.TxConfig.TxEncoder()(tx)
		s.Require().NoError(err)
		s.Require().NotNil(bz)

		// decode fails because the provided extension option does not implement the correct TxExtensionOptionI interface
		txWithExt, err := decodeTx(bz)
		s.Require().Error(err)
		s.Require().ErrorContains(err, "failed to decode tx: no concrete type registered for type URL /cosmos.bank.v1beta1.MsgMultiSend against interface *tx.TxExtensionOptionI")
		s.Require().Nil(txWithExt)
	})
}
