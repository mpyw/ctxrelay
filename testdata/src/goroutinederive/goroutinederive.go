package goroutinederive

import (
	"context"

	"github.com/my-example-app/telemetry/apm"
)

// Test cases for goroutine-derive checker with -goroutine-deriver=github.com/my-example-app/telemetry/apm.NewGoroutineContext

// ===== SHOULD NOT REPORT =====

// [GOOD]: Basic - calls deriver.
//
// Goroutine properly calls the required context deriver function.
func goodCallsDeriver(ctx context.Context) {
	go func() {
		ctx := apm.NewGoroutineContext(ctx)
		_ = ctx
	}()
}

// [GOOD]: Basic - nested goroutines both call deriver.
//
// Both outer and inner goroutines call the deriver function.
func goodNestedBothCallDeriver(ctx context.Context) {
	go func() {
		ctx := apm.NewGoroutineContext(ctx)
		go func() {
			ctx := apm.NewGoroutineContext(ctx)
			_ = ctx
		}()
		_ = ctx
	}()
}

// [NOTCHECKED]: Basic - has own context param.
//
// Function declares its own context parameter, so outer context not required.
func notCheckedOwnContextParam(ctx context.Context) {
	go func(ctx context.Context) {
		_ = ctx
	}(ctx)
}

// [NOTCHECKED]: Basic - named function call (not checked).
//
// Named function call pattern is not checked for deriver.
func notCheckedNamedFuncCall(ctx context.Context) {
	go namedFunc(ctx)
}

//vt:helper
func namedFunc(ctx context.Context) {
	_ = ctx
}

// ===== SHOULD REPORT =====

// [BAD]: Basic - no deriver call.
//
// Goroutine does not call the required context deriver function.
func badNoDeriverCall(ctx context.Context) {
	go func() { // want "goroutine should call github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		_ = ctx
	}()
}

// [BAD]: Basic - uses different function (not deriver).
//
// Goroutine calls a function, but not the required deriver.
func badUsesDifferentFunc(ctx context.Context) {
	go func() { // want "goroutine should call github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		ctx := context.WithValue(ctx, "key", "value")
		_ = ctx
	}()
}

// [BAD]: Basic - nested, inner missing deriver.
//
// Inner goroutine does not call the required deriver.
func badNestedInnerMissingDeriver(ctx context.Context) {
	go func() {
		ctx := apm.NewGoroutineContext(ctx)
		go func() { // want "goroutine should call github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
			_ = ctx
		}()
		_ = ctx
	}()
}
