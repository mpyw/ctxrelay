package slog

import (
	"context"
	"log/slog"
)

// ===== SHOULD REPORT =====

func badInfo(ctx context.Context) {
	slog.Info("hello") // want `use slog.InfoContext with context "ctx" instead of slog.Info`
}

func badDebug(ctx context.Context) {
	slog.Debug("debug message") // want `use slog.DebugContext with context "ctx" instead of slog.Debug`
}

func badWarn(ctx context.Context) {
	slog.Warn("warning") // want `use slog.WarnContext with context "ctx" instead of slog.Warn`
}

func badError(ctx context.Context) {
	slog.Error("error occurred") // want `use slog.ErrorContext with context "ctx" instead of slog.Error`
}

func badLoggerMethod(ctx context.Context, logger *slog.Logger) {
	logger.Info("hello") // want `use slog.InfoContext with context "ctx" instead of slog.Info`
}

func badLoggerMethodDebug(ctx context.Context, logger *slog.Logger) {
	logger.Debug("debug") // want `use slog.DebugContext with context "ctx" instead of slog.Debug`
}

// ===== NESTED FUNCTIONS - SHOULD REPORT =====

func badNestedInnerFunc(ctx context.Context) {
	innerFunc := func() {
		slog.Info("inner") // want `use slog.InfoContext with context "ctx" instead of slog.Info`
	}
	innerFunc()
}

func badNestedClosure(ctx context.Context) {
	func() {
		slog.Info("closure") // want `use slog.InfoContext with context "ctx" instead of slog.Info`
	}()
}

func badNestedDeep(ctx context.Context) {
	func() {
		func() {
			slog.Info("deep") // want `use slog.InfoContext with context "ctx" instead of slog.Info`
		}()
	}()
}

func badNestedWithLogger(ctx context.Context, logger *slog.Logger) {
	func() {
		logger.Warn("nested") // want `use slog.WarnContext with context "ctx" instead of slog.Warn`
	}()
}

// ===== SHOULD NOT REPORT =====

func goodInfoContext(ctx context.Context) {
	slog.InfoContext(ctx, "hello") // OK
}

func goodDebugContext(ctx context.Context) {
	slog.DebugContext(ctx, "debug") // OK
}

func goodWarnContext(ctx context.Context) {
	slog.WarnContext(ctx, "warning") // OK
}

func goodErrorContext(ctx context.Context) {
	slog.ErrorContext(ctx, "error") // OK
}

func goodLoggerInfoContext(ctx context.Context, logger *slog.Logger) {
	logger.InfoContext(ctx, "hello") // OK
}

func goodNoContextParam() {
	slog.Info("hello")
}

func goodNoContextParamLogger(logger *slog.Logger) {
	logger.Info("hello")
}

// ===== NESTED - SHOULD NOT REPORT =====

func goodNestedWithCtx(ctx context.Context) {
	innerFunc := func() {
		slog.InfoContext(ctx, "inner") // OK
	}
	innerFunc()
}

func goodNestedInnerHasOwnCtx(outerCtx context.Context) {
	innerFunc := func(ctx context.Context) {
		slog.InfoContext(ctx, "inner") // OK - uses inner ctx
	}
	innerFunc(outerCtx)
}

// ===== ADVANCED NESTED PATTERNS =====

// Deep nesting (3 levels)
func badNest3Level(ctx context.Context) {
	func() {
		func() {
			func() {
				slog.Info("level 3") // want `use slog.InfoContext with context "ctx" instead of slog.Info`
			}()
		}()
	}()
}

func goodNest3Level(ctx context.Context) {
	func() {
		func() {
			func() {
				slog.InfoContext(ctx, "level 3")
			}()
		}()
	}()
}

// Shadowing - two levels of ctx argument passing
func goodShadowingTwoLevels(ctx1 context.Context) {
	func(ctx2 context.Context) {
		func(ctx3 context.Context) {
			slog.InfoContext(ctx3, "uses ctx3") // OK
		}(ctx2)
	}(ctx1)
}

// Deferred closures
func badDeferred(ctx context.Context) {
	defer func() {
		slog.Info("deferred") // want `use slog.InfoContext with context "ctx" instead of slog.Info`
	}()
}

func goodDeferred(ctx context.Context) {
	defer func() {
		slog.InfoContext(ctx, "deferred")
	}()
}

func badDeferredNested(ctx context.Context) {
	defer func() {
		defer func() {
			slog.Info("nested defer") // want `use slog.InfoContext with context "ctx" instead of slog.Info`
		}()
	}()
}

// Conditional closures
func badConditionalClosure(ctx context.Context, cond bool) {
	if cond {
		func() {
			slog.Info("conditional") // want `use slog.InfoContext with context "ctx" instead of slog.Info`
		}()
	}
}

// Loop closures
func badLoopClosure(ctx context.Context) {
	for i := 0; i < 3; i++ {
		func() {
			slog.Info("loop") // want `use slog.InfoContext with context "ctx" instead of slog.Info`
		}()
	}
}

// Returned closures
func badReturnedClosure(ctx context.Context) func() {
	return func() {
		slog.Info("returned") // want `use slog.InfoContext with context "ctx" instead of slog.Info`
	}
}

func goodReturnedClosure(ctx context.Context) func() {
	return func() {
		slog.InfoContext(ctx, "returned")
	}
}

// Closure in slice
func badClosureInSlice(ctx context.Context) {
	funcs := []func(){
		func() {
			slog.Info("slice 0") // want `use slog.InfoContext with context "ctx" instead of slog.Info`
		},
	}
	for _, f := range funcs {
		f()
	}
}

// Partial usage - ctx used in one place but not another
func badPartialUsage(ctx context.Context) {
	slog.InfoContext(ctx, "first call") // OK
	func() {
		slog.Info("second call") // want `use slog.InfoContext with context "ctx" instead of slog.Info`
	}()
}

// Middle layer introduces ctx (outer has none)
func badMiddleLayerIntroducesCtx() {
	func(ctx context.Context) {
		slog.InfoContext(ctx, "middle layer") // OK
		func() {
			slog.Info("inner") // want `use slog.InfoContext with context "ctx" instead of slog.Info`
		}()
	}(context.Background())
}

func goodMiddleLayerIntroducesCtx() {
	func(ctx context.Context) {
		func() {
			slog.InfoContext(ctx, "inner uses middle ctx") // OK
		}()
	}(context.Background())
}

// Interleaved layers: ctx -> no ctx -> ctx (shadowing)
func badInterleavedLayers(outerCtx context.Context) {
	func() {
		func(middleCtx context.Context) {
			func() {
				slog.Info("interleaved") // want `use slog.InfoContext with context "middleCtx" instead of slog.Info`
			}()
		}(outerCtx)
	}()
}

func goodInterleavedLayers(outerCtx context.Context) {
	func() {
		func(middleCtx context.Context) {
			func() {
				slog.InfoContext(middleCtx, "interleaved") // OK
			}()
		}(outerCtx)
	}()
}

// Multiple ctx parameters
func badMultipleCtxParams(ctx1 context.Context, ctx2 context.Context) {
	slog.Info("multiple ctx") // want `use slog.InfoContext with context "ctx1" instead of slog.Info`
}

func badMultipleCtxParamsNested(ctx1 context.Context, ctx2 context.Context) {
	func() {
		slog.Info("nested multiple") // want `use slog.InfoContext with context "ctx1" instead of slog.Info`
	}()
}

func goodMultipleCtxParams(ctx1 context.Context, ctx2 context.Context) {
	slog.InfoContext(ctx1, "uses ctx1") // OK
}
