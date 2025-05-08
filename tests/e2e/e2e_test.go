package e2e

import (
	"fmt"
	"os"
)

var (
	runBankTest                   = true
	runEncodeTest                 = true
	runEvidenceTest               = true
	runFeeGrantTest               = true
	runGovTest                    = true
	runIBCTest                    = true
	runSlashingTest               = true
	runStakingAndDistributionTest = true
	runVestingTest                = true
	runRestInterfacesTest         = true
	runRateLimitTest              = true
	runTokenFactoryTest           = true
	runEVMTest                    = true
	runERC20Test                  = true
	runWasmTest                   = true

	// skipIBCTests skips tests that uses IBC
	skipIBCTests = os.Getenv("SKIP_IBC_TESTS") == "true"
)

// TestRestInterfaces runs the rest interfaces tests. It is skipped if the variable is set
func (s *IntegrationTestSuite) TestRestInterfaces() {
	if !runRestInterfacesTest {
		s.T().Skip()
	}
	s.testRestInterfaces()
}

// TestBank runs the bank tests. It is skipped if the variable is set
func (s *IntegrationTestSuite) TestBank() {
	if !runBankTest {
		s.T().Skip()
	}
	s.testBankTokenTransfer()
}

// TestEncode runs the encode tests. It is skipped if the variable is set
func (s *IntegrationTestSuite) TestEncode() {
	if !runEncodeTest {
		s.T().Skip()
	}
	s.testEncode()
	s.testDecode()
}

// TestEvidence runs the evidence tests. It is skipped if the variable is set
func (s *IntegrationTestSuite) TestEvidence() {
	if !runEvidenceTest {
		s.T().Skip()
	}
	s.testEvidence()
}

// TestFeeGrant runs the fee grant tests. It is skipped if the variable is set
func (s *IntegrationTestSuite) TestFeeGrant() {
	if !runFeeGrantTest {
		s.T().Skip()
	}
	s.testFeeGrant()
}

// TestGov runs the governance tests. It is skipped if the variable is set
func (s *IntegrationTestSuite) TestGov() {
	if !runGovTest {
		s.T().Skip()
	}

	s.GovCancelSoftwareUpgrade()
	s.GovCommunityPoolSpend()

	s.GovSoftwareUpgradeExpedited()
}

// TestIBC runs the IBC tests. It is skipped if the variable is set
func (s *IntegrationTestSuite) TestIBC() {
	if !runIBCTest || skipIBCTests {
		s.T().Log("skipping IBC e2e tests...")
		s.T().Skip()
	}

	s.testIBCTokenTransfer()
	s.testMultihopIBCTokenTransfer()
	s.testFailedMultihopIBCTokenTransfer()
	s.testICARegisterAccountAndSendTx()
}

// TestSlashing runs the slashing tests. It is skipped if the variable is set
func (s *IntegrationTestSuite) TestSlashing() {
	if !runSlashingTest {
		s.T().Skip()
	}
	chainAPI := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	s.testSlashing(chainAPI)
}

// todo add fee test with wrong denom order
func (s *IntegrationTestSuite) TestStakingAndDistribution() {
	if !runStakingAndDistributionTest {
		s.T().Skip()
	}
	s.testStaking()
	s.testDistribution()
}

// TestVesting runs the vesting tests. It is skipped if the variable is set
func (s *IntegrationTestSuite) TestVesting() {
	if !runVestingTest {
		s.T().Skip()
	}
	chainAAPI := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
	s.testDelayedVestingAccount(chainAAPI)
	s.testContinuousVestingAccount(chainAAPI)
}

// TestTokenFactory runs the token factory tests. It is skipped if the variable is set
func (s *IntegrationTestSuite) TestTokenFactory() {
	if !runTokenFactoryTest {
		s.T().Log("skipping token factory e2e tests...")
		s.T().Skip()
	}
	s.testTokenFactory()
}

// TestRateLimit runs the rate limit tests. It is skipped if the variable is set
func (s *IntegrationTestSuite) TestRateLimit() {
	if !runRateLimitTest || skipIBCTests {
		s.T().Log("skipping rate limit e2e tests...")
		s.T().Skip()
	}
	s.testAddRateLimits()
	s.testIBCTransfer(true)
	s.testUpdateRateLimit()
	s.testIBCTransfer(false)
	s.testResetRateLimit()
	s.testRemoveRateLimit()
}

// TestEVM runs basic EVM tests. It is skipped if the variable is set
func (s *IntegrationTestSuite) TestEVM() {
	if !runEVMTest {
		s.T().Log("skipping evm e2e tests...")
		s.T().Skip()
	}
	jsonRPC := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("8545/tcp"))
	s.testEVMQueries(jsonRPC)
	s.testEVM(jsonRPC)
}

// TestERC20 runs the ERC20 tests. It is skipped if the variable is set
func (s *IntegrationTestSuite) TestERC20() {
	if !runERC20Test {
		s.T().Skip()
	}
	jsonRPC := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("8545/tcp"))
	s.testERC20(jsonRPC)
}

// TestWasm runs the Wasm tests. It is skipped if the variable is set
func (s *IntegrationTestSuite) TestWasmd() {
	if !runWasmTest {
		s.T().Log("skipping wasm e2e tests...")
		s.T().Skip()
	}
	s.testWasmdCounter()
}
