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

// GO20: Defer without ctx
func badGoroutineWithDefer(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		defer fmt.Println("deferred")
		fmt.Println("body")
	}()
}

// GO21: LIMITATION - ctx in deferred nested closure not detected
func badGoroutineUsesCtxOnlyInDeferredClosure(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		defer func() {
			_ = ctx.Done() // ctx in deferred closure doesn't count
		}()
	}()
}

// GO20b: Defer with recovery, no ctx
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

// GO21b: Ctx only in recovery closure (LIMITATION)
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

// GO22: Go in for loop without ctx
func badGoroutinesInLoop(ctx context.Context) {
	for i := 0; i < 3; i++ {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println("loop iteration")
		}()
	}
}

// GO22b: Goroutine in for loop with ctx
func goodGoroutinesInLoopWithCtx(ctx context.Context) {
	for i := 0; i < 3; i++ {
		go func() {
			_ = ctx
		}()
	}
}

// GO23: Go in range loop without ctx
func badGoroutinesInRangeLoop(ctx context.Context) {
	items := []int{1, 2, 3}
	for _, item := range items {
		go func() { // want `goroutine does not propagate context "ctx"`
			fmt.Println(item)
		}()
	}
}

// ===== CONDITIONAL GOROUTINE =====

// GO24: Conditional Go without ctx
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

// GO24b: Conditional goroutine with ctx
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

// GO25: Channel send without ctx
func badGoroutineWithChannelSend(ctx context.Context) {
	ch := make(chan int)
	go func() { // want `goroutine does not propagate context "ctx"`
		ch <- 42
	}()
	<-ch
}

// GO25b: Channel with select on ctx.Done()
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

// GO26: Channel result without ctx
func badGoroutineReturnsViaChannel(ctx context.Context) {
	result := make(chan int)
	go func() { // want `goroutine does not propagate context "ctx"`
		result <- compute()
	}()
	<-result
}

// GO26b: Channel result with ctx
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

// GO27: Select without ctx.Done() case
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

// GO27b: Select with ctx.Done() case
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

// GO28: WaitGroup (traditional Add/Done) without ctx
func badGoroutineWithWaitGroup(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { // want `goroutine does not propagate context "ctx"`
		defer wg.Done()
		fmt.Println("work")
	}()
	wg.Wait()
}

// GO28b: WaitGroup (traditional Add/Done) with ctx
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

// GO29: Method call without ctx
func badGoroutineCallsMethodWithoutCtx(ctx context.Context) {
	w := &worker{name: "test"}
	go func() { // want `goroutine does not propagate context "ctx"`
		w.run()
	}()
}

// GO29b: Method call with ctx
func goodGoroutineCallsMethodWithCtx(ctx context.Context) {
	w := &worker{name: "test"}
	go func() {
		w.runWithCtx(ctx)
	}()
}

// ===== MULTIPLE VARIABLE CAPTURE =====

// GO30: Captures other vars but not ctx
func badGoroutineCapturesOtherButNotCtx(ctx context.Context) {
	x := 42
	y := "hello"
	go func() { // want `goroutine does not propagate context "ctx"`
		fmt.Println(x, y) // captures x, y but NOT ctx
	}()
}

// GO30b: Captures ctx among other vars
func goodGoroutineCapturesCtxAmongOthers(ctx context.Context) {
	x := 42
	y := "hello"
	go func() {
		fmt.Println(x, y)
		_ = ctx
	}()
}

// ===== CONTROL FLOW =====

// GO31: Loop inside goroutine without ctx
func badGoroutineWithLoop(ctx context.Context) {
	go func() { // want `goroutine does not propagate context "ctx"`
		for i := 0; i < 10; i++ {
			fmt.Println(i)
		}
	}()
}

// GO31b: Loop inside goroutine with ctx
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

// GO32: Switch inside goroutine without ctx
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

// GO32b: Switch inside goroutine with ctx
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
