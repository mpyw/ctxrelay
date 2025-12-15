// Package goroutine contains test fixtures for the goroutine context propagation checker.
// This file covers basic/daily patterns - single goroutine, shadowing, ignore directives.
// See advanced.go for real-world complex patterns and evil.go for adversarial tests.
package goroutine

import (
	"context"
	"fmt"
)

// ===== SHOULD REPORT =====

// GO01: Literal without ctx - basic bad case
func badGoroutineNoCapture(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		fmt.Println("no context")
	}()
}

// GO01b: Literal without ctx - variant
func badGoroutineIgnoresCtx(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		x := 1
		_ = x
	}()
}

// ===== SHOULD NOT REPORT =====

// GO02: Literal with ctx - basic good case
func goodGoroutineCapturesCtx(ctx context.Context) {
	go func() {
		_ = ctx.Done()
	}()
}

// GO02b: Literal with ctx - via function call
func goodGoroutineUsesCtxInCall(ctx context.Context) {
	go func() {
		doSomething(ctx)
	}()
}

// GO03: No ctx param - not checked
func goodNoContextParam() {
	go func() {
		fmt.Println("hello")
	}()
}

// GO02c: Literal with derived ctx
func goodGoroutineWithDerivedCtx(ctx context.Context) {
	go func() {
		ctx2, cancel := context.WithCancel(ctx)
		defer cancel()
		_ = ctx2
	}()
}

// GO02d: Literal with ctx in select
func goodGoroutineSelectOnCtx(ctx context.Context) {
	go func() {
		select {
		case <-ctx.Done():
			return
		default:
		}
	}()
}

// ===== SHADOWING TESTS =====

// GO04: Shadow with non-ctx type (string)
func badShadowingNonContext(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		ctx := "not a context" // shadows with string
		_ = ctx
	}()
}

// GO10: Inner func has own ctx param
func goodShadowingInnerCtxParam(outerCtx context.Context) {
	go func(ctx context.Context) {
		_ = ctx.Done() // uses inner ctx - OK
	}(outerCtx)
}

// GO04b: Shadow with non-ctx type (channel)
func badShadowingWithDifferentType(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		ctx := make(chan int) // shadows with channel
		close(ctx)
	}()
}

// GO04c: Shadow with non-ctx type (function)
func badShadowingWithFunction(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		ctx := func() {} // shadows with function
		ctx()
	}()
}

// GO04d: Shadow in nested block
func badShadowingInNestedBlock(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		if true {
			ctx := "shadowed in block"
			_ = ctx
		}
	}()
}

// GO05: Uses ctx before shadow - valid usage
func goodUsesCtxBeforeShadowing(ctx context.Context) {
	go func() {
		_ = ctx.Done() // use ctx before shadowing
		ctx := "shadow"
		_ = ctx
	}()
}

// ===== MULTIPLE CONTEXT PARAMETERS =====

// GO08: Multiple ctx params - reports first (bad)
func twoContextParams(ctx1, ctx2 context.Context) {
	go func() { // want `goroutine does not propagate context "ctx1"`
		fmt.Println("ignoring both contexts")
	}()
}

// GO09: Multiple ctx params - uses first (good)
func goodUsesOneOfTwoContexts(ctx1, ctx2 context.Context) {
	go func() {
		_ = ctx1 // uses first context
	}()
}

// GO09b: Multiple ctx params - uses second (good)
func goodUsesSecondOfTwoContexts(ctx1, ctx2 context.Context) {
	go func() {
		_ = ctx2 // uses second context - should NOT report
	}()
}

// GO14: Context as non-first param (good)
func goodCtxAsSecondParam(logger interface{}, ctx context.Context) {
	go func() {
		_ = ctx // ctx is second param but still detected
	}()
}

// GO14b: Context as non-first param without use (bad)
func badCtxAsSecondParam(logger interface{}, ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		_ = logger
	}()
}

// ===== CONTEXT FROM LOCAL VARIABLE =====

// GO03b: No ctx param (local var) - not checked
func notCheckedLocalContextVariable() {
	// No context parameter, so not checked
	ctx := context.Background()
	go func() {
		fmt.Println("local context not checked")
	}()
	_ = ctx
}

// ===== CONTEXT PASSED AS ARGUMENT =====

// GO10b: Ctx passed as argument to goroutine
func goodGoroutinePassesCtxAsArg(ctx context.Context) {
	go func(c context.Context) {
		_ = c.Done() // uses its own param
	}(ctx)
}

// ===== DIRECT FUNCTION CALL =====

// GO11: Direct function call - not a func literal
func goodDirectFunctionCall(ctx context.Context) {
	go doSomething(ctx) // not a func literal
}

func doSomething(ctx context.Context) {
	_ = ctx
}

// ===== IGNORE DIRECTIVES =====

// GO06: Ignore directive - same line
func goodIgnoredSameLine(ctx context.Context) {
	go func() { //ctxrelay:ignore
		fmt.Println("ignored")
	}()
}

// GO07: Ignore directive - previous line
func goodIgnoredPreviousLine(ctx context.Context) {
	//ctxrelay:ignore
	go func() {
		fmt.Println("ignored")
	}()
}

// GO07b: Ignore directive - with reason
func goodIgnoredWithReason(ctx context.Context) {
	go func() { //ctxrelay:ignore - intentionally fire-and-forget
		fmt.Println("background task")
	}()
}
