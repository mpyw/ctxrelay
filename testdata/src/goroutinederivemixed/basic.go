package goroutinederivemixed

import (
	"context"

	"github.com/my-example-app/telemetry/apm"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// =============================================================================
// BASIC: Mixed AND/OR - (A+B),C means (A AND B) OR C
// Test flag: -goroutine-deriver=github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext
// =============================================================================

// ===== SHOULD NOT REPORT =====

// [GOOD]: Mixed - satisfies first AND group (both Transaction.NewGoroutine and NewContext).
//
// Both required deriver functions are called, satisfying AND condition.
func goodMixedSatisfiesAndGroup(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
}

// [GOOD]: Mixed - satisfies OR alternative (NewGoroutineContext).
//
// Satisfies the mixed requirement via OR alternative path.
func goodMixedSatisfiesOrAlternative(ctx context.Context) {
	go func() {
		ctx = apm.NewGoroutineContext(ctx)
		_ = ctx
	}()
}

// [GOOD]: Mixed - satisfies both (AND group and OR alternative).
//
// Both required deriver functions are called, satisfying AND condition.
func goodMixedSatisfiesBoth(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		ctx = apm.NewGoroutineContext(ctx)
		_ = ctx
	}()
}

// [NOTCHECKED]: Mixed - has own context param.
//
// Function with own context parameter is not checked.
func notCheckedMixedOwnContextParam(ctx context.Context) {
	go func(ctx context.Context) {
		_ = ctx
	}(ctx)
}

// ===== SHOULD REPORT =====

// [BAD]: Mixed - only calls first of AND group (incomplete).
//
// Only one of the required deriver functions is called.
func badMixedOnlyFirstOfAnd(ctx context.Context, txn *newrelic.Transaction) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		txn = txn.NewGoroutine()
		_ = ctx
		_ = txn
	}()
}

// [BAD]: Mixed - only calls second of AND group (incomplete).
//
// Only one of the required deriver functions is called.
func badMixedOnlySecondOfAnd(ctx context.Context, txn *newrelic.Transaction) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
}

// [BAD]: Mixed - calls nothing.
//
// Goroutine does not call any deriver function.
func badMixedCallsNothing(ctx context.Context) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		_ = ctx
	}()
}
