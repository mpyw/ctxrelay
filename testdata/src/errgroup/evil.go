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

// Variable func without ctx
// Function stored in variable does not use context
// see also: waitgroup
func badVariableFunc(ctx context.Context) {
	g := new(errgroup.Group)
	fn := func() error {
		fmt.Println("no ctx")
		return nil
	}
	g.Go(fn) // want `errgroup.Group.Go\(\) closure should use context "ctx"`
	_ = g.Wait()
}

// Variable func with ctx
// Function stored in variable uses context
// see also: waitgroup
func goodVariableFuncWithCtx(ctx context.Context) {
	g := new(errgroup.Group)
	fn := func() error {
		_ = ctx
		return nil
	}
	g.Go(fn) // OK - fn uses ctx
	_ = g.Wait()
}

// Higher-order func without ctx
// Higher-order function that returns closure without context
// see also: waitgroup
func badHigherOrderFunc(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(makeWorker()) // want `errgroup.Group.Go\(\) closure should use context "ctx"`
	_ = g.Wait()
}

// Higher-order func with ctx
// Higher-order function that returns closure with context
// see also: waitgroup
func goodHigherOrderFuncWithCtx(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(makeWorkerWithCtx(ctx)) // OK - makeWorkerWithCtx captures ctx
	_ = g.Wait()
}

// ===== STRUCT FIELD / SLICE / MAP TRACKING =====
// These patterns CAN be tracked when defined in the same function.

// Struct field with ctx
// Function from struct field that uses context
// see also: goroutine, waitgroup
type taskHolderWithCtx struct {
	task func() error
}

// Struct field func with ctx
// Function from struct field that uses context
// see also: waitgroup
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

// Slice index with ctx
// Function from slice that uses context
// see also: waitgroup
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

// Map key with ctx
// Function from map that uses context
// see also: waitgroup
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

// Interface method with ctx argument
// When ctx is passed as argument, analyzer detects ctx usage
// see also: goroutine, waitgroup
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

// Interface method with ctx arg
// Interface method receives context as argument
// see also: waitgroup
func goodInterfaceMethodWithCtxArg(ctx context.Context, factory WorkerFactory) {
	g := new(errgroup.Group)
	// ctx IS passed as argument to CreateWorker - analyzer detects ctx usage
	g.Go(factory.CreateWorker(ctx)) // OK - ctx passed as argument
	_ = g.Wait()
}

// Interface method without ctx argument
// Interface method that does not receive context
// see also: goroutine, waitgroup
type WorkerFactoryNoCtx interface {
	CreateWorker() func() error
}

// Interface method without ctx arg
// Interface method that does not receive context
// see also: waitgroup
func badInterfaceMethodWithoutCtxArg(ctx context.Context, factory WorkerFactoryNoCtx) {
	g := new(errgroup.Group)
	// ctx NOT passed to CreateWorker - expected to fail
	g.Go(factory.CreateWorker()) // want `errgroup.Group.Go\(\) closure should use context "ctx"`
	_ = g.Wait()
}

// ===== REMAINING LIMITATIONS =====
// These patterns cannot be tracked statically.
// LIMITATION = false positive: ctx IS used but analyzer can't detect it.

// Function passed through parameter - NOW SUPPORTED via directive
//
//goroutinectx:goroutine_creator
func runWithGroup(g *errgroup.Group, fn func() error) {
	g.Go(fn)
}

// Function with ctx passed through creator
// Function with context passed through goroutine creator helper
// see also: waitgroup
func goodFuncPassedThroughCreator(ctx context.Context) {
	g := new(errgroup.Group)
	fn := func() error {
		_ = ctx // fn uses ctx
		return nil
	}
	runWithGroup(g, fn) // OK - fn uses ctx, and runWithGroup is marked as creator
	_ = g.Wait()
}

// Function without ctx passed through creator
// Function without context passed through goroutine creator helper
// see also: waitgroup
func badFuncPassedThroughCreator(ctx context.Context) {
	g := new(errgroup.Group)
	fn := func() error {
		fmt.Println("no ctx")
		return nil
	}
	runWithGroup(g, fn) // want `runWithGroup\(\) func argument should use context "ctx"`
	_ = g.Wait()
}

// LIMITATION - Function from channel
// Context captured but not traced through channel receive
// see also: waitgroup
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

// Function from struct field without ctx
// Function from struct field that does not use context
// see also: goroutine, waitgroup
type taskHolder struct {
	task func() error
}

// Struct field func without ctx
// Function from struct field that does not use context
// see also: waitgroup
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

// Function from map without ctx
// Function from map that does not use context
// see also: waitgroup
func badMapValueWithoutCtx(ctx context.Context) {
	g := new(errgroup.Group)
	tasks := map[string]func() error{
		"task1": func() error { return nil },
	}
	g.Go(tasks["task1"]) // want `errgroup.Group.Go\(\) closure should use context "ctx"`
	_ = g.Wait()
}

// Function from slice without ctx
// Function from slice that does not use context
// see also: waitgroup
func badSliceValueWithoutCtx(ctx context.Context) {
	g := new(errgroup.Group)
	tasks := []func() error{
		func() error { return nil },
	}
	g.Go(tasks[0]) // want `errgroup.Group.Go\(\) closure should use context "ctx"`
	_ = g.Wait()
}

// LIMITATION - Function through interface{} type assertion
// Context captured but not traced through interface{} assertion
// see also: waitgroup
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

// Function through interface{} - control case without ctx
// Function through interface{} that does not use context
// see also: waitgroup
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

// Three contexts - uses middle one
// Function with three context parameters, uses middle one
// see also: goroutine, waitgroup
func goodUsesMiddleOfThreeContexts(ctx1, ctx2, ctx3 context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx2 // uses middle context
		return nil
	})
	_ = g.Wait()
}

// Three contexts - uses last one
// Function with three context parameters, uses last one
// see also: goroutine, waitgroup
func goodUsesLastOfThreeContexts(ctx1, ctx2, ctx3 context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx3 // uses last context
		return nil
	})
	_ = g.Wait()
}

// Multiple ctx in separate param groups
// Context parameters in separate groups, uses second
// see also: goroutine, waitgroup
func goodMultipleCtxSeparateGroups(a int, ctx1 context.Context, b string, ctx2 context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx2 // uses second ctx from different param group
		return nil
	})
	_ = g.Wait()
}

// Multiple ctx in separate param groups - none used
// Context parameters in separate groups, none used
// see also: goroutine, waitgroup
func badMultipleCtxSeparateGroups(a int, ctx1 context.Context, b string, ctx2 context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx1"`
		fmt.Println(a, b) // uses other params but not ctx
		return nil
	})
	_ = g.Wait()
}

// Both contexts used
// Function with two context parameters, both used
// see also: goroutine, waitgroup
func goodUsesBothContexts(ctx1, ctx2 context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx1
		_ = ctx2
		return nil
	})
	_ = g.Wait()
}

// Higher-order with multiple ctx - factory receives ctx1
// Higher-order function with multiple contexts, uses first
// see also: goroutine, waitgroup
func goodHigherOrderMultipleCtx(ctx1, ctx2 context.Context) {
	g := new(errgroup.Group)
	g.Go(makeWorkerWithCtx(ctx1)) // factory uses ctx1
	_ = g.Wait()
}

// Higher-order with multiple ctx - factory receives ctx2
// Higher-order function with multiple contexts, uses second
// see also: goroutine, waitgroup
func goodHigherOrderMultipleCtxSecond(ctx1, ctx2 context.Context) {
	g := new(errgroup.Group)
	g.Go(makeWorkerWithCtx(ctx2)) // factory uses ctx2
	_ = g.Wait()
}

// ===== ADVANCED NESTED PATTERNS (SHADOWING, ARGUMENT PASSING) =====

// Shadowing - inner ctx shadows outer
// Inner function with its own context parameter uses it
// see also: waitgroup
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

// Shadowing - inner ignores ctx
// Inner function with its own context parameter ignores it
// see also: waitgroup
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
// Two levels of argument passing with context shadowing
// see also: waitgroup
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

// Two levels of shadowing - innermost ignores ctx
// Two levels of shadowing where innermost closure ignores context
// see also: waitgroup
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

// Middle layer introduces ctx - mixed usage
// Middle layer introduces context, some closures use it and some don't
// see also: waitgroup
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

// Middle layer introduces ctx - good
// Middle layer introduces context and nested closure properly uses it
// see also: waitgroup
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

// Interleaved layers - goroutine ignores shadowing ctx
// Interleaved layers where goroutine ignores the shadowing context
// see also: waitgroup
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

// Interleaved layers - good
// Interleaved layers where goroutine properly uses shadowing context
// see also: waitgroup
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
