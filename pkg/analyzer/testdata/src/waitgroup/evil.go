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

// GW12: Variable func without ctx
func badVariableFunc(ctx context.Context) {
	var wg sync.WaitGroup
	fn := func() {
		fmt.Println("no ctx")
	}
	wg.Go(fn) // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	wg.Wait()
}

// GW12b: Variable func with ctx
func goodVariableFuncWithCtx(ctx context.Context) {
	var wg sync.WaitGroup
	fn := func() {
		_ = ctx
	}
	wg.Go(fn) // OK - fn uses ctx
	wg.Wait()
}

// GW13: Higher-order func without ctx
func badHigherOrderFunc(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(makeWorker()) // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	wg.Wait()
}

// GW13b: Higher-order func with ctx
func goodHigherOrderFuncWithCtx(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(makeWorkerWithCtx(ctx)) // OK - makeWorkerWithCtx captures ctx
	wg.Wait()
}

// ===== STRUCT FIELD / SLICE / MAP TRACKING =====
// These patterns CAN be tracked when defined in the same function.

// GW18: Struct field with ctx
type taskHolderWithCtx struct {
	task func()
}

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

// GW15: Slice index with ctx
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

// GW16: Map key with ctx
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

// GW100: Interface method with ctx argument (good)
// When ctx is passed as argument, analyzer detects ctx usage.
type WorkerFactory interface {
	CreateWorker(ctx context.Context) func()
}

type myFactory struct{}

func (f *myFactory) CreateWorker(ctx context.Context) func() {
	return func() {
		_ = ctx // Implementation captures ctx
	}
}

func goodInterfaceMethodWithCtxArg(ctx context.Context, factory WorkerFactory) {
	var wg sync.WaitGroup
	// ctx IS passed as argument to CreateWorker - analyzer detects ctx usage
	wg.Go(factory.CreateWorker(ctx)) // OK - ctx passed as argument
	wg.Wait()
}

// GW100b: Interface method without ctx argument (bad)
type WorkerFactoryNoCtx interface {
	CreateWorker() func()
}

func badInterfaceMethodWithoutCtxArg(ctx context.Context, factory WorkerFactoryNoCtx) {
	var wg sync.WaitGroup
	// ctx NOT passed to CreateWorker - expected to fail
	wg.Go(factory.CreateWorker()) // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	wg.Wait()
}

// ===== REMAINING LIMITATIONS =====
// These patterns cannot be tracked statically.
// LIMITATION = false positive: ctx IS used but analyzer can't detect it.

// GW101: Function passed through parameter - NOW SUPPORTED via directive
//
//ctxrelay:goroutine_creator
func runWithWaitGroup(wg *sync.WaitGroup, fn func()) {
	wg.Go(fn)
}

// GW101a: Function with ctx passed through creator - should pass
func goodFuncPassedThroughCreator(ctx context.Context) {
	var wg sync.WaitGroup
	fn := func() {
		_ = ctx // fn uses ctx
	}
	runWithWaitGroup(&wg, fn) // OK - fn uses ctx, and runWithWaitGroup is marked as creator
	wg.Wait()
}

// GW101b: Function without ctx passed through creator - should report
func badFuncPassedThroughCreator(ctx context.Context) {
	var wg sync.WaitGroup
	fn := func() {
		fmt.Println("no ctx")
	}
	runWithWaitGroup(&wg, fn) // want `runWithWaitGroup\(\) func argument should use context "ctx"`
	wg.Wait()
}

// GW102: LIMITATION - Function from channel - ctx captured but not traced
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

// GW103: Function from struct field without ctx - NOW TRACKED
type taskHolder struct {
	task func()
}

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

// GW104: Function from map without ctx - NOW TRACKED
func badMapValueWithoutCtx(ctx context.Context) {
	var wg sync.WaitGroup
	tasks := map[string]func(){
		"task1": func() {},
	}
	wg.Go(tasks["task1"]) // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	wg.Wait()
}

// GW105: Function from slice without ctx - NOW TRACKED
func badSliceValueWithoutCtx(ctx context.Context) {
	var wg sync.WaitGroup
	tasks := []func(){
		func() {},
	}
	wg.Go(tasks[0]) // want `sync.WaitGroup.Go\(\) closure should use context "ctx"`
	wg.Wait()
}

// GW108: LIMITATION - Function through interface{} type assertion - ctx captured but not traced
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

// Control: same pattern without ctx
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

// GW70: Three contexts - uses middle one (good)
func goodUsesMiddleOfThreeContexts(ctx1, ctx2, ctx3 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx2 // uses middle context
	})
	wg.Wait()
}

// GW71: Three contexts - uses last one (good)
func goodUsesLastOfThreeContexts(ctx1, ctx2, ctx3 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx3 // uses last context
	})
	wg.Wait()
}

// GW72: Multiple ctx in separate param groups (good)
func goodMultipleCtxSeparateGroups(a int, ctx1 context.Context, b string, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx2 // uses second ctx from different param group
	})
	wg.Wait()
}

// GW73: Multiple ctx in separate param groups - none used (bad)
func badMultipleCtxSeparateGroups(a int, ctx1 context.Context, b string, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() { // want `sync.WaitGroup.Go\(\) closure should use context "ctx1"`
		fmt.Println(a, b) // uses other params but not ctx
	})
	wg.Wait()
}

// GW74: Both contexts used (good)
func goodUsesBothContexts(ctx1, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		_ = ctx1
		_ = ctx2
	})
	wg.Wait()
}

// GW85: Higher-order with multiple ctx - factory receives ctx1 (good)
func goodHigherOrderMultipleCtx(ctx1, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(makeWorkerWithCtx(ctx1)) // factory uses ctx1
	wg.Wait()
}

// GW86: Higher-order with multiple ctx - factory receives ctx2 (good)
func goodHigherOrderMultipleCtxSecond(ctx1, ctx2 context.Context) {
	var wg sync.WaitGroup
	wg.Go(makeWorkerWithCtx(ctx2)) // factory uses ctx2
	wg.Wait()
}

// ===== ADVANCED NESTED PATTERNS (SHADOWING, ARGUMENT PASSING) =====

// Shadowing - inner ctx shadows outer
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
