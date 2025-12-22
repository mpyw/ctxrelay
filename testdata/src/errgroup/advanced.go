// Package errgroup contains test fixtures for the errgroup context propagation checker.
// This file covers advanced patterns - real-world complex patterns: nested functions,
// conditionals, loops. See basic.go for daily patterns and evil.go for adversarial tests.
package errgroup

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

// ===== NESTED FUNCTIONS =====

// [BAD]: Go call inside inner named func without ctx
//
// Go call in nested function does not use the outer context.
//
// See also:
//   conc: badNestedInnerFunc
//   waitgroup: badNestedInnerFunc
func badNestedInnerFunc(ctx context.Context) {
	g := new(errgroup.Group)
	innerFunc := func() {
		g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
			return nil
		})
	}
	innerFunc()
	_ = g.Wait()
}

// [BAD]: Go call inside IIFE without ctx
//
// Go call inside immediately invoked function expression without context.
//
// See also:
//   conc: badNestedClosure
//   waitgroup: badNestedClosure
func badNestedClosure(ctx context.Context) {
	g := new(errgroup.Group)
	func() {
		g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
			return nil
		})
	}()
	_ = g.Wait()
}

// [BAD]: Go call inside deeply nested IIFE without ctx
//
// Go call inside immediately invoked function expression without context.
//
// See also:
//   conc: badNestedDeep
//   waitgroup: badNestedDeep
func badNestedDeep(ctx context.Context) {
	g := new(errgroup.Group)
	func() {
		func() {
			g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
				return nil
			})
		}()
	}()
	_ = g.Wait()
}

// [GOOD]: Go call inside inner func with ctx
//
// Nested function properly captures and uses the outer context.
//
// See also:
//   conc: goodNestedWithCtx
//   waitgroup: goodNestedWithCtx
func goodNestedWithCtx(ctx context.Context) {
	g := new(errgroup.Group)
	innerFunc := func() {
		g.Go(func() error {
			_ = ctx
			return nil
		})
	}
	innerFunc()
	_ = g.Wait()
}

// [GOOD]: Inner func has own ctx param
//
// Inner function declares its own context parameter and uses it.
//
// See also:
//   conc: goodNestedInnerHasOwnCtx
//   goroutine: goodShadowingInnerCtxParam
//   waitgroup: goodNestedInnerHasOwnCtx
func goodNestedInnerHasOwnCtx(outerCtx context.Context) {
	innerFunc := func(ctx context.Context) {
		g := new(errgroup.Group)
		g.Go(func() error {
			_ = ctx // uses inner ctx
			return nil
		})
		_ = g.Wait()
	}
	innerFunc(outerCtx)
}

// ===== CONDITIONAL PATTERNS =====

// [BAD]: Conditional Go without ctx
//
// Conditional branches spawn goroutines without using context.
//
// See also:
//   conc: badConditionalGo
//   goroutine: badConditionalGoroutine
//   waitgroup: badConditionalGo
func badConditionalGo(ctx context.Context, flag bool) {
	g := new(errgroup.Group)
	if flag {
		g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
			return nil
		})
	} else {
		g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
			return nil
		})
	}
	_ = g.Wait()
}

// [GOOD]: Conditional goroutine with ctx
//
// All conditional branches properly use context in goroutines.
//
// See also:
//   conc: goodConditionalGo
func goodConditionalGo(ctx context.Context, flag bool) {
	g := new(errgroup.Group)
	if flag {
		g.Go(func() error {
			_ = ctx
			return nil
		})
	} else {
		g.Go(func() error {
			_ = ctx
			return nil
		})
	}
	_ = g.Wait()
}

// ===== LOOP PATTERNS =====

// [BAD]: Go in for loop without ctx
//
// Goroutines spawned in loop iterations do not use context.
//
// See also:
//   conc: badLoopGo
//   goroutine: badGoroutinesInLoop
//   waitgroup: badLoopGo
func badLoopGo(ctx context.Context) {
	g := new(errgroup.Group)
	for i := 0; i < 3; i++ {
		g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
			return nil
		})
	}
	_ = g.Wait()
}

// [GOOD]: Goroutine in for loop with ctx
//
// Goroutines in loop properly capture and use context.
//
// See also:
//   conc: goodLoopGo
func goodLoopGo(ctx context.Context) {
	g := new(errgroup.Group)
	for i := 0; i < 3; i++ {
		g.Go(func() error {
			_ = ctx
			return nil
		})
	}
	_ = g.Wait()
}

// [BAD]: Go in range loop without ctx
//
// Goroutines spawned in loop iterations do not use context.
//
// See also:
//   conc: badRangeLoopGo
//   goroutine: badGoroutinesInRangeLoop
//   waitgroup: badRangeLoopGo
func badRangeLoopGo(ctx context.Context) {
	g := new(errgroup.Group)
	items := []int{1, 2, 3}
	for _, item := range items {
		g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
			fmt.Println(item)
			return nil
		})
	}
	_ = g.Wait()
}

// ===== DEFER PATTERNS =====

// [BAD]: Closure with defer but no ctx
//
// Closure with defer statement does not use context.
//
// See also:
//   conc: badDeferInClosure
//   waitgroup: badDeferInClosure
func badDeferInClosure(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
		defer fmt.Println("deferred")
		return nil
	})
	_ = g.Wait()
}

// [GOOD]: Literal with ctx in select - with defer
//
// Closure with ctx and defer
//
// See also:
//   conc: goodDeferWithCtxDirect
func goodDeferWithCtxDirect(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx // use ctx directly
		defer fmt.Println("cleanup")
		return nil
	})
	_ = g.Wait()
}

// [LIMITATION]: Ctx in deferred nested closure not detected
//
// Context used only in deferred nested closure is not detected.
//
// See also:
//   conc: limitationDeferNestedClosure
//   goroutine: limitationDeferNestedClosure
//   waitgroup: limitationDeferNestedClosure
func limitationDeferNestedClosure(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
		// ctx is only in nested closure - not detected
		defer func() { _ = ctx }()
		return nil
	})
	_ = g.Wait()
}
