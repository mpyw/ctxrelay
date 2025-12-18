// Package waitgroup contains test fixtures for the waitgroup context propagation checker.
// This file covers adversarial patterns - tests analyzer limits: higher-order functions,
// non-literal function arguments, interface methods.
// See basic.go for daily patterns and advanced.go for real-world complex patterns.
package waitgroup

import (
	"context"
	"fmt"
	"sync"
)

// ===== HIGHER-ORDER FUNCTION PATTERNS =====

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

// Variable func without ctx
// Function stored in variable does not use context
// see also: errgroup
func badVariableFunc(ctx context.Context) {
	var wg sync.WaitGroup
	fn := func() {
		fmt.Println("no ctx")
	}
	wg.Go(fn) // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	wg.Wait()
}

// Variable func with ctx
// Function stored in variable uses context
// see also: errgroup
func goodVariableFuncWithCtx(ctx context.Context) {
	var wg sync.WaitGroup
	fn := func() {
		_ = ctx
	}
	wg.Go(fn) // OK - fn uses ctx
	wg.Wait()
}

// Higher-order func without ctx
// Higher-order function that returns closure without context
// see also: errgroup
func badHigherOrderFunc(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(makeWorker()) // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	wg.Wait()
}

// Higher-order func with ctx
// Higher-order function that returns closure with context
// see also: errgroup
func goodHigherOrderFuncWithCtx(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(makeWorkerWithCtx(ctx)) // OK - makeWorkerWithCtx captures ctx
	wg.Wait()
}

// ===== STRUCT FIELD / SLICE / MAP TRACKING =====
// These patterns CAN be tracked when defined in the same function.

// Struct field with ctx
// Function from struct field that uses context
// see also: goroutine, errgroup
type taskHolderWithCtx struct {
	task func()
}

// Struct field func with ctx
// Function from struct field that uses context
// see also: errgroup
func goodStructFieldWithCtx(ctx context.Context) {
	var wg sync.WaitGroup
	holder := taskHolderWithCtx{
		task: func() {
			_ = ctx // Uses ctx
		},
	}
	wg.Go(holder.task) // OK - now tracked
	wg.Wait()
}

// Slice index with ctx
// Function from slice that uses context
// see also: errgroup
func goodSliceIndexWithCtx(ctx context.Context) {
	var wg sync.WaitGroup
	tasks := []func(){
		func() {
			_ = ctx // Uses ctx
		},
	}
	wg.Go(tasks[0]) // OK - now tracked
	wg.Wait()
}

// Map key with ctx
// Function from map that uses context
// see also: errgroup
func goodMapKeyWithCtx(ctx context.Context) {
	var wg sync.WaitGroup
	tasks := map[string]func(){
		"key": func() {
			_ = ctx // Uses ctx
		},
	}
	wg.Go(tasks["key"]) // OK - now tracked
	wg.Wait()
}

// ===== INTERFACE METHOD PATTERNS =====
// ctx passed as argument to interface method IS detected by the analyzer.

// Interface method with ctx argument
// When ctx is passed as argument, analyzer detects ctx usage
// see also: goroutine, errgroup
type WorkerFactory interface {
	CreateWorker(ctx context.Context) func()
}

type myFactory struct{}

func (f *myFactory) CreateWorker(ctx context.Context) func() {
	return func() {
		_ = ctx // Implementation captures ctx
	}
}

// Interface method with ctx arg
// Interface method receives context as argument
// see also: errgroup
func goodInterfaceMethodWithCtxArg(ctx context.Context, factory WorkerFactory) {
	var wg sync.WaitGroup
	// ctx IS passed as argument to CreateWorker - analyzer detects ctx usage
	wg.Go(factory.CreateWorker(ctx)) // OK - ctx passed as argument
	wg.Wait()
}

// Interface method without ctx argument
// Interface method that does not receive context
// see also: goroutine, errgroup
type WorkerFactoryNoCtx interface {
	CreateWorker() func()
}

// Interface method without ctx arg
// Interface method that does not receive context
// see also: errgroup
func badInterfaceMethodWithoutCtxArg(ctx context.Context, factory WorkerFactoryNoCtx) {
	var wg sync.WaitGroup
	// ctx NOT passed to CreateWorker - expected to fail
	wg.Go(factory.CreateWorker()) // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	wg.Wait()
}

// ===== REMAINING LIMITATIONS =====
// These patterns cannot be tracked statically.
// LIMITATION = false positive: ctx IS used but analyzer can't detect it.

// Function passed through parameter - NOW SUPPORTED via directive
//
//goroutinectx:goroutine_creator
func runWithWaitGroup(wg *sync.WaitGroup, fn func()) {
	wg.Go(fn)
}

// Function with ctx passed through creator
// Function with context passed through goroutine creator helper
// see also: errgroup
func goodFuncPassedThroughCreator(ctx context.Context) {
	var wg sync.WaitGroup
	fn := func() {
		_ = ctx // fn uses ctx
	}
	runWithWaitGroup(&wg, fn) // OK - fn uses ctx, and runWithWaitGroup is marked as creator
	wg.Wait()
}

// Function without ctx passed through creator
// Function without context passed through goroutine creator helper
// see also: errgroup
func badFuncPassedThroughCreator(ctx context.Context) {
	var wg sync.WaitGroup
	fn := func() {
		fmt.Println("no ctx")
	}
	runWithWaitGroup(&wg, fn) // want `runWithWaitGroup\(\) func argument should use context "ctx"`
	wg.Wait()
}

// LIMITATION - Function from channel
// Context captured but not traced through channel receive
// see also: errgroup
func limitationFuncFromChannel(ctx context.Context) {
	var wg sync.WaitGroup
	ch := make(chan func(), 1)
	ch <- func() {
		_ = ctx // The func DOES capture ctx
	}
	fn := <-ch
	// LIMITATION: fn captures ctx, but analyzer can't trace through channel receive
	wg.Go(fn) // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	wg.Wait()
}

// Function from struct field without ctx
// Function from struct field that does not use context
// see also: goroutine, errgroup
type taskHolder struct {
	task func()
}

// Struct field func without ctx
// Function from struct field that does not use context
// see also: errgroup
func badStructFieldWithoutCtx(ctx context.Context) {
	var wg sync.WaitGroup
	holder := taskHolder{
		task: func() {
			fmt.Println("no ctx")
		},
	}
	wg.Go(holder.task) // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	wg.Wait()
}

// Function from map without ctx
// Function from map that does not use context
// see also: errgroup
func badMapValueWithoutCtx(ctx context.Context) {
	var wg sync.WaitGroup
	tasks := map[string]func(){
		"task1": func() {},
	}
	wg.Go(tasks["task1"]) // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	wg.Wait()
}

// Function from slice without ctx
// Function from slice that does not use context
// see also: errgroup
func badSliceValueWithoutCtx(ctx context.Context) {
	var wg sync.WaitGroup
	tasks := []func(){
		func() {},
	}
	wg.Go(tasks[0]) // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	wg.Wait()
}

// LIMITATION - Function through interface{} type assertion
// Context captured but not traced through interface{} assertion
// see also: errgroup
func limitationFuncThroughInterfaceWithCtx(ctx context.Context) {
	var wg sync.WaitGroup

	var i interface{} = func() {
		_ = ctx // fn DOES capture ctx
	}

	// Type assert to get func back
	fn := i.(func())
	// LIMITATION: fn captures ctx, but analyzer can't trace through interface{} assertion
	wg.Go(fn) // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	wg.Wait()
}

// Function through interface{} - control case without ctx
// Function through interface{} that does not use context
// see also: errgroup
func badFuncThroughInterfaceWithoutCtx(ctx context.Context) {
	var wg sync.WaitGroup

	var i interface{} = func() {
		fmt.Println("no ctx") // fn does NOT use ctx
	}

	fn := i.(func())
	wg.Go(fn) // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	wg.Wait()
}

// ===== MULTIPLE CONTEXT EVIL PATTERNS =====

// Three contexts - uses middle one
// Function with three context parameters, uses middle one
// see also: goroutine, errgroup
func goodUsesMiddleOfThreeContexts(ctx1, ctx2, ctx3 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx2 // uses middle context
	})
	wg.Wait()
}

// Three contexts - uses last one
// Function with three context parameters, uses last one
// see also: goroutine, errgroup
func goodUsesLastOfThreeContexts(ctx1, ctx2, ctx3 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx3 // uses last context
	})
	wg.Wait()
}

// Multiple ctx in separate param groups
// Context parameters in separate groups, uses second
// see also: goroutine, errgroup
func goodMultipleCtxSeparateGroups(a int, ctx1 context.Context, b string, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx2 // uses second ctx from different param group
	})
	wg.Wait()
}

// Multiple ctx in separate param groups - none used
// Context parameters in separate groups, none used
// see also: goroutine, errgroup
func badMultipleCtxSeparateGroups(a int, ctx1 context.Context, b string, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx1"`
		fmt.Println(a, b) // uses other params but not ctx
	})
	wg.Wait()
}

// Both contexts used
// Function with two context parameters, both used
// see also: goroutine, errgroup
func goodUsesBothContexts(ctx1, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx1
		_ = ctx2
	})
	wg.Wait()
}

// Higher-order with multiple ctx - factory receives ctx1
// Higher-order function with multiple contexts, uses first
// see also: goroutine, errgroup
func goodHigherOrderMultipleCtx(ctx1, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(makeWorkerWithCtx(ctx1)) // factory uses ctx1
	wg.Wait()
}

// Higher-order with multiple ctx - factory receives ctx2
// Higher-order function with multiple contexts, uses second
// see also: goroutine, errgroup
func goodHigherOrderMultipleCtxSecond(ctx1, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(makeWorkerWithCtx(ctx2)) // factory uses ctx2
	wg.Wait()
}

// ===== ADVANCED NESTED PATTERNS (SHADOWING, ARGUMENT PASSING) =====

// Shadowing - inner ctx shadows outer
// Inner function with its own context parameter uses it
// see also: errgroup
func evilShadowingInnerHasCtx(outerCtx context.Context) {
	innerFunc := func(ctx context.Context) {
		var wg sync.WaitGroup
		wg.Go(func() {
			_ = ctx // uses inner ctx
		})
		wg.Wait()
	}
	innerFunc(outerCtx)
}

// Shadowing - inner ignores ctx
// Inner function with its own context parameter ignores it
// see also: errgroup
func evilShadowingInnerIgnoresCtx(outerCtx context.Context) {
	innerFunc := func(ctx context.Context) {
		var wg sync.WaitGroup
		wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
		})
		wg.Wait()
	}
	innerFunc(outerCtx)
}

// Two levels of shadowing
// Two levels of argument passing with context shadowing
// see also: errgroup
func evilShadowingTwoLevels(ctx1 context.Context) {
	func(ctx2 context.Context) {
		func(ctx3 context.Context) {
			var wg sync.WaitGroup
			wg.Go(func() {
				_ = ctx3 // uses ctx3
			})
			wg.Wait()
		}(ctx2)
	}(ctx1)
}

// Two levels of shadowing - innermost ignores ctx
// Two levels of shadowing where innermost closure ignores context
// see also: errgroup
func evilShadowingTwoLevelsBad(ctx1 context.Context) {
	func(ctx2 context.Context) {
		func(ctx3 context.Context) {
			var wg sync.WaitGroup
			wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx3"`
			})
			wg.Wait()
		}(ctx2)
	}(ctx1)
}

// ===== MIDDLE LAYER INTRODUCES CTX (OUTER HAS NONE) =====

// Middle layer introduces ctx - mixed usage
// Middle layer introduces context, some closures use it and some don't
// see also: errgroup
func evilMiddleLayerIntroducesCtx() {
	func(ctx context.Context) {
		var wg sync.WaitGroup
		wg.Go(func() {
			_ = ctx
		})
		func() {
			wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
			})
		}()
		wg.Wait()
	}(context.Background())
}

// Middle layer introduces ctx - good
// Middle layer introduces context and nested closure properly uses it
// see also: errgroup
func evilMiddleLayerIntroducesCtxGood() {
	func(ctx context.Context) {
		var wg sync.WaitGroup
		func() {
			wg.Go(func() {
				_ = ctx
			})
		}()
		wg.Wait()
	}(context.Background())
}

// ===== INTERLEAVED LAYERS (ctx -> no ctx -> ctx shadowing) =====

// Interleaved layers - goroutine ignores shadowing ctx
// Interleaved layers where goroutine ignores the shadowing context
// see also: errgroup
func evilInterleavedLayers(outerCtx context.Context) {
	func() {
		func(middleCtx context.Context) {
			var wg sync.WaitGroup
			func() {
				wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "middleCtx"`
				})
			}()
			wg.Wait()
		}(outerCtx)
	}()
}

// Interleaved layers - good
// Interleaved layers where goroutine properly uses shadowing context
// see also: errgroup
func evilInterleavedLayersGood(outerCtx context.Context) {
	func() {
		func(middleCtx context.Context) {
			var wg sync.WaitGroup
			func() {
				wg.Go(func() {
					_ = middleCtx
				})
			}()
			wg.Wait()
		}(outerCtx)
	}()
}
