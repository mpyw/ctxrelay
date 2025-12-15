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

// GW01: Literal without ctx - basic bad case
func badWaitGroupGo(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		fmt.Println("no context")
	})
	wg.Wait()
}

// GW01b: Pointer receiver variant
func badWaitGroupGoPtr(ctx context.Context) {
	wg := new(sync.WaitGroup)
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		fmt.Println("no context")
	})
	wg.Wait()
}

// GW01c: Multiple Go calls without ctx
func badWaitGroupGoMultiple(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	})
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	})
	wg.Wait()
}

// ===== SHOULD NOT REPORT =====

// GW02: Literal with ctx - basic good case
func goodWaitGroupGoWithCtx(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx.Done()
	})
	wg.Wait()
}

// GW02b: Literal with ctx - via function call
func goodWaitGroupGoCallsWithCtx(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		doSomething(ctx)
	})
	wg.Wait()
}

// GW03: No ctx param - not checked
func goodNoContextParam() {
	var wg sync.WaitGroup
	wg.Go(func() {
		fmt.Println("hello")
	})
	wg.Wait()
}

// GW17: Traditional pattern (Add/Done) - not checked by waitgroup checker
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

// GW04: Shadow with non-ctx type (string)
func badShadowingNonContext(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		ctx := "not a context"
		_ = ctx
	})
	wg.Wait()
}

// GW05: Uses ctx before shadow - valid usage
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

// GW06: Ignore directive - same line
func goodIgnoredSameLine(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { //ctxrelay:ignore
	})
	wg.Wait()
}

// GW07: Ignore directive - previous line
func goodIgnoredPreviousLine(ctx context.Context) {
	var wg sync.WaitGroup
	//ctxrelay:ignore
	wg.Go(func() {
	})
	wg.Wait()
}

// ===== MULTIPLE CONTEXT PARAMETERS =====

// GW08: Multiple ctx params - reports first (bad)
func twoContextParams(ctx1, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx1"`
	})
	wg.Wait()
}

// GW09: Multiple ctx params - uses first (good)
func goodUsesOneOfTwoContexts(ctx1, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx1
	})
	wg.Wait()
}

// GW09b: Multiple ctx params - uses second (good)
func goodUsesSecondOfTwoContexts(ctx1, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx2 // uses second context - should NOT report
	})
	wg.Wait()
}

// GW14: Context as non-first param (good)
func goodCtxAsSecondParam(logger interface{}, ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx // ctx is second param but still detected
	})
	wg.Wait()
}

// GW14b: Context as non-first param without use (bad)
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
