// Package goroutinederivemixed contains test fixtures for the goroutine-derive mixed AND/OR mode.
// This file covers adversarial patterns with mixed (A+B),C mode.
// Test flag: -goroutine-deriver=github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext
package goroutinederivemixed

import (
	"context"

	"github.com/my-example-app/telemetry/apm"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// ===== SHOULD NOT REPORT =====

// Mixed - nested 2-level, outer satisfies AND group, inner satisfies OR alternative
// Nested with different approaches at each level
func m40MixedNested2LevelDifferentApproaches(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		go func() {
			ctx = apm.NewGoroutineContext(ctx) // Inner satisfies OR alternative
			_ = ctx
		}()
		_ = ctx
	}()
}

// Mixed - nested 2-level, outer satisfies OR alternative, inner satisfies AND group
// Nested with reversed approaches at each level
func m41MixedNested2LevelReversedApproaches(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		ctx = apm.NewGoroutineContext(ctx)
		go func() {
			txn = txn.NewGoroutine()
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
		_ = ctx
	}()
}

// ===== SHOULD REPORT =====

// Mixed - nested 2-level, inner satisfies neither
// Inner goroutine satisfies neither AND nor OR
func m42MixedNested2LevelInnerSatisfiesNeither(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		ctx = apm.NewGoroutineContext(ctx)
		go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
			ctx = newrelic.NewContext(ctx, txn) // Only second of AND, not OR alt
			_ = ctx
		}()
		_ = ctx
	}()
}

// Mixed - AND group split between outer and IIFE
// LIMITATION: Split derivers across levels not counted
func m43MixedSplitDeriversAcrossLevels(ctx context.Context, txn *newrelic.Transaction) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		txn = txn.NewGoroutine() // Only first of AND
		func() {
			ctx = newrelic.NewContext(ctx, txn) // Second of AND in IIFE - doesn't count
			_ = ctx
		}()
		_ = txn
	}()
}

// Mixed - OR alternative only in nested IIFE
// LIMITATION: OR alternative in nested IIFE not counted
func m44MixedOrAlternativeInNestedIIFE(ctx context.Context) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		func() {
			ctx = apm.NewGoroutineContext(ctx)
			_ = ctx
		}()
	}()
}

// Mixed - nested 3-level, outer only has first of AND
// Nested 3-level with partial derivers at each level
func m45MixedNested3LevelOuterPartial(ctx context.Context, txn *newrelic.Transaction) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		txn = txn.NewGoroutine() // Only first of AND
		go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
			ctx = newrelic.NewContext(ctx, txn) // Only second of AND
			go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
				_ = ctx // Neither AND nor OR
			}()
			_ = ctx
		}()
		_ = txn
	}()
}

// ===== HIGHER-ORDER PATTERNS =====

// Higher-order go fn()() - returned func only has first of AND, not OR alternative
// Higher-order with returned func having partial derivers
func m46HigherOrderReturnedFuncPartialDeriver(ctx context.Context, txn *newrelic.Transaction) {
	makeWorker := func() func() {
		return func() {
			txn = txn.NewGoroutine() // Only first of AND, not OR alt
			_ = ctx
			_ = txn
		}
	}
	go makeWorker()() // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
}

// ===== VARIABLE REASSIGNMENT =====

// Variable reassignment - last assignment with incomplete derivers should warn
// Reassigned variable with incomplete derivers
func m47ReassignedFuncIncompleteDeriver(ctx context.Context, txn *newrelic.Transaction) {
	fn := func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}
	fn = func() {
		txn = txn.NewGoroutine() // Only first of AND, not OR alt
		_ = ctx
		_ = txn
	}
	go fn() // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
}

// Variable reassignment - last assignment satisfies OR alternative should pass
// Reassigned variable satisfying OR alternative
func m48ReassignedFuncOrAlternative(ctx context.Context) {
	fn := func() {
		_ = ctx // First assignment has no deriver
	}
	fn = func() {
		ctx = apm.NewGoroutineContext(ctx) // OR alternative
		_ = ctx
	}
	go fn() // OK - last assignment satisfies OR alternative
}
