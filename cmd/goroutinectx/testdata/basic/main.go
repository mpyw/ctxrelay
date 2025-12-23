package main

import (
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()
	good(ctx)
	bad(ctx)
}

// good passes context to goroutine
func good(ctx context.Context) {
	go func() {
		_ = ctx
		fmt.Println("work")
	}()
}

// bad does not pass context to goroutine
func bad(ctx context.Context) {
	go func() {
		fmt.Println("work")
	}()
}
