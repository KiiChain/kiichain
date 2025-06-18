package e2e

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/cosmos/cosmos-sdk/client/flags"

	erc20types "github.com/cosmos/evm/x/erc20/types"

	"github.com/kiichain/kiichain/v2/tests/e2e/mock"
)

// testEVM Tests EVM send and contract usage
func (s *IntegrationTestSuite) testERC20(jsonRCP string) {
	var (
		err           error
		valIdx        = 0
		c             = s.chainA
		chainEndpoint = fmt.Sprintf("http://%s", s.valResources[c.id][valIdx].GetHostPort("1317/tcp"))
	)

	// Get a funded EVM account and check balance transactions
	key, evmAddress := s.setupEVMwithFunds(jsonRCP, chainEndpoint, valIdx)

	// Setup client
	client, err := ethclient.Dial(jsonRCP)
	s.Require().NoError(err)

	// 1. Deploy ERC20 contract
	// Prepare auth
	auth, err := bind.NewKeyedTransactorWithChainID(key, big.NewInt(1010))
	if err != nil {
		log.Fatal(err)
	}

	// Set optional params
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000) // gas limit
	auth.GasPrice, _ = client.SuggestGasPrice(context.Background())

	// Deploy
	contractAddress, tx, erc20, err := mock.DeployERC20Mock(auth, client)
	s.Require().NoError(err)
	s.waitForTransaction(client, tx)

	// Setup alice information
	publicKey, err := s.chainA.genesisAccounts[1].keyInfo.GetPubKey()
	s.Require().NoError(err)
	aliceEvmAddress, err := CosmosPubKeyToEVMAddress(publicKey)
	s.Require().NoError(err)
	alice, err := s.chainA.genesisAccounts[1].keyInfo.GetAddress()
	s.Require().NoError(err)

	// Test minting and balance change
	amount := big.NewInt(1000000000000000000)
	doubleAmount := big.NewInt(2000000000000000000)
	s.Run("Interact w/ ERC20", func() {
		auth.Nonce = big.NewInt(int64(tx.Nonce() + 1)) // update nonce

		// Mint some amount
		mintTx, err := erc20.Mint(auth, evmAddress, doubleAmount)
		s.Require().NoError(err)
		s.waitForTransaction(client, mintTx)

		// Setup call options
		callOpts := &bind.CallOpts{
			Pending: false,
			Context: context.Background(),
		}

		// Balance should have changed
		newBalance, err := erc20.BalanceOf(callOpts, evmAddress)
		s.Require().NoError(err)
		s.Require().Equal(doubleAmount, newBalance)

		// Transfer some to alice
		auth.Nonce = big.NewInt(int64(mintTx.Nonce() + 1)) // update nounce
		transferTx, err := erc20.Transfer(auth, aliceEvmAddress, amount)
		s.Require().NoError(err)
		s.waitForTransaction(client, transferTx)

		aliceBalance, err := erc20.BalanceOf(callOpts, aliceEvmAddress)
		s.Require().NoError(err)
		s.Require().Equal(amount, aliceBalance)
	})

	chainAAPIEndpoint := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	senderAddress, _ := s.chainA.validators[0].keyInfo.GetAddress()
	sender := senderAddress.String()

	s.Run("Register ERC20 proposal", func() {
		proposalCounter++
		s.writeERC20RegisterProposal(c, contractAddress)
		submitGovFlags := []string{configFile(proposalRegisterERC20)}

		depositGovFlags := []string{strconv.Itoa(proposalCounter), depositAmount.String()}
		voteGovFlags := []string{strconv.Itoa(proposalCounter), "yes"}
		s.submitGovProposal(chainAAPIEndpoint, sender, proposalCounter, "RegisterERC20", submitGovFlags, depositGovFlags, voteGovFlags, "vote")
	})

	s.Run("ConvertERC20 to native", func() {
		// Convert amount to native coins
		s.convertERC20(c, valIdx, contractAddress, alice.String(), amount)

		// Get specific erc20 native balance
		denom := fmt.Sprintf("erc20/%s", contractAddress)
		erc20Balance, err := getSpecificBalance(chainAAPIEndpoint, alice.String(), denom)
		s.Require().NoError(err)
		s.T().Logf("ERC20 Balance: %s", erc20Balance)
		// converting to string since one is big int and the other is math int
		s.Require().Equal(amount.String(), erc20Balance.Amount.String())
	})
}

// convertERC20 calls the CLI to transfer the given erc20 contract coin to the linked native pair
func (s *IntegrationTestSuite) convertERC20(c *chain, valIdx int, contractAddress common.Address, sender string, amount *big.Int) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	kiichainCommand := []string{
		kiichaindBinary,
		txCommand,
		erc20types.ModuleName,
		"convert-erc20",
		contractAddress.String(),
		amount.String(), // not a coin
		fmt.Sprintf("--from=%s", sender),
		fmt.Sprintf("--%s=%s", flags.FlagChainID, c.id),
		fmt.Sprintf("--%s=%s", flags.FlagGasPrices, "300000000akii"),
		fmt.Sprintf("--%s=%s", flags.FlagGas, "5000000"),
		"--keyring-backend=test",
		"--output=json",
		"-y",
	}

	s.executeKiichainTxCommand(ctx, c, kiichainCommand, valIdx, s.defaultExecValidation(c, valIdx))
}

// writeERC20RegisterProposal stores a file with the ERC20 Register proposal
func (s *IntegrationTestSuite) writeERC20RegisterProposal(c *chain, erc20Address common.Address) {
	body := `{
		"messages": [
		 {
		  "@type": "/cosmos.evm.erc20.v1.MsgRegisterERC20",
		  "authority": "kii10d07y265gmmuvt4z0w9aw880jnsr700jrff0qv",
		  "erc20addresses": [
		    "%s"
		  ]
		 }
		],
		"metadata": "ipfs://CID",
		"deposit": "100akii",
		"title": "title",
		"summary": "test"
	   }`

	propMsgBody := fmt.Sprintf(body, erc20Address.String())

	err := writeFile(filepath.Join(c.validators[0].configDir(), "config", proposalRegisterERC20), []byte(propMsgBody))
	s.Require().NoError(err)
}
