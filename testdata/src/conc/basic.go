// Package conc contains test fixtures for sourcegraph/conc context propagation checker.
// This file tests that the analyzer correctly detects context usage in conc pool APIs,
// including generic types like ResultPool[T].
package conc

import (
	"context"
	"fmt"

	"github.com/sourcegraph/conc"
	"github.com/sourcegraph/conc/pool"
)

// ===== conc.Pool =====

// [BAD]: conc.Pool.Go without ctx
func badConcPoolGo(ctx context.Context) {
	p := &conc.Pool{}
	p.Go(func() { // want `conc.Pool.Go\(\) closure should use context "ctx"`
		fmt.Println("no context")
	})
	p.Wait()
}

// [GOOD]: conc.Pool.Go with ctx
func goodConcPoolGo(ctx context.Context) {
	p := &conc.Pool{}
	p.Go(func() {
		_ = ctx
	})
	p.Wait()
}

// ===== conc.WaitGroup =====

// [BAD]: conc.WaitGroup.Go without ctx
func badConcWaitGroupGo(ctx context.Context) {
	wg := &conc.WaitGroup{}
	wg.Go(func() { // want `conc.WaitGroup.Go\(\) closure should use context "ctx"`
		fmt.Println("no context")
	})
	wg.Wait()
}

// [GOOD]: conc.WaitGroup.Go with ctx
func goodConcWaitGroupGo(ctx context.Context) {
	wg := &conc.WaitGroup{}
	wg.Go(func() {
		_ = ctx
	})
	wg.Wait()
}

// ===== pool.Pool =====

// [BAD]: pool.Pool.Go without ctx
func badPoolPoolGo(ctx context.Context) {
	p := &pool.Pool{}
	p.Go(func() { // want `pool.Pool.Go\(\) closure should use context "ctx"`
		fmt.Println("no context")
	})
	p.Wait()
}

// [GOOD]: pool.Pool.Go with ctx
func goodPoolPoolGo(ctx context.Context) {
	p := &pool.Pool{}
	p.Go(func() {
		_ = ctx
	})
	p.Wait()
}

// ===== pool.ResultPool[T] (generic) =====

// [BAD]: pool.ResultPool[T].Go without ctx
func badResultPoolGo(ctx context.Context) {
	p := &pool.ResultPool[int]{}
	p.Go(func() int { // want `pool.ResultPool.Go\(\) closure should use context "ctx"`
		return 42
	})
	_ = p.Wait()
}

// [GOOD]: pool.ResultPool[T].Go with ctx
func goodResultPoolGo(ctx context.Context) {
	p := &pool.ResultPool[int]{}
	p.Go(func() int {
		_ = ctx
		return 42
	})
	_ = p.Wait()
}

// ===== pool.ContextPool =====

// [BAD]: pool.ContextPool.Go without ctx capture (callback receives ctx as arg)
// Note: ContextPool passes ctx to callback, so this is actually OK
// The callback receives ctx as argument, not capturing from outside
func goodContextPoolGo(ctx context.Context) {
	p := &pool.ContextPool{}
	p.Go(func(ctx context.Context) error {
		// ctx is passed as argument, not captured - this is fine
		return nil
	})
	_ = p.Wait()
}

// ===== pool.ResultContextPool[T] (generic) =====

// [GOOD]: pool.ResultContextPool[T].Go - callback receives ctx
func goodResultContextPoolGo(ctx context.Context) {
	p := &pool.ResultContextPool[int]{}
	p.Go(func(ctx context.Context) (int, error) {
		// ctx is passed as argument
		return 42, nil
	})
	_, _ = p.Wait()
}

// ===== pool.ErrorPool =====

// [BAD]: pool.ErrorPool.Go without ctx
func badErrorPoolGo(ctx context.Context) {
	p := &pool.ErrorPool{}
	p.Go(func() error { // want `pool.ErrorPool.Go\(\) closure should use context "ctx"`
		return nil
	})
	_ = p.Wait()
}

// [GOOD]: pool.ErrorPool.Go with ctx
func goodErrorPoolGo(ctx context.Context) {
	p := &pool.ErrorPool{}
	p.Go(func() error {
		_ = ctx
		return nil
	})
	_ = p.Wait()
}

// ===== pool.ResultErrorPool[T] (generic) =====

// [BAD]: pool.ResultErrorPool[T].Go without ctx
func badResultErrorPoolGo(ctx context.Context) {
	p := &pool.ResultErrorPool[string]{}
	p.Go(func() (string, error) { // want `pool.ResultErrorPool.Go\(\) closure should use context "ctx"`
		return "result", nil
	})
	_, _ = p.Wait()
}

// [GOOD]: pool.ResultErrorPool[T].Go with ctx
func goodResultErrorPoolGo(ctx context.Context) {
	p := &pool.ResultErrorPool[string]{}
	p.Go(func() (string, error) {
		_ = ctx
		return "result", nil
	})
	_, _ = p.Wait()
}
