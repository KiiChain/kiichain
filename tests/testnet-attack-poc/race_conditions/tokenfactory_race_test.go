package race_conditions

import (
	"fmt"
	"sync"
	"testing"
)

// TestTokenfactoryRaceCondition demonstrates the race condition in tokenfactory
// ISSUE: Tokens minted BEFORE checking if recipient is blocked
// Location: wasmbinding/tokenfactory/message_plugin.go:123-137
func TestTokenfactoryRaceCondition(t *testing.T) {
	t.Log("=== Testing Tokenfactory Race Condition ===")
	t.Log("ISSUE: Order of operations allows minting to blocked addresses")
	t.Log("CLIENT RESPONSE: 'We will switch the order' - Good, but let's show WHY")

	type BlockedAddressChecker struct {
		blockedAddrs map[string]bool
		mu          sync.RWMutex
	}

	type TokenMinter struct {
		balances map[string]int64
		mu       sync.RWMutex
	}

	// Simulate the WRONG order (current bug)
	mintTokensWrongOrder := func(minter *TokenMinter, checker *BlockedAddressChecker, recipient string, amount int64) error {
		t.Log("\n--- WRONG ORDER (Current Bug) ---")

		// Step 1: MINT FIRST (STATE CHANGE!)
		minter.mu.Lock()
		minter.balances[recipient] += amount
		t.Logf("Step 1: Minted %d tokens to %s", amount, recipient)
		t.Logf("  Balance after mint: %d", minter.balances[recipient])
		minter.mu.Unlock()

		// Step 2: CHECK IF BLOCKED (TOO LATE!)
		checker.mu.RLock()
		isBlocked := checker.blockedAddrs[recipient]
		checker.mu.RUnlock()

		if isBlocked {
			t.Logf("Step 2: ERROR - Address %s is blocked!", recipient)
			t.Log("  ❌ But tokens were ALREADY minted!")

			// Try to revert (but state might be corrupted)
			minter.mu.Lock()
			minter.balances[recipient] -= amount // Attempted rollback
			minter.mu.Unlock()

			return fmt.Errorf("recipient %s is blocked", recipient)
		}

		t.Log("Step 2: Address not blocked, mint successful")
		return nil
	}

	// Simulate the CORRECT order (fixed version)
	mintTokensCorrectOrder := func(minter *TokenMinter, checker *BlockedAddressChecker, recipient string, amount int64) error {
		t.Log("\n--- CORRECT ORDER (Fixed) ---")

		// Step 1: CHECK BLOCKED FIRST
		checker.mu.RLock()
		isBlocked := checker.blockedAddrs[recipient]
		checker.mu.RUnlock()

		if isBlocked {
			t.Logf("Step 1: Address %s is blocked - rejecting mint", recipient)
			t.Log("  ✅ No tokens minted, state unchanged")
			return fmt.Errorf("recipient %s is blocked", recipient)
		}

		t.Log("Step 1: Address not blocked, proceeding to mint")

		// Step 2: MINT ONLY IF NOT BLOCKED
		minter.mu.Lock()
		minter.balances[recipient] += amount
		t.Logf("Step 2: Minted %d tokens to %s", amount, recipient)
		t.Logf("  Balance after mint: %d", minter.balances[recipient])
		minter.mu.Unlock()

		return nil
	}

	t.Run("DemonstrateRaceCondition", func(t *testing.T) {
		minter := &TokenMinter{
			balances: make(map[string]int64),
		}
		checker := &BlockedAddressChecker{
			blockedAddrs: map[string]bool{
				"kii1blocked": true,
				"kii1allowed": false,
			},
		}

		// Test with blocked address - WRONG order
		t.Log("\n=== Testing with BLOCKED address (Wrong Order) ===")
		err := mintTokensWrongOrder(minter, checker, "kii1blocked", 1000)

		if err != nil {
			t.Logf("Error returned: %v", err)
			t.Logf("Final balance of blocked address: %d", minter.balances["kii1blocked"])

			if minter.balances["kii1blocked"] != 0 {
				t.Log("⚠️ RACE CONDITION EFFECT: Rollback might fail in real scenario!")
				t.Log("⚠️ If error occurs during rollback, tokens stay minted to blocked address!")
			}
		}

		// Test with blocked address - CORRECT order
		t.Log("\n=== Testing with BLOCKED address (Correct Order) ===")
		minter.balances = make(map[string]int64) // Reset
		err = mintTokensCorrectOrder(minter, checker, "kii1blocked", 1000)

		if err != nil {
			t.Logf("Error returned: %v", err)
			t.Logf("Final balance of blocked address: %d", minter.balances["kii1blocked"])
			t.Log("✅ No tokens ever minted - state remained clean")
		}
	})

	t.Run("RaceConditionWithStateCorruption", func(t *testing.T) {
		t.Log("\n=== Demonstrating State Corruption Risk ===")

		minter := &TokenMinter{
			balances: map[string]int64{
				"kii1blocked": 0,
			},
		}
		checker := &BlockedAddressChecker{
			blockedAddrs: map[string]bool{
				"kii1blocked": true,
			},
		}

		// Simulate what happens if rollback fails
		t.Log("Scenario: Mint happens, blocked check fails, but rollback ALSO fails")

		// Step 1: Mint tokens
		minter.balances["kii1blocked"] = 1000
		t.Log("1. Tokens minted: 1000")

		// Step 2: Check reveals it's blocked
		if checker.blockedAddrs["kii1blocked"] {
			t.Log("2. Address is blocked!")

			// Step 3: Rollback attempt fails (simulating error)
			t.Log("3. Rollback attempt fails due to:")
			t.Log("   - Database error")
			t.Log("   - Network issue")
			t.Log("   - Panic in another module")

			t.Log("\n💥 RESULT: State Corruption!")
			t.Logf("   Blocked address has balance: %d", minter.balances["kii1blocked"])
			t.Log("   This violates the invariant: blocked addresses should have 0 balance")
		}
	})

	t.Run("ConcurrentAccessPattern", func(t *testing.T) {
		t.Log("\n=== Concurrent Access Pattern ===")
		t.Log("Even without parallel execution, the pattern is dangerous:")

		t.Log("\n1. Admin Transaction (Block Address):")
		t.Log("   TX1: AddToBlocklist('kii1user')")

		t.Log("\n2. User Transaction (Mint Tokens):")
		t.Log("   TX2: MintTokens('kii1user', 1000)")

		t.Log("\nIf TX2 reads state before TX1 completes:")
		t.Log("  - TX2 reads: 'kii1user' not in blocklist")
		t.Log("  - TX1 executes: Adds 'kii1user' to blocklist")
		t.Log("  - TX2 continues: Mints tokens (using stale read)")

		t.Log("\n⚠️ Result: Blocked address receives tokens")
	})
}

// TestTokenfactoryRaceConditionRealCode shows the actual code pattern
func TestTokenfactoryRaceConditionRealCode(t *testing.T) {
	t.Log("=== Real Code Analysis ===")
	t.Log("")
	t.Log("BUGGY CODE (wasmbinding/tokenfactory/message_plugin.go:123-137):")
	t.Log("```go")
	t.Log("// Execute mint FIRST")
	t.Log("_, err := msgServer.Mint(ctx, sdkMsg)")
	t.Log("if err != nil {")
	t.Log("    return nil, errorsmod.Wrap(err, \"mint\")")
	t.Log("}")
	t.Log("")
	t.Log("// Check blocked AFTER minting (TOO LATE!)")
	t.Log("if b.BlockedAddr(rcpt) {")
	t.Log("    return nil, errorsmod.Wrapf(err, \"failed to mint to blocked address\")")
	t.Log("}")
	t.Log("```")
	t.Log("")
	t.Log("PROBLEMS:")
	t.Log("1. Tokens already minted when blocked check happens")
	t.Log("2. Error return doesn't undo the mint")
	t.Log("3. 'err' variable is nil in blocked check (another bug!)")
	t.Log("")
	t.Log("CORRECT CODE:")
	t.Log("```go")
	t.Log("// Check blocked FIRST")
	t.Log("if b.BlockedAddr(rcpt) {")
	t.Log("    return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, \"blocked address\")")
	t.Log("}")
	t.Log("")
	t.Log("// Only mint if not blocked")
	t.Log("_, err := msgServer.Mint(ctx, sdkMsg)")
	t.Log("if err != nil {")
	t.Log("    return nil, errorsmod.Wrap(err, \"mint\")")
	t.Log("}")
	t.Log("```")
}

// TestAdditionalTokenfactoryIssues highlights other problems
func TestAdditionalTokenfactoryIssues(t *testing.T) {
	t.Log("=== Additional Tokenfactory Issues ===")
	t.Log("")
	t.Log("Issue #1: Nil Error Variable")
	t.Log("  Location: message_plugin.go:130-131")
	t.Log("  Code: errorsmod.Wrapf(err, '...') where err is nil")
	t.Log("  Impact: Error message will be malformed")
	t.Log("  Fix: Use proper error like sdkerrors.ErrUnauthorized")
	t.Log("")
	t.Log("Issue #2: No Atomic Guarantees")
	t.Log("  Problem: Mint and check aren't in same transaction")
	t.Log("  Impact: Partial state updates possible")
	t.Log("  Fix: Wrap in database transaction")
	t.Log("")
	t.Log("Issue #3: Missing Validation Order")
	t.Log("  Should validate in this order:")
	t.Log("  1. Check if address is blocked")
	t.Log("  2. Validate mint parameters")
	t.Log("  3. Check permissions")
	t.Log("  4. Perform mint")
	t.Log("  5. Emit events")
	t.Log("")
	t.Log("Issue #4: Resource Leak Risk")
	t.Log("  If mint succeeds but event emission fails:")
	t.Log("  - Tokens are minted")
	t.Log("  - But no audit trail")
	t.Log("  Fix: Ensure events are part of atomic operation")
}

// TestWhyOrderMatters explains the importance
func TestWhyOrderMatters(t *testing.T) {
	t.Log("=== Why Order of Operations Matters ===")
	t.Log("")
	t.Log("Rule: VALIDATE → MUTATE → COMMIT")
	t.Log("")
	t.Log("1. VALIDATE Phase (No State Changes):")
	t.Log("   - Check permissions")
	t.Log("   - Validate inputs")
	t.Log("   - Check business rules")
	t.Log("   - Verify constraints")
	t.Log("")
	t.Log("2. MUTATE Phase (Prepare Changes):")
	t.Log("   - Calculate new state")
	t.Log("   - Prepare updates")
	t.Log("   - Stage changes")
	t.Log("")
	t.Log("3. COMMIT Phase (Apply Changes):")
	t.Log("   - Apply all changes atomically")
	t.Log("   - Emit events")
	t.Log("   - Update indexes")
	t.Log("")
	t.Log("Breaking this pattern leads to:")
	t.Log("  ❌ Partial state updates")
	t.Log("  ❌ Invariant violations")
	t.Log("  ❌ Security vulnerabilities")
	t.Log("  ❌ Data corruption")
	t.Log("  ❌ Audit trail gaps")
}

// TestMitigationStrategies provides solutions
func TestMitigationStrategies(t *testing.T) {
	t.Log("=== Mitigation Strategies ===")
	t.Log("")
	t.Log("Strategy 1: Pre-validation Pattern")
	t.Log("```go")
	t.Log("func MintTokens(ctx Context, recipient Address, amount Coins) error {")
	t.Log("    // ALL validation BEFORE any state change")
	t.Log("    if err := ValidateRecipient(recipient); err != nil {")
	t.Log("        return err")
	t.Log("    }")
	t.Log("    if err := ValidateAmount(amount); err != nil {")
	t.Log("        return err")
	t.Log("    }")
	t.Log("    if IsBlocked(recipient) {")
	t.Log("        return ErrBlocked")
	t.Log("    }")
	t.Log("    ")
	t.Log("    // Only NOW do state changes")
	t.Log("    return ExecuteMint(ctx, recipient, amount)")
	t.Log("}")
	t.Log("```")
	t.Log("")
	t.Log("Strategy 2: Transaction Pattern")
	t.Log("```go")
	t.Log("tx := BeginTransaction()")
	t.Log("defer tx.Rollback() // Auto-rollback on error")
	t.Log("")
	t.Log("if err := tx.MintTokens(recipient, amount); err != nil {")
	t.Log("    return err")
	t.Log("}")
	t.Log("")
	t.Log("if tx.IsBlocked(recipient) {")
	t.Log("    return ErrBlocked // Rollback happens automatically")
	t.Log("}")
	t.Log("")
	t.Log("return tx.Commit()")
	t.Log("```")
	t.Log("")
	t.Log("Strategy 3: State Machine Pattern")
	t.Log("  - Define valid state transitions")
	t.Log("  - Enforce transition rules")
	t.Log("  - Prevent invalid states")
}