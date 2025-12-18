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

// GE35: Go call inside inner named func without ctx
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

// GE35b: Go call inside IIFE without ctx
func badNestedClosure(ctx context.Context) {
	g := new(errgroup.Group)
	func() {
		g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
			return nil
		})
	}()
	_ = g.Wait()
}

// GE35c: Go call inside deeply nested IIFE without ctx
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

// GE35d: Go call inside inner func with ctx
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

// GE10: Inner func has own ctx param
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

// GE24: Conditional Go without ctx
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

// GE24b: Conditional Go with ctx
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

// GE22: Go in for loop without ctx
func badLoopGo(ctx context.Context) {
	g := new(errgroup.Group)
	for i := 0; i < 3; i++ {
		g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
			return nil
		})
	}
	_ = g.Wait()
}

// GE22b: Go in for loop with ctx
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

// GE23: Go in range loop without ctx
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

// GE35e: Closure with defer but no ctx
func badDeferInClosure(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
		defer fmt.Println("deferred")
		return nil
	})
	_ = g.Wait()
}

// GE02d: Closure with ctx and defer
func goodDeferWithCtxDirect(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx // use ctx directly
		defer fmt.Println("cleanup")
		return nil
	})
	_ = g.Wait()
}

// GE21: LIMITATION - ctx in deferred nested closure not detected
func limitationDeferNestedClosure(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
		// ctx is only in nested closure - not detected
		defer func() { _ = ctx }()
		return nil
	})
	_ = g.Wait()
}
