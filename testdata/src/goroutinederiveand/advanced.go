package goroutinederiveand

import (
	"context"
	"sync"

	"github.com/newrelic/go-agent/v3/newrelic"
)

// =============================================================================
// ADVANCED: AND (plus) - complex patterns
// Test flag: -goroutine-deriver=github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine+github.com/newrelic/go-agent/v3/newrelic.NewContext
// =============================================================================

// ===== SHOULD NOT REPORT =====

// [GOOD]: AND - Defer pattern
//
// AND - defer with both derivers.
func goodAndDeferWithBothDerivers(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		defer func() {
			recover()
		}()
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
}

// [GOOD]: AND - For loop pattern
//
// AND - for loop with both derivers.
func goodAndForLoopWithBothDerivers(ctx context.Context, txn *newrelic.Transaction) {
	for i := 0; i < 3; i++ {
		go func() {
			txn = txn.NewGoroutine()
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
	}
}

// [GOOD]: AND - WaitGroup pattern
//
// Both required deriver functions are called, satisfying AND condition.
func goodAndWaitGroupWithBothDerivers(ctx context.Context, txn *newrelic.Transaction) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
	wg.Wait()
}

// [GOOD]: AND - Conditional with both derivers in both branches.
//
// AND - conditional with both derivers in both branches.
func goodAndConditionalBothBranches(ctx context.Context, txn *newrelic.Transaction, cond bool) {
	if cond {
		go func() {
			txn = txn.NewGoroutine()
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
	} else {
		go func() {
			ctx = newrelic.NewContext(ctx, txn)
			txn = txn.NewGoroutine()
			_ = ctx
			_ = txn
		}()
	}
}

// [GOOD]: AND - Higher-order go fn()() where returned func has both derivers.
//
// AND - higher-order go fn()() where returned func has both derivers.
func goodAndHigherOrderReturnedFuncWithBothDerivers(ctx context.Context, txn *newrelic.Transaction) {
	makeWorker := func() func() {
		return func() {
			txn = txn.NewGoroutine()
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}
	}
	go makeWorker()() // Returned func calls both derivers
}

// [GOOD]: AND - Higher-order go fn() where fn is variable with both derivers.
//
// AND - higher-order go fn() where fn is variable with both derivers.
func goodAndHigherOrderVariableWithBothDerivers(ctx context.Context, txn *newrelic.Transaction) {
	fn := func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}
	go fn() // Variable func calls both derivers
}

// ===== SHOULD REPORT =====

// [BAD]: AND - Defer pattern
//
// AND - defer with only one deriver.
func badAndDeferWithOneDeriver(ctx context.Context, txn *newrelic.Transaction) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
		defer func() {
			recover()
		}()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
}

// [BAD]: AND - For loop pattern
//
// AND - for loop with only one deriver.
func badAndForLoopWithOneDeriver(ctx context.Context, txn *newrelic.Transaction) {
	for i := 0; i < 3; i++ {
		go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
	}
}

// [BAD]: AND - WaitGroup pattern
//
// Only one of the required deriver functions is called.
func badAndWaitGroupWithOneDeriver(ctx context.Context, txn *newrelic.Transaction) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
		defer wg.Done()
		txn = txn.NewGoroutine()
		_ = ctx
		_ = txn
	}()
	wg.Wait()
}

// [BAD]: AND - Conditional with one branch incomplete.
//
// AND - conditional with one branch incomplete.
func badAndConditionalOneBranchIncomplete(ctx context.Context, txn *newrelic.Transaction, cond bool) {
	if cond {
		go func() {
			txn = txn.NewGoroutine()
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
	} else {
		go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
	}
}

// [BAD]: AND - Multiple goroutines, one incomplete.
//
// AND - multiple goroutines, one incomplete.
func badAndMultipleGoroutinesOneIncomplete(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
}
