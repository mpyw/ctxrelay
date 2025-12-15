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

// GO40: Nested goroutine - outer uses ctx, inner doesn't
func badNestedInner(ctx context.Context) {
	go func() {
		_ = ctx // outer goroutine uses ctx, but inner doesn't
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("inner")
		}()
	}()
}

// GO41: 3-level nesting - only first uses ctx
func badNestedDeep(ctx context.Context) {
	go func() {
		_ = ctx
		go func() { // want `goroutine does not propagate context "ctx"`
			go func() { // want `goroutine does not propagate context "ctx"`
				fmt.Println("deep")
			}()
		}()
	}()
}

// GO42: Nested goroutines - all use ctx
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

// GO43: 4-level nesting - last level missing ctx
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

// GO50: go fn()() - higher-order without ctx
func badGoHigherOrder(ctx context.Context) {
	go makeWorker()() // want `goroutine does not propagate context "ctx"`
}

// GO50b: go fn(ctx)() - higher-order with ctx
func goodGoHigherOrder(ctx context.Context) {
	go makeWorkerWithCtx(ctx)()
}

// GO51: go fn()()() - triple higher-order without ctx
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

// GO51b: go fn(ctx)()() - triple higher-order with ctx
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

// GO52: Arbitrary depth go fn()()()...() without ctx
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

// GO52b: Arbitrary depth go fn(ctx)()()...() with ctx
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

// GO60: IIFE inside goroutine without ctx
func badIIFEInsideGoroutine(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		func() {
			fmt.Println("iife")
		}()
	}()
}

// GO61: Ctx only in IIFE nested closure (LIMITATION)
func badIIFEUsesCtxInNestedFunc(ctx context.Context) {
	// ctx used only in nested IIFE, not by goroutine's direct body
	go func() { // want `goroutine does not propagate context "ctx"`
		func() {
			_ = ctx.Done()
		}()
	}()
}

// GO60b: IIFE with ctx in goroutine body
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

// GO62: Interface method without ctx
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

// GO62b: Interface method with ctx
func goodGoroutineCallsInterfaceMethodWithCtx(ctx context.Context, r CtxRunner) {
	go func() {
		r.RunWithCtx(ctx)
	}()
}

// ===== TYPE ASSERTION IN GOROUTINE =====

// GO63: Type assertion without ctx
func badGoroutineWithTypeAssertion(ctx context.Context) {
	var x interface{} = "hello"
	go func() { // want `goroutine does not propagate context "ctx"`
		if s, ok := x.(string); ok {
			fmt.Println(s)
		}
	}()
}

// ===== GOROUTINE IN EXPRESSION =====

// GO64: Goroutine in expression (not immediately invoked)
func badGoroutineInExpression(ctx context.Context) {
	_ = func() {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("in expression")
		}()
	}
}

// ===== MULTIPLE GOROUTINES PARALLEL =====

// GO65: Multiple parallel goroutines - mixed ctx usage
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

// GO90: Goroutine in nested closure WITH ctx (good)
func goodNestedClosureWithCtx(ctx context.Context) {
	wrapper := func() {
		go func() {
			_ = ctx // ctx IS used via FreeVar chain - analyzer detects it
		}()
	}
	wrapper()
}

// GO90b: Goroutine in nested closure WITHOUT ctx (bad)
func badNestedClosureWithoutCtx(ctx context.Context) {
	wrapper := func() {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("no ctx") // ctx NOT used
		}()
	}
	wrapper()
}

// GO91: Goroutine in deferred function WITH ctx (good)
func goodDeferredGoroutineWithCtx(ctx context.Context) {
	defer func() {
		go func() {
			_ = ctx // ctx IS used - analyzer detects it
		}()
	}()
}

// GO91b: Goroutine in deferred function WITHOUT ctx (bad)
func badDeferredGoroutineWithoutCtx(ctx context.Context) {
	defer func() {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("no ctx") // ctx NOT used
		}()
	}()
}

// GO92: Goroutine in init expression WITH ctx (good)
func goodGoroutineInInitWithCtx(ctx context.Context) {
	_ = func() func() {
		go func() {
			_ = ctx // ctx IS used - analyzer detects it
		}()
		return nil
	}()
}

// GO92b: Goroutine in init expression WITHOUT ctx (bad)
func badGoroutineInInitWithoutCtx(ctx context.Context) {
	_ = func() func() {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("no ctx") // ctx NOT used
		}()
		return nil
	}()
}

// ===== MULTIPLE CONTEXT EVIL PATTERNS =====

// GO70: Three contexts - uses middle one (good)
func goodUsesMiddleOfThreeContexts(ctx1, ctx2, ctx3 context.Context) {
	go func() {
		_ = ctx2 // uses middle context
	}()
}

// GO71: Three contexts - uses last one (good)
func goodUsesLastOfThreeContexts(ctx1, ctx2, ctx3 context.Context) {
	go func() {
		_ = ctx3 // uses last context
	}()
}

// GO72: Multiple ctx in separate param groups (good)
func goodMultipleCtxSeparateGroups(a int, ctx1 context.Context, b string, ctx2 context.Context) {
	go func() {
		_ = ctx2 // uses second ctx from different param group
	}()
}

// GO73: Multiple ctx in separate param groups - none used (bad)
func badMultipleCtxSeparateGroups(a int, ctx1 context.Context, b string, ctx2 context.Context) {
	go func() { // want `goroutine does not propagate context "ctx1"`
		fmt.Println(a, b) // uses other params but not ctx
	}()
}

// GO74: Both contexts used (good)
func goodUsesBothContexts(ctx1, ctx2 context.Context) {
	go func() {
		_ = ctx1
		_ = ctx2
	}()
}

// GO75: Nested goroutine - outer uses ctx1, inner uses ctx2 (good)
func goodNestedDifferentContexts(ctx1, ctx2 context.Context) {
	go func() {
		_ = ctx1 // outer uses ctx1
		go func() {
			_ = ctx2 // inner uses ctx2 - still valid!
		}()
	}()
}

// GO76: Nested goroutine - outer uses ctx2, inner uses neither (bad inner)
func badNestedOnlyOuterUsesCtx(ctx1, ctx2 context.Context) {
	go func() {
		_ = ctx2 // outer uses ctx2
		go func() { // want `goroutine does not propagate context "ctx1"`
			fmt.Println("inner uses neither")
		}()
	}()
}

// GO77: Higher-order with multiple ctx - factory receives ctx1 (good)
func goodHigherOrderMultipleCtx(ctx1, ctx2 context.Context) {
	go makeWorkerWithCtx(ctx1)() // factory uses ctx1
}

// GO78: Higher-order with multiple ctx - factory receives ctx2 (good)
func goodHigherOrderMultipleCtxSecond(ctx1, ctx2 context.Context) {
	go makeWorkerWithCtx(ctx2)() // factory uses ctx2
}

// ===== IIFE AND ARGUMENT-PASSING PATTERNS =====
// These test goroutines inside IIFEs and context passed via arguments

// GO80: Goroutine inside IIFE without ctx
func go80GoroutineInIIFEBad(ctx context.Context) {
	func() {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("goroutine in IIFE")
		}()
	}()
}

// GO80b: Goroutine inside IIFE with ctx
func go80GoroutineInIIFEGood(ctx context.Context) {
	func() {
		go func() {
			_ = ctx.Done()
		}()
	}()
}

// GO81: Context passed via argument - inner shadows outer
func go81ArgumentShadowing(outerCtx context.Context) {
	func(ctx context.Context) {
		go func() {
			_ = ctx // uses inner ctx (shadowing)
		}()
	}(outerCtx)
}

// GO81b: Context passed via argument - inner ignores it
func go81ArgumentShadowingBad(outerCtx context.Context) {
	func(ctx context.Context) {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("ignores inner ctx")
		}()
	}(outerCtx)
}

// GO82: Two levels of argument passing
func go82TwoLevelArguments(ctx1 context.Context) {
	func(ctx2 context.Context) {
		func(ctx3 context.Context) {
			go func() {
				_ = ctx3 // uses innermost ctx
			}()
		}(ctx2)
	}(ctx1)
}

// GO82b: Two levels - innermost doesn't use
func go82TwoLevelArgumentsBad(ctx1 context.Context) {
	func(ctx2 context.Context) {
		func(ctx3 context.Context) {
			go func() { // want `goroutine does not propagate context "ctx3"`
				fmt.Println("ignores ctx3")
			}()
		}(ctx2)
	}(ctx1)
}

// GO83: Middle layer introduces ctx (outer has none)
func go83MiddleLayerIntroducesCtx() {
	func(ctx context.Context) {
		go func() {
			_ = ctx
		}()
	}(context.Background())
}

// GO83b: Middle layer introduces ctx - goroutine ignores
func go83MiddleLayerIntroducesCtxBad() {
	func(ctx context.Context) {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("ignores middle ctx")
		}()
	}(context.Background())
}

// GO84: Interleaved layers - ctx -> no ctx -> ctx (shadowing) -> goroutine
func go84InterleavedLayers(outerCtx context.Context) {
	func() {
		func(middleCtx context.Context) {
			go func() {
				_ = middleCtx // uses shadowing ctx
			}()
		}(outerCtx)
	}()
}

// GO84b: Interleaved layers - goroutine ignores shadowing ctx
func go84InterleavedLayersBad(outerCtx context.Context) {
	func() {
		func(middleCtx context.Context) {
			go func() { // want `goroutine does not propagate context "middleCtx"`
				fmt.Println("ignores middleCtx")
			}()
		}(outerCtx)
	}()
}
