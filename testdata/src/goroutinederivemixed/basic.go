// Package goroutinederivemixed contains test fixtures for the goroutine-derive mixed AND/OR mode.
// This file covers basic patterns with mixed (A+B),C mode - (A AND B) OR C.
// Test flag: -goroutine-deriver=github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext
package goroutinederivemixed

import (
	"context"

	"github.com/my-example-app/telemetry/apm"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// ===== SHOULD NOT REPORT =====

// Mixed - satisfies first AND group
// Satisfies first AND group (both Transaction.NewGoroutine and NewContext)
func m01MixedSatisfiesAndGroup(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
}

// Mixed - satisfies OR alternative
// Satisfies OR alternative (NewGoroutineContext)
func m02MixedSatisfiesOrAlternative(ctx context.Context) {
	go func() {
		ctx = apm.NewGoroutineContext(ctx)
		_ = ctx
	}()
}

// Mixed - satisfies both
// Satisfies both AND group and OR alternative
func m03MixedSatisfiesBoth(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		ctx = apm.NewGoroutineContext(ctx)
		_ = ctx
	}()
}

// Mixed - has own context param
// Goroutine has its own context parameter
func m04MixedOwnContextParam(ctx context.Context) {
	go func(ctx context.Context) {
		_ = ctx
	}(ctx)
}

// ===== SHOULD REPORT =====

// Mixed - only calls first of AND group
// Only calls first of AND group (incomplete)
func m05MixedOnlyFirstOfAnd(ctx context.Context, txn *newrelic.Transaction) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		txn = txn.NewGoroutine()
		_ = ctx
		_ = txn
	}()
}

// Mixed - only calls second of AND group
// Only calls second of AND group (incomplete)
func m06MixedOnlySecondOfAnd(ctx context.Context, txn *newrelic.Transaction) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
}

// Mixed - calls nothing
// Calls neither AND group nor OR alternative
func m07MixedCallsNothing(ctx context.Context) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		_ = ctx
	}()
}
