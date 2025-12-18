// Package apm provides application-specific APM wrappers.
// This wraps New Relic functions for easier use in the application.
package apm

import (
	"context"

	"github.com/newrelic/go-agent/v3/newrelic"
)

// NewGoroutineContext creates a new context for a goroutine with proper APM tracing.
// This extracts the transaction from the context and calls NewGoroutine on it.
func NewGoroutineContext(ctx context.Context) context.Context {
	txn := newrelic.FromContext(ctx)
	if txn == nil {
		return ctx
	}
	return newrelic.NewContext(ctx, txn.NewGoroutine())
}

// StartSpan starts a new span in the context.
// This is an alternative entry point for APM tracing.
func StartSpan(ctx context.Context, name string) context.Context {
	// In real implementation, this would create a span
	return ctx
}
