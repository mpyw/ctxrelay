// Package pool provides stub types for sourcegraph/conc/pool.
package pool

import "context"

// Pool is a stub for pool.Pool.
type Pool struct{}

// Go submits a task to the pool.
func (*Pool) Go(f func()) {}

// Wait waits for all tasks to complete.
func (*Pool) Wait() {}

// ResultPool is a stub for pool.ResultPool[T] (generic).
type ResultPool[T any] struct{}

// Go submits a task that returns a result.
func (*ResultPool[T]) Go(f func() T) {}

// Wait waits for all tasks and returns results.
func (*ResultPool[T]) Wait() []T { return nil }

// ContextPool is a stub for pool.ContextPool.
type ContextPool struct{}

// Go submits a task with context.
func (*ContextPool) Go(f func(context.Context) error) {}

// Wait waits for all tasks to complete.
func (*ContextPool) Wait() error { return nil }

// ResultContextPool is a stub for pool.ResultContextPool[T] (generic).
type ResultContextPool[T any] struct{}

// Go submits a task with context that returns a result.
func (*ResultContextPool[T]) Go(f func(context.Context) (T, error)) {}

// Wait waits for all tasks and returns results.
func (*ResultContextPool[T]) Wait() ([]T, error) { return nil, nil }

// ErrorPool is a stub for pool.ErrorPool.
type ErrorPool struct{}

// Go submits a task that may return an error.
func (*ErrorPool) Go(f func() error) {}

// Wait waits for all tasks to complete.
func (*ErrorPool) Wait() error { return nil }

// ResultErrorPool is a stub for pool.ResultErrorPool[T] (generic).
type ResultErrorPool[T any] struct{}

// Go submits a task that returns a result or error.
func (*ResultErrorPool[T]) Go(f func() (T, error)) {}

// Wait waits for all tasks and returns results.
func (*ResultErrorPool[T]) Wait() ([]T, error) { return nil, nil }
