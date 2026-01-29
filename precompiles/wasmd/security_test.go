package wasmd_test

import (
	"bytes"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	cmn "github.com/cosmos/evm/precompiles/common"
	"github.com/cosmos/evm/precompiles/testutil"

	wasmdprecompile "github.com/kiichain/kiichain/v7/precompiles/wasmd"
)

const (
	reentrancyCallError = "reentrant call"
)

// TestInvalidCodeIDValidation tests if the precompile properly validates code_id
// Bug Report Issue #4: No Input Validation
func (s *WasmdPrecompileTestSuite) TestInvalidCodeIDValidation() {
	method := s.Precompile.Methods[wasmdprecompile.InstantiateMethod]
	account := s.keyring.GetKey(0)

	testCases := []struct {
		name        string
		codeID      uint64
		expectError bool
		errContains string
	}{
		{
			name:        "zero code_id should fail",
			codeID:      0,
			expectError: true,
			errContains: "code id",
		},
		{
			name:        "valid code_id should succeed",
			codeID:      s.CounterCodeID,
			expectError: false,
		},
		{
			name:        "very large code_id should fail",
			codeID:      999999999,
			expectError: true,
			errContains: "no such code",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			stateDB := s.GetStateDB()
			contract, ctx := testutil.NewPrecompileContract(
				s.T(),
				s.Ctx,
				account.Addr,
				s.Precompile.Address(),
				200000,
			)

			args := []any{
				account.Addr,
				tc.codeID,
				"Test Label",
				[]byte(`"zero"`),
				[]cmn.Coin{},
			}

			_, err := s.Precompile.Instantiate(ctx, account.Addr, contract, stateDB, &method, args)

			if tc.expectError {
				s.Require().Error(err, "Expected error for code_id %d", tc.codeID)
				if tc.errContains != "" {
					s.Require().Contains(err.Error(), tc.errContains)
				}
			} else {
				s.Require().NoError(err, "Should succeed with valid code_id %d", tc.codeID)
			}
		})
	}
}

// TestUnsafeLogging tests if binary data in msg.Msg is logged safely
// Bug Report Issue #3: Unsafe Logging
func (s *WasmdPrecompileTestSuite) TestUnsafeLogging() {
	contractAddr := s.instantiateContract()
	method := s.Precompile.Methods[wasmdprecompile.ExecuteMethod]
	account := s.keyring.GetKey(0)

	testCases := []struct {
		name        string
		msg         []byte
		expectError bool
		description string
	}{
		{
			name:        "binary data with null bytes",
			msg:         []byte{0x00, 0xFF, 0xAB, 0xCD},
			expectError: true,
			description: "Raw binary should be handled safely in logs",
		},
		{
			name:        "very long message",
			msg:         make([]byte, 10000),
			expectError: true,
			description: "Very long messages should be truncated in logs",
		},
		{
			name:        "mixed binary and json",
			msg:         append([]byte(`{"set": `), []byte{0x00, 0xFF}...),
			expectError: true,
			description: "Mixed content should be sanitized",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Start a new logger with buffer to capture logs
			var buf bytes.Buffer
			testLogger := log.NewLogger(&buf)

			// Add to the context
			ctx := s.Ctx.WithLogger(testLogger)

			// Start the contract
			stateDB := s.GetStateDB()
			contract, ctx := testutil.NewPrecompileContract(
				s.T(),
				ctx,
				account.Addr,
				s.Precompile.Address(),
				200000,
			)

			args := []any{
				contractAddr,
				tc.msg,
				[]cmn.Coin{},
			}

			// This call will log the message
			// The issue is that msg.Msg is logged as-is, which could expose binary data
			_, err := s.Precompile.ExecuteWasm(ctx, account.Addr, contract, stateDB, &method, args)

			// Make the buf into a string for inspection
			logOutput := buf.String()

			// Logs should never surpass a reasonable length
			s.Require().Less(len(logOutput), 200, "Logs should be truncated to avoid overflow")

			// We expect these to fail because of invalid CosmWasm messages
			// But the real issue is that they get logged with raw binary
			if tc.expectError {
				s.Require().Error(err, tc.description)
			}
		})
	}
}

// TestGasHandling tests if gas limits are properly enforced
// Bug Report Issue #2: Missing Gas Handling
func (s *WasmdPrecompileTestSuite) TestGasHandling() {
	contractAddr := s.instantiateContract()
	method := s.Precompile.Methods[wasmdprecompile.ExecuteMethod]
	account := s.keyring.GetKey(0)

	testCases := []struct {
		name              string
		gasLimit          uint64
		shouldPanicOutGas bool
	}{
		{
			name:              "zero gas limit",
			gasLimit:          0,
			shouldPanicOutGas: true,
		},
		{
			name:              "very low gas limit",
			gasLimit:          1000,
			shouldPanicOutGas: true,
		},
		{
			name:     "adequate gas limit",
			gasLimit: 200000,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			stateDB := s.GetStateDB()
			contract, ctx := testutil.NewPrecompileContract(
				s.T(),
				s.Ctx,
				account.Addr,
				s.Precompile.Address(),
				tc.gasLimit,
			)

			args := []any{
				contractAddr,
				[]byte(`{"set": 42}`),
				[]cmn.Coin{},
			}

			// Set the gas limit on the context
			ctx = ctx.WithGasMeter(storetypes.NewGasMeter(tc.gasLimit))

			if tc.shouldPanicOutGas {
				require.Panics(s.T(), func() {
					// The call panics if out of gas
					_, _ = s.Precompile.ExecuteWasm(ctx, account.Addr, contract, stateDB, &method, args)
				}, "Expected out-of-gas panic for %s", tc.name)
			} else {
				// Should not panic
				_, err := s.Precompile.ExecuteWasm(ctx, account.Addr, contract, stateDB, &method, args)
				s.Require().NoError(err, "Should succeed with adequate gas for %s", tc.name)
			}
		})
	}
}

// TestReentrancy attempts to test for reentrancy vulnerabilities
// Bug Report Issue #5: Reentrancy Risk
func (s *WasmdPrecompileTestSuite) TestReentrancy() {
	contractAddr := s.instantiateContract()
	account := s.keyring.GetKey(0)

	s.Run("test_concurrent_execution_without_guard", func() {
		// This simulates what would happen if two transactions
		// tried to execute the same contract simultaneously
		contract, ctx := testutil.NewPrecompileContract(
			s.T(),
			s.Ctx,
			account.Addr,
			s.Precompile.Address(),
			800000,
		)

		args := []any{
			contractAddr,
			[]byte(`{"set": 100}`),
			[]cmn.Coin{},
		}

		// First execution starts
		contract.Input = s.PrepareInputData(wasmdprecompile.ExecuteMethod, args)
		_, err1 := s.Precompile.Run(s.NewVMInstance(ctx), contract, false)
		s.Require().NoError(err1)

		// Second execution on same contract (simulating reentrancy)
		// In a real attack, this would be triggered from within the first execution
		_, err2 := s.Precompile.Run(s.NewVMInstance(ctx), contract, false)

		// If there's no reentrancy guard, this should succeed
		// If there IS a guard, this should fail with a reentrancy error
		if err2 == nil {
			s.T().Log("⚠️  VULNERABILITY: No reentrancy protection detected!")
			s.T().Log("    Second execution succeeded without any guard")
			s.T().Log("    A malicious contract could exploit this")
		} else if err2.Error() == reentrancyCallError || err2.Error() == "reentrancy detected" {
			s.T().Log("✓ PROTECTED: Reentrancy guard is working")
		}
	})

	s.Run("test_nested_execution_attack_scenario", func() {
		// Restart the env
		s.SetupTest()
		contractAddr := s.instantiateContract()

		// Simulate a realistic attack scenario:
		// 1. Attacker calls Execute on Contract A
		// 2. Contract A calls back to Execute on Contract B
		// 3. Contract B tries to call Execute on Contract A again
		// This should be blocked by a reentrancy guard

		s.T().Log("ATTACK SCENARIO: Nested Contract Execution")
		s.T().Log("═══════════════════════════════════════════")

		contract, ctx := testutil.NewPrecompileContract(
			s.T(),
			s.Ctx,
			account.Addr,
			s.Precompile.Address(),
			800000, // More gas for nested calls
		)

		// Track execution depth
		executionDepth := 0
		maxDepth := 5

		args := []any{
			contractAddr,
			[]byte(`{"set": 42}`),
			[]cmn.Coin{},
		}
		contract.Input = s.PrepareInputData(wasmdprecompile.ExecuteMethod, args)

		s.T().Logf("Attempting %d nested executions...", maxDepth)

		// Attempt nested executions
		for i := 0; i < maxDepth; i++ {
			_, err := s.Precompile.Run(s.NewVMInstance(ctx), contract, false)
			if err != nil {
				s.T().Logf("  Depth %d: BLOCKED - %v", i, err)
				if err.Error() == reentrancyCallError {
					s.T().Log("✓ Reentrancy guard successfully prevented nested execution")
					return
				}
				break
			}
			executionDepth++
			s.T().Logf("  Depth %d: ALLOWED ⚠️", i)
		}

		if executionDepth == maxDepth {
			s.T().Log("")
			s.T().Log("❌ CRITICAL: No reentrancy guard detected!")
			s.T().Logf("   All %d nested executions were allowed", maxDepth)
			s.T().Log("")
			s.T().Log("EXPLOITATION SCENARIO:")
			s.T().Log("  1. Attacker deploys malicious contract with callback")
			s.T().Log("  2. Contract calls Execute repeatedly before state settles")
			s.T().Log("  3. Could drain funds or corrupt state")
			s.T().Log("")
			s.T().Log("IMPACT: HIGH - Cross-contract call exploitation possible")
			require.Fail(s.T(), "Reentrancy vulnerability exists")
		}
	})

	s.Run("test_state_consistency_under_reentrancy", func() {
		// Restart the env
		s.SetupTest()
		contractAddr := s.instantiateContract()

		// Test if state remains consistent during potential reentrancy
		contract, ctx := testutil.NewPrecompileContract(
			s.T(),
			s.Ctx,
			account.Addr,
			s.Precompile.Address(),
			800000,
		)

		// Set initial value
		args1 := []any{
			contractAddr,
			[]byte(`{"set": 50}`),
			[]cmn.Coin{},
		}

		// Prepare the contract input data
		contract.Input = s.PrepareInputData(wasmdprecompile.ExecuteMethod, args1)
		_, err := s.Precompile.Run(s.NewVMInstance(ctx), contract, false)
		s.Require().NoError(err)

		// Try to execute again before state is committed (simulating reentrancy)
		args2 := []any{
			contractAddr,
			[]byte(`{"set": 75}`),
			[]cmn.Coin{},
		}
		contract.Input = s.PrepareInputData(wasmdprecompile.ExecuteMethod, args2)
		_, err = s.Precompile.Run(s.NewVMInstance(ctx), contract, false)

		if err == nil {
			s.T().Log("⚠️  State modification allowed during potential reentrancy")
			s.T().Log("    This could lead to state inconsistencies")
		}
	})
}

// TestKeeperReference tests that the correct keeper is referenced
// Bug Report Issue #1: Typo in Keeper Reference (ALREADY FIXED)
func (s *WasmdPrecompileTestSuite) TestKeeperReference() {
	// This test verifies that the correct keeper (wasmdKeeper) is used
	// The typo (pasmdKeeper) has been fixed in the current codebase

	method := s.Precompile.Methods[wasmdprecompile.ExecuteMethod]
	account := s.keyring.GetKey(0)
	contractAddr := s.instantiateContract()

	s.Run("keeper should be accessible and functional", func() {
		stateDB := s.GetStateDB()
		contract, ctx := testutil.NewPrecompileContract(
			s.T(),
			s.Ctx,
			account.Addr,
			s.Precompile.Address(),
			200000,
		)

		args := []any{
			contractAddr,
			[]byte(`{"set": 100}`),
			[]cmn.Coin{},
		}

		_, err := s.Precompile.ExecuteWasm(ctx, account.Addr, contract, stateDB, &method, args)
		s.Require().NoError(err, "Keeper should work correctly without typo")
	})
}
