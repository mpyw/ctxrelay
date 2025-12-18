// Package waitgroup contains test fixtures for the waitgroup context propagation checker.
// This file covers basic/daily patterns - simple good/bad cases, shadowing, ignore directives.
// Note: sync.WaitGroup.Go() was added in Go 1.25.
// See advanced.go for real-world complex patterns and evil.go for adversarial tests.
package waitgroup

import (
	"context"
	"fmt"
	"sync"
)

// ===== SHOULD REPORT =====

// Literal without ctx - basic bad case
// sync.WaitGroup.Go() closure does not use context
func badWaitGroupGo(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		fmt.Println("no context")
	})
	wg.Wait()
}

// Pointer receiver variant
// sync.WaitGroup.Go() with pointer receiver, closure does not use context
func badWaitGroupGoPtr(ctx context.Context) {
	wg := new(sync.WaitGroup)
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		fmt.Println("no context")
	})
	wg.Wait()
}

// Multiple Go calls without ctx
// Multiple sync.WaitGroup.Go() calls without context
func badWaitGroupGoMultiple(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	})
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	})
	wg.Wait()
}

// ===== SHOULD NOT REPORT =====

// Literal with ctx - basic good case
// sync.WaitGroup.Go() closure properly uses context
func goodWaitGroupGoWithCtx(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx.Done()
	})
	wg.Wait()
}

// Literal with ctx - via function call
// sync.WaitGroup.Go() closure uses context via function call
func goodWaitGroupGoCallsWithCtx(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		doSomething(ctx)
	})
	wg.Wait()
}

// No ctx param
// Function has no context parameter - not checked
// see also: goroutine, errgroup
func goodNoContextParam() {
	var wg sync.WaitGroup
	wg.Go(func() {
		fmt.Println("hello")
	})
	wg.Wait()
}

// Traditional pattern (Add/Done)
// Traditional sync.WaitGroup pattern with Add/Done - not checked by waitgroup checker
func goodTraditionalPattern(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = ctx.Done()
	}()
	wg.Wait()
}

// ===== SHADOWING TESTS =====

// Shadow with non-ctx type
// Context variable is shadowed with non-context type (string)
// see also: goroutine, errgroup
func badShadowingNonContext(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		ctx := "not a context"
		_ = ctx
	})
	wg.Wait()
}

// Uses ctx before shadow
// Uses context before shadowing it - valid usage
// see also: goroutine, errgroup
func goodUsesCtxBeforeShadowing(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx.Done() // use ctx before shadowing
		ctx := "shadow"
		_ = ctx
	})
	wg.Wait()
}

// ===== IGNORE DIRECTIVES =====

// Ignore directive - same line
// Ignore directive on the same line suppresses warning
// see also: goroutine, errgroup
func goodIgnoredSameLine(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { //goroutinectx:ignore
	})
	wg.Wait()
}

// Ignore directive - previous line
// Ignore directive on the previous line suppresses warning
// see also: goroutine, errgroup
func goodIgnoredPreviousLine(ctx context.Context) {
	var wg sync.WaitGroup
	//goroutinectx:ignore
	wg.Go(func() {
	})
	wg.Wait()
}

// ===== MULTIPLE CONTEXT PARAMETERS =====

// Multiple ctx params - reports first
// Function has two context parameters, reports first one when neither used
// see also: goroutine, errgroup
func twoContextParams(ctx1, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx1"`
	})
	wg.Wait()
}

// Multiple ctx params - uses first
// Function has multiple context parameters and uses first
// see also: goroutine, errgroup
func goodUsesOneOfTwoContexts(ctx1, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx1
	})
	wg.Wait()
}

// Multiple ctx params - uses second
// Function has multiple context parameters and uses second - should NOT report
// see also: goroutine, errgroup
func goodUsesSecondOfTwoContexts(ctx1, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx2 // uses second context - should NOT report
	})
	wg.Wait()
}

// Context as non-first param
// Context is second parameter and is properly used
// see also: goroutine, errgroup
func goodCtxAsSecondParam(logger interface{}, ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx // ctx is second param but still detected
	})
	wg.Wait()
}

// Context as non-first param without use
// Context is second parameter but not used in closure
// see also: goroutine, errgroup
func badCtxAsSecondParam(logger interface{}, ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		_ = logger
	})
	wg.Wait()
}

func doSomething(ctx context.Context) {
	_ = ctx
}
