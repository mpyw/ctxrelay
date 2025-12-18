// Package gotask contains evil edge case tests for the gotask context derivation checker.
package gotask

import (
	"context"

	"github.com/my-example-app/telemetry/apm"
	gotask "github.com/siketyan/gotask/v2"
)

// ===== VARIADIC EXPANSION - SHOULD REPORT =====

// Variadic expansion without deriver
// Variadic expansion of tasks without deriver
func badVariadicExpansion(ctx context.Context) {
	tasks := []func(context.Context) error{
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return nil },
	}
	_ = gotask.DoAllFnsSettled(ctx, tasks...) // want `gotask\.DoAllFnsSettled\(\) variadic argument should call goroutine deriver`
}

// ===== VARIABLE TASK - SHOULD REPORT =====

// Task stored in variable (func literal without deriver)
// Variable task without deriver
func badVariableTaskNoDeriver(ctx context.Context) {
	fn := func(ctx context.Context) error {
		return nil
	}
	_ = gotask.DoAllFnsSettled(ctx, fn) // want `gotask\.DoAllFnsSettled\(\) 2nd argument should call goroutine deriver`
}

// NewTask stored in variable
// NewTask in variable without deriver
func badNewTaskVariableNoDeriver(ctx context.Context) {
	task := gotask.NewTask(func(ctx context.Context) error {
		return nil
	})
	_ = gotask.DoAllSettled(ctx, task) // want `gotask\.DoAllSettled\(\) 2nd argument should call goroutine deriver`
}

// ===== NESTED CLOSURE - SHOULD REPORT (deriver in nested closure doesn't count) =====

// Deriver only in nested closure
// LIMITATION: Deriver in nested closure is not detected at top level
func badDerivedInNestedClosure(ctx context.Context) {
	_ = gotask.DoAllFnsSettled( // want `gotask\.DoAllFnsSettled\(\) 2nd argument should call goroutine deriver`
		ctx,
		func(ctx context.Context) error {
			// Deriver is in a nested closure - won't be detected at top level
			go func() {
				_ = apm.NewGoroutineContext(ctx)
			}()
			return nil
		},
	)
}

// ===== CONDITIONAL DERIVER - SHOULD NOT REPORT =====

// Deriver in if branch
// Deriver in if branch (any presence should satisfy)
func goodDerivedInIfBranch(ctx context.Context) {
	_ = gotask.DoAllFnsSettled(
		ctx,
		func(ctx context.Context) error {
			if true {
				_ = apm.NewGoroutineContext(ctx)
			}
			return nil
		},
	)
}

// ===== METHOD CHAINING - SHOULD REPORT =====

// Chained task creation - DoAsync on result of method chain
// DoAsync on method chain result without deriver
func badChainedTaskDoAsync(ctx context.Context) {
	gotask.NewTask(func(ctx context.Context) error {
		return nil
	}).DoAsync(ctx, nil) // want `\(\*gotask\.Task\)\.DoAsync\(\) 1st argument should call goroutine deriver`
}

// Cancelable chain DoAsync without deriver
// Cancelable chain DoAsync without deriver
func badCancelableChainDoAsync(ctx context.Context) {
	gotask.NewTask(func(ctx context.Context) error {
		return nil
	}).Cancelable().DoAsync(ctx, nil) // want `\(\*gotask\.CancelableTask\)\.DoAsync\(\) 1st argument should call goroutine deriver`
}

// ===== METHOD CHAINING - SHOULD NOT REPORT =====

// Chained task creation with derived ctx
// DoAsync on method chain with derived ctx
func goodChainedTaskDoAsyncWithDeriver(ctx context.Context) {
	gotask.NewTask(func(ctx context.Context) error {
		return nil
	}).DoAsync(apm.NewGoroutineContext(ctx), nil)
}

// Cancelable chain DoAsync with derived ctx
// Cancelable chain DoAsync with derived ctx
func goodCancelableChainDoAsyncWithDeriver(ctx context.Context) {
	gotask.NewTask(func(ctx context.Context) error {
		return nil
	}).Cancelable().DoAsync(apm.NewGoroutineContext(ctx), nil)
}

// ===== VARIADIC EXPANSION FROM VARIABLE - LIMITATION (can't trace variable) =====

// LIMITATION - Variable slice expansion can't be traced
// Variable slice expansion with deriver can't be traced
func limitationVariadicExpansionVariable(ctx context.Context) {
	tasks := []func(context.Context) error{
		func(ctx context.Context) error {
			_ = apm.NewGoroutineContext(ctx)
			return nil
		},
	}
	// Reports because we can't trace into variable assignment
	_ = gotask.DoAllFnsSettled(ctx, tasks...) // want `gotask\.DoAllFnsSettled\(\) variadic argument should call goroutine deriver`
}

// ===== VARIABLE TASK - SHOULD NOT REPORT (variable tracing works) =====

// Variable func assignment with deriver is traced correctly
// Variable task with deriver is traced
func goodVariableTaskWithDeriver(ctx context.Context) {
	fn := func(ctx context.Context) error {
		_ = apm.NewGoroutineContext(ctx)
		return nil
	}
	_ = gotask.DoAllFnsSettled(ctx, fn)
}

// NewTask in variable with deriver is traced correctly
// NewTask in variable with deriver is traced
func goodNewTaskVariableWithDeriver(ctx context.Context) {
	task := gotask.NewTask(func(ctx context.Context) error {
		_ = apm.NewGoroutineContext(ctx)
		return nil
	})
	_ = gotask.DoAllSettled(ctx, task)
}

// ===== DERIVER IN DEFER CLOSURE - LIMITATION (defer closure is a nested FuncLit) =====

// LIMITATION - Deriver in defer closure not detected
// Defer closure is treated as nested FuncLit, not traversed
func limitationDerivedInDeferClosure(ctx context.Context) {
	_ = gotask.DoAllFnsSettled( // want `gotask\.DoAllFnsSettled\(\) 2nd argument should call goroutine deriver`
		ctx,
		func(ctx context.Context) error {
			// The defer's func() is a FuncLit, so we don't look inside
			defer func() {
				_ = apm.NewGoroutineContext(ctx)
			}()
			return nil
		},
	)
}

// ===== MIXED DERIVER AND NON-DERIVER - SHOULD REPORT =====

// Multiple tasks, only some have deriver
// Mixed deriver and non-deriver tasks
func badMixedDerivers(ctx context.Context) {
	_ = gotask.DoAllSettled( // want `gotask\.DoAllSettled\(\) 3rd argument should call goroutine deriver`
		ctx,
		gotask.NewTask(func(ctx context.Context) error {
			_ = apm.NewGoroutineContext(ctx)
			return nil
		}),
		gotask.NewTask(func(ctx context.Context) error {
			return nil // No deriver!
		}),
	)
}

// ===== DOASYNC ON POINTER - SHOULD REPORT =====

// Task pointer DoAsync
// DoAsync on task pointer without deriver
func badTaskPointerDoAsync(ctx context.Context) {
	task := gotask.NewTask(func(ctx context.Context) error {
		return nil
	})
	taskPtr := &task
	taskPtr.DoAsync(ctx, nil) // want `\(\*gotask\.Task\)\.DoAsync\(\) 1st argument should call goroutine deriver`
}

// ===== DOASYNC ON POINTER - SHOULD NOT REPORT =====

// Task pointer DoAsync with deriver
// DoAsync on task pointer with deriver
func goodTaskPointerDoAsyncWithDeriver(ctx context.Context) {
	task := gotask.NewTask(func(ctx context.Context) error {
		return nil
	})
	taskPtr := &task
	taskPtr.DoAsync(apm.NewGoroutineContext(ctx), nil)
}

// ===== HIGHER-ORDER FUNCTIONS - SHOULD REPORT/NOT REPORT =====

// Higher-order function returning task WITHOUT deriver
// Higher-order task factory without deriver
func badHigherOrderTaskFactoryNoDeriver(ctx context.Context) {
	makeTask := func() gotask.Task[error] {
		return gotask.NewTask(func(ctx context.Context) error {
			return nil // No deriver
		})
	}
	_ = gotask.DoAllSettled(ctx, makeTask()) // want `gotask\.DoAllSettled\(\) 2nd argument should call goroutine deriver`
}

// Higher-order function returning task WITH deriver
// Higher-order task factory with deriver
func goodHigherOrderTaskFactoryWithDeriver(ctx context.Context) {
	makeTask := func() gotask.Task[error] {
		return gotask.NewTask(func(ctx context.Context) error {
			_ = apm.NewGoroutineContext(ctx)
			return nil
		})
	}
	_ = gotask.DoAllSettled(ctx, makeTask())
}

// ===== INTERFACE - LIMITATION (reports because can't trace) =====

type taskMaker interface {
	MakeTask() gotask.Task[error]
}

// LIMITATION - Interface method returns can't be traced
// Interface task maker can't be traced
func limitationInterfaceTaskMaker(ctx context.Context, maker taskMaker) {
	// Reports because maker.MakeTask() can't be traced
	_ = gotask.DoAllSettled(ctx, maker.MakeTask()) // want `gotask\.DoAllSettled\(\) 2nd argument should call goroutine deriver`
}

// ===== EDGE CASES - SHOULD NOT REPORT (not gotask or edge behavior) =====

// Empty call (less than 2 args)
// Empty DoAll call - not checked
func goodEmptyDoAll(ctx context.Context) {
	_ = gotask.DoAll[int](ctx)
}

// Only ctx arg
// Only ctx arg - not checked
func goodOnlyCtxArg(ctx context.Context) {
	// This would be invalid Go code if DoAll required args, but tests analyzer edge
}

// Multiple DoAsync calls in same function
// Multiple DoAsync calls - each reported independently
func badMultipleDoAsync(ctx context.Context) {
	task1 := gotask.NewTask(func(ctx context.Context) error { return nil })
	task2 := gotask.NewTask(func(ctx context.Context) error { return nil })

	task1.DoAsync(ctx, nil)                         // want `\(\*gotask\.Task\)\.DoAsync\(\) 1st argument should call goroutine deriver`
	task2.DoAsync(apm.NewGoroutineContext(ctx), nil) // OK - has deriver
	task1.DoAsync(ctx, nil)                         // want `\(\*gotask\.Task\)\.DoAsync\(\) 1st argument should call goroutine deriver`
}

// Context with different param name
// Different ctx param name
func badDifferentCtxName(c context.Context) {
	_ = gotask.DoAllFnsSettled( // want `gotask\.DoAllFnsSettled\(\) 2nd argument should call goroutine deriver`
		c,
		func(ctx context.Context) error {
			return nil
		},
	)
}

// Context param with unusual name
// Unusual ctx param name
func badContextParamUnusualName(myCtx context.Context) {
	_ = gotask.DoAllFnsSettled( // want `gotask\.DoAllFnsSettled\(\) 2nd argument should call goroutine deriver`
		myCtx,
		func(ctx context.Context) error {
			return nil
		},
	)
}

// Good with different ctx param names
// Different ctx names with deriver
func goodDifferentCtxNames(c context.Context) {
	_ = gotask.DoAllFnsSettled(
		c,
		func(ctx context.Context) error {
			_ = apm.NewGoroutineContext(ctx)
			return nil
		},
	)
}

// ===== RECURSIVE TASKS - SHOULD REPORT =====

// Task that creates another task (nested gotask call)
// Nested task creation - outer lacks deriver
func badNestedTaskCreation(ctx context.Context) {
	_ = gotask.DoAllFnsSettled( // want `gotask\.DoAllFnsSettled\(\) 2nd argument should call goroutine deriver`
		ctx,
		func(ctx context.Context) error {
			// This outer task doesn't call deriver
			_ = gotask.DoAllFnsSettled(
				ctx,
				func(ctx context.Context) error {
					_ = apm.NewGoroutineContext(ctx) // Inner has deriver but outer doesn't
					return nil
				},
			)
			return nil
		},
	)
}

// ===== DERIVER CALL IN EXPRESSION CONTEXT =====

// Deriver result used directly in expression
// Deriver result used in expression
func goodDerivedUsedInExpression(ctx context.Context) {
	_ = gotask.DoAllFnsSettled(
		ctx,
		func(ctx context.Context) error {
			// Deriver called in expression context
			doSomethingWithContext(apm.NewGoroutineContext(ctx))
			return nil
		},
	)
}

func doSomethingWithContext(_ context.Context) {}

// Deriver result stored and used
// Deriver result stored then used
func goodDerivedStoredAndUsed(ctx context.Context) {
	_ = gotask.DoAllFnsSettled(
		ctx,
		func(ctx context.Context) error {
			derivedCtx := apm.NewGoroutineContext(ctx)
			doSomethingWithContext(derivedCtx)
			return nil
		},
	)
}

// ===== EARLY RETURN PATHS =====

// Deriver called before early return
// Deriver before early return
func goodDerivedBeforeEarlyReturn(ctx context.Context) {
	_ = gotask.DoAllFnsSettled(
		ctx,
		func(ctx context.Context) error {
			_ = apm.NewGoroutineContext(ctx)
			if true {
				return nil // Early return
			}
			return nil
		},
	)
}

// Deriver only on one branch
// Deriver on one branch (any call counts)
func goodDerivedOnOneBranch(ctx context.Context) {
	_ = gotask.DoAllFnsSettled(
		ctx,
		func(ctx context.Context) error {
			if false {
				_ = apm.NewGoroutineContext(ctx) // Only called conditionally but detected
			}
			return nil
		},
	)
}
