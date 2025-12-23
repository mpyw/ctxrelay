package apm

import "context"

// NewGoroutineContext derives a new context for goroutine instrumentation.
func NewGoroutineContext(ctx context.Context) context.Context {
	return ctx
}
