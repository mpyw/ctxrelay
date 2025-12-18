// Package goroutine contains test fixtures for the goroutine context propagation checker.
// This file covers advanced patterns - real-world complex patterns that are not daily
// but commonly seen in production code: defer, loops, channels, WaitGroup, method calls.
// See basic.go for daily patterns and evil.go for adversarial tests.
package goroutine

import (
	"context"
	"fmt"
	"sync"
)

// ===== DEFER PATTERNS =====

// Defer without ctx
// Goroutine with defer statement but no context usage
func badGoroutineWithDefer(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		defer fmt.Println("deferred")
		fmt.Println("body")
	}()
}

// Ctx in deferred nested closure
// LIMITATION: Context used only in deferred nested closure is not detected
func badGoroutineUsesCtxOnlyInDeferredClosure(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		defer func() {
			_ = ctx.Done() // ctx in deferred closure doesn't count
		}()
	}()
}

// Defer with recovery, no ctx
// Goroutine with defer/recover but no context usage
func badGoroutineWithRecovery(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("recovered:", r)
			}
		}()
		panic("test")
	}()
}

// Ctx only in recovery closure
// LIMITATION: Context used only in recovery closure is not detected
func badGoroutineUsesCtxOnlyInRecoveryClosure(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		defer func() {
			if r := recover(); r != nil {
				_ = ctx // ctx in recovery closure doesn't count
			}
		}()
		panic("test")
	}()
}

// ===== GOROUTINE IN LOOP =====

// Go in for loop without ctx
// Goroutine spawned in for loop without context
// see also: errgroup, waitgroup
func badGoroutinesInLoop(ctx context.Context) {
	for i := 0; i < 3; i++ {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("loop iteration")
		}()
	}
}

// Goroutine in for loop with ctx
// Goroutine spawned in for loop with context properly captured
// see also: errgroup, waitgroup
func goodGoroutinesInLoopWithCtx(ctx context.Context) {
	for i := 0; i < 3; i++ {
		go func() {
			_ = ctx
		}()
	}
}

// Go in range loop without ctx
// Goroutine spawned in range loop without context
// see also: errgroup, waitgroup
func badGoroutinesInRangeLoop(ctx context.Context) {
	items := []int{1, 2, 3}
	for _, item := range items {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println(item)
		}()
	}
}

// ===== CONDITIONAL GOROUTINE =====

// Conditional Go without ctx
// Goroutine spawned conditionally without context
// see also: errgroup, waitgroup
func badConditionalGoroutine(ctx context.Context, flag bool) {
	if flag {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("if branch")
		}()
	} else {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("else branch")
		}()
	}
}

// Conditional goroutine with ctx
// Goroutine spawned conditionally with context properly captured
// see also: errgroup, waitgroup
func goodConditionalGoroutine(ctx context.Context, flag bool) {
	if flag {
		go func() {
			_ = ctx
		}()
	} else {
		go func() {
			_ = ctx
		}()
	}
}

// ===== CHANNEL OPERATIONS =====

// Channel send without ctx
// Goroutine sending to channel without context
func badGoroutineWithChannelSend(ctx context.Context) {
	ch := make(chan int)
	go func() { // want `goroutine does not propagate context "ctx"`
		ch <- 42
	}()
	<-ch
}

// Channel with select on ctx.Done
// Goroutine with channel and proper context handling in select
func goodGoroutineWithChannelAndCtx(ctx context.Context) {
	ch := make(chan int)
	go func() {
		select {
		case ch <- 42:
		case <-ctx.Done():
			return
		}
	}()
	<-ch
}

// Channel result without ctx
// Goroutine returning result via channel without context
func badGoroutineReturnsViaChannel(ctx context.Context) {
	result := make(chan int)
	go func() { // want `goroutine does not propagate context "ctx"`
		result <- compute()
	}()
	<-result
}

// Channel result with ctx
// Goroutine returning result via channel with proper context handling
func goodGoroutineReturnsWithCtx(ctx context.Context) {
	result := make(chan int)
	go func() {
		select {
		case result <- compute():
		case <-ctx.Done():
		}
	}()
	<-result
}

func compute() int { return 42 }

// ===== SELECT PATTERNS =====

// Select without ctx.Done case
// Goroutine with multi-case select but no ctx.Done case
func badGoroutineWithMultiCaseSelect(ctx context.Context) {
	ch1 := make(chan int)
	ch2 := make(chan int)
	go func() { // want `goroutine does not propagate context "ctx"`
		select {
		case <-ch1:
			fmt.Println("ch1")
		case <-ch2:
			fmt.Println("ch2")
		}
	}()
}

// Select with ctx.Done case
// Goroutine with select properly handling context cancellation
func goodGoroutineWithCtxInSelect(ctx context.Context) {
	ch1 := make(chan int)
	go func() {
		select {
		case <-ch1:
			fmt.Println("ch1")
		case <-ctx.Done():
			return
		}
	}()
}

// ===== WAITGROUP PATTERN =====

// WaitGroup traditional without ctx
// Traditional WaitGroup (Add/Done) pattern without context
func badGoroutineWithWaitGroup(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { // want `goroutine does not propagate context "ctx"`
		defer wg.Done()
		fmt.Println("work")
	}()
	wg.Wait()
}

// WaitGroup traditional with ctx
// Traditional WaitGroup (Add/Done) pattern with proper context handling
func goodGoroutineWithWaitGroupAndCtx(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			return
		default:
			fmt.Println("work")
		}
	}()
	wg.Wait()
}

// ===== METHOD CALLS =====

type worker struct {
	name string
}

func (w *worker) run() {
	fmt.Println("running:", w.name)
}

func (w *worker) runWithCtx(ctx context.Context) {
	_ = ctx
	fmt.Println("running:", w.name)
}

// Method call without ctx
// Goroutine calling method without passing context
func badGoroutineCallsMethodWithoutCtx(ctx context.Context) {
	w := &worker{name: "test"}
	go func() { // want `goroutine does not propagate context "ctx"`
		w.run()
	}()
}

// Method call with ctx
// Goroutine calling method with context passed
func goodGoroutineCallsMethodWithCtx(ctx context.Context) {
	w := &worker{name: "test"}
	go func() {
		w.runWithCtx(ctx)
	}()
}

// ===== MULTIPLE VARIABLE CAPTURE =====

// Captures other vars but not ctx
// Goroutine captures other variables but not context
func badGoroutineCapturesOtherButNotCtx(ctx context.Context) {
	x := 42
	y := "hello"
	go func() { // want `goroutine does not propagate context "ctx"`
		fmt.Println(x, y) // captures x, y but NOT ctx
	}()
}

// Captures ctx among other vars
// Goroutine captures context along with other variables
func goodGoroutineCapturesCtxAmongOthers(ctx context.Context) {
	x := 42
	y := "hello"
	go func() {
		fmt.Println(x, y)
		_ = ctx
	}()
}

// ===== CONTROL FLOW =====

// Loop inside goroutine without ctx
// Goroutine with internal loop but no context
func badGoroutineWithLoop(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		for i := 0; i < 10; i++ {
			fmt.Println(i)
		}
	}()
}

// Loop inside goroutine with ctx
// Goroutine with internal loop properly checking context
func goodGoroutineUsesCtxInLoop(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// work
			}
		}
	}()
}

// Switch inside goroutine without ctx
// Goroutine with switch statement but no context
func badGoroutineWithSwitch(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		switch x := 1; x {
		case 1:
			fmt.Println("one")
		default:
			fmt.Println("other")
		}
	}()
}

// Switch inside goroutine with ctx
// Goroutine with switch checking context
func goodGoroutineUsesCtxInSwitch(ctx context.Context) {
	go func() {
		switch {
		case ctx.Err() != nil:
			return
		default:
			// continue
		}
	}()
}

// ===== DEEPLY NESTED PATTERNS =====

// Deep nested without ctx
// Three levels of nested goroutines where only the first uses context
// see also: errgroup, waitgroup
func badNestedDeep(ctx context.Context) {
	go func() {
		_ = ctx
		go func() { // want `goroutine does not propagate context "ctx"`
			go func() { // want `goroutine does not propagate context "ctx"`
				fmt.Println("deep")
			}()
		}()
	}()
}
