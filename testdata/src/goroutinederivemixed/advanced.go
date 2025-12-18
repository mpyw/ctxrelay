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

// [GOOD]: Mixed - defer satisfies AND group.
//
// Closure with defer statement properly uses context.
func goodMixedDeferSatisfiesAndGroup(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		defer func() {
			recover()
		}()
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
}

// [GOOD]: Mixed - defer satisfies OR alternative.
//
// Closure with defer statement properly uses context.
func goodMixedDeferSatisfiesOrAlternative(ctx context.Context) {
	go func() {
		defer func() {
			recover()
		}()
		ctx = apm.NewGoroutineContext(ctx)
		_ = ctx
	}()
}

// [GOOD]: Mixed - for loop satisfies AND group.
//
// Goroutines in loop properly capture and use context.
func goodMixedForLoopSatisfiesAndGroup(ctx context.Context, txn *newrelic.Transaction) {
	for i := 0; i < 3; i++ {
		go func() {
			txn = txn.NewGoroutine()
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
	}
}

// [GOOD]: Mixed - WaitGroup satisfies OR alternative.
//
// Satisfies the mixed requirement via OR alternative path.
func goodMixedWaitGroupSatisfiesOrAlternative(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx = apm.NewGoroutineContext(ctx)
		_ = ctx
	}()
	wg.Wait()
}

// [GOOD]: Mixed - conditional with different valid approaches per branch.
//
// All conditional branches properly use context in goroutines.
func goodMixedConditionalDifferentApproaches(ctx context.Context, txn *newrelic.Transaction, cond bool) {
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

// [GOOD]: Mixed - multiple goroutines with different valid approaches.
//
// Multiple goroutines each satisfying requirements differently.
func goodMixedMultipleGoroutinesDifferentApproaches(ctx context.Context, txn *newrelic.Transaction) {
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

// [GOOD]: Mixed - higher-order go fn()() where returned func satisfies AND group.
//
// Satisfies the mixed requirement by completing AND group.
func goodMixedHigherOrderReturnedFuncSatisfiesAndGroup(ctx context.Context, txn *newrelic.Transaction) {
	makeWorker := func() func() {
		return func() {
			txn = txn.NewGoroutine()
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}
	}
	go makeWorker()() // Returned func satisfies AND group
}

// [GOOD]: Mixed - higher-order go fn()() where returned func satisfies OR alternative.
//
// Satisfies the mixed requirement via OR alternative path.
func goodMixedHigherOrderReturnedFuncSatisfiesOrAlternative(ctx context.Context) {
	makeWorker := func() func() {
		return func() {
			ctx = apm.NewGoroutineContext(ctx)
			_ = ctx
		}
	}
	go makeWorker()() // Returned func satisfies OR alternative
}

// [GOOD]: Mixed - higher-order go fn() where fn is variable satisfying AND group.
//
// Satisfies the mixed requirement by completing AND group.
func goodMixedHigherOrderVariableSatisfiesAndGroup(ctx context.Context, txn *newrelic.Transaction) {
	fn := func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}
	go fn() // Variable func satisfies AND group
}

// ===== SHOULD REPORT =====

// [BAD]: Mixed - defer with only first of AND group.
//
// Closure with defer statement does not use context.
func badMixedDeferOnlyFirstOfAndGroup(ctx context.Context, txn *newrelic.Transaction) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		defer func() {
			recover()
		}()
		txn = txn.NewGoroutine()
		_ = ctx
		_ = txn
	}()
}

// [BAD]: Mixed - for loop with incomplete AND group.
//
// Goroutines spawned in loop iterations do not use context.
func badMixedForLoopIncompleteAndGroup(ctx context.Context, txn *newrelic.Transaction) {
	for i := 0; i < 3; i++ {
		go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
	}
}

// [BAD]: Mixed - WaitGroup with nothing.
//
// WaitGroup pattern without any deriver calls.
func badMixedWaitGroupWithNothing(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		defer wg.Done()
		_ = ctx
	}()
	wg.Wait()
}

// [BAD]: Mixed - conditional with one branch failing both conditions.
//
// Conditional branches spawn goroutines without using context.
func badMixedConditionalOneBranchFails(ctx context.Context, txn *newrelic.Transaction, cond bool) {
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

// [BAD]: Mixed - multiple goroutines, one fails.
//
// One of multiple goroutines fails to meet deriver requirements.
func badMixedMultipleGoroutinesOneFails(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		ctx = apm.NewGoroutineContext(ctx)
		_ = ctx
	}()
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext,github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		ctx = newrelic.NewContext(ctx, txn) // Only second of AND group
		_ = ctx
	}()
}
