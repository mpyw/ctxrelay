package main

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := context.Background()
	good(ctx)
	bad(ctx)
}

// good passes context to errgroup closure
func good(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx
		fmt.Println("work")
		return nil
	})
	_ = g.Wait()
}

// bad does not pass context to errgroup closure
func bad(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		fmt.Println("work")
		return nil
	})
	_ = g.Wait()
}
