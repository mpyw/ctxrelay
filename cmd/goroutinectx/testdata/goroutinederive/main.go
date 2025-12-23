package main

import (
	"context"
	"fmt"

	"example.com/goroutinederive/apm"
)

func main() {
	ctx := context.Background()
	goodSimple(ctx)
	badSimple(ctx)
	goodComplex(ctx)
	badComplex(ctx)
}

// ===== SIMPLE PATTERN =====

// goodSimple: goroutine calls deriver
func goodSimple(ctx context.Context) {
	go func() {
		ctx := apm.NewGoroutineContext(ctx)
		fmt.Println("work", ctx)
	}()
}

// badSimple: goroutine does not call deriver
func badSimple(ctx context.Context) {
	go func() {
		_ = ctx
		fmt.Println("work without deriver")
	}()
}

// ===== COMPLEX PATTERN =====

// goodComplex: nested goroutines all call deriver
func goodComplex(ctx context.Context) {
	go func() {
		ctx := apm.NewGoroutineContext(ctx)
		fmt.Println("outer", ctx)

		go func() {
			ctx := apm.NewGoroutineContext(ctx)
			fmt.Println("inner", ctx)
		}()
	}()
}

// badComplex: outer calls deriver but inner does not
func badComplex(ctx context.Context) {
	go func() {
		ctx := apm.NewGoroutineContext(ctx)
		fmt.Println("outer", ctx)

		go func() {
			_ = ctx
			fmt.Println("inner missing deriver")
		}()
	}()
}
