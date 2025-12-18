// Package gotask contains test fixtures for the gotask context derivation checker.
package gotask

import (
	"context"

	"github.com/my-example-app/telemetry/apm"
	gotask "github.com/siketyan/gotask/v2"
)

// ===== DoAllFnsSettled - SHOULD REPORT =====

// DoAllFnsSettled - func literal without deriver
// DoAllFnsSettled with func literal that doesn't call goroutine deriver
func badDoAllFnsSettledNoDeriver(ctx context.Context) {
	_ = gotask.DoAllFnsSettled( // want `gotask\.DoAllFnsSettled\(\) 2nd argument should call goroutine deriver`
		ctx,
		func(ctx context.Context) bool {
			return true
		},
	)
}

// DoAllFnsSettled - multiple args, some without deriver
// DoAllFnsSettled with multiple args where some don't call goroutine deriver
func badDoAllFnsSettledPartialDeriver(ctx context.Context) {
	_ = gotask.DoAllFnsSettled( // want `gotask\.DoAllFnsSettled\(\) 3rd argument should call goroutine deriver` `gotask\.DoAllFnsSettled\(\) 5th argument should call goroutine deriver`
		ctx,
		func(ctx context.Context) error {
			_ = apm.NewGoroutineContext(ctx)
			return nil
		},
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			_ = apm.NewGoroutineContext(ctx)
			return nil
		},
		func(ctx context.Context) error {
			return nil
		},
	)
}

// DoAllFnsSettled - deriver called on parent ctx
// Deriver called on parent ctx is still bad - deriver must be inside task body
func badDoAllFnsSettledDerivedParentCtx(ctx context.Context) {
	_ = gotask.DoAllFnsSettled( // want `gotask\.DoAllFnsSettled\(\) 2nd argument should call goroutine deriver`
		apm.NewGoroutineContext(ctx),
		func(ctx context.Context) error {
			return nil
		},
	)
}

// ===== DoAllSettled with NewTask - SHOULD REPORT =====

// DoAllSettled - NewTask without deriver
// DoAllSettled with NewTask that doesn't call goroutine deriver
func badDoAllSettledNewTaskNoDeriver(ctx context.Context) {
	_ = gotask.DoAllSettled( // want `gotask\.DoAllSettled\(\) 2nd argument should call goroutine deriver`
		ctx,
		gotask.NewTask(func(ctx context.Context) bool {
			return true
		}),
	)
}

// DoAllSettled - NewTask with deriver on parent ctx
// NewTask with deriver on parent ctx is still bad - deriver must be inside task body
func badDoAllSettledNewTaskDerivedParentCtx(ctx context.Context) {
	_ = gotask.DoAllSettled( // want `gotask\.DoAllSettled\(\) 2nd argument should call goroutine deriver`
		apm.NewGoroutineContext(ctx),
		gotask.NewTask(func(ctx context.Context) bool {
			return true
		}),
	)
}

// ===== DoAsync - SHOULD REPORT =====

// Task.DoAsync without deriver
// Task.DoAsync called without deriver on ctx
func badTaskDoAsyncNoDeriver(ctx context.Context) {
	task := gotask.NewTask(func(ctx context.Context) error {
		return nil
	})
	errChan := make(chan error)

	task.DoAsync(ctx, errChan) // want `\(\*gotask\.Task\)\.DoAsync\(\) 1st argument should call goroutine deriver`
}

// Task.DoAsync with nil channel
// Task.DoAsync with nil channel still needs deriver on ctx
func badTaskDoAsyncNilChannel(ctx context.Context) {
	task := gotask.NewTask(func(ctx context.Context) error {
		return nil
	})

	task.DoAsync(ctx, nil) // want `\(\*gotask\.Task\)\.DoAsync\(\) 1st argument should call goroutine deriver`
}

// CancelableTask.DoAsync without deriver
// CancelableTask.DoAsync called without deriver on ctx
func badCancelableTaskDoAsyncNoDeriver(ctx context.Context) {
	task := gotask.NewTask(func(ctx context.Context) error {
		return nil
	}).Cancelable()
	errChan := make(chan error)

	task.DoAsync(ctx, errChan) // want `\(\*gotask\.CancelableTask\)\.DoAsync\(\) 1st argument should call goroutine deriver`
}

// ===== DoAllFnsSettled - SHOULD NOT REPORT =====

// DoAllFnsSettled - func literal with deriver
// DoAllFnsSettled with func literal that calls goroutine deriver
func goodDoAllFnsSettledWithDeriver(ctx context.Context) {
	_ = gotask.DoAllFnsSettled(
		ctx,
		func(ctx context.Context) bool {
			ctx = apm.NewGoroutineContext(ctx)
			_ = ctx
			return true
		},
	)
}

// DoAllFnsSettled - deriver called but result assigned
// DoAllFnsSettled with deriver result assigned to variable
func goodDoAllFnsSettledDerivedAssigned(ctx context.Context) {
	_ = gotask.DoAllFnsSettled(
		ctx,
		func(ctx context.Context) error {
			_ = apm.NewGoroutineContext(ctx)
			return nil
		},
	)
}

// DoAllFnsSettled - all args have deriver
// DoAllFnsSettled with all args calling goroutine deriver
func goodDoAllFnsSettledAllWithDeriver(ctx context.Context) {
	_ = gotask.DoAllFnsSettled(
		ctx,
		func(ctx context.Context) error {
			_ = apm.NewGoroutineContext(ctx)
			return nil
		},
		func(ctx context.Context) error {
			_ = apm.NewGoroutineContext(ctx)
			return nil
		},
	)
}

// ===== DoAllSettled with NewTask - SHOULD NOT REPORT =====

// DoAllSettled - NewTask with deriver
// DoAllSettled with NewTask that calls goroutine deriver
func goodDoAllSettledNewTaskWithDeriver(ctx context.Context) {
	_ = gotask.DoAllSettled(
		ctx,
		gotask.NewTask(func(ctx context.Context) bool {
			ctx = apm.NewGoroutineContext(ctx)
			_ = ctx
			return true
		}),
	)
}

// ===== DoAsync - SHOULD NOT REPORT =====

// Task.DoAsync with deriver
// Task.DoAsync called with deriver on ctx
func goodTaskDoAsyncWithDeriver(ctx context.Context) {
	task := gotask.NewTask(func(ctx context.Context) error {
		return nil
	})
	errChan := make(chan error)

	task.DoAsync(apm.NewGoroutineContext(ctx), errChan)
}

// CancelableTask.DoAsync with deriver
// CancelableTask.DoAsync called with deriver on ctx
func goodCancelableTaskDoAsyncWithDeriver(ctx context.Context) {
	task := gotask.NewTask(func(ctx context.Context) error {
		return nil
	}).Cancelable()
	errChan := make(chan error)

	task.DoAsync(apm.NewGoroutineContext(ctx), errChan)
}

// ===== Other Do* functions - SHOULD REPORT =====

// DoAll without deriver
// DoAll with task that doesn't call goroutine deriver
func badDoAllNoDeriver(ctx context.Context) {
	_ = gotask.DoAll( // want `gotask\.DoAll\(\) 2nd argument should call goroutine deriver`
		ctx,
		gotask.NewTask(func(ctx context.Context) gotask.Result[int] {
			return gotask.Result[int]{Value: 1}
		}),
	)
}

// DoAllFns without deriver
// DoAllFns with func that doesn't call goroutine deriver
func badDoAllFnsNoDeriver(ctx context.Context) {
	_ = gotask.DoAllFns( // want `gotask\.DoAllFns\(\) 2nd argument should call goroutine deriver`
		ctx,
		func(ctx context.Context) gotask.Result[int] {
			return gotask.Result[int]{Value: 1}
		},
	)
}

// DoRace without deriver
// DoRace with task that doesn't call goroutine deriver
func badDoRaceNoDeriver(ctx context.Context) {
	_ = gotask.DoRace( // want `gotask\.DoRace\(\) 2nd argument should call goroutine deriver`
		ctx,
		gotask.NewTask(func(ctx context.Context) int {
			return 1
		}),
	)
}

// DoRaceFns without deriver
// DoRaceFns with func that doesn't call goroutine deriver
func badDoRaceFnsNoDeriver(ctx context.Context) {
	_ = gotask.DoRaceFns( // want `gotask\.DoRaceFns\(\) 2nd argument should call goroutine deriver`
		ctx,
		func(ctx context.Context) int {
			return 1
		},
	)
}

// ===== Other Do* functions - SHOULD NOT REPORT =====

// DoAll with deriver
// DoAll with task that calls goroutine deriver
func goodDoAllWithDeriver(ctx context.Context) {
	_ = gotask.DoAll(
		ctx,
		gotask.NewTask(func(ctx context.Context) gotask.Result[int] {
			_ = apm.NewGoroutineContext(ctx)
			return gotask.Result[int]{Value: 1}
		}),
	)
}

// DoAllFns with deriver
// DoAllFns with func that calls goroutine deriver
func goodDoAllFnsWithDeriver(ctx context.Context) {
	_ = gotask.DoAllFns(
		ctx,
		func(ctx context.Context) gotask.Result[int] {
			_ = apm.NewGoroutineContext(ctx)
			return gotask.Result[int]{Value: 1}
		},
	)
}

// DoRace with deriver
// DoRace with task that calls goroutine deriver
func goodDoRaceWithDeriver(ctx context.Context) {
	_ = gotask.DoRace(
		ctx,
		gotask.NewTask(func(ctx context.Context) int {
			_ = apm.NewGoroutineContext(ctx)
			return 1
		}),
	)
}

// DoRaceFns with deriver
// DoRaceFns with func that calls goroutine deriver
func goodDoRaceFnsWithDeriver(ctx context.Context) {
	_ = gotask.DoRaceFns(
		ctx,
		func(ctx context.Context) int {
			_ = apm.NewGoroutineContext(ctx)
			return 1
		},
	)
}

// ===== Ignore directive =====

// Ignore directive on DoAllFnsSettled
// Ignore directive suppresses warning on DoAllFnsSettled
func goodIgnoreDoAllFnsSettled(ctx context.Context) {
	//goroutinectx:ignore
	_ = gotask.DoAllFnsSettled(
		ctx,
		func(ctx context.Context) bool {
			return true
		},
	)
}

// Ignore directive on Task.DoAsync
// Ignore directive suppresses warning on Task.DoAsync
func goodIgnoreTaskDoAsync(ctx context.Context) {
	task := gotask.NewTask(func(ctx context.Context) error { return nil })

	//goroutinectx:ignore
	task.DoAsync(ctx, nil)
}

// ===== No ctx param - SHOULD NOT REPORT =====

// No ctx param
// Function has no context parameter - not checked
// see also: goroutinecreator
func goodNoCtxParam() {
	_ = gotask.DoAllFnsSettled(
		context.Background(),
		func(ctx context.Context) bool {
			return true
		},
	)
}
