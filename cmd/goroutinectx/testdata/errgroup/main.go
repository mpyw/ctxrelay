package main

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := context.Background()
	goodSimple(ctx)
	badSimple(ctx)
	goodComplex(ctx)
	badComplex(ctx)
}

// ===== SIMPLE PATTERN =====

// goodSimple: errgroup closure uses context
func goodSimple(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx
		fmt.Println("work")
		return nil
	})
	_ = g.Wait()
}

// badSimple: errgroup closure does not use context
func badSimple(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		fmt.Println("work")
		return nil
	})
	_ = g.Wait()
}

// ===== COMPLEX PATTERN =====

// goodComplex: errgroup with factory function
func goodComplex(ctx context.Context) {
	makeTask := func(name string) func() error {
		return func() error {
			_ = ctx
			fmt.Println(name)
			return nil
		}
	}

	g := new(errgroup.Group)
	g.Go(makeTask("task1"))
	g.Go(makeTask("task2"))
	g.TryGo(makeTask("task3"))
	_ = g.Wait()
}

// badComplex: errgroup with factory function missing context
func badComplex(ctx context.Context) {
	makeTask := func(name string) func() error {
		return func() error {
			fmt.Println(name)
			return nil
		}
	}

	g := new(errgroup.Group)
	g.Go(makeTask("task1"))
	g.Go(makeTask("task2"))
	g.TryGo(makeTask("task3"))
	_ = g.Wait()
}
