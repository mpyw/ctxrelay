// Package goroutinederiveand contains test fixtures for the goroutine-derive AND mode.
// This file covers adversarial patterns with AND (plus) mode.
// Test flag: -goroutine-deriver=github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine+github.com/newrelic/go-agent/v3/newrelic.NewContext
package goroutinederiveand

import (
	"context"

	"github.com/newrelic/go-agent/v3/newrelic"
)

// ===== SHOULD NOT REPORT =====

// AND - nested 2-level, both have both derivers
// Nested goroutines both call both derivers
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

// AND - both derivers in different order across conditional branches
// Both derivers called in different order per branch
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

// AND - nested 2-level, inner missing one deriver
// Inner goroutine missing one deriver
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

// AND - both derivers in nested IIFE (not at outer level)
// LIMITATION: Derivers in nested IIFE not counted for outer goroutine
func a43AndBothDeriverInNestedIIFE(ctx context.Context, txn *newrelic.Transaction) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
		func() {
			txn = txn.NewGoroutine()
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
	}()
}

// AND - split derivers across levels (outer has first, IIFE has second)
// LIMITATION: Split derivers across levels not counted
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

// AND - nested 3-level, outer only has first deriver
// Nested 3-level with partial derivers at each level
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

// Higher-order go fn()() - returned func only has first deriver
// Higher-order with returned func having partial derivers
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

// Variable reassignment - last assignment with incomplete derivers should warn
// Reassigned variable with incomplete derivers
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

// Variable reassignment - last assignment with both derivers should pass
// Reassigned variable with complete derivers
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
