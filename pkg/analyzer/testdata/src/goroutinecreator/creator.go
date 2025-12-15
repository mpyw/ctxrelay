// Package goroutinecreator tests the //ctxrelay:goroutine_creator directive.
package goroutinecreator

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

// ===== GOROUTINE CREATOR FUNCTIONS =====

//ctxrelay:goroutine_creator
func runWithGroup(g *errgroup.Group, fn func() error) {
	g.Go(fn)
}

//ctxrelay:goroutine_creator
func runWithWaitGroup(wg *sync.WaitGroup, fn func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		fn()
	}()
}

//ctxrelay:goroutine_creator
func runMultipleFuncs(fn1, fn2 func()) {
	go fn1()
	go fn2()
}

// ===== SHOULD REPORT =====

// GC01: Basic - func doesn't use ctx
func badBasicErrgroup(ctx context.Context) {
	g := new(errgroup.Group)
	fn := func() error {
		fmt.Println("no ctx")
		return nil
	}
	runWithGroup(g, fn) // want `runWithGroup\(\) func argument should use context "ctx"`
	_ = g.Wait()
}

// GC02: Basic - func doesn't use ctx (waitgroup)
func badBasicWaitGroup(ctx context.Context) {
	var wg sync.WaitGroup
	fn := func() {
		fmt.Println("no ctx")
	}
	runWithWaitGroup(&wg, fn) // want `runWithWaitGroup\(\) func argument should use context "ctx"`
	wg.Wait()
}

// GC03: Inline func literal doesn't use ctx
func badInlineFuncLiteral(ctx context.Context) {
	g := new(errgroup.Group)
	runWithGroup(g, func() error { // want `runWithGroup\(\) func argument should use context "ctx"`
		fmt.Println("no ctx")
		return nil
	})
	_ = g.Wait()
}

// GC04: Multiple func args - both bad
func badMultipleFuncs(ctx context.Context) {
	runMultipleFuncs(
		func() { fmt.Println("no ctx 1") }, // want `runMultipleFuncs\(\) func argument should use context "ctx"`
		func() { fmt.Println("no ctx 2") }, // want `runMultipleFuncs\(\) func argument should use context "ctx"`
	)
}

// GC05: Multiple func args - first bad
func badFirstFuncOnly(ctx context.Context) {
	runMultipleFuncs(
		func() { fmt.Println("no ctx") }, // want `runMultipleFuncs\(\) func argument should use context "ctx"`
		func() { _ = ctx },
	)
}

// GC06: Multiple func args - second bad
func badSecondFuncOnly(ctx context.Context) {
	runMultipleFuncs(
		func() { _ = ctx },
		func() { fmt.Println("no ctx") }, // want `runMultipleFuncs\(\) func argument should use context "ctx"`
	)
}

// ===== SHOULD NOT REPORT =====

// GC10: Basic - func uses ctx
func goodBasicErrgroup(ctx context.Context) {
	g := new(errgroup.Group)
	fn := func() error {
		_ = ctx
		return nil
	}
	runWithGroup(g, fn) // OK - fn uses ctx
	_ = g.Wait()
}

// GC11: Basic - func uses ctx (waitgroup)
func goodBasicWaitGroup(ctx context.Context) {
	var wg sync.WaitGroup
	fn := func() {
		_ = ctx
	}
	runWithWaitGroup(&wg, fn) // OK - fn uses ctx
	wg.Wait()
}

// GC12: Inline func literal uses ctx
func goodInlineFuncLiteral(ctx context.Context) {
	g := new(errgroup.Group)
	runWithGroup(g, func() error {
		_ = ctx
		return nil
	}) // OK
	_ = g.Wait()
}

// GC13: Multiple func args - both good
func goodMultipleFuncs(ctx context.Context) {
	runMultipleFuncs(
		func() { _ = ctx },
		func() { _ = ctx },
	) // OK
}

// GC14: No ctx param - not checked
func goodNoCtxParam() {
	g := new(errgroup.Group)
	runWithGroup(g, func() error {
		fmt.Println("no ctx")
		return nil
	}) // OK - no ctx in scope
	_ = g.Wait()
}

// GC15: Func has own ctx param
func goodFuncHasOwnCtx(ctx context.Context) {
	g := new(errgroup.Group)
	fn := func(innerCtx context.Context) error {
		_ = innerCtx
		return nil
	}
	// Note: runWithGroup expects func() error, not func(context.Context) error
	// This pattern is valid when the function declares its own context
	_ = fn
	_ = g
}

// ===== NON-CREATOR FUNCTIONS (should not be checked) =====

func normalHelper(g *errgroup.Group, fn func() error) {
	g.Go(fn)
}

// GC20: Call to non-creator function - not checked
func goodNonCreatorFunction(ctx context.Context) {
	g := new(errgroup.Group)
	fn := func() error {
		fmt.Println("no ctx")
		return nil
	}
	normalHelper(g, fn) // OK - normalHelper is not marked as goroutine_creator
	_ = g.Wait()
}
