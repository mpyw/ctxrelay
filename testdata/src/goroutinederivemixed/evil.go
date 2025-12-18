package goroutinederivemixed

import (
	"context"

	"github.com/my-example-app/telemetry/apm"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// =============================================================================
// EVIL: Mixed AND/OR - adversarial patterns
// Test flag: -goroutine-deriver=github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext
// =============================================================================

// ===== SHOULD NOT REPORT =====

// DM40: Mixed - nested 2-level, outer satisfies AND group, inner satisfies OR alternative.
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

// DM41: Mixed - nested 2-level, outer satisfies OR alternative, inner satisfies AND group.
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

// DM42: Mixed - nested 2-level, inner satisfies neither.
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

// DM43: Mixed - AND group split between outer and IIFE.
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

// DM44: Mixed - OR alternative only in nested IIFE.
func m44MixedOrAlternativeInNestedIIFE(ctx context.Context) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		func() {
			ctx = apm.NewGoroutineContext(ctx)
			_ = ctx
		}()
	}()
}

// DM45: Mixed - nested 3-level, outer only has first of AND.
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

// DM46: Higher-order go fn()() - returned func only has first of AND, not OR alternative.
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

// DM47: Variable reassignment - last assignment with incomplete derivers should warn.
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

// DM48: Variable reassignment - last assignment satisfies OR alternative should pass.
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
