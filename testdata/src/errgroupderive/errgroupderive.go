package errgroupderive

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/my-example-app/telemetry/apm"
)

// Test cases for errgroup checker with -goroutine-deriver flag.
// When deriver is configured, callbacks should either capture context OR call deriver.

// ===== GOOD: Context captured =====

// [GOOD]: Callback captures context
func goodCapturesContext(ctx context.Context) {
	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		_ = ctx // captures context
		return nil
	})
}

// ===== GOOD: Deriver called =====

// [GOOD]: Callback calls deriver
func goodCallsDeriver(ctx context.Context) {
	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		newCtx := apm.NewGoroutineContext(ctx)
		_ = newCtx
		return nil
	})
}

// [GOOD]: Callback calls deriver without capturing ctx
func goodCallsDeriverOnly(ctx context.Context) {
	var g errgroup.Group
	g.Go(func() error {
		newCtx := apm.NewGoroutineContext(context.Background())
		_ = newCtx
		return nil
	})
}

// ===== BAD: Neither context nor deriver =====

// [BAD]: Callback does not capture context or call deriver
func badNoContextNoDeriver(ctx context.Context) {
	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx" or call goroutine deriver`
		return nil
	})
}

// [BAD]: TryGo without context or deriver
func badTryGoNoContextNoDeriver(ctx context.Context) {
	g, _ := errgroup.WithContext(ctx)
	g.TryGo(func() error { // want `errgroup.Group.TryGo\(\) closure should use context "ctx" or call goroutine deriver`
		return nil
	})
}

// ===== Variable patterns =====

// [GOOD]: Variable callback calls deriver
func goodVariableCallsDeriver(ctx context.Context) {
	g, _ := errgroup.WithContext(ctx)
	fn := func() error {
		newCtx := apm.NewGoroutineContext(ctx)
		_ = newCtx
		return nil
	}
	g.Go(fn)
}

// [BAD]: Variable callback without deriver
func badVariableNoDeriver(ctx context.Context) {
	g, _ := errgroup.WithContext(ctx)
	fn := func() error {
		return nil
	}
	g.Go(fn) // want `errgroup.Group.Go\(\) closure should use context "ctx" or call goroutine deriver`
}
