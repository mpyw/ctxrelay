// Package goroutine contains test fixtures for the goroutine context propagation checker.
// This file covers adversarial patterns - unusual code that tests analyzer limits:
// 2+ level nesting, go fn()(), IIFE, interface methods, LIMITATION cases.
// See basic.go for daily patterns and advanced.go for real-world complex patterns.
package goroutine

import (
	"context"
	"fmt"
)

// ===== 2-LEVEL GOROUTINE NESTING =====

// Nested goroutine - outer uses ctx, inner doesn't
// Outer goroutine uses context but inner goroutine does not
func badNestedInner(ctx context.Context) {
	go func() {
		_ = ctx // outer goroutine uses ctx, but inner doesn't
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("inner")
		}()
	}()
}

// Nested goroutines - all use ctx
// All levels of nested goroutines properly use context
func goodNestedAllUseCtx(ctx context.Context) {
	go func() {
		_ = ctx
		go func() {
			_ = ctx
			go func() {
				_ = ctx
			}()
		}()
	}()
}

// 4-level nesting - last level missing ctx
// Four levels of nested goroutines where only the last level misses context
func badDeeplyNestedGoroutines(ctx context.Context) {
	go func() {
		_ = ctx
		go func() {
			_ = ctx
			go func() {
				_ = ctx
				go func() { // want `goroutine does not propagate context "ctx"`
					fmt.Println("level 4")
				}()
			}()
		}()
	}()
}

// ===== go fn()() HIGHER-ORDER PATTERNS =====

func makeWorker() func() {
	return func() {
		fmt.Println("worker")
	}
}

func makeWorkerWithCtx(ctx context.Context) func() {
	return func() {
		_ = ctx
	}
}

// go fn()() higher-order without ctx
// Higher-order function call in go statement without context
func badGoHigherOrder(ctx context.Context) {
	go makeWorker()() // want `goroutine does not propagate context "ctx"`
}

// go fn(ctx)() higher-order with ctx
// Higher-order function call in go statement with context
func goodGoHigherOrder(ctx context.Context) {
	go makeWorkerWithCtx(ctx)()
}

// go fn()()() triple higher-order without ctx
// Triple higher-order function call without context
func badGoHigherOrderTriple(ctx context.Context) {
	makeMaker := func() func() func() {
		return func() func() {
			return func() {
				fmt.Println("triple")
			}
		}
	}
	go makeMaker()()() // want `goroutine does not propagate context "ctx"`
}

// go fn(ctx)()() triple higher-order with ctx
// Triple higher-order function call with context
func goodGoHigherOrderTriple(ctx context.Context) {
	makeMaker := func(c context.Context) func() func() {
		return func() func() {
			return func() {
				_ = c
			}
		}
	}
	go makeMaker(ctx)()()
}

// Arbitrary depth go fn()()()...() without ctx
// Deep chain of function calls in go statement without context
func badGoInfiniteChain(ctx context.Context) {
	f := func() func() func() func() {
		return func() func() func() {
			return func() func() {
				return func() {
					fmt.Println("deep chain")
				}
			}
		}
	}
	go f()()()() // want `goroutine does not propagate context "ctx"`
}

// Arbitrary depth go fn(ctx)()()...() with ctx
// Deep chain of function calls in go statement with context
func goodGoInfiniteChain(ctx context.Context) {
	f := func(c context.Context) func() func() func() {
		return func() func() func() {
			return func() func() {
				return func() {
					_ = c
				}
			}
		}
	}
	go f(ctx)()()()
}

// ===== IIFE PATTERNS =====

// IIFE inside goroutine without ctx
// Goroutine with IIFE inside but no context usage
func badIIFEInsideGoroutine(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		func() {
			fmt.Println("iife")
		}()
	}()
}

// Ctx only in IIFE nested closure
// LIMITATION: Context used only in nested IIFE is not detected
func badIIFEUsesCtxInNestedFunc(ctx context.Context) {
	// ctx used only in nested IIFE, not by goroutine's direct body
	go func() { // want `goroutine does not propagate context "ctx"`
		func() {
			_ = ctx.Done()
		}()
	}()
}

// IIFE with ctx in goroutine body
// Goroutine uses context directly before IIFE
func goodIIFEWithCtxInGoroutineBody(ctx context.Context) {
	go func() {
		_ = ctx // ctx used directly
		func() {
			fmt.Println("iife")
		}()
	}()
}

// ===== INTERFACE METHOD PATTERNS =====

type Runner interface {
	Run()
}

type myRunner struct{}

func (r *myRunner) Run() {
	fmt.Println("running")
}

// Interface method without ctx
// Goroutine calls interface method without context
func badGoroutineCallsInterfaceMethod(ctx context.Context, r Runner) {
	go func() { // want `goroutine does not propagate context "ctx"`
		r.Run()
	}()
}

type CtxRunner interface {
	RunWithCtx(ctx context.Context)
}

type myCtxRunner struct{}

func (r *myCtxRunner) RunWithCtx(ctx context.Context) {
	_ = ctx
}

// Interface method with ctx
// Goroutine calls interface method with context
func goodGoroutineCallsInterfaceMethodWithCtx(ctx context.Context, r CtxRunner) {
	go func() {
		r.RunWithCtx(ctx)
	}()
}

// ===== TYPE ASSERTION IN GOROUTINE =====

// Type assertion without ctx
// Goroutine with type assertion but no context
func badGoroutineWithTypeAssertion(ctx context.Context) {
	var x interface{} = "hello"
	go func() { // want `goroutine does not propagate context "ctx"`
		if s, ok := x.(string); ok {
			fmt.Println(s)
		}
	}()
}

// ===== GOROUTINE IN EXPRESSION =====

// Goroutine in expression
// Goroutine inside function expression (not immediately invoked)
func badGoroutineInExpression(ctx context.Context) {
	_ = func() {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("in expression")
		}()
	}
}

// ===== MULTIPLE GOROUTINES PARALLEL =====

// Multiple parallel goroutines - mixed ctx usage
// Multiple parallel goroutines with mixed context usage
func badMultipleGoroutinesParallel(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		fmt.Println("first")
	}()

	go func() { // want `goroutine does not propagate context "ctx"`
		fmt.Println("second")
	}()

	go func() {
		_ = ctx
	}()
}

// ===== NESTED CLOSURE GOROUTINE PATTERNS =====
// These test goroutines inside nested closures - analyzer CAN trace ctx through FreeVar chains.

// Goroutine in nested closure WITH ctx
// Goroutine in nested closure with context properly captured via FreeVar chain
func goodNestedClosureWithCtx(ctx context.Context) {
	wrapper := func() {
		go func() {
			_ = ctx // ctx IS used via FreeVar chain - analyzer detects it
		}()
	}
	wrapper()
}

// Goroutine in nested closure WITHOUT ctx
// Goroutine in nested closure without context
func badNestedClosureWithoutCtx(ctx context.Context) {
	wrapper := func() {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("no ctx") // ctx NOT used
		}()
	}
	wrapper()
}

// Goroutine in deferred function WITH ctx
// Goroutine spawned from deferred function with context
func goodDeferredGoroutineWithCtx(ctx context.Context) {
	defer func() {
		go func() {
			_ = ctx // ctx IS used - analyzer detects it
		}()
	}()
}

// Goroutine in deferred function WITHOUT ctx
// Goroutine spawned from deferred function without context
func badDeferredGoroutineWithoutCtx(ctx context.Context) {
	defer func() {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("no ctx") // ctx NOT used
		}()
	}()
}

// Goroutine in init expression WITH ctx
// Goroutine in init expression with context properly captured
func goodGoroutineInInitWithCtx(ctx context.Context) {
	_ = func() func() {
		go func() {
			_ = ctx // ctx IS used - analyzer detects it
		}()
		return nil
	}()
}

// Goroutine in init expression WITHOUT ctx
// Goroutine in init expression without context
func badGoroutineInInitWithoutCtx(ctx context.Context) {
	_ = func() func() {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("no ctx") // ctx NOT used
		}()
		return nil
	}()
}

// ===== MULTIPLE CONTEXT EVIL PATTERNS =====

// Three contexts - uses middle one
// Function with three context parameters, uses middle one
// see also: errgroup, waitgroup
func goodUsesMiddleOfThreeContexts(ctx1, ctx2, ctx3 context.Context) {
	go func() {
		_ = ctx2 // uses middle context
	}()
}

// Three contexts - uses last one
// Function with three context parameters, uses last one
// see also: errgroup, waitgroup
func goodUsesLastOfThreeContexts(ctx1, ctx2, ctx3 context.Context) {
	go func() {
		_ = ctx3 // uses last context
	}()
}

// Multiple ctx in separate param groups
// Context parameters in separate groups, uses second
// see also: errgroup, waitgroup
func goodMultipleCtxSeparateGroups(a int, ctx1 context.Context, b string, ctx2 context.Context) {
	go func() {
		_ = ctx2 // uses second ctx from different param group
	}()
}

// Multiple ctx in separate param groups - none used
// Context parameters in separate groups, none used
// see also: errgroup, waitgroup
func badMultipleCtxSeparateGroups(a int, ctx1 context.Context, b string, ctx2 context.Context) {
	go func() { // want `goroutine does not propagate context "ctx1"`
		fmt.Println(a, b) // uses other params but not ctx
	}()
}

// Both contexts used
// Function with two context parameters, both used
// see also: errgroup, waitgroup
func goodUsesBothContexts(ctx1, ctx2 context.Context) {
	go func() {
		_ = ctx1
		_ = ctx2
	}()
}

// Nested goroutine - outer uses ctx1, inner uses ctx2
// Nested goroutines each using different context parameter
func goodNestedDifferentContexts(ctx1, ctx2 context.Context) {
	go func() {
		_ = ctx1 // outer uses ctx1
		go func() {
			_ = ctx2 // inner uses ctx2 - still valid!
		}()
	}()
}

// Nested goroutine - outer uses ctx2, inner uses neither
// Nested goroutine where inner uses neither context
func badNestedOnlyOuterUsesCtx(ctx1, ctx2 context.Context) {
	go func() {
		_ = ctx2 // outer uses ctx2
		go func() { // want `goroutine does not propagate context "ctx1"`
			fmt.Println("inner uses neither")
		}()
	}()
}

// Higher-order with multiple ctx - factory receives ctx1
// Higher-order function with multiple contexts, uses first
// see also: errgroup, waitgroup
func goodHigherOrderMultipleCtx(ctx1, ctx2 context.Context) {
	go makeWorkerWithCtx(ctx1)() // factory uses ctx1
}

// Higher-order with multiple ctx - factory receives ctx2
// Higher-order function with multiple contexts, uses second
// see also: errgroup, waitgroup
func goodHigherOrderMultipleCtxSecond(ctx1, ctx2 context.Context) {
	go makeWorkerWithCtx(ctx2)() // factory uses ctx2
}

// ===== IIFE AND ARGUMENT-PASSING PATTERNS =====
// These test goroutines inside IIFEs and context passed via arguments

// Goroutine inside IIFE without ctx
// Goroutine inside IIFE without context
func go80GoroutineInIIFEBad(ctx context.Context) {
	func() {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("goroutine in IIFE")
		}()
	}()
}

// Goroutine inside IIFE with ctx
// Goroutine inside IIFE with context properly captured
func go80GoroutineInIIFEGood(ctx context.Context) {
	func() {
		go func() {
			_ = ctx.Done()
		}()
	}()
}

// Context passed via argument - inner shadows outer
// Context passed to inner function that uses it (shadowing pattern)
func go81ArgumentShadowing(outerCtx context.Context) {
	func(ctx context.Context) {
		go func() {
			_ = ctx // uses inner ctx (shadowing)
		}()
	}(outerCtx)
}

// Context passed via argument - inner ignores it
// Context passed to inner function but ignored
func go81ArgumentShadowingBad(outerCtx context.Context) {
	func(ctx context.Context) {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("ignores inner ctx")
		}()
	}(outerCtx)
}

// Two levels of argument passing
// Context passed through two levels of function calls
func go82TwoLevelArguments(ctx1 context.Context) {
	func(ctx2 context.Context) {
		func(ctx3 context.Context) {
			go func() {
				_ = ctx3 // uses innermost ctx
			}()
		}(ctx2)
	}(ctx1)
}

// Two levels - innermost doesn't use
// Context passed through two levels but innermost doesn't use it
func go82TwoLevelArgumentsBad(ctx1 context.Context) {
	func(ctx2 context.Context) {
		func(ctx3 context.Context) {
			go func() { // want `goroutine does not propagate context "ctx3"`
				fmt.Println("ignores ctx3")
			}()
		}(ctx2)
	}(ctx1)
}

// Middle layer introduces ctx
// Middle layer introduces context when outer has none
func go83MiddleLayerIntroducesCtx() {
	func(ctx context.Context) {
		go func() {
			_ = ctx
		}()
	}(context.Background())
}

// Middle layer introduces ctx - goroutine ignores
// Middle layer introduces context but goroutine ignores it
func go83MiddleLayerIntroducesCtxBad() {
	func(ctx context.Context) {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("ignores middle ctx")
		}()
	}(context.Background())
}

// Interleaved layers - ctx -> no ctx -> ctx (shadowing) -> goroutine
// Interleaved layers with context shadowing
func go84InterleavedLayers(outerCtx context.Context) {
	func() {
		func(middleCtx context.Context) {
			go func() {
				_ = middleCtx // uses shadowing ctx
			}()
		}(outerCtx)
	}()
}

// Interleaved layers - goroutine ignores shadowing ctx
// Interleaved layers where goroutine ignores the shadowing context
func go84InterleavedLayersBad(outerCtx context.Context) {
	func() {
		func(middleCtx context.Context) {
			go func() { // want `goroutine does not propagate context "middleCtx"`
				fmt.Println("ignores middleCtx")
			}()
		}(outerCtx)
	}()
}
