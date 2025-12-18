// Package gotask contains test fixtures for the gotask context derivation checker.
package gotask

import (
	"context"

	"github.com/my-example-app/telemetry/apm"
	gotask "github.com/siketyan/gotask/v2"
)

// ===== DoAllFnsSettled - SHOULD REPORT =====

// GT01: DoAllFnsSettled - func literal without deriver
func badDoAllFnsSettledNoDeriver(ctx context.Context) {
	_ = gotask.DoAllFnsSettled( // want `gotask\.DoAllFnsSettled\(\) 2nd argument should call goroutine deriver`
		ctx,
		func(ctx context.Context) bool {
			return true
		},
	)
}

// GT02: DoAllFnsSettled - multiple args, some without deriver
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

// GT03: DoAllFnsSettled - deriver called on parent ctx (still bad - deriver must be inside task body)
func badDoAllFnsSettledDerivedParentCtx(ctx context.Context) {
	_ = gotask.DoAllFnsSettled( // want `gotask\.DoAllFnsSettled\(\) 2nd argument should call goroutine deriver`
		apm.NewGoroutineContext(ctx),
		func(ctx context.Context) error {
			return nil
		},
	)
}

// ===== DoAllSettled with NewTask - SHOULD REPORT =====

// GT10: DoAllSettled - NewTask without deriver
func badDoAllSettledNewTaskNoDeriver(ctx context.Context) {
	_ = gotask.DoAllSettled( // want `gotask\.DoAllSettled\(\) 2nd argument should call goroutine deriver`
		ctx,
		gotask.NewTask(func(ctx context.Context) bool {
			return true
		}),
	)
}

// GT11: DoAllSettled - NewTask with deriver on parent ctx (still bad - deriver must be inside task body)
func badDoAllSettledNewTaskDerivedParentCtx(ctx context.Context) {
	_ = gotask.DoAllSettled( // want `gotask\.DoAllSettled\(\) 2nd argument should call goroutine deriver`
		apm.NewGoroutineContext(ctx),
		gotask.NewTask(func(ctx context.Context) bool {
			return true
		}),
	)
}

// ===== DoAsync - SHOULD REPORT =====

// GT20: Task.DoAsync without deriver on ctx
func badTaskDoAsyncNoDeriver(ctx context.Context) {
	task := gotask.NewTask(func(ctx context.Context) error {
		return nil
	})
	errChan := make(chan error)

	task.DoAsync(ctx, errChan) // want `\(\*gotask\.Task\)\.DoAsync\(\) 1st argument should call goroutine deriver`
}

// GT21: Task.DoAsync with nil channel (ctx still needs deriver)
func badTaskDoAsyncNilChannel(ctx context.Context) {
	task := gotask.NewTask(func(ctx context.Context) error {
		return nil
	})

	task.DoAsync(ctx, nil) // want `\(\*gotask\.Task\)\.DoAsync\(\) 1st argument should call goroutine deriver`
}

// GT22: CancelableTask.DoAsync without deriver on ctx
func badCancelableTaskDoAsyncNoDeriver(ctx context.Context) {
	task := gotask.NewTask(func(ctx context.Context) error {
		return nil
	}).Cancelable()
	errChan := make(chan error)

	task.DoAsync(ctx, errChan) // want `\(\*gotask\.CancelableTask\)\.DoAsync\(\) 1st argument should call goroutine deriver`
}

// ===== DoAllFnsSettled - SHOULD NOT REPORT =====

// GT30: DoAllFnsSettled - func literal with deriver
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

// GT31: DoAllFnsSettled - deriver called but result assigned
func goodDoAllFnsSettledDerivedAssigned(ctx context.Context) {
	_ = gotask.DoAllFnsSettled(
		ctx,
		func(ctx context.Context) error {
			_ = apm.NewGoroutineContext(ctx)
			return nil
		},
	)
}

// GT32: DoAllFnsSettled - all args have deriver
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

// GT40: DoAllSettled - NewTask with deriver
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

// GT50: Task.DoAsync with deriver on ctx
func goodTaskDoAsyncWithDeriver(ctx context.Context) {
	task := gotask.NewTask(func(ctx context.Context) error {
		return nil
	})
	errChan := make(chan error)

	task.DoAsync(apm.NewGoroutineContext(ctx), errChan)
}

// GT51: CancelableTask.DoAsync with deriver on ctx
func goodCancelableTaskDoAsyncWithDeriver(ctx context.Context) {
	task := gotask.NewTask(func(ctx context.Context) error {
		return nil
	}).Cancelable()
	errChan := make(chan error)

	task.DoAsync(apm.NewGoroutineContext(ctx), errChan)
}

// ===== Other Do* functions - SHOULD REPORT =====

// GT60: DoAll without deriver
func badDoAllNoDeriver(ctx context.Context) {
	_ = gotask.DoAll( // want `gotask\.DoAll\(\) 2nd argument should call goroutine deriver`
		ctx,
		gotask.NewTask(func(ctx context.Context) gotask.Result[int] {
			return gotask.Result[int]{Value: 1}
		}),
	)
}

// GT61: DoAllFns without deriver
func badDoAllFnsNoDeriver(ctx context.Context) {
	_ = gotask.DoAllFns( // want `gotask\.DoAllFns\(\) 2nd argument should call goroutine deriver`
		ctx,
		func(ctx context.Context) gotask.Result[int] {
			return gotask.Result[int]{Value: 1}
		},
	)
}

// GT62: DoRace without deriver
func badDoRaceNoDeriver(ctx context.Context) {
	_ = gotask.DoRace( // want `gotask\.DoRace\(\) 2nd argument should call goroutine deriver`
		ctx,
		gotask.NewTask(func(ctx context.Context) int {
			return 1
		}),
	)
}

// GT63: DoRaceFns without deriver
func badDoRaceFnsNoDeriver(ctx context.Context) {
	_ = gotask.DoRaceFns( // want `gotask\.DoRaceFns\(\) 2nd argument should call goroutine deriver`
		ctx,
		func(ctx context.Context) int {
			return 1
		},
	)
}

// ===== Other Do* functions - SHOULD NOT REPORT =====

// GT70: DoAll with deriver
func goodDoAllWithDeriver(ctx context.Context) {
	_ = gotask.DoAll(
		ctx,
		gotask.NewTask(func(ctx context.Context) gotask.Result[int] {
			_ = apm.NewGoroutineContext(ctx)
			return gotask.Result[int]{Value: 1}
		}),
	)
}

// GT71: DoAllFns with deriver
func goodDoAllFnsWithDeriver(ctx context.Context) {
	_ = gotask.DoAllFns(
		ctx,
		func(ctx context.Context) gotask.Result[int] {
			_ = apm.NewGoroutineContext(ctx)
			return gotask.Result[int]{Value: 1}
		},
	)
}

// GT72: DoRace with deriver
func goodDoRaceWithDeriver(ctx context.Context) {
	_ = gotask.DoRace(
		ctx,
		gotask.NewTask(func(ctx context.Context) int {
			_ = apm.NewGoroutineContext(ctx)
			return 1
		}),
	)
}

// GT73: DoRaceFns with deriver
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

// GT80: Ignore directive on DoAllFnsSettled
func goodIgnoreDoAllFnsSettled(ctx context.Context) {
	//goroutinectx:ignore
	_ = gotask.DoAllFnsSettled(
		ctx,
		func(ctx context.Context) bool {
			return true
		},
	)
}

// GT81: Ignore directive on Task.DoAsync
func goodIgnoreTaskDoAsync(ctx context.Context) {
	task := gotask.NewTask(func(ctx context.Context) error { return nil })

	//goroutinectx:ignore
	task.DoAsync(ctx, nil)
}

// ===== No ctx param - SHOULD NOT REPORT =====

// GT90: No ctx param - not checked
func goodNoCtxParam() {
	_ = gotask.DoAllFnsSettled(
		context.Background(),
		func(ctx context.Context) bool {
			return true
		},
	)
}
