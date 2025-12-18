// Package goroutinederivemixed contains test fixtures for the goroutine-derive mixed AND/OR mode.
// This file covers advanced patterns with mixed (A+B),C mode.
// Test flag: -goroutine-deriver=github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext
package goroutinederivemixed

import (
	"context"
	"sync"

	"github.com/my-example-app/telemetry/apm"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// ===== SHOULD NOT REPORT =====

// Mixed - defer satisfies AND group
// Defer goroutine satisfies AND group
func m20MixedDeferSatisfiesAndGroup(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		defer func() {
			recover()
		}()
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
}

// Mixed - defer satisfies OR alternative
// Defer goroutine satisfies OR alternative
func m21MixedDeferSatisfiesOrAlternative(ctx context.Context) {
	go func() {
		defer func() {
			recover()
		}()
		ctx = apm.NewGoroutineContext(ctx)
		_ = ctx
	}()
}

// Mixed - for loop satisfies AND group
// For loop goroutine satisfies AND group
func m22MixedForLoopSatisfiesAndGroup(ctx context.Context, txn *newrelic.Transaction) {
	for i := 0; i < 3; i++ {
		go func() {
			txn = txn.NewGoroutine()
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
	}
}

// Mixed - WaitGroup satisfies OR alternative
// WaitGroup pattern satisfies OR alternative
func m23MixedWaitGroupSatisfiesOrAlternative(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx = apm.NewGoroutineContext(ctx)
		_ = ctx
	}()
	wg.Wait()
}

// Mixed - conditional with different valid approaches per branch
// Conditional with AND group in one branch, OR alternative in other
func m24MixedConditionalDifferentApproaches(ctx context.Context, txn *newrelic.Transaction, cond bool) {
	if cond {
		// Satisfies via AND group
		go func() {
			txn = txn.NewGoroutine()
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
	} else {
		// Satisfies via OR alternative
		go func() {
			ctx = apm.NewGoroutineContext(ctx)
			_ = ctx
		}()
	}
}

// Mixed - multiple goroutines with different valid approaches
// Multiple goroutines with different approaches
func m25MixedMultipleGoroutinesDifferentApproaches(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
	go func() {
		ctx = apm.NewGoroutineContext(ctx)
		_ = ctx
	}()
}

// Mixed - higher-order go fn()() where returned func satisfies AND group
// Higher-order with returned func satisfying AND group
func m26MixedHigherOrderReturnedFuncSatisfiesAndGroup(ctx context.Context, txn *newrelic.Transaction) {
	makeWorker := func() func() {
		return func() {
			txn = txn.NewGoroutine()
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}
	}
	go makeWorker()() // Returned func satisfies AND group
}

// Mixed - higher-order go fn()() where returned func satisfies OR alternative
// Higher-order with returned func satisfying OR alternative
func m27MixedHigherOrderReturnedFuncSatisfiesOrAlternative(ctx context.Context) {
	makeWorker := func() func() {
		return func() {
			ctx = apm.NewGoroutineContext(ctx)
			_ = ctx
		}
	}
	go makeWorker()() // Returned func satisfies OR alternative
}

// Mixed - higher-order go fn() where fn is variable satisfying AND group
// Variable function satisfying AND group
func m28MixedHigherOrderVariableSatisfiesAndGroup(ctx context.Context, txn *newrelic.Transaction) {
	fn := func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}
	go fn() // Variable func satisfies AND group
}

// ===== SHOULD REPORT =====

// Mixed - defer with only first of AND group
// Defer with only first of AND group (incomplete)
func m29MixedDeferOnlyFirstOfAndGroup(ctx context.Context, txn *newrelic.Transaction) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		defer func() {
			recover()
		}()
		txn = txn.NewGoroutine()
		_ = ctx
		_ = txn
	}()
}

// Mixed - for loop with incomplete AND group
// For loop with incomplete AND group
func m30MixedForLoopIncompleteAndGroup(ctx context.Context, txn *newrelic.Transaction) {
	for i := 0; i < 3; i++ {
		go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
	}
}

// Mixed - WaitGroup with nothing
// WaitGroup with no derivers
func m31MixedWaitGroupWithNothing(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		defer wg.Done()
		_ = ctx
	}()
	wg.Wait()
}

// Mixed - conditional with one branch failing both conditions
// Conditional with one branch failing both
func m32MixedConditionalOneBranchFails(ctx context.Context, txn *newrelic.Transaction, cond bool) {
	if cond {
		go func() {
			txn = txn.NewGoroutine()
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
	} else {
		go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
			ctx = newrelic.NewContext(ctx, txn) // Only second of AND group
			_ = ctx
		}()
	}
}

// Mixed - multiple goroutines, one fails
// Multiple goroutines where one fails both conditions
func m33MixedMultipleGoroutinesOneFails(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		ctx = apm.NewGoroutineContext(ctx)
		_ = ctx
	}()
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		ctx = newrelic.NewContext(ctx, txn) // Only second of AND group
		_ = ctx
	}()
}
