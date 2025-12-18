// Package goroutinederiveand contains test fixtures for the goroutine-derive AND mode.
// This file covers basic patterns with AND (plus) mode - all derivers must be called.
// Test flag: -goroutine-deriver=github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine+github.com/newrelic/go-agent/v3/newrelic.NewContext
package goroutinederiveand

import (
	"context"

	"github.com/newrelic/go-agent/v3/newrelic"
)

// ===== SHOULD NOT REPORT =====

// AND - calls both functions
// Goroutine calls both required deriver functions
func a01AndCallsBoth(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
}

// AND - calls both in different order
// Goroutine calls both derivers in reversed order
func a02AndCallsBothReversed(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		ctx = newrelic.NewContext(ctx, txn)
		txn = txn.NewGoroutine()
		_ = ctx
		_ = txn
	}()
}

// AND - calls both with other code between
// Goroutine calls both derivers with interleaved code
func a03AndCallsBothInterleaved(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		txn = txn.NewGoroutine()
		doSomething()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
}

// AND - has own context param
// Goroutine has its own context parameter
func a04AndOwnContextParam(ctx context.Context) {
	go func(ctx context.Context) {
		_ = ctx
	}(ctx)
}

// ===== SHOULD REPORT =====

// AND - calls only first (method)
// Goroutine calls only first of AND group
func a05AndCallsOnlyFirst(ctx context.Context, txn *newrelic.Transaction) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
		txn = txn.NewGoroutine()
		_ = ctx
		_ = txn
	}()
}

// AND - calls only second (function)
// Goroutine calls only second of AND group
func a06AndCallsOnlySecond(ctx context.Context, txn *newrelic.Transaction) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
}

// AND - calls neither function
// Goroutine calls neither deriver
func a07AndCallsNeither(ctx context.Context) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
		_ = ctx
	}()
}

func doSomething() {}
