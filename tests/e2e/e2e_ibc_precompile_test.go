package e2e

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *IntegrationTestSuite) testIBCPrecompileTransfer() {
	s.Run("send_akii_to_chainB", func() {
		// require the recipient account receives the IBC tokens (IBC packets ACKd)
		var (
			balances      sdk.Coins
			err           error
			beforeBalance int64
			ibcStakeDenom string
		)

		address, _ := s.chainA.validators[0].keyInfo.GetAddress()
		sender := address.String()

		address, _ = s.chainB.validators[0].keyInfo.GetAddress()
		recipient := address.String()

		chainBAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainB.id][0].GetHostPort("1317/tcp"))

		s.Require().Eventually(
			func() bool {
				balances, err = queryKiichainAllBalances(chainBAPIEndpoint, recipient)
				s.Require().NoError(err)
				return balances.Len() != 0
			},
			time.Minute,
			5*time.Second,
		)
		for _, c := range balances {
			if strings.Contains(c.Denom, "ibc/") {
				beforeBalance = c.Amount.Int64()
				break
			}
		}

		tokenAmt := tokenAmount.Amount // 3,300 Kii
		s.sendIBCPrecompile(s.chainA, 0, sender, recipient, tokenAmount.String()+akiiDenom, standardFees.String(), "", false)

		pass := s.hermesClearPacket(hermesConfigWithGasPrices, s.chainA.id, transferPort, transferChannel)
		s.Require().True(pass)

		s.Require().Eventually(
			func() bool {
				balances, err = queryKiichainAllBalances(chainBAPIEndpoint, recipient)
				s.Require().NoError(err)
				return balances.Len() != 0
			},
			time.Minute,
			5*time.Second,
		)
		for _, c := range balances {
			if strings.Contains(c.Denom, "ibc/") {
				ibcStakeDenom = c.Denom
				s.Require().Equal((tokenAmt.Add(math.NewInt(beforeBalance))), c.Amount)
				break
			}
		}

		s.Require().NotEmpty(ibcStakeDenom)
	})
}

func (s *IntegrationTestSuite) sendIBCPrecompile(c *chain, valIdx int, sender, recipient, token, fees, note string, expErr bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	ibcCmd := []string{
		kiichaindBinary,
		txCommand,
		"ibc-transfer",
		"transfer",
		"transfer",
		"channel-0",
		recipient,
		token,
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagFees, fees),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		// fmt.Sprintf("--%s=%s", flags.FlagNote, note),
		fmt.Sprintf("--memo=%s", note),
		"--keyring-backend=test",
		"--broadcast-mode=sync",
		"--output=json",
		"-y",
	}
	s.T().Logf("sending %s from %s (%s) to %s (%s) with memo %s", token, s.chainA.id, sender, s.chainB.id, recipient, note)
	if expErr {
		s.executeKiichainTxCommand(ctx, c, ibcCmd, valIdx, s.expectErrExecValidation(c, valIdx, true))
		s.T().Log("unsuccessfully sent IBC tokens")
	} else {
		s.executeKiichainTxCommand(ctx, c, ibcCmd, valIdx, s.defaultExecValidation(c, valIdx))
		s.T().Log("successfully sent IBC tokens")
	}
}
