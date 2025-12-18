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

// GW35: Go call inside inner named func without ctx
func badNestedInnerFunc(ctx context.Context) {
	var wg sync.WaitGroup
	innerFunc := func() {
		wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		})
	}
	innerFunc()
	wg.Wait()
}

// GW35b: Go call inside IIFE without ctx
func badNestedClosure(ctx context.Context) {
	var wg sync.WaitGroup
	func() {
		wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		})
	}()
	wg.Wait()
}

// GW35c: Go call inside deeply nested IIFE without ctx
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

// GW35d: Go call inside inner func with ctx
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

// GW10: Inner func has own ctx param
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

// GW24: Conditional Go without ctx
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

// GW24b: Conditional Go with ctx
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

// GW22: Go in for loop without ctx
func badLoopGo(ctx context.Context) {
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		})
	}
	wg.Wait()
}

// GW22b: Go in for loop with ctx
func goodLoopGo(ctx context.Context) {
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Go(func() {
			_ = ctx
		})
	}
	wg.Wait()
}

// GW23: Go in range loop without ctx
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

// GW35e: Closure with defer but no ctx
func badDeferInClosure(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		defer fmt.Println("deferred")
	})
	wg.Wait()
}

// GW02c: Closure with ctx and defer
func goodDeferWithCtxDirect(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx // use ctx directly
		defer fmt.Println("cleanup")
	})
	wg.Wait()
}

// GW21: LIMITATION - ctx in deferred nested closure not detected
func limitationDeferNestedClosure(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		// ctx is only in nested closure - not detected
		defer func() { _ = ctx }()
	})
	wg.Wait()
}
