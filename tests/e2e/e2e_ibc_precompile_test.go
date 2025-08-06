package e2e

import (
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kiichain/kiichain/v4/tests/e2e/precompiles"
)

const (
	IBCPrecompileAddress = "0x0000000000000000000000000000000000001002"
)

// testIBCPrecompileTransfer tests transfer with the ibc precompile
func (s *IntegrationTestSuite) testIBCPrecompileTransfer(jsonRPC string) {
	s.Run("send_akii_to_chainB via precompile", func() {
		// require the recipient account receives the IBC tokens (IBC packets ACKd)
		var (
			balances      sdk.Coins
			err           error
			beforeBalance math.Int
			ibcStakeDenom string
		)

		evmAccount := s.chainA.evmAccount

		address, _ := s.chainB.validators[0].keyInfo.GetAddress()
		recipient := address.String()

		chainBAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainB.id][0].GetHostPort("1317/tcp"))

		s.Require().Eventually(
			func() bool {
				balances, err = queryKiichainAllBalances(chainBAPIEndpoint, recipient)
				s.Require().NoError(err)

				// There should be some IBC balance from previous tests
				for _, c := range balances {
					if strings.Contains(c.Denom, "ibc/") {
						beforeBalance = c.Amount
						return true
					}
				}
				return false
			},
			time.Minute,
			5*time.Second,
		)

		// Send via precompile
		tokenAmt := standardFees.Amount // 0.33 Kii
		s.sendIBCPrecompile(jsonRPC, evmAccount, recipient, standardFees, "")

		// Apply packet changes
		pass := s.hermesClearPacket(hermesConfigWithGasPrices, s.chainA.id, transferPort, transferChannel)
		s.Require().True(pass)

		s.Require().Eventually(
			func() bool {
				balances, err = queryKiichainAllBalances(chainBAPIEndpoint, recipient)
				s.Require().NoError(err)

				// Check if the balance has increased
				for _, c := range balances {
					if strings.Contains(c.Denom, "ibc/") {
						ibcStakeDenom = c.Denom
						s.Require().Equal((tokenAmt.Add(beforeBalance)), c.Amount)
						return true
					}
				}
				return false
			},
			time.Minute,
			5*time.Second,
		)

		s.Require().NotEmpty(ibcStakeDenom)
	})
}

// sendIBCPrecompile sends funds via IBC precompile to a receipient using default timeout options
func (s *IntegrationTestSuite) sendIBCPrecompile(jsonRPC string, senderEvmAccount EVMAccount, recipient string, token sdk.Coin, note string) {
	// Setup client
	client, err := ethclient.Dial(jsonRPC)
	s.Require().NoError(err)

	// Deploy contract
	s.Run("send to IBC precompile transfer", func() {
		// Bind
		ibcPrecompile, err := precompiles.NewIbcPrecompile(common.HexToAddress(IBCPrecompileAddress), client)
		s.Require().NoError(err)

		// Call transfer
		tx, err := ibcPrecompile.TransferWithDefaultTimeout(
			setupDefaultAuth(client, senderEvmAccount.key),
			recipient,
			transferPort,
			transferChannel,
			token.Denom,
			token.Amount.BigInt(),
			note,
		)
		s.Require().NoError(err)

		// Wait and check tx
		s.waitForTransaction(client, tx, senderEvmAccount.address)
	})
}
