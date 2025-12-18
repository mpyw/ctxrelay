// Package waitgroup contains test fixtures for the waitgroup context propagation checker.
// This file covers advanced patterns - real-world complex patterns: nested functions,
// conditionals, loops. See basic.go for daily patterns and evil.go for adversarial tests.
package waitgroup

import (
	"context"
	"fmt"
	"sync"
)

// ===== NESTED FUNCTIONS =====

// Go call inside inner named func without ctx
// sync.WaitGroup.Go() called from inner named function without context
// see also: errgroup
func badNestedInnerFunc(ctx context.Context) {
	var wg sync.WaitGroup
	innerFunc := func() {
		wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		})
	}
	innerFunc()
	wg.Wait()
}

// Go call inside IIFE without ctx
// sync.WaitGroup.Go() called from immediately invoked function without context
// see also: errgroup
func badNestedClosure(ctx context.Context) {
	var wg sync.WaitGroup
	func() {
		wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		})
	}()
	wg.Wait()
}

// Deep nested without ctx
// sync.WaitGroup.Go() called from deeply nested closure without context
// see also: goroutine, errgroup
func badNestedDeep(ctx context.Context) {
	var wg sync.WaitGroup
	func() {
		func() {
			wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
			})
		}()
	}()
	wg.Wait()
}

// Go call inside inner func with ctx
// sync.WaitGroup.Go() called from inner function with context properly captured
// see also: errgroup
func goodNestedWithCtx(ctx context.Context) {
	var wg sync.WaitGroup
	innerFunc := func() {
		wg.Go(func() {
			_ = ctx
		})
	}
	innerFunc()
	wg.Wait()
}

// Inner func has own ctx param
// Inner function has its own context parameter
// see also: errgroup
func goodNestedInnerHasOwnCtx(outerCtx context.Context) {
	innerFunc := func(ctx context.Context) {
		var wg sync.WaitGroup
		wg.Go(func() {
			_ = ctx // uses inner ctx
		})
		wg.Wait()
	}
	innerFunc(outerCtx)
}

// ===== CONDITIONAL PATTERNS =====

// Conditional Go without ctx
// sync.WaitGroup.Go() called conditionally without context
// see also: errgroup
func badConditionalGo(ctx context.Context, flag bool) {
	var wg sync.WaitGroup
	if flag {
		wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		})
	} else {
		wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		})
	}
	wg.Wait()
}

// Conditional Go with ctx
// sync.WaitGroup.Go() called conditionally with context properly captured
// see also: errgroup
func goodConditionalGo(ctx context.Context, flag bool) {
	var wg sync.WaitGroup
	if flag {
		wg.Go(func() {
			_ = ctx
		})
	} else {
		wg.Go(func() {
			_ = ctx
		})
	}
	wg.Wait()
}

// ===== LOOP PATTERNS =====

// Go in for loop without ctx
// sync.WaitGroup.Go() called in for loop without context
// see also: errgroup
func badLoopGo(ctx context.Context) {
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		})
	}
	wg.Wait()
}

// Go in for loop with ctx
// sync.WaitGroup.Go() called in for loop with context properly captured
// see also: errgroup
func goodLoopGo(ctx context.Context) {
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Go(func() {
			_ = ctx
		})
	}
	wg.Wait()
}

// Go in range loop without ctx
// sync.WaitGroup.Go() called in range loop without context
// see also: errgroup
func badRangeLoopGo(ctx context.Context) {
	var wg sync.WaitGroup
	items := []int{1, 2, 3}
	for _, item := range items {
		wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
			fmt.Println(item)
		})
	}
	wg.Wait()
}

// ===== DEFER PATTERNS =====

// Closure with defer but no ctx
// sync.WaitGroup.Go() closure with defer but no context
// see also: errgroup
func badDeferInClosure(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		defer fmt.Println("deferred")
	})
	wg.Wait()
}

// Closure with ctx and defer
// sync.WaitGroup.Go() closure with context and defer
// see also: errgroup
func goodDeferWithCtxDirect(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx // use ctx directly
		defer fmt.Println("cleanup")
	})
	wg.Wait()
}

// Ctx in deferred nested closure
// LIMITATION: Context used only in deferred nested closure is not detected
// see also: errgroup
func limitationDeferNestedClosure(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		// ctx is only in nested closure - not detected
		defer func() { _ = ctx }()
	})
	wg.Wait()
}
