package e2e

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/kiichain/kiichain/v2/tests/e2e/precompiles"
)

const (
	IBCPrecompileAddress = "0x0000000000000000000000000000000000001002"
)

// testIBCPrecompileTransfer tests transfer with the ibc precompile
func (s *IntegrationTestSuite) testIBCPrecompileTransfer(jsonRPC string) {
	s.Run("send_akii_to_chainB", func() {
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
				return balances.Len() != 0
			},
			time.Minute,
			5*time.Second,
		)
		for _, c := range balances {
			if strings.Contains(c.Denom, "ibc/") {
				beforeBalance = c.Amount
				break
			}
		}

		tokenAmt := tokenAmount.Amount // 3,300 Kii
		s.sendIBCPrecompile(jsonRPC, evmAccount, recipient, tokenAmount, "precompile ibc transfer")
		s.sendIBCPrecompile(jsonRPC, evmAccount, recipient, tokenAmount, "")

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
				s.Require().Equal((tokenAmt.Add(beforeBalance)), c.Amount)
				break
			}
		}

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
		// Prepare auth
		auth, err := bind.NewKeyedTransactorWithChainID(senderEvmAccount.key, big.NewInt(1010))
		if err != nil {
			log.Fatal(err)
		}

		// Set optional params
		auth.Value = big.NewInt(0)
		auth.GasLimit = uint64(3000000) // gas limit
		auth.GasPrice, _ = client.SuggestGasPrice(context.Background())

		// Deploy
		ibcPrecompile, err := precompiles.NewIbcPrecompile(common.HexToAddress(IBCPrecompileAddress), client)
		s.Require().NoError(err)

		// Get height + 25 blocks
		height := 25 + s.getLatestBlockHeight(s.chainA, 0)

		tx, err := ibcPrecompile.Transfer(
			auth,
			recipient,
			transferPort,
			transferChannel,
			token.Denom,
			token.Amount.BigInt(),
			1, // revisionNumber
			uint64(height),
			0, // timeoutTimestamp
			note,
		)
		s.Require().NoError(err)

		// Wait and check tx
		receipt := s.waitForTransaction(client, tx)
		s.Require().False(receipt.Status == geth.ReceiptStatusFailed)
	})
}
