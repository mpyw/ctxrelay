// Package goroutine contains test fixtures for the goroutine context propagation checker.
// This file covers basic/daily patterns - single goroutine, shadowing, ignore directives.
// See advanced.go for real-world complex patterns and evil.go for adversarial tests.
package goroutine

import (
	"context"
	"fmt"
)

// ===== SHOULD REPORT =====

// Literal without ctx - basic bad case
// Goroutine func literal does not capture context
func badGoroutineNoCapture(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		fmt.Println("no context")
	}()
}

// Literal without ctx - variant
// Goroutine ignores available context
func badGoroutineIgnoresCtx(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		x := 1
		_ = x
	}()
}

// ===== SHOULD NOT REPORT =====

// Literal with ctx - basic good case
// Goroutine captures and uses context
func goodGoroutineCapturesCtx(ctx context.Context) {
	go func() {
		_ = ctx.Done()
	}()
}

// Literal with ctx - via function call
// Goroutine uses context via function call
func goodGoroutineUsesCtxInCall(ctx context.Context) {
	go func() {
		doSomething(ctx)
	}()
}

// No ctx param
// Function has no context parameter - not checked
// see also: errgroup, waitgroup
func goodNoContextParam() {
	go func() {
		fmt.Println("hello")
	}()
}

// Literal with derived ctx
// Goroutine uses derived context (WithCancel, etc.)
func goodGoroutineWithDerivedCtx(ctx context.Context) {
	go func() {
		ctx2, cancel := context.WithCancel(ctx)
		defer cancel()
		_ = ctx2
	}()
}

// Literal with ctx in select
// Goroutine uses context in select statement
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

// Shadow with non-ctx type
// Context variable is shadowed with non-context type (string)
// see also: errgroup, waitgroup
func badShadowingNonContext(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		ctx := "not a context" // shadows with string
		_ = ctx
	}()
}

// Inner func has own ctx param
// Inner function has its own context parameter
func goodShadowingInnerCtxParam(outerCtx context.Context) {
	go func(ctx context.Context) {
		_ = ctx.Done() // uses inner ctx - OK
	}(outerCtx)
}

// Shadow with non-ctx type (channel)
// Context shadowed with channel type
func badShadowingWithDifferentType(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		ctx := make(chan int) // shadows with channel
		close(ctx)
	}()
}

// Shadow with non-ctx type (function)
// Context shadowed with function type
func badShadowingWithFunction(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		ctx := func() {} // shadows with function
		ctx()
	}()
}

// Shadow in nested block
// Context shadowed in nested block with non-context type
func badShadowingInNestedBlock(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		if true {
			ctx := "shadowed in block"
			_ = ctx
		}
	}()
}

// Uses ctx before shadow
// Uses context before shadowing it - valid usage
// see also: errgroup, waitgroup
func goodUsesCtxBeforeShadowing(ctx context.Context) {
	go func() {
		_ = ctx.Done() // use ctx before shadowing
		ctx := "shadow"
		_ = ctx
	}()
}

// ===== MULTIPLE CONTEXT PARAMETERS =====

// Multiple ctx params - reports first
// Function has two context parameters, reports first one when neither used
// see also: errgroup, waitgroup
func twoContextParams(ctx1, ctx2 context.Context) {
	go func() { // want `goroutine does not propagate context "ctx1"`
		fmt.Println("ignoring both contexts")
	}()
}

// Multiple ctx params - uses first
// Function has multiple context parameters and uses first
// see also: errgroup, waitgroup
func goodUsesOneOfTwoContexts(ctx1, ctx2 context.Context) {
	go func() {
		_ = ctx1 // uses first context
	}()
}

// Multiple ctx params - uses second
// Function has multiple context parameters and uses second - should NOT report
// see also: errgroup, waitgroup
func goodUsesSecondOfTwoContexts(ctx1, ctx2 context.Context) {
	go func() {
		_ = ctx2 // uses second context - should NOT report
	}()
}

// Context as non-first param
// Context is second parameter and is properly used
// see also: errgroup, waitgroup
func goodCtxAsSecondParam(logger interface{}, ctx context.Context) {
	go func() {
		_ = ctx // ctx is second param but still detected
	}()
}

// Context as non-first param without use
// Context is second parameter but not used in closure
// see also: errgroup, waitgroup
func badCtxAsSecondParam(logger interface{}, ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		_ = logger
	}()
}

// ===== CONTEXT FROM LOCAL VARIABLE =====

// No ctx param (local var)
// Context from local variable is not checked (no param)
func notCheckedLocalContextVariable() {
	// No context parameter, so not checked
	ctx := context.Background()
	go func() {
		fmt.Println("local context not checked")
	}()
	_ = ctx
}

// ===== CONTEXT PASSED AS ARGUMENT =====

// Ctx passed as argument to goroutine
// Context passed as argument to goroutine function
func goodGoroutinePassesCtxAsArg(ctx context.Context) {
	go func(c context.Context) {
		_ = c.Done() // uses its own param
	}(ctx)
}

// ===== DIRECT FUNCTION CALL =====

// Direct function call
// go statement with direct function call (not func literal)
func goodDirectFunctionCall(ctx context.Context) {
	go doSomething(ctx) // not a func literal
}

func doSomething(ctx context.Context) {
	_ = ctx
}

// ===== IGNORE DIRECTIVES =====

// Ignore directive - same line
// Ignore directive on the same line suppresses warning
// see also: errgroup, waitgroup
func goodIgnoredSameLine(ctx context.Context) {
	go func() { //goroutinectx:ignore
		fmt.Println("ignored")
	}()
}

// Ignore directive - previous line
// Ignore directive on the previous line suppresses warning
// see also: errgroup, waitgroup
func goodIgnoredPreviousLine(ctx context.Context) {
	//goroutinectx:ignore
	go func() {
		fmt.Println("ignored")
	}()
}

// Ignore directive - with reason
// Ignore directive with explanatory reason
func goodIgnoredWithReason(ctx context.Context) {
	go func() { //goroutinectx:ignore - intentionally fire-and-forget
		fmt.Println("background task")
	}()
}
