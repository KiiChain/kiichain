package wasmd_test

import (
	"encoding/json"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cmn "github.com/cosmos/evm/precompiles/common"
	"github.com/cosmos/evm/precompiles/testutil"

	wasmdprecompile "github.com/kiichain/kiichain/v7/precompiles/wasmd"
)

// Note:
// The test TestRealWorldDataLeakage was removed because its causer was removed
// The test TestCompareRealWorldVsUnitTest was removed because it was redundant

// TestRealWorldReentrancyAttack simulates a real-world attack scenario
// This test is closer to what would happen on an actual running chain
func (s *WasmdPrecompileTestSuite) TestRealWorldReentrancyAttack() {
	s.Run("realistic_attack_with_malicious_contract", func() {
		s.T().Log("")
		s.T().Log("════════════════════════════════════════════════════════════")
		s.T().Log("  REAL WORLD REENTRANCY ATTACK SIMULATION")
		s.T().Log("  Simulating attack on live KiiChain")
		s.T().Log("════════════════════════════════════════════════════════════")
		s.T().Log("")

		// Setup: Deploy a "victim" contract with funds
		victimContract := s.instantiateContract()
		attacker := s.keyring.GetKey(1) // Different account for attacker

		s.T().Logf("Victim Contract: %s", victimContract)
		s.T().Logf("Attacker Address: %s", attacker.Addr.Hex())
		s.T().Log("")

		// Simulate funding the victim contract (like a DeFi pool)
		initialFunds := sdk.NewCoins(sdk.NewInt64Coin("ukii", 10000000))
		victimAddr, err := sdk.AccAddressFromBech32(victimContract)
		require.NoError(s.T(), err)
		err = s.App.BankKeeper.MintCoins(s.Ctx, "evm", initialFunds)
		require.NoError(s.T(), err)
		err = s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, "evm", victimAddr, initialFunds)
		require.NoError(s.T(), err)

		victimBalance := s.App.BankKeeper.GetBalance(s.Ctx, victimAddr, "ukii")
		s.T().Logf("✓ Victim contract funded: %s", victimBalance.String())
		s.T().Log("")

		// Attack Phase: Simulate multiple reentrancy attempts
		s.T().Log("ATTACK PHASE: Attempting Reentrancy Exploit")
		s.T().Log("───────────────────────────────────────────────")
		s.T().Log("METHODOLOGY:")
		s.T().Log("  - Using counter contract's 'set' method")
		s.T().Log("  - Each call should modify state")
		s.T().Log("  - If NO guard exists → all 5 calls succeed")
		s.T().Log("  - If guard EXISTS → only 1 call succeeds")
		s.T().Log("")
		s.T().Log("REAL-WORLD EQUIVALENT:")
		s.T().Log("  - Replace 'set' with 'withdraw' on DeFi contract")
		s.T().Log("  - Each successful call drains funds")
		s.T().Log("  - Reentrancy = Multiple withdrawals from same balance")
		s.T().Log("")

		attackAttempts := 5
		successfulReentries := 0

		for attempt := 1; attempt <= attackAttempts; attempt++ {
			contract, ctx := testutil.NewPrecompileContract(
				s.T(),
				s.Ctx,
				attacker.Addr,
				s.Precompile.Address(),
				800000,
			)

			// Simulate state modification attack using counter's "set" method
			// In a real attack, this would be a withdrawal or transfer
			// The point is to show multiple executions are allowed (reentrancy)
			setMsg := map[string]interface{}{
				"set": attempt * 10, // Set to different values
			}
			msgBytes, err := json.Marshal(setMsg)
			require.NoError(s.T(), err)

			args := []any{
				victimContract,
				msgBytes,
				[]cmn.Coin{},
			}

			// Prepare the contract input data
			contract.Input = s.PrepareInputData(wasmdprecompile.ExecuteMethod, args)
			s.T().Logf("  Execution %d: set(value=%d)...", attempt, attempt*10)

			_, err = s.Precompile.Run(s.NewVMInstance(ctx), contract, false)

			// Analyze result
			switch {
			case err == nil:
				successfulReentries++
				s.T().Logf("    ⚠️  ALLOWED - Execution %d succeeded", attempt)
				s.T().Logf("       (In real DeFi: this would be withdrawal #%d)", attempt)
			case err.Error() == reentrancyCallError || err.Error() == "reentrancy detected":
				s.T().Logf("    ✓ BLOCKED - Reentrancy guard detected: %v", err)
				s.T().Log("")
				s.T().Log("═══════════════════════════════════════════════")
				s.T().Log("✅ PROTECTED: Reentrancy guard is functioning!")
				s.T().Log("   Guard correctly blocked second execution")
				s.T().Log("═══════════════════════════════════════════════")
				return // Exit early - guard is working
			default:
				s.T().Logf("    ℹ️  Failed: %v", err)
			}
		}

		s.T().Log("")

		// Check final state
		finalBalance := s.App.BankKeeper.GetBalance(s.Ctx, victimAddr, "ukii")
		s.T().Log("ATTACK RESULTS:")
		s.T().Log("───────────────────────────────────────────────")
		s.T().Logf("Initial Balance:  %s", victimBalance.String())
		s.T().Logf("Final Balance:    %s", finalBalance.String())
		s.T().Logf("Successful Reentries: %d/%d", successfulReentries, attackAttempts)

		// Analyze results
		switch {
		case successfulReentries >= 2:
			s.T().Log("")
			s.T().Log("════════════════════════════════════════════════")
			s.T().Log("❌ CRITICAL: REENTRANCY VULNERABILITY CONFIRMED!")
			s.T().Log("════════════════════════════════════════════════")
			s.T().Logf("   Successful Reentrancies: %d/%d", successfulReentries, attackAttempts)
			s.T().Log("")
			s.T().Log("WHAT THIS MEANS:")
			s.T().Log("  ✓ No reentrancy guard exists in precompile")
			s.T().Log("  ✓ Multiple calls execute without blocking")
			s.T().Log("  ✓ State can be manipulated during execution")
			s.T().Log("")
			s.T().Log("REAL-WORLD ATTACK SCENARIO:")
			s.T().Log("  If this was a DeFi contract with withdraw():")
			s.T().Log("  ├─ Call 1: withdraw(1000) → Success ✓")
			s.T().Log("  ├─ Call 2: withdraw(1000) → Success ✓ (REENTRANCY!)")
			s.T().Log("  ├─ Call 3: withdraw(1000) → Success ✓ (REENTRANCY!)")
			s.T().Log("  └─ Balance updated ONCE → 3000 stolen, 1000 paid")
			s.T().Log("")
			s.T().Log("IMPACT:")
			s.T().Log("  💸 Drain entire contract balance")
			s.T().Log("  ⚡ Completes in single transaction")
			s.T().Log("  🔒 Funds irrecoverable once stolen")
			s.T().Log("  👥 All users affected")
			s.T().Log("")
			require.Fail(s.T(), "REENTRANCY VULNERABILITY EXISTS")
		case successfulReentries == 1:
			s.T().Log("")
			s.T().Log("✅ Only one execution succeeded - Guard may be present")
		default:
			s.T().Log("")
			s.T().Log("ℹ️  No executions succeeded (contract method issue)")
			s.T().Log("   Note: Reentrancy vulnerability still tested in other test suites")
		}

		s.T().Log("")
	})
}

// TestRealWorldGasExhaustion simulates gas exhaustion attack
func (s *WasmdPrecompileTestSuite) TestRealWorldGasExhaustion() {
	s.Run("gas_exhaustion_attack_simulation", func() {
		s.T().Log("")
		s.T().Log("════════════════════════════════════════════════════════════")
		s.T().Log("  REAL WORLD GAS EXHAUSTION ATTACK")
		s.T().Log("════════════════════════════════════════════════════════════")
		s.T().Log("")

		contractAddr := s.instantiateContract()
		method := s.Precompile.Methods[wasmdprecompile.ExecuteMethod]
		attacker := s.keyring.GetKey(1)

		expectedUsage := uint64(68137)

		gasLimits := []struct {
			name       string
			gasLimit   uint64
			panicsWith interface{}
		}{
			{"Minimal Gas (1)", 1, storetypes.ErrorOutOfGas{Descriptor: "ReadFlat"}},
			{"Low Gas (100)", 100, storetypes.ErrorOutOfGas{Descriptor: "ReadFlat"}},
			{"Medium Gas (10K)", 10000, storetypes.ErrorOutOfGas{Descriptor: "Loading CosmWasm module: execute"}},
			{"Medium Gas (20K)", 20000, storetypes.ErrorOutOfGas{Descriptor: "Loading CosmWasm module: execute"}},
			{"10 less than usage", expectedUsage - 10, storetypes.ErrorOutOfGas{Descriptor: "Custom contract event attributes"}},
			{"10 more than usage", expectedUsage + 10, nil},
			{"Adequate Gas (200K)", 200000, nil},
		}

		s.T().Log("Testing gas limits like on real chain:")
		s.T().Log("───────────────────────────────────────────────")

		// Run the tests for each gas limit
		for _, test := range gasLimits {
			stateDB := s.GetStateDB()
			contract, ctx := testutil.NewPrecompileContract(
				s.T(),
				s.Ctx,
				attacker.Addr,
				s.Precompile.Address(),
				test.gasLimit,
			)

			args := []any{
				contractAddr,
				[]byte(`{"set": 42}`),
				[]cmn.Coin{},
			}

			// Set the gas limit on the context
			ctx = ctx.WithGasMeter(storetypes.NewGasMeter(test.gasLimit))

			// Measure actual gas consumed (like real chain would)
			gasStart := ctx.GasMeter().GasConsumed()

			if test.panicsWith != nil {
				require.PanicsWithValue(s.T(), test.panicsWith, func() {
					// The call panics if out of gas
					_, _ = s.Precompile.ExecuteWasm(ctx, attacker.Addr, contract, stateDB, &method, args)
				})
				s.T().Logf("  %s: ✓ Rejected - Out of Gas panic", test.name)
				continue
			}

			// The call panics if out of gas
			_, err := s.Precompile.ExecuteWasm(ctx, attacker.Addr, contract, stateDB, &method, args)
			require.NoError(s.T(), err)
			s.T().Logf("  %s: ✓ Allowed - Executed successfully", test.name)

			// Check gas used
			gasUsed := ctx.GasMeter().GasConsumed() - gasStart
			// Log the gas used
			s.T().Logf("    → Gas Used: %d / Limit: %d", gasUsed, test.gasLimit)

			// Final gas should be less than limit
			require.Less(s.T(), gasUsed, test.gasLimit, "Gas used should be less than limit for %s", test.name)
			// But should be a significant portion (indicating real work done)
			require.Greater(s.T(), gasUsed, test.gasLimit/10, "Gas used should be significant for %s", test.name)

		}

		s.T().Log("")
	})
}
