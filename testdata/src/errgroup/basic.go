// Package errgroup contains test fixtures for the errgroup context propagation checker.
// This file covers basic/daily patterns - simple good/bad cases, shadowing, ignore directives.
// See advanced.go for real-world complex patterns and evil.go for adversarial tests.
package errgroup

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

// ===== SHOULD REPORT =====

// GE01: Literal without ctx - basic bad case
func badErrgroupGo(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
		fmt.Println("no context")
		return nil
	})
	_ = g.Wait()
}

// GE01b: TryGo without ctx
func badErrgroupTryGo(ctx context.Context) {
	g := new(errgroup.Group)
	g.TryGo(func() error { // want `errgroup.Group.TryGo\(\) closure should use context "ctx"`
		fmt.Println("no context")
		return nil
	})
	_ = g.Wait()
}

// GE01c: Multiple Go calls without ctx
func badErrgroupGoMultiple(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
		return nil
	})
	g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
		return nil
	})
	_ = g.Wait()
}

// ===== SHOULD NOT REPORT =====

// GE02: Literal with ctx - basic good case
func goodErrgroupGoWithCtx(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx.Done()
		return nil
	})
	_ = g.Wait()
}

// GE02b: Literal with ctx - via function call
func goodErrgroupGoCallsWithCtx(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		return doWork(ctx)
	})
	_ = g.Wait()
}

// GE02c: errgroup.WithContext pattern
func goodErrgroupWithContext(ctx context.Context) {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		_ = ctx.Done()
		return nil
	})
	_ = g.Wait()
}

// GE03: No ctx param - not checked
func goodNoContextParam() {
	g := new(errgroup.Group)
	g.Go(func() error {
		return nil
	})
	_ = g.Wait()
}

// ===== SHADOWING TESTS =====

// GE04: Shadow with non-ctx type (string)
func badShadowingNonContext(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
		ctx := "not a context"
		_ = ctx
		return nil
	})
	_ = g.Wait()
}

// GE05: Uses ctx before shadow - valid usage
func goodUsesCtxBeforeShadowing(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx.Done() // use ctx before shadowing
		ctx := "shadow"
		_ = ctx
		return nil
	})
	_ = g.Wait()
}

// ===== IGNORE DIRECTIVES =====

// GE06: Ignore directive - same line
func goodIgnoredSameLine(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error { //goroutinectx:ignore
		return nil
	})
	_ = g.Wait()
}

// GE07: Ignore directive - previous line
func goodIgnoredPreviousLine(ctx context.Context) {
	g := new(errgroup.Group)
	//goroutinectx:ignore
	g.Go(func() error {
		return nil
	})
	_ = g.Wait()
}

// ===== MULTIPLE CONTEXT PARAMETERS =====

// GE08: Multiple ctx params - reports first (bad)
func twoContextParams(ctx1, ctx2 context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx1"`
		return nil
	})
	_ = g.Wait()
}

// GE09: Multiple ctx params - uses first (good)
func goodUsesOneOfTwoContexts(ctx1, ctx2 context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx1
		return nil
	})
	_ = g.Wait()
}

// GE09b: Multiple ctx params - uses second (good)
func goodUsesSecondOfTwoContexts(ctx1, ctx2 context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx2 // uses second context - should NOT report
		return nil
	})
	_ = g.Wait()
}

// GE14: Context as non-first param (good)
func goodCtxAsSecondParam(logger interface{}, ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx // ctx is second param but still detected
		return nil
	})
	_ = g.Wait()
}

// GE14b: Context as non-first param without use (bad)
func badCtxAsSecondParam(logger interface{}, ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
		_ = logger
		return nil
	})
	_ = g.Wait()
}

func doWork(ctx context.Context) error {
	_ = ctx
	return nil
}
