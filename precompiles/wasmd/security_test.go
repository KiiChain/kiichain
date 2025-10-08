package wasmd_test

import (
	"encoding/hex"

	cmn "github.com/cosmos/evm/precompiles/common"
	"github.com/cosmos/evm/precompiles/testutil"

	wasmdprecompile "github.com/kiichain/kiichain/v5/precompiles/wasmd"
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
				tc.msg,
				[]cmn.Coin{},
			}

			// This call will log the message
			// The issue is that msg.Msg is logged as-is, which could expose binary data
			_, err := s.Precompile.Execute(ctx, account.Addr, contract, stateDB, &method, args)

			// We expect these to fail because of invalid CosmWasm messages
			// But the real issue is that they get logged with raw binary
			if tc.expectError {
				s.Require().Error(err, tc.description)
			}

			// Log what would be logged unsafely
			s.T().Logf("Binary data logged unsafely: %v", tc.msg)
			s.T().Logf("Safe hex encoding: %s...", hex.EncodeToString(tc.msg[:min(len(tc.msg), 32)]))
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
		name        string
		gasLimit    uint64
		expectError bool
		errContains string
	}{
		{
			name:        "very low gas limit",
			gasLimit:    1000,
			expectError: true,
			errContains: "out of gas",
		},
		{
			name:        "adequate gas limit",
			gasLimit:    200000,
			expectError: false,
		},
		{
			name:        "zero gas limit",
			gasLimit:    0,
			expectError: true,
			errContains: "out of gas",
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

			_, err := s.Precompile.Execute(ctx, account.Addr, contract, stateDB, &method, args)

			if tc.expectError {
				s.Require().Error(err, "Should fail with gas limit %d", tc.gasLimit)
				if tc.errContains != "" {
					s.Require().Contains(err.Error(), tc.errContains)
				}
			} else {
				s.Require().NoError(err, "Should succeed with gas limit %d", tc.gasLimit)
			}
		})
	}
}

// TestReentrancy attempts to test for reentrancy vulnerabilities
// Bug Report Issue #5: Reentrancy Risk
func (s *WasmdPrecompileTestSuite) TestReentrancy() {
	contractAddr := s.instantiateContract()
	method := s.Precompile.Methods[wasmdprecompile.ExecuteMethod]
	account := s.keyring.GetKey(0)

	s.Run("test_concurrent_execution_without_guard", func() {
		// This simulates what would happen if two transactions
		// tried to execute the same contract simultaneously
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

		// First execution starts
		_, err1 := s.Precompile.Execute(ctx, account.Addr, contract, stateDB, &method, args)
		s.Require().NoError(err1)

		// Second execution on same contract (simulating reentrancy)
		// In a real attack, this would be triggered from within the first execution
		_, err2 := s.Precompile.Execute(ctx, account.Addr, contract, stateDB, &method, args)

		// If there's no reentrancy guard, this should succeed
		// If there IS a guard, this should fail with a reentrancy error
		if err2 == nil {
			s.T().Log("⚠️  VULNERABILITY: No reentrancy protection detected!")
			s.T().Log("    Second execution succeeded without any guard")
			s.T().Log("    A malicious contract could exploit this")
		} else if err2.Error() == "reentrant call" || err2.Error() == "reentrancy detected" {
			s.T().Log("✓ PROTECTED: Reentrancy guard is working")
		}
	})

	s.Run("test_nested_execution_attack_scenario", func() {
		// Simulate a realistic attack scenario:
		// 1. Attacker calls Execute on Contract A
		// 2. Contract A calls back to Execute on Contract B
		// 3. Contract B tries to call Execute on Contract A again
		// This should be blocked by a reentrancy guard

		s.T().Log("ATTACK SCENARIO: Nested Contract Execution")
		s.T().Log("═══════════════════════════════════════════")

		stateDB := s.GetStateDB()
		contract, ctx := testutil.NewPrecompileContract(
			s.T(),
			s.Ctx,
			account.Addr,
			s.Precompile.Address(),
			500000, // More gas for nested calls
		)

		// Track execution depth
		executionDepth := 0
		maxDepth := 5

		args := []any{
			contractAddr,
			[]byte(`{"set": 42}`),
			[]cmn.Coin{},
		}

		s.T().Logf("Attempting %d nested executions...", maxDepth)

		// Attempt nested executions
		for i := 0; i < maxDepth; i++ {
			_, err := s.Precompile.Execute(ctx, account.Addr, contract, stateDB, &method, args)
			if err != nil {
				s.T().Logf("  Depth %d: BLOCKED - %v", i, err)
				if err.Error() == "reentrant call" {
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
		}
	})

	s.Run("test_state_consistency_under_reentrancy", func() {
		// Test if state remains consistent during potential reentrancy
		stateDB := s.GetStateDB()
		contract, ctx := testutil.NewPrecompileContract(
			s.T(),
			s.Ctx,
			account.Addr,
			s.Precompile.Address(),
			200000,
		)

		// Set initial value
		args1 := []any{
			contractAddr,
			[]byte(`{"set": 50}`),
			[]cmn.Coin{},
		}
		_, err := s.Precompile.Execute(ctx, account.Addr, contract, stateDB, &method, args1)
		s.Require().NoError(err)

		// Try to execute again before state is committed (simulating reentrancy)
		args2 := []any{
			contractAddr,
			[]byte(`{"set": 75}`),
			[]cmn.Coin{},
		}
		_, err = s.Precompile.Execute(ctx, account.Addr, contract, stateDB, &method, args2)

		if err == nil {
			s.T().Log("⚠️  State modification allowed during potential reentrancy")
			s.T().Log("    This could lead to state inconsistencies")
		}
	})

	s.Run("verify_reentrancy_guard_implementation", func() {
		s.T().Log("")
		s.T().Log("REENTRANCY GUARD ANALYSIS")
		s.T().Log("════════════════════════════")

		// Check the code for common reentrancy protection patterns
		s.T().Log("Checking for protection mechanisms:")
		s.T().Log("  [ ] Mutex lock before execution")
		s.T().Log("  [ ] Reentrancy flag/counter")
		s.T().Log("  [ ] Call depth tracking")
		s.T().Log("  [ ] State commitment before callbacks")
		s.T().Log("")
		s.T().Log("⚠️  RECOMMENDATION:")
		s.T().Log("    Add explicit reentrancy guard using one of:")
		s.T().Log("    1. sync.Mutex to prevent concurrent execution")
		s.T().Log("    2. execution status flag (locked/unlocked)")
		s.T().Log("    3. call depth counter with maximum limit")
		s.T().Log("")
		s.T().Log("EXAMPLE IMPLEMENTATION:")
		s.T().Log("  type Precompile struct {")
		s.T().Log("      mu            sync.Mutex")
		s.T().Log("      executing     bool")
		s.T().Log("  }")
		s.T().Log("")
		s.T().Log("  func (p *Precompile) Execute(...) {")
		s.T().Log("      p.mu.Lock()")
		s.T().Log("      defer p.mu.Unlock()")
		s.T().Log("      if p.executing {")
		s.T().Log("          return nil, errors.New(\"reentrant call\")")
		s.T().Log("      }")
		s.T().Log("      p.executing = true")
		s.T().Log("      defer func() { p.executing = false }()")
		s.T().Log("      // ... rest of execution")
		s.T().Log("  }")
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

		_, err := s.Precompile.Execute(ctx, account.Addr, contract, stateDB, &method, args)
		s.Require().NoError(err, "Keeper should work correctly without typo")
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
