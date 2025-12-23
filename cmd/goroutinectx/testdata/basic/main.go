package main

import (
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()
	goodSimple(ctx)
	badSimple(ctx)
	goodComplex(ctx)
	badComplex(ctx)
}

// ===== SIMPLE PATTERN =====

// goodSimple: goroutine uses context
func goodSimple(ctx context.Context) {
	go func() {
		_ = ctx
		fmt.Println("work")
	}()
}

// badSimple: goroutine does not use context
func badSimple(ctx context.Context) {
	go func() {
		fmt.Println("work")
	}()
}

// ===== COMPLEX PATTERN =====

// goodComplex: higher-order function with context
func goodComplex(ctx context.Context) {
	makeWorker := func() func() {
		return func() {
			_ = ctx
			fmt.Println("work")
		}
	}
	go makeWorker()()
}

// badComplex: higher-order function without context
func badComplex(ctx context.Context) {
	makeWorker := func() func() {
		return func() {
			fmt.Println("work without ctx")
		}
	}
	go makeWorker()()
}
