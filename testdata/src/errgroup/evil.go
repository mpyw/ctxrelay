// Package errgroup contains test fixtures for the errgroup context propagation checker.
// This file covers adversarial patterns - tests analyzer limits: higher-order functions,
// non-literal function arguments, interface methods.
// See basic.go for daily patterns and advanced.go for real-world complex patterns.
package errgroup

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

// ===== HIGHER-ORDER FUNCTION PATTERNS =====

func makeWorker() func() error {
	return func() error {
		fmt.Println("worker")
		return nil
	}
}

func makeWorkerWithCtx(ctx context.Context) func() error {
	return func() error {
		_ = ctx
		return nil
	}
}

// GE12: Variable func without ctx
func badVariableFunc(ctx context.Context) {
	g := new(errgroup.Group)
	fn := func() error {
		fmt.Println("no ctx")
		return nil
	}
	g.Go(fn) // want `errgroup.Group.Go\(\) closure should use context "ctx"`
	_ = g.Wait()
}

// GE12b: Variable func with ctx
func goodVariableFuncWithCtx(ctx context.Context) {
	g := new(errgroup.Group)
	fn := func() error {
		_ = ctx
		return nil
	}
	g.Go(fn) // OK - fn uses ctx
	_ = g.Wait()
}

// GE13: Higher-order func without ctx
func badHigherOrderFunc(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(makeWorker()) // want `errgroup.Group.Go\(\) closure should use context "ctx"`
	_ = g.Wait()
}

// GE13b: Higher-order func with ctx
func goodHigherOrderFuncWithCtx(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(makeWorkerWithCtx(ctx)) // OK - makeWorkerWithCtx captures ctx
	_ = g.Wait()
}

// ===== STRUCT FIELD / SLICE / MAP TRACKING =====
// These patterns CAN be tracked when defined in the same function.

// GE18: Struct field with ctx
type taskHolderWithCtx struct {
	task func() error
}

func goodStructFieldWithCtx(ctx context.Context) {
	g := new(errgroup.Group)
	holder := taskHolderWithCtx{
		task: func() error {
			_ = ctx // Uses ctx
			return nil
		},
	}
	g.Go(holder.task) // OK - now tracked
	_ = g.Wait()
}

// GE15: Slice index with ctx
func goodSliceIndexWithCtx(ctx context.Context) {
	g := new(errgroup.Group)
	tasks := []func() error{
		func() error {
			_ = ctx // Uses ctx
			return nil
		},
	}
	g.Go(tasks[0]) // OK - now tracked
	_ = g.Wait()
}

// GE16: Map key with ctx
func goodMapKeyWithCtx(ctx context.Context) {
	g := new(errgroup.Group)
	tasks := map[string]func() error{
		"key": func() error {
			_ = ctx // Uses ctx
			return nil
		},
	}
	g.Go(tasks["key"]) // OK - now tracked
	_ = g.Wait()
}

// ===== INTERFACE METHOD PATTERNS =====
// ctx passed as argument to interface method IS detected by the analyzer.

// GE100: Interface method with ctx argument (good)
// When ctx is passed as argument, analyzer detects ctx usage.
type WorkerFactory interface {
	CreateWorker(ctx context.Context) func() error
}

type myFactory struct{}

func (f *myFactory) CreateWorker(ctx context.Context) func() error {
	return func() error {
		_ = ctx // Implementation captures ctx
		return nil
	}
}

func goodInterfaceMethodWithCtxArg(ctx context.Context, factory WorkerFactory) {
	g := new(errgroup.Group)
	// ctx IS passed as argument to CreateWorker - analyzer detects ctx usage
	g.Go(factory.CreateWorker(ctx)) // OK - ctx passed as argument
	_ = g.Wait()
}

// GE100b: Interface method without ctx argument (bad)
type WorkerFactoryNoCtx interface {
	CreateWorker() func() error
}

func badInterfaceMethodWithoutCtxArg(ctx context.Context, factory WorkerFactoryNoCtx) {
	g := new(errgroup.Group)
	// ctx NOT passed to CreateWorker - expected to fail
	g.Go(factory.CreateWorker()) // want `errgroup.Group.Go\(\) closure should use context "ctx"`
	_ = g.Wait()
}

// ===== REMAINING LIMITATIONS =====
// These patterns cannot be tracked statically.
// LIMITATION = false positive: ctx IS used but analyzer can't detect it.

// GE101: Function passed through parameter - NOW SUPPORTED via directive
//
//goroutinectx:goroutine_creator
func runWithGroup(g *errgroup.Group, fn func() error) {
	g.Go(fn)
}

// GE101a: Function with ctx passed through creator - should pass
func goodFuncPassedThroughCreator(ctx context.Context) {
	g := new(errgroup.Group)
	fn := func() error {
		_ = ctx // fn uses ctx
		return nil
	}
	runWithGroup(g, fn) // OK - fn uses ctx, and runWithGroup is marked as creator
	_ = g.Wait()
}

// GE101b: Function without ctx passed through creator - should report
func badFuncPassedThroughCreator(ctx context.Context) {
	g := new(errgroup.Group)
	fn := func() error {
		fmt.Println("no ctx")
		return nil
	}
	runWithGroup(g, fn) // want `runWithGroup\(\) func argument should use context "ctx"`
	_ = g.Wait()
}

// GE102: LIMITATION - Function from channel - ctx captured but not traced
func limitationFuncFromChannel(ctx context.Context) {
	g := new(errgroup.Group)
	ch := make(chan func() error, 1)
	ch <- func() error {
		_ = ctx // The func DOES capture ctx
		return nil
	}
	fn := <-ch
	// LIMITATION: fn captures ctx, but analyzer can't trace through channel receive
	g.Go(fn) // want `errgroup.Group.Go\(\) closure should use context "ctx"`
	_ = g.Wait()
}

// GE103: Function from struct field without ctx - NOW TRACKED
type taskHolder struct {
	task func() error
}

func badStructFieldWithoutCtx(ctx context.Context) {
	g := new(errgroup.Group)
	holder := taskHolder{
		task: func() error {
			fmt.Println("no ctx")
			return nil
		},
	}
	g.Go(holder.task) // want `errgroup.Group.Go\(\) closure should use context "ctx"`
	_ = g.Wait()
}

// GE104: Function from map without ctx - NOW TRACKED
func badMapValueWithoutCtx(ctx context.Context) {
	g := new(errgroup.Group)
	tasks := map[string]func() error{
		"task1": func() error { return nil },
	}
	g.Go(tasks["task1"]) // want `errgroup.Group.Go\(\) closure should use context "ctx"`
	_ = g.Wait()
}

// GE105: Function from slice without ctx - NOW TRACKED
func badSliceValueWithoutCtx(ctx context.Context) {
	g := new(errgroup.Group)
	tasks := []func() error{
		func() error { return nil },
	}
	g.Go(tasks[0]) // want `errgroup.Group.Go\(\) closure should use context "ctx"`
	_ = g.Wait()
}

// GE108: LIMITATION - Function through interface{} type assertion - ctx captured but not traced
func limitationFuncThroughInterfaceWithCtx(ctx context.Context) {
	g := new(errgroup.Group)

	var i interface{} = func() error {
		_ = ctx // fn DOES capture ctx
		return nil
	}

	// Type assert to get func back
	fn := i.(func() error)
	// LIMITATION: fn captures ctx, but analyzer can't trace through interface{} assertion
	g.Go(fn) // want `errgroup.Group.Go\(\) closure should use context "ctx"`
	_ = g.Wait()
}

// Control: same pattern without ctx
func badFuncThroughInterfaceWithoutCtx(ctx context.Context) {
	g := new(errgroup.Group)

	var i interface{} = func() error {
		fmt.Println("no ctx") // fn does NOT use ctx
		return nil
	}

	fn := i.(func() error)
	g.Go(fn) // want `errgroup.Group.Go\(\) closure should use context "ctx"`
	_ = g.Wait()
}

// ===== MULTIPLE CONTEXT EVIL PATTERNS =====

// GE70: Three contexts - uses middle one (good)
func goodUsesMiddleOfThreeContexts(ctx1, ctx2, ctx3 context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx2 // uses middle context
		return nil
	})
	_ = g.Wait()
}

// GE71: Three contexts - uses last one (good)
func goodUsesLastOfThreeContexts(ctx1, ctx2, ctx3 context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx3 // uses last context
		return nil
	})
	_ = g.Wait()
}

// GE72: Multiple ctx in separate param groups (good)
func goodMultipleCtxSeparateGroups(a int, ctx1 context.Context, b string, ctx2 context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx2 // uses second ctx from different param group
		return nil
	})
	_ = g.Wait()
}

// GE73: Multiple ctx in separate param groups - none used (bad)
func badMultipleCtxSeparateGroups(a int, ctx1 context.Context, b string, ctx2 context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx1"`
		fmt.Println(a, b) // uses other params but not ctx
		return nil
	})
	_ = g.Wait()
}

// GE74: Both contexts used (good)
func goodUsesBothContexts(ctx1, ctx2 context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx1
		_ = ctx2
		return nil
	})
	_ = g.Wait()
}

// GE85: Higher-order with multiple ctx - factory receives ctx1 (good)
func goodHigherOrderMultipleCtx(ctx1, ctx2 context.Context) {
	g := new(errgroup.Group)
	g.Go(makeWorkerWithCtx(ctx1)) // factory uses ctx1
	_ = g.Wait()
}

// GE86: Higher-order with multiple ctx - factory receives ctx2 (good)
func goodHigherOrderMultipleCtxSecond(ctx1, ctx2 context.Context) {
	g := new(errgroup.Group)
	g.Go(makeWorkerWithCtx(ctx2)) // factory uses ctx2
	_ = g.Wait()
}

// ===== ADVANCED NESTED PATTERNS (SHADOWING, ARGUMENT PASSING) =====

// Shadowing - inner ctx shadows outer
func evilShadowingInnerHasCtx(outerCtx context.Context) {
	innerFunc := func(ctx context.Context) {
		g := new(errgroup.Group)
		g.Go(func() error {
			_ = ctx // uses inner ctx
			return nil
		})
		_ = g.Wait()
	}
	innerFunc(outerCtx)
}

func evilShadowingInnerIgnoresCtx(outerCtx context.Context) {
	innerFunc := func(ctx context.Context) {
		g := new(errgroup.Group)
		g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
			return nil
		})
		_ = g.Wait()
	}
	innerFunc(outerCtx)
}

// Two levels of shadowing
func evilShadowingTwoLevels(ctx1 context.Context) {
	func(ctx2 context.Context) {
		func(ctx3 context.Context) {
			g := new(errgroup.Group)
			g.Go(func() error {
				_ = ctx3 // uses ctx3
				return nil
			})
			_ = g.Wait()
		}(ctx2)
	}(ctx1)
}

func evilShadowingTwoLevelsBad(ctx1 context.Context) {
	func(ctx2 context.Context) {
		func(ctx3 context.Context) {
			g := new(errgroup.Group)
			g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx3"`
				return nil
			})
			_ = g.Wait()
		}(ctx2)
	}(ctx1)
}

// ===== MIDDLE LAYER INTRODUCES CTX (OUTER HAS NONE) =====

func evilMiddleLayerIntroducesCtx() {
	func(ctx context.Context) {
		g := new(errgroup.Group)
		g.Go(func() error {
			_ = ctx
			return nil
		})
		func() {
			g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
				return nil
			})
		}()
		_ = g.Wait()
	}(context.Background())
}

func evilMiddleLayerIntroducesCtxGood() {
	func(ctx context.Context) {
		g := new(errgroup.Group)
		func() {
			g.Go(func() error {
				_ = ctx
				return nil
			})
		}()
		_ = g.Wait()
	}(context.Background())
}

// ===== INTERLEAVED LAYERS (ctx -> no ctx -> ctx shadowing) =====

func evilInterleavedLayers(outerCtx context.Context) {
	func() {
		func(middleCtx context.Context) {
			g := new(errgroup.Group)
			func() {
				g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "middleCtx"`
					return nil
				})
			}()
			_ = g.Wait()
		}(outerCtx)
	}()
}

func evilInterleavedLayersGood(outerCtx context.Context) {
	func() {
		func(middleCtx context.Context) {
			g := new(errgroup.Group)
			func() {
				g.Go(func() error {
					_ = middleCtx
					return nil
				})
			}()
			_ = g.Wait()
		}(outerCtx)
	}()
}
