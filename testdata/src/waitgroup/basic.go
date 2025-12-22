//go:build go1.25

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

// [BAD]: Literal without ctx
//
// Literal without ctx - basic bad case
//
// See also:
//   conc: badConcWaitGroupGo
//   errgroup: badErrgroupGo
//   goroutine: badGoroutineNoCapture
func badWaitGroupGo(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		fmt.Println("no context")
	})
	wg.Wait()
}

// [BAD]: Literal without ctx - pointer receiver
//
// Pointer receiver variant
func badWaitGroupGoPtr(ctx context.Context) {
	wg := new(sync.WaitGroup)
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		fmt.Println("no context")
	})
	wg.Wait()
}

// [BAD]: Multiple Go calls without ctx
//
// Multiple goroutine closures all fail to use the available context.
//
// See also:
//   conc: badConcWaitGroupGoMultiple
//   errgroup: badErrgroupGoMultiple
func badWaitGroupGoMultiple(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	})
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	})
	wg.Wait()
}

// ===== SHOULD NOT REPORT =====

// [GOOD]: Literal with ctx - basic good case
//
// Closure directly references the context variable from enclosing scope.
//
// See also:
//   conc: goodConcWaitGroupGoWithCtx
//   errgroup: goodErrgroupGoWithCtx
//   goroutine: goodGoroutineCapturesCtx
func goodWaitGroupGoWithCtx(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx.Done()
	})
	wg.Wait()
}

// [GOOD]: Literal with ctx - via function call
//
// Context is passed to helper function inside closure.
//
// See also:
//   conc: goodConcWaitGroupGoCallsWithCtx
//   errgroup: goodErrgroupGoCallsWithCtx
//   goroutine: goodGoroutineUsesCtxInCall
func goodWaitGroupGoCallsWithCtx(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		doSomething(ctx)
	})
	wg.Wait()
}

// [GOOD]: No ctx param
//
// No ctx param - not checked
//
// See also:
//   conc: goodNoContextParam
//   errgroup: goodNoContextParam
//   goroutine: goodNoContextParam
func goodNoContextParam() {
	var wg sync.WaitGroup
	wg.Go(func() {
		fmt.Println("hello")
	})
	wg.Wait()
}

// [GOOD]: Traditional pattern (Add/Done)
//
// Traditional pattern (Add/Done) - not checked by waitgroup checker
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

// [BAD]: Shadow with non-ctx type - string
//
// Shadow with non-ctx type (string)
//
// See also:
//   conc: badShadowingNonContext
//   errgroup: badShadowingNonContext
//   goroutine: badShadowingNonContext
func badShadowingNonContext(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		ctx := "not a context"
		_ = ctx
	})
	wg.Wait()
}

// [GOOD]: Uses ctx before shadow
//
// Uses ctx before shadow - valid usage
//
// See also:
//   conc: goodUsesCtxBeforeShadowing
//   errgroup: goodUsesCtxBeforeShadowing
//   goroutine: goodUsesCtxBeforeShadowing
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

// [GOOD]: Ignore directive - same line
//
// The //goroutinectx:ignore directive suppresses the warning.
//
// See also:
//   conc: goodIgnoredSameLine
//   errgroup: goodIgnoredSameLine
//   goroutine: goodIgnoredSameLine
func goodIgnoredSameLine(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { //goroutinectx:ignore
	})
	wg.Wait()
}

// [GOOD]: Ignore directive - previous line
//
// The //goroutinectx:ignore directive suppresses the warning.
//
// See also:
//   conc: goodIgnoredPreviousLine
//   errgroup: goodIgnoredPreviousLine
//   goroutine: goodIgnoredPreviousLine
func goodIgnoredPreviousLine(ctx context.Context) {
	var wg sync.WaitGroup
	//goroutinectx:ignore
	wg.Go(func() {
	})
	wg.Wait()
}

// ===== MULTIPLE CONTEXT PARAMETERS =====

// [BAD]: Multiple ctx params - reports first
//
// Multiple context parameters available but none are used.
//
// See also:
//   conc: twoContextParams
//   errgroup: twoContextParams
//   goroutine: twoContextParams
func twoContextParams(ctx1, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx1"`
	})
	wg.Wait()
}

// [GOOD]: Multiple ctx params - uses first
//
// One of the available context parameters is properly used.
//
// See also:
//   conc: goodUsesOneOfTwoContexts
//   errgroup: goodUsesOneOfTwoContexts
//   goroutine: goodUsesOneOfTwoContexts
func goodUsesOneOfTwoContexts(ctx1, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx1
	})
	wg.Wait()
}

// [GOOD]: Multiple ctx params - uses second
//
// One of the available context parameters is properly used.
//
// See also:
//   conc: goodUsesSecondOfTwoContexts
//   errgroup: goodUsesSecondOfTwoContexts
//   goroutine: goodUsesSecondOfTwoContexts
func goodUsesSecondOfTwoContexts(ctx1, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx2 // uses second context - should NOT report
	})
	wg.Wait()
}

// [GOOD]: Context as non-first param
//
// Context is detected and used even when not the first parameter.
//
// See also:
//   conc: goodCtxAsSecondParam
//   errgroup: goodCtxAsSecondParam
//   goroutine: goodCtxAsSecondParam
func goodCtxAsSecondParam(logger interface{}, ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx // ctx is second param but still detected
	})
	wg.Wait()
}

// [BAD]: Context as non-first param without use
//
// Context parameter exists but is not used in the closure.
//
// See also:
//   conc: badCtxAsSecondParam
//   errgroup: badCtxAsSecondParam
//   goroutine: badCtxAsSecondParam
func badCtxAsSecondParam(logger interface{}, ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		_ = logger
	})
	wg.Wait()
}

//vt:helper
func doSomething(ctx context.Context) {
	_ = ctx
}
