package e2e

import (
	"context"
	"fmt"
	"math/big"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	geth "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"cosmossdk.io/math"

	circuittypes "github.com/cosmos/cosmos-sdk/contrib/x/circuit/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	stakingprecompile "github.com/cosmos/evm/precompiles/staking"
	evmtypes "github.com/cosmos/evm/x/vm/types"

	"github.com/kiichain/kiichain/v7/tests/e2e/mock"
)

const (
	// circuitAdminAccountIndex is the genesis account granted LEVEL_SUPER_ADMIN on x/circuit
	circuitAdminAccountIndex = 4

	// stakingPrecompileAddress is the cosmos/evm staking precompile
	stakingPrecompileAddress = "0x0000000000000000000000000000000000000800"

	// circuitAnteErr is returned by the circuit ante decorator for a tripped msg
	circuitAnteErr = "tx type not allowed"
	// circuitPrecompileErr is returned by the precompile MsgServer wrappers for a tripped msg
	circuitPrecompileErr = "circuit breaker disables execution of this message"

	// circuitTxGas is fixed so a tripped tx is broadcast instead of failing in gas simulation
	circuitTxGas = "300000"
)

var (
	msgSendTypeURL     = sdk.MsgTypeURL(&banktypes.MsgSend{})
	msgDelegateTypeURL = sdk.MsgTypeURL(&stakingtypes.MsgDelegate{})
	msgEthereumTxURL   = sdk.MsgTypeURL(&evmtypes.MsgEthereumTx{})
)

// testCircuitAuthorize grants an account LEVEL_SOME_MSGS through the CLI, which
// uses the flat permissions JSON AutoCLI override in cmd/kiichaind
func (s *IntegrationTestSuite) testCircuitAuthorize() {
	c := s.chainA
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[c.id][0].GetHostPort("1317/tcp"))
	admin := s.circuitAdminAddress()

	grantee, err := PubKeyBytesToCosmosAddress(c.evmAccount.address.Bytes())
	s.Require().NoError(err)

	s.Run("super admin is set from genesis", func() {
		res, err := queryCircuitAccount(chainEndpoint, admin)
		s.Require().NoError(err)
		s.Require().Equal(circuittypes.Permissions_LEVEL_SUPER_ADMIN, res.Permission.Level)
	})

	s.Run("authorize LEVEL_SOME_MSGS via CLI", func() {
		permissions := fmt.Sprintf(`{"level":"LEVEL_SOME_MSGS","limit_type_urls":[%q]}`, msgSendTypeURL)
		s.execCircuitTx(c, 0, admin, []string{"authorize", grantee, permissions}, s.defaultExecValidation(c, 0))

		s.Require().Eventually(
			func() bool {
				res, err := queryCircuitAccount(chainEndpoint, grantee)
				if err != nil || res.Permission == nil {
					return false
				}
				return res.Permission.Level == circuittypes.Permissions_LEVEL_SOME_MSGS &&
					slices.Equal(res.Permission.LimitTypeUrls, []string{msgSendTypeURL})
			},
			20*time.Second,
			5*time.Second,
		)
	})
}

// testCircuitCosmosTx trips MsgSend and checks a Cosmos tx is rejected by the
// ante handler, then resets it and checks the same send goes through
func (s *IntegrationTestSuite) testCircuitCosmosTx() {
	c := s.chainA
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[c.id][0].GetHostPort("1317/tcp"))
	admin := s.circuitAdminAddress()
	defer s.resetCircuitBreakers(chainEndpoint, admin)

	recipient, err := PubKeyBytesToCosmosAddress(c.evmAccount.address.Bytes())
	s.Require().NoError(err)
	amount := sdk.NewCoin(akiiDenom, math.NewInt(1000))

	s.Run("trip MsgSend", func() {
		s.tripCircuitBreaker(chainEndpoint, admin, msgSendTypeURL)
	})

	s.Run("bank send is rejected while tripped", func() {
		before, err := getSpecificBalance(chainEndpoint, recipient, akiiDenom)
		s.Require().NoError(err)

		s.execCircuitBankSend(c, 0, admin, recipient, amount.String(), s.execValidationWithError(c, 0, circuitAnteErr))

		after, err := getSpecificBalance(chainEndpoint, recipient, akiiDenom)
		s.Require().NoError(err)
		s.Require().Equal(before.Amount.String(), after.Amount.String())
	})

	s.Run("reset MsgSend", func() {
		s.resetCircuitBreaker(chainEndpoint, admin, msgSendTypeURL)
	})

	s.Run("bank send succeeds after reset", func() {
		before, err := getSpecificBalance(chainEndpoint, recipient, akiiDenom)
		s.Require().NoError(err)

		s.execCircuitBankSend(c, 0, admin, recipient, amount.String(), s.defaultExecValidation(c, 0))

		s.Require().Eventually(
			func() bool {
				after, err := getSpecificBalance(chainEndpoint, recipient, akiiDenom)
				s.Require().NoError(err)
				return after.Amount.Equal(before.Amount.Add(amount.Amount))
			},
			20*time.Second,
			5*time.Second,
		)
	})
}

// testCircuitEVMPrecompileTx trips MsgDelegate and checks an EVM tx calling the
// staking precompile reverts, since precompiles skip the ante handler and the
// BaseApp router. It then resets it and checks the same call delegates
func (s *IntegrationTestSuite) testCircuitEVMPrecompileTx(jsonRPC string) {
	c := s.chainA
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[c.id][0].GetHostPort("1317/tcp"))
	admin := s.circuitAdminAddress()
	defer s.resetCircuitBreakers(chainEndpoint, admin)

	evmAccount := c.evmAccount
	delegator, err := PubKeyBytesToCosmosAddress(evmAccount.address.Bytes())
	s.Require().NoError(err)

	valAddr, err := c.validators[0].keyInfo.GetAddress()
	s.Require().NoError(err)
	validator := sdk.ValAddress(valAddr).String()

	client, err := ethclient.Dial(jsonRPC)
	s.Require().NoError(err)
	staking := bind.NewBoundContract(common.HexToAddress(stakingPrecompileAddress), stakingprecompile.ABI, client, client, client)

	amount := big.NewInt(1_000_000_000_000_000_000) // 1 KII

	s.Run("trip MsgDelegate", func() {
		s.tripCircuitBreaker(chainEndpoint, admin, msgDelegateTypeURL)
	})

	s.Run("staking precompile delegate reverts while tripped", func() {
		before := s.delegationShares(chainEndpoint, validator, delegator)

		// setupDefaultAuth uses a fixed gas limit, so the reverting call is
		// broadcast and mined instead of failing in gas estimation
		tx, err := staking.Transact(setupDefaultAuth(client, evmAccount.key), "delegate", evmAccount.address, validator, amount)
		s.Require().NoError(err)

		receipt, err := bind.WaitMined(context.Background(), client, tx)
		s.Require().NoError(err)
		s.Require().Equal(geth.ReceiptStatusFailed, receipt.Status)

		reason, err := getRevertReason(client, tx.Hash(), evmAccount.address)
		s.Require().NoError(err)
		s.Require().Contains(reason, circuitPrecompileErr)

		after := s.delegationShares(chainEndpoint, validator, delegator)
		s.Require().True(after.Equal(before), "delegation changed while MsgDelegate was tripped: %s -> %s", before, after)
	})

	s.Run("reset MsgDelegate", func() {
		s.resetCircuitBreaker(chainEndpoint, admin, msgDelegateTypeURL)
	})

	s.Run("staking precompile delegate succeeds after reset", func() {
		before := s.delegationShares(chainEndpoint, validator, delegator)

		tx, err := staking.Transact(setupDefaultAuth(client, evmAccount.key), "delegate", evmAccount.address, validator, amount)
		s.Require().NoError(err)
		s.waitForTransaction(client, tx, evmAccount.address)

		s.Require().Eventually(
			func() bool {
				return s.delegationShares(chainEndpoint, validator, delegator).GT(before)
			},
			20*time.Second,
			5*time.Second,
		)
	})
}

// testCircuitEVMTx trips MsgEthereumTx and checks every EVM tx (a native
// transfer and a contract call) is rejected at submission without charging a
// fee or consuming a nonce, then resets it and checks the contract call works
func (s *IntegrationTestSuite) testCircuitEVMTx(jsonRPC string) {
	c := s.chainA
	chainEndpoint := fmt.Sprintf("http://%s", s.valResources[c.id][0].GetHostPort("1317/tcp"))
	admin := s.circuitAdminAddress()
	defer s.resetCircuitBreakers(chainEndpoint, admin)

	evmAccount := c.evmAccount
	client, err := ethclient.Dial(jsonRPC)
	s.Require().NoError(err)

	// Deploy the counter before tripping, so the blocked call targets a live contract
	_, tx, counter, err := mock.DeployCounter(setupDefaultAuth(client, evmAccount.key), client)
	s.Require().NoError(err)
	s.waitForTransaction(client, tx, evmAccount.address)

	counterBefore, err := counter.GetCounter(nil)
	s.Require().NoError(err)

	s.Run("trip MsgEthereumTx", func() {
		s.tripCircuitBreaker(chainEndpoint, admin, msgEthereumTxURL)
	})

	s.Run("EVM txs are rejected while tripped", func() {
		ctx := context.Background()
		nonceBefore, err := client.PendingNonceAt(ctx, evmAccount.address)
		s.Require().NoError(err)
		balanceBefore, err := client.BalanceAt(ctx, evmAccount.address, nil)
		s.Require().NoError(err)

		_, err = EVMSendWithNonce(client, evmAccount.key, evmAccount.address, s.chainB.evmAccount.address, big.NewInt(1), nil, nonceBefore)
		s.Require().ErrorContains(err, circuitAnteErr)

		_, err = counter.Increment(setupDefaultAuth(client, evmAccount.key))
		s.Require().ErrorContains(err, circuitAnteErr)

		// Nothing was charged: the circuit decorator runs before the EVM ante
		// deducts fees and bumps the nonce
		nonceAfter, err := client.PendingNonceAt(ctx, evmAccount.address)
		s.Require().NoError(err)
		s.Require().Equal(nonceBefore, nonceAfter)
		balanceAfter, err := client.BalanceAt(ctx, evmAccount.address, nil)
		s.Require().NoError(err)
		s.Require().Equal(balanceBefore.String(), balanceAfter.String())

		counterValue, err := counter.GetCounter(nil)
		s.Require().NoError(err)
		s.Require().Equal(counterBefore, counterValue)
	})

	s.Run("reset MsgEthereumTx", func() {
		s.resetCircuitBreaker(chainEndpoint, admin, msgEthereumTxURL)
	})

	s.Run("EVM contract call succeeds after reset", func() {
		tx, err := counter.Increment(setupDefaultAuth(client, evmAccount.key))
		s.Require().NoError(err)
		s.waitForTransaction(client, tx, evmAccount.address)

		counterValue, err := counter.GetCounter(nil)
		s.Require().NoError(err)
		s.Require().Equal(new(big.Int).Add(counterBefore, big.NewInt(1)), counterValue)
	})
}

// circuitAdminAddress returns the bech32 address of the genesis circuit super admin
func (s *IntegrationTestSuite) circuitAdminAddress() string {
	addr, err := s.chainA.genesisAccounts[circuitAdminAccountIndex].keyInfo.GetAddress()
	s.Require().NoError(err)
	return addr.String()
}

// delegationShares returns the delegator's shares on a validator, or zero if there is no delegation
func (s *IntegrationTestSuite) delegationShares(endpoint, validator, delegator string) math.LegacyDec {
	res, err := queryDelegation(endpoint, validator, delegator)
	if err != nil || res.DelegationResponse == nil {
		return math.LegacyZeroDec()
	}
	return res.DelegationResponse.Delegation.Shares
}

// tripCircuitBreaker disables msgTypeURL and waits until it shows in the disabled list
func (s *IntegrationTestSuite) tripCircuitBreaker(endpoint, admin, msgTypeURL string) {
	s.execCircuitTx(s.chainA, 0, admin, []string{"disable", msgTypeURL}, s.defaultExecValidation(s.chainA, 0))
	s.Require().Eventually(
		func() bool {
			disabled, err := queryCircuitDisabledList(endpoint)
			return err == nil && slices.Contains(disabled, msgTypeURL)
		},
		20*time.Second,
		5*time.Second,
	)
}

// resetCircuitBreaker re-enables msgTypeURL and waits until it leaves the disabled list
func (s *IntegrationTestSuite) resetCircuitBreaker(endpoint, admin, msgTypeURL string) {
	s.execCircuitTx(s.chainA, 0, admin, []string{"reset", msgTypeURL}, s.defaultExecValidation(s.chainA, 0))
	s.Require().Eventually(
		func() bool {
			disabled, err := queryCircuitDisabledList(endpoint)
			return err == nil && !slices.Contains(disabled, msgTypeURL)
		},
		20*time.Second,
		5*time.Second,
	)
}

// resetCircuitBreakers re-enables anything still disabled, so a failed
// circuit test does not leave msgs tripped for the rest of the suite
func (s *IntegrationTestSuite) resetCircuitBreakers(endpoint, admin string) {
	disabled, err := queryCircuitDisabledList(endpoint)
	s.Require().NoError(err)
	for _, msgTypeURL := range disabled {
		s.resetCircuitBreaker(endpoint, admin, msgTypeURL)
	}
}

// execCircuitTx runs a `kiichaind tx circuit` command
func (s *IntegrationTestSuite) execCircuitTx(c *chain, valIdx int, from string, args []string, validation func([]byte, []byte) bool) {
	opts := applyOptions(c.id, []flagOption{
		withKeyValue(flagFrom, from),
		withKeyValue(flagGas, circuitTxGas),
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	kiichainCommand := []string{
		kiichaindBinary,
		txCommand,
		circuittypes.ModuleName,
	}
	kiichainCommand = append(kiichainCommand, args...)
	kiichainCommand = append(kiichainCommand, "-y")
	for flag, value := range opts {
		kiichainCommand = append(kiichainCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.T().Logf("Executing kiichaind tx circuit %v on chain %s", args, c.id)
	s.executeKiichainTxCommand(ctx, c, kiichainCommand, valIdx, validation)
}

// execCircuitBankSend runs a bank send with a fixed gas limit and a custom validation
func (s *IntegrationTestSuite) execCircuitBankSend(c *chain, valIdx int, from, to, amt string, validation func([]byte, []byte) bool) {
	opts := applyOptions(c.id, []flagOption{
		withKeyValue(flagFrom, from),
		withKeyValue(flagGas, circuitTxGas),
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	kiichainCommand := []string{
		kiichaindBinary,
		txCommand,
		banktypes.ModuleName,
		"send",
		from,
		to,
		amt,
		"-y",
	}
	for flag, value := range opts {
		kiichainCommand = append(kiichainCommand, fmt.Sprintf("--%s=%v", flag, value))
	}

	s.T().Logf("sending %s from %s to %s on chain %s", amt, from, to, c.id)
	s.executeKiichainTxCommand(ctx, c, kiichainCommand, valIdx, validation)
}
