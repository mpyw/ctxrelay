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

// Go call inside inner named func without ctx
// errgroup.Go() called from inner named function without context
// see also: waitgroup
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

// Go call inside IIFE without ctx
// errgroup.Go() called from immediately invoked function without context
// see also: waitgroup
func badNestedClosure(ctx context.Context) {
	g := new(errgroup.Group)
	func() {
		g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
			return nil
		})
	}()
	_ = g.Wait()
}

// Deep nested without ctx
// errgroup.Go() called from deeply nested closure without context
// see also: goroutine, waitgroup
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

// Go call inside inner func with ctx
// errgroup.Go() called from inner function with context properly captured
// see also: waitgroup
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

// Inner func has own ctx param
// Inner function has its own context parameter
// see also: waitgroup
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

// Conditional Go without ctx
// errgroup.Go() called conditionally without context
// see also: waitgroup
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

// Conditional Go with ctx
// errgroup.Go() called conditionally with context properly captured
// see also: waitgroup
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

// Go in for loop without ctx
// errgroup.Go() called in for loop without context
// see also: waitgroup
func badLoopGo(ctx context.Context) {
	g := new(errgroup.Group)
	for i := 0; i < 3; i++ {
		g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
			return nil
		})
	}
	_ = g.Wait()
}

// Go in for loop with ctx
// errgroup.Go() called in for loop with context properly captured
// see also: waitgroup
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

// Go in range loop without ctx
// errgroup.Go() called in range loop without context
// see also: waitgroup
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

// Closure with defer but no ctx
// errgroup.Go() closure with defer but no context
// see also: waitgroup
func badDeferInClosure(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
		defer fmt.Println("deferred")
		return nil
	})
	_ = g.Wait()
}

// Closure with ctx and defer
// errgroup.Go() closure with context and defer
// see also: waitgroup
func goodDeferWithCtxDirect(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx // use ctx directly
		defer fmt.Println("cleanup")
		return nil
	})
	_ = g.Wait()
}

// Ctx in deferred nested closure
// LIMITATION: Context used only in deferred nested closure is not detected
// see also: waitgroup
func limitationDeferNestedClosure(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
		// ctx is only in nested closure - not detected
		defer func() { _ = ctx }()
		return nil
	})
	_ = g.Wait()
}
