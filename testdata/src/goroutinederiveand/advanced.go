// Package goroutinederiveand contains test fixtures for the goroutine-derive AND mode.
// This file covers advanced patterns with AND (plus) mode.
// Test flag: -goroutine-deriver=github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine+github.com/newrelic/go-agent/v3/newrelic.NewContext
package goroutinederiveand

import (
	"context"
	"sync"

	"github.com/newrelic/go-agent/v3/newrelic"
)

// ===== SHOULD NOT REPORT =====

// AND - defer with both derivers
// Goroutine with defer calls both derivers
func a20AndDeferWithBothDerivers(ctx context.Context, txn *newrelic.Transaction) {
	go func() {
		defer func() {
			recover()
		}()
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
}

// AND - for loop with both derivers
// Goroutine in for loop calls both derivers
func a21AndForLoopWithBothDerivers(ctx context.Context, txn *newrelic.Transaction) {
	for i := 0; i < 3; i++ {
		go func() {
			txn = txn.NewGoroutine()
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
	}
}

// AND - WaitGroup pattern with both derivers
// WaitGroup pattern with both derivers
func a22AndWaitGroupWithBothDerivers(ctx context.Context, txn *newrelic.Transaction) {
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

// AND - conditional with both derivers in both branches
// Conditional with both derivers in each branch
func a23AndConditionalBothBranches(ctx context.Context, txn *newrelic.Transaction, cond bool) {
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

// AND - higher-order go fn()() where returned func has both derivers
// Higher-order function with both derivers in returned func
func a24AndHigherOrderReturnedFuncWithBothDerivers(ctx context.Context, txn *newrelic.Transaction) {
	makeWorker := func() func() {
		return func() {
			txn = txn.NewGoroutine()
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}
	}
	go makeWorker()() // Returned func calls both derivers
}

// AND - higher-order go fn() where fn is variable with both derivers
// Variable function with both derivers
func a25AndHigherOrderVariableWithBothDerivers(ctx context.Context, txn *newrelic.Transaction) {
	fn := func() {
		txn = txn.NewGoroutine()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}
	go fn() // Variable func calls both derivers
}

// ===== SHOULD REPORT =====

// AND - defer with only one deriver
// Goroutine with defer calls only one deriver
func a26AndDeferWithOneDeriver(ctx context.Context, txn *newrelic.Transaction) {
	go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
		defer func() {
			recover()
		}()
		ctx = newrelic.NewContext(ctx, txn)
		_ = ctx
	}()
}

// AND - for loop with only one deriver
// Goroutine in for loop calls only one deriver
func a27AndForLoopWithOneDeriver(ctx context.Context, txn *newrelic.Transaction) {
	for i := 0; i < 3; i++ {
		go func() { // want "goroutine should call github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine\\+github.com/newrelic/go-agent/v3/newrelic.NewContext to derive context"
			ctx = newrelic.NewContext(ctx, txn)
			_ = ctx
		}()
	}
}

// AND - WaitGroup pattern with only one deriver
// WaitGroup pattern with only one deriver
func a28AndWaitGroupWithOneDeriver(ctx context.Context, txn *newrelic.Transaction) {
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

// AND - conditional with one branch incomplete
// Conditional with one branch missing a deriver
func a29AndConditionalOneBranchIncomplete(ctx context.Context, txn *newrelic.Transaction, cond bool) {
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

// AND - multiple goroutines, one incomplete
// Multiple goroutines where one is incomplete
func a30AndMultipleGoroutinesOneIncomplete(ctx context.Context, txn *newrelic.Transaction) {
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
