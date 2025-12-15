package newrelic

import "context"

// Transaction represents a New Relic transaction.
type Transaction struct{}

// NewContext creates a new context containing the transaction.
// This is used to propagate transaction context to child goroutines.
func NewContext(ctx context.Context, txn *Transaction) context.Context {
	return ctx
}

// NewGoroutine creates a thread-safe copy of the transaction for use
// in another goroutine.
func (txn *Transaction) NewGoroutine() *Transaction {
	return txn
}

// FromContext extracts the transaction from the context.
func FromContext(ctx context.Context) *Transaction {
	_ = ctx
	return nil
}
