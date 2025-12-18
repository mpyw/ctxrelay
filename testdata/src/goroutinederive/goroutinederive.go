// Package goroutinederive contains test fixtures for the goroutine-derive checker.
// This file covers basic patterns with -goroutine-deriver flag.
package goroutinederive

import (
	"context"

	"github.com/my-example-app/telemetry/apm"
)

// Test cases for goroutine-derive checker with -goroutine-deriver=github.com/my-example-app/telemetry/apm.NewGoroutineContext

// ===== SHOULD NOT REPORT =====

// Basic - calls deriver
// Goroutine calls deriver function
func d01CallsDeriver(ctx context.Context) {
	go func() {
		ctx := apm.NewGoroutineContext(ctx)
		_ = ctx
	}()
}

// Nested goroutines both call deriver
// Both nested goroutines call deriver function
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

// Has own context param
// Goroutine has its own context parameter
func d03OwnContextParam(ctx context.Context) {
	go func(ctx context.Context) {
		_ = ctx
	}(ctx)
}

// Named function call
// go statement with named function call (not checked)
func d04NamedFuncCall(ctx context.Context) {
	go namedFunc(ctx)
}

func namedFunc(ctx context.Context) {
	_ = ctx
}

// ===== SHOULD REPORT =====

// No deriver call
// Goroutine does not call deriver function
func d05NoDeriverCall(ctx context.Context) {
	go func() { // want "goroutine should call github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		_ = ctx
	}()
}

// Uses different function
// Goroutine uses context.WithValue instead of deriver
func d06UsesDifferentFunc(ctx context.Context) {
	go func() { // want "goroutine should call github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
		ctx := context.WithValue(ctx, "key", "value")
		_ = ctx
	}()
}

// Nested, inner missing deriver
// Outer goroutine calls deriver but inner does not
func d07NestedInnerMissingDeriver(ctx context.Context) {
	go func() {
		ctx := apm.NewGoroutineContext(ctx)
		go func() { // want "goroutine should call github.com/my-example-app/telemetry/apm.NewGoroutineContext to derive context"
			_ = ctx
		}()
		_ = ctx
	}()
}
