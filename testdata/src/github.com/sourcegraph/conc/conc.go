// Package conc provides stub types for sourcegraph/conc.
package conc

// Pool is a stub for conc.Pool.
type Pool struct{}

// Go submits a task to the pool.
func (*Pool) Go(f func()) {}

// Wait waits for all tasks to complete.
func (*Pool) Wait() {}

// WaitGroup is a stub for conc.WaitGroup.
type WaitGroup struct{}

// Go submits a task to the wait group.
func (*WaitGroup) Go(f func()) {}

// Wait waits for all tasks to complete.
func (*WaitGroup) Wait() {}
