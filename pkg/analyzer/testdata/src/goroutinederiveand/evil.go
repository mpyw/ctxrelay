package goroutinederiveand

import (
	"context"

	"github.com/newrelic/go-agent/v3/newrelic"
)

// =============================================================================
// EVIL: AND (plus) - adversarial patterns
// Test flag: -goroutine-deriver=github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine+github.com/newrelic/go-agent/v3/newrelic.NewContext
// =============================================================================

// ===== SHOULD NOT REPORT =====

// DA40: AND - nested 2-level, both have both derivers.
func a40AndNested2LevelBothHaveBothDerivers(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		go func() {
			txn = txn.NewGoroutine()
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
		_ = ctx
	}()
}

// DA41: AND - both derivers in different order across conditional branches.
func a41AndDifferentOrderInBranches(ctx context.Context, txn *newrelic.Transaction, cond bool) {
	go func() {
		if cond {
			txn = txn.NewGoroutine()
			ctx = newrelic.NewContext(ctx, txn)
		} else {
			ctx = newrelic.NewContext(ctx, txn)
			txn = txn.NewGoroutine()
		}
		_ = ctx
		_ = txn
	}()
}

// ===== SHOULD REPORT =====

// DA42: AND - nested 2-level, inner missing one deriver.
func a42AndNested2LevelInnerMissingOneDeriver(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
			ctx = newrelic.NewContext(ctx, txn) // Missing NewGoroutine
			_ = ctx
		}()
		_ = ctx
	}()
}

// DA43: AND - both derivers in nested IIFE (not at outer level).
func a43AndBothDeriverInNestedIIFE(ctx context.Context, txn *newrelic.Transaction) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
		func() {
			txn = txn.NewGoroutine()
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
	}()
}

// DA44: AND - split derivers across levels (outer has first, IIFE has second).
func a44AndSplitDeriversAcrossLevels(ctx context.Context, txn *newrelic.Transaction) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
		txn = txn.NewGoroutine() // First deriver at outer level
		func() {
			ctx = newrelic.NewContext(ctx, txn) // Second deriver in IIFE - not counted for outer
			_ = ctx
		}()
		_ = txn
	}()
}

// DA45: AND - nested 3-level, outer only has first deriver.
func a45AndNested3LevelOuterPartial(ctx context.Context, txn *newrelic.Transaction) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
		txn = txn.NewGoroutine() // Only first deriver
		go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
			ctx = newrelic.NewContext(ctx, txn) // Only second deriver
			go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
				_ = ctx // Neither deriver
			}()
			_ = ctx
		}()
		_ = txn
	}()
}

// ===== HIGHER-ORDER PATTERNS =====

// DA46: Higher-order go fn()() - returned func only has first deriver.
func a46HigherOrderReturnedFuncPartialDeriver(ctx context.Context, txn *newrelic.Transaction) {
	makeWorker := func() func() {
		return func() {
			txn = txn.NewGoroutine() // Only first deriver, missing NewContext
			_ = ctx
			_ = txn
		}
	}
	go makeWorker()() // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
}

// ===== VARIABLE REASSIGNMENT =====

// DA47: Variable reassignment - last assignment with incomplete derivers should warn.
func a47ReassignedFuncIncompleteDeriver(ctx context.Context, txn *newrelic.Transaction) {
	fn := func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}
	fn = func() {
		txn = txn.NewGoroutine() // Only first deriver
		_ = ctx
		_ = txn
	}
	go fn() // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
}

// DA48: Variable reassignment - last assignment with both derivers should pass.
func a48ReassignedFuncBothDerivers(ctx context.Context, txn *newrelic.Transaction) {
	fn := func() {
		_ = ctx // First assignment has no deriver
	}
	fn = func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}
	go fn() // OK - last assignment has both derivers
}
