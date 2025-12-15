package goroutinederivemixed

import (
	"context"
	"sync"

	"github.com/my-example-app/telemetry/apm"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// =============================================================================
// ADVANCED: Mixed AND/OR - complex patterns
// Test flag: -goroutine-deriver=github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext
// =============================================================================

// ===== SHOULD NOT REPORT =====

// DM20: Mixed - defer satisfies AND group.
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

// DM21: Mixed - defer satisfies OR alternative.
func m21MixedDeferSatisfiesOrAlternative(ctx context.Context) {
	go func() {
		defer func() {
			recover()
		}()
		ctx = apm.NewGoroutineContext(ctx)
		_ = ctx
	}()
}

// DM22: Mixed - for loop satisfies AND group.
func m22MixedForLoopSatisfiesAndGroup(ctx context.Context, txn *newrelic.Transaction) {
	for i := 0; i < 3; i++ {
		go func() {
			txn = txn.NewGoroutine()
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
	}
}

// DM23: Mixed - WaitGroup satisfies OR alternative.
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

// DM24: Mixed - conditional with different valid approaches per branch.
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

// DM25: Mixed - multiple goroutines with different valid approaches.
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

// DM26: Mixed - higher-order go fn()() where returned func satisfies AND group.
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

// DM27: Mixed - higher-order go fn()() where returned func satisfies OR alternative.
func m27MixedHigherOrderReturnedFuncSatisfiesOrAlternative(ctx context.Context) {
	makeWorker := func() func() {
		return func() {
			ctx = apm.NewGoroutineContext(ctx)
			_ = ctx
		}
	}
	go makeWorker()() // Returned func satisfies OR alternative
}

// DM28: Mixed - higher-order go fn() where fn is variable satisfying AND group.
func m28MixedHigherOrderVariableSatisfiesAndGroup(ctx context.Context, txn *newrelic.Transaction) {
	fn := func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}
	go fn() // Variable func satisfies AND group
}

// ===== SHOULD REPORT =====

// DM29: Mixed - defer with only first of AND group.
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

// DM30: Mixed - for loop with incomplete AND group.
func m30MixedForLoopIncompleteAndGroup(ctx context.Context, txn *newrelic.Transaction) {
	for i := 0; i < 3; i++ {
		go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
	}
}

// DM31: Mixed - WaitGroup with nothing.
func m31MixedWaitGroupWithNothing(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		defer wg.Done()
		_ = ctx
	}()
	wg.Wait()
}

// DM32: Mixed - conditional with one branch failing both conditions.
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

// DM33: Mixed - multiple goroutines, one fails.
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
