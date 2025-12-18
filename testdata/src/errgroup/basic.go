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

// Literal without ctx - basic bad case
// errgroup.Group.Go() closure does not use context
func badErrgroupGo(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
		fmt.Println("no context")
		return nil
	})
	_ = g.Wait()
}

// TryGo without ctx
// errgroup.Group.TryGo() closure does not use context
func badErrgroupTryGo(ctx context.Context) {
	g := new(errgroup.Group)
	g.TryGo(func() error { // want `errgroup.Group.TryGo\(\) closure should use context "ctx"`
		fmt.Println("no context")
		return nil
	})
	_ = g.Wait()
}

// Multiple Go calls without ctx
// Multiple errgroup.Group.Go() calls without context
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

// Literal with ctx - basic good case
// errgroup.Group.Go() closure properly uses context
func goodErrgroupGoWithCtx(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx.Done()
		return nil
	})
	_ = g.Wait()
}

// Literal with ctx - via function call
// errgroup.Group.Go() closure uses context via function call
func goodErrgroupGoCallsWithCtx(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		return doWork(ctx)
	})
	_ = g.Wait()
}

// errgroup.WithContext pattern
// Using errgroup.WithContext to create group and context
func goodErrgroupWithContext(ctx context.Context) {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		_ = ctx.Done()
		return nil
	})
	_ = g.Wait()
}

// No ctx param
// Function has no context parameter - not checked
// see also: goroutine, waitgroup
func goodNoContextParam() {
	g := new(errgroup.Group)
	g.Go(func() error {
		return nil
	})
	_ = g.Wait()
}

// ===== SHADOWING TESTS =====

// Shadow with non-ctx type
// Context variable is shadowed with non-context type (string)
// see also: goroutine, waitgroup
func badShadowingNonContext(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx"`
		ctx := "not a context"
		_ = ctx
		return nil
	})
	_ = g.Wait()
}

// Uses ctx before shadow
// Uses context before shadowing it - valid usage
// see also: goroutine, waitgroup
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

// Ignore directive - same line
// Ignore directive on the same line suppresses warning
// see also: goroutine, waitgroup
func goodIgnoredSameLine(ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error { //goroutinectx:ignore
		return nil
	})
	_ = g.Wait()
}

// Ignore directive - previous line
// Ignore directive on the previous line suppresses warning
// see also: goroutine, waitgroup
func goodIgnoredPreviousLine(ctx context.Context) {
	g := new(errgroup.Group)
	//goroutinectx:ignore
	g.Go(func() error {
		return nil
	})
	_ = g.Wait()
}

// ===== MULTIPLE CONTEXT PARAMETERS =====

// Multiple ctx params - reports first
// Function has two context parameters, reports first one when neither used
// see also: goroutine, waitgroup
func twoContextParams(ctx1, ctx2 context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error { // want `errgroup.Group.Go\(\) closure should use context "ctx1"`
		return nil
	})
	_ = g.Wait()
}

// Multiple ctx params - uses first
// Function has multiple context parameters and uses first
// see also: goroutine, waitgroup
func goodUsesOneOfTwoContexts(ctx1, ctx2 context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx1
		return nil
	})
	_ = g.Wait()
}

// Multiple ctx params - uses second
// Function has multiple context parameters and uses second - should NOT report
// see also: goroutine, waitgroup
func goodUsesSecondOfTwoContexts(ctx1, ctx2 context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx2 // uses second context - should NOT report
		return nil
	})
	_ = g.Wait()
}

// Context as non-first param
// Context is second parameter and is properly used
// see also: goroutine, waitgroup
func goodCtxAsSecondParam(logger interface{}, ctx context.Context) {
	g := new(errgroup.Group)
	g.Go(func() error {
		_ = ctx // ctx is second param but still detected
		return nil
	})
	_ = g.Wait()
}

// Context as non-first param without use
// Context is second parameter but not used in closure
// see also: goroutine, waitgroup
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
