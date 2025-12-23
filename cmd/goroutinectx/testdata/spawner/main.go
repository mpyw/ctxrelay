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

//goroutinectx:spawner
func runTask(g *errgroup.Group, fn func() error) {
	g.Go(fn)
}

// goodSimple: spawner with context propagation
func goodSimple(ctx context.Context) {
	g := new(errgroup.Group)
	runTask(g, func() error {
		_ = ctx
		fmt.Println("work")
		return nil
	})
	_ = g.Wait()
}

// badSimple: spawner without context propagation
func badSimple(ctx context.Context) {
	g := new(errgroup.Group)
	runTask(g, func() error {
		fmt.Println("work")
		return nil
	})
	_ = g.Wait()
}

// ===== COMPLEX PATTERN =====

//goroutinectx:spawner
func runMultipleTasks(g *errgroup.Group, tasks ...func() error) {
	for _, task := range tasks {
		g.Go(task)
	}
}

// goodComplex: spawner with multiple tasks, all use context
func goodComplex(ctx context.Context) {
	g := new(errgroup.Group)
	runMultipleTasks(g,
		func() error {
			_ = ctx
			fmt.Println("task1")
			return nil
		},
		func() error {
			_ = ctx
			fmt.Println("task2")
			return nil
		},
	)
	_ = g.Wait()
}

// badComplex: spawner with multiple tasks, one missing context
func badComplex(ctx context.Context) {
	g := new(errgroup.Group)
	runMultipleTasks(g,
		func() error {
			_ = ctx
			fmt.Println("task1")
			return nil
		},
		func() error {
			fmt.Println("task2 missing ctx")
			return nil
		},
	)
	_ = g.Wait()
}
