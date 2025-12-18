package goroutinederiveand

import (
	"context"

	"github.com/newrelic/go-agent/v3/newrelic"
)

// =============================================================================
// BASIC: AND (plus) - all must be called
// Test flag: -goroutine-deriver=github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine+github.com/newrelic/go-agent/v3/newrelic.NewContext
// =============================================================================

// ===== SHOULD NOT REPORT =====

// DA01: AND - calls both functions.
func a01AndCallsBoth(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
}

// DA02: AND - calls both in different order.
func a02AndCallsBothReversed(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		ctx = newrelic.NewContext(ctx, txn)
		txn = txn.NewGoroutine()
		_ = ctx
		_ = txn
	}()
}

// DA03: AND - calls both with other code between.
func a03AndCallsBothInterleaved(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		txn = txn.NewGoroutine()
		doSomething()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
}

// DA04: AND - has own context param.
func a04AndOwnContextParam(ctx context.Context) {
	go func(ctx context.Context) {
		_ = ctx
	}(ctx)
}

// ===== SHOULD REPORT =====

// DA05: AND - calls only first (method).
func a05AndCallsOnlyFirst(ctx context.Context, txn *newrelic.Transaction) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
		txn = txn.NewGoroutine()
		_ = ctx
		_ = txn
	}()
}

// DA06: AND - calls only second (function).
func a06AndCallsOnlySecond(ctx context.Context, txn *newrelic.Transaction) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
}

// DA07: AND - calls neither function.
func a07AndCallsNeither(ctx context.Context) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
		_ = ctx
	}()
}

func doSomething() {}
