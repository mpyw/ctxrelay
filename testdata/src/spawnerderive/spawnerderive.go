package spawnerderive

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/my-example-app/telemetry/apm"
)

// Test cases for spawner checker with -goroutine-deriver flag.
// When deriver is configured, callbacks should either capture context OR call deriver.

// ===== SPAWNER FUNCTIONS =====

//goroutinectx:spawner
func runWithGroup(g *errgroup.Group, fn func() error) {
	g.Go(fn)
}

//goroutinectx:spawner
func runWithWaitGroup(wg *sync.WaitGroup, fn func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		fn()
	}()
}

// ===== GOOD: Context captured =====

// [GOOD]: Callback captures context
func goodCapturesContext(ctx context.Context) {
	g := new(errgroup.Group)
	runWithGroup(g, func() error {
		_ = ctx // captures context
		return nil
	})
}

// ===== GOOD: Deriver called =====

// [GOOD]: Callback calls deriver
func goodCallsDeriver(ctx context.Context) {
	g := new(errgroup.Group)
	runWithGroup(g, func() error {
		newCtx := apm.NewGoroutineContext(ctx)
		_ = newCtx
		return nil
	})
}

// [GOOD]: Callback calls deriver without capturing ctx
func goodCallsDeriverOnly(ctx context.Context) {
	g := new(errgroup.Group)
	runWithGroup(g, func() error {
		newCtx := apm.NewGoroutineContext(context.Background())
		_ = newCtx
		return nil
	})
}

// ===== BAD: Neither context nor deriver =====

// [BAD]: Callback does not capture context or call deriver
func badNoContextNoDeriver(ctx context.Context) {
	g := new(errgroup.Group)
	runWithGroup(g, func() error { // want `runWithGroup\(\) func argument should use context "ctx" or call goroutine deriver`
		return nil
	})
}

// [BAD]: WaitGroup callback without context or deriver
func badWaitGroupNoDeriver(ctx context.Context) {
	var wg sync.WaitGroup
	runWithWaitGroup(&wg, func() { // want `runWithWaitGroup\(\) func argument should use context "ctx" or call goroutine deriver`
	})
	wg.Wait()
}

// ===== Variable patterns =====

// [GOOD]: Variable callback calls deriver
func goodVariableCallsDeriver(ctx context.Context) {
	g := new(errgroup.Group)
	fn := func() error {
		newCtx := apm.NewGoroutineContext(ctx)
		_ = newCtx
		return nil
	}
	runWithGroup(g, fn)
}

// [BAD]: Variable callback without deriver
func badVariableNoDeriver(ctx context.Context) {
	g := new(errgroup.Group)
	fn := func() error {
		return nil
	}
	runWithGroup(g, fn) // want `runWithGroup\(\) func argument should use context "ctx" or call goroutine deriver`
}
