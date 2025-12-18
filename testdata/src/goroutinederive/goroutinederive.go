package goroutinederive

import (
	"context"

	"github.com/my-example-app/telemetry/apm"
)

// Test cases for goroutine-derive checker with -goroutine-deriver=github.com/my-example-app/telemetry/apm.NewGoroutineContext

// ===== SHOULD NOT REPORT =====

// DD01: Basic - calls deriver.
func d01CallsDeriver(ctx context.Context) {
	go func() {
		ctx := apm.NewGoroutineContext(ctx)
		_ = ctx
	}()
}

// DD02: Basic - nested goroutines both call deriver.
func d02NestedBothCallDeriver(ctx context.Context) {
	go func() {
		ctx := apm.NewGoroutineContext(ctx)
		go func() {
			ctx := apm.NewGoroutineContext(ctx)
			_ = ctx
		}()
		_ = ctx
	}()
}

// DD03: Basic - has own context param.
func d03OwnContextParam(ctx context.Context) {
	go func(ctx context.Context) {
		_ = ctx
	}(ctx)
}

// DD04: Basic - named function call (not checked).
func d04NamedFuncCall(ctx context.Context) {
	go namedFunc(ctx)
}

func namedFunc(ctx context.Context) {
	_ = ctx
}

// ===== SHOULD REPORT =====

// DD05: Basic - no deriver call.
func d05NoDeriverCall(ctx context.Context) {
	go func() { // want "goroutine should call github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		_ = ctx
	}()
}

// DD06: Basic - uses different function (not deriver).
func d06UsesDifferentFunc(ctx context.Context) {
	go func() { // want "goroutine should call github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		ctx := context.WithValue(ctx, "key", "value")
		_ = ctx
	}()
}

// DD07: Basic - nested, inner missing deriver.
func d07NestedInnerMissingDeriver(ctx context.Context) {
	go func() {
		ctx := apm.NewGoroutineContext(ctx)
		go func() { // want "goroutine should call github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
			_ = ctx
		}()
		_ = ctx
	}()
}
