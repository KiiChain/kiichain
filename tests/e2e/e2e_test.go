// e2e_test.go
package e2e

import (
	"fmt"
	"os"
	"testing"
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
	runRewardsTest                = true
	runEVMTest                    = true
	runERC20Test                  = true
	runWasmTest                   = true
	runOracleTest                 = true

	// skipIBCTests skips tests that uses IBC
	skipIBCTests = os.Getenv("SKIP_IBC_TESTS") == "true"
)

// TestParallelE2E runs all E2E tests in parallel with isolated networks
func TestParallelE2E(t *testing.T) {
	// Define all test configurations
	tests := []struct {
		name    string
		enabled bool
		runner  func(*testing.T)
	}{
		{
			name:    "TestRestInterfaces",
			enabled: runRestInterfacesTest,
			runner: func(t *testing.T) {
				s := NewIntegrationTestSuite(t)
				s.SetupSuite()
				defer s.TearDownSuite()
				s.testRestInterfaces()
			},
		},
		{
			name:    "TestBank",
			enabled: runBankTest,
			runner: func(t *testing.T) {
				s := NewIntegrationTestSuite(t)
				s.SetupSuite()
				defer s.TearDownSuite()
				s.testBankTokenTransfer()
			},
		},
		{
			name:    "TestEncode",
			enabled: runEncodeTest,
			runner: func(t *testing.T) {
				s := NewIntegrationTestSuite(t)
				s.SetupSuite()
				defer s.TearDownSuite()
				s.testEncode()
				s.testDecode()
			},
		},
		{
			name:    "TestEvidence",
			enabled: runEvidenceTest,
			runner: func(t *testing.T) {
				s := NewIntegrationTestSuite(t)
				s.SetupSuite()
				defer s.TearDownSuite()
				s.testEvidence()
			},
		},
		{
			name:    "TestFeeGrant",
			enabled: runFeeGrantTest,
			runner: func(t *testing.T) {
				s := NewIntegrationTestSuite(t)
				s.SetupSuite()
				defer s.TearDownSuite()
				s.testFeeGrant()
			},
		},
		{
			name:    "TestGov",
			enabled: runGovTest,
			runner: func(t *testing.T) {
				s := NewIntegrationTestSuite(t)
				s.SetupSuite()
				defer s.TearDownSuite()
				s.GovCancelSoftwareUpgrade()
				s.GovCommunityPoolSpend()
				s.GovSoftwareUpgradeExpedited()
			},
		},
		{
			name:    "TestIBC",
			enabled: runIBCTest && !skipIBCTests,
			runner: func(t *testing.T) {
				s := NewIntegrationTestSuite(t)
				s.SetupSuite()
				defer s.TearDownSuite()
				s.testIBCTokenTransfer()
				s.testMultihopIBCTokenTransfer()
				s.testFailedMultihopIBCTokenTransfer()
				s.testICARegisterAccountAndSendTx()
			},
		},
		{
			name:    "TestSlashing",
			enabled: runSlashingTest,
			runner: func(t *testing.T) {
				s := NewIntegrationTestSuite(t)
				s.SetupSuite()
				defer s.TearDownSuite()
				chainAPI := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
				s.testSlashing(chainAPI)
			},
		},
		{
			name:    "TestStakingAndDistribution",
			enabled: runStakingAndDistributionTest,
			runner: func(t *testing.T) {
				s := NewIntegrationTestSuite(t)
				s.SetupSuite()
				defer s.TearDownSuite()
				s.testStaking()
				s.testDistribution()
			},
		},
		{
			name:    "TestVesting",
			enabled: runVestingTest,
			runner: func(t *testing.T) {
				s := NewIntegrationTestSuite(t)
				s.SetupSuite()
				defer s.TearDownSuite()
				chainAAPI := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("1317/tcp"))
				s.testDelayedVestingAccount(chainAAPI)
				s.testContinuousVestingAccount(chainAAPI)
			},
		},
		{
			name:    "TestRewards",
			enabled: runRewardsTest,
			runner: func(t *testing.T) {
				s := NewIntegrationTestSuite(t)
				s.SetupSuite()
				defer s.TearDownSuite()
				s.testRewardUpdate()
			},
		},
		{
			name:    "TestTokenFactory",
			enabled: runTokenFactoryTest,
			runner: func(t *testing.T) {
				s := NewIntegrationTestSuite(t)
				s.SetupSuite()
				defer s.TearDownSuite()
				s.testTokenFactory()
			},
		},
		{
			name:    "TestRateLimit",
			enabled: runRateLimitTest && !skipIBCTests,
			runner: func(t *testing.T) {
				s := NewIntegrationTestSuite(t)
				s.SetupSuite()
				defer s.TearDownSuite()
				s.testAddRateLimits()
				s.testIBCTransfer(true)
				s.testUpdateRateLimit()
				s.testIBCTransfer(false)
				s.testResetRateLimit()
				s.testRemoveRateLimit()
			},
		},
		{
			name:    "TestEVM",
			enabled: runEVMTest,
			runner: func(t *testing.T) {
				s := NewIntegrationTestSuite(t)
				s.SetupSuite()
				defer s.TearDownSuite()
				jsonRPC := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("8545/tcp"))
				s.testEVMQueries(jsonRPC)
				s.testEVM(jsonRPC)
				// Disabled until mempool bug is fixed
				// s.testMempoolEVM(jsonRPC)
			},
		},
		{
			name:    "TestERC20",
			enabled: runERC20Test,
			runner: func(t *testing.T) {
				s := NewIntegrationTestSuite(t)
				s.SetupSuite()
				defer s.TearDownSuite()
				jsonRPC := fmt.Sprintf("http://%s", s.valResources[s.chainA.id][0].GetHostPort("8545/tcp"))
				s.testERC20(jsonRPC)
			},
		},
		{
			name:    "TestWasm",
			enabled: runWasmTest,
			runner: func(t *testing.T) {
				s := NewIntegrationTestSuite(t)
				s.SetupSuite()
				defer s.TearDownSuite()
				s.testWasmdCounter()
				// Test precompile
				s.testWasmdPrecompile()
			},
		},
		{
			name:    "TestOracle",
			enabled: runOracleTest,
			runner: func(t *testing.T) {
				s := NewIntegrationTestSuite(t)
				s.SetupSuite()
				defer s.TearDownSuite()
				s.testFeelessTx()
				s.testFeeder()
				s.testSlash()
			},
		},
	}

	// Run all tests in parallel
	for _, tc := range tests {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			if !tc.enabled {
				t.Skipf("%s is disabled", tc.name)
			}
			t.Parallel() // Each test runs in parallel
			tc.runner(t)
		})
	}
}
