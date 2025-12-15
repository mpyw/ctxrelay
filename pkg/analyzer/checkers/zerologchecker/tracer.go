package zerologchecker

import (
	"go/types"

	"golang.org/x/tools/go/ssa"
)

// =============================================================================
// Value Tracing - Strategy Pattern
// =============================================================================
//
// The tracing system follows SSA values backwards to find if context was set.
// It uses a Strategy Pattern with three tracers for zerolog types:
//
//   eventTracer   - traces *zerolog.Event values
//   loggerTracer  - traces zerolog.Logger values
//   contextTracer - traces zerolog.Context values (from With())
//
// Each tracer implements ssaTracer interface with type-specific context checks.
// Cross-type delegation happens when values flow between types (e.g., Event
// created from Logger).
//
// Architecture:
//
//   ┌─────────────┐     ┌─────────────┐     ┌───────────────┐
//   │ eventTracer │────▶│loggerTracer │────▶│ contextTracer │
//   │  (Event)    │◀────│  (Logger)   │◀────│   (Context)   │
//   └─────────────┘     └─────────────┘     └───────────────┘
//         │                   │                    │
//         └───────────────────┴────────────────────┘
//                             │
//                      ┌──────▼──────┐
//                      │ traceCommon │  (handles Phi, UnOp, etc.)
//                      └─────────────┘

// tracer defines the strategy for tracing a specific zerolog type.
// Each implementation knows how to check for context on its type and
// when to delegate to other tracers.
type tracer interface {
	// hasContext checks if this call sets or inherits context.
	// Returns:
	//   - found=true if context is definitely set
	//   - delegate=non-nil to continue tracing with another tracer
	//   - found=false, delegate=nil to continue with current tracer
	hasContext(
		call *ssa.Call,
		callee *ssa.Function,
		recv *types.Var,
	) (found bool, delegate tracer, delegateVal ssa.Value)

	// continueOnReceiverType returns true if this tracer should continue
	// tracing when the receiver matches its type (for chained method calls).
	continueOnReceiverType(recv *types.Var) bool
}

// =============================================================================
// Event Tracer
// =============================================================================

// eventTracer traces *zerolog.Event values for context.
//
// Context sources:
//   - Event.Ctx(ctx): direct context setting
//   - Context.Ctx(ctx): inherited from Context builder
//   - zerolog.Ctx(ctx): Logger returned already has context
//   - Logger.Info/Debug/etc(): inherits from parent Logger
//   - Context.Logger(): inherits from Context builder
type eventTracer struct {
	logger  *loggerTracer
	context *contextTracer
}

func (t *eventTracer) hasContext(
	call *ssa.Call,
	callee *ssa.Function,
	recv *types.Var,
) (bool, tracer, ssa.Value) {
	// Event.Ctx(ctx) or Context.Ctx(ctx) - direct context setting
	if callee.Name() == ctxMethod && recv != nil {
		if isEvent(recv.Type()) || isContext(recv.Type()) {
			return true, nil, nil
		}
	}

	// zerolog.Ctx(ctx) - returns Logger with context
	if isCtxFunc(callee) {
		return true, nil, nil
	}

	// Logger.Info/Debug/etc() - delegate to logger tracer
	if isLogLevelMethod(callee.Name()) && recv != nil && isLogger(recv.Type()) {
		if len(call.Call.Args) > 0 {
			return false, t.logger, call.Call.Args[0]
		}
	}

	// Context.Logger() - delegate to context tracer
	if callee.Name() == loggerMethod && recv != nil && isContext(recv.Type()) {
		if len(call.Call.Args) > 0 {
			return false, t.context, call.Call.Args[0]
		}
	}

	return false, nil, nil
}

func (t *eventTracer) continueOnReceiverType(recv *types.Var) bool {
	return recv != nil && isEvent(recv.Type())
}

// =============================================================================
// Logger Tracer
// =============================================================================

// loggerTracer traces zerolog.Logger values for context.
//
// Context sources:
//   - zerolog.Ctx(ctx): returns Logger from context
//   - Context.Logger(): inherits from Context builder
//   - Logger.With(): inherits from parent Logger (via Context)
type loggerTracer struct {
	context *contextTracer
}

func (t *loggerTracer) hasContext(
	call *ssa.Call,
	callee *ssa.Function,
	recv *types.Var,
) (bool, tracer, ssa.Value) {
	// zerolog.Ctx(ctx) - returns Logger with context
	if isCtxFunc(callee) {
		return true, nil, nil
	}

	// Context.Logger() - delegate to context tracer
	if callee.Name() == loggerMethod && recv != nil && isContext(recv.Type()) {
		if len(call.Call.Args) > 0 {
			return false, t.context, call.Call.Args[0]
		}
	}

	// Logger.With() - trace parent Logger (self-delegation)
	if callee.Name() == withMethod && recv != nil && isLogger(recv.Type()) {
		if len(call.Call.Args) > 0 {
			return false, t, call.Call.Args[0]
		}
	}

	return false, nil, nil
}

func (t *loggerTracer) continueOnReceiverType(recv *types.Var) bool {
	return recv != nil && isLogger(recv.Type())
}

// =============================================================================
// Context Tracer
// =============================================================================

// contextTracer traces zerolog.Context values for context.
//
// Context sources:
//   - Context.Ctx(ctx): direct context setting
//   - Logger.With(): inherits from parent Logger
type contextTracer struct {
	logger *loggerTracer
}

func (t *contextTracer) hasContext(
	call *ssa.Call,
	callee *ssa.Function,
	recv *types.Var,
) (bool, tracer, ssa.Value) {
	// Context.Ctx(ctx) - direct context setting
	if callee.Name() == ctxMethod && recv != nil && isContext(recv.Type()) {
		return true, nil, nil
	}

	// Logger.With() - delegate to logger tracer
	if callee.Name() == withMethod && recv != nil && isLogger(recv.Type()) {
		if len(call.Call.Args) > 0 {
			return false, t.logger, call.Call.Args[0]
		}
	}

	return false, nil, nil
}

func (t *contextTracer) continueOnReceiverType(recv *types.Var) bool {
	return recv != nil && isContext(recv.Type())
}

// =============================================================================
// Tracer Factory
// =============================================================================

// newTracers creates the interconnected tracer instances.
func newTracers() *eventTracer {
	context := &contextTracer{}
	logger := &loggerTracer{context: context}
	context.logger = logger
	event := &eventTracer{logger: logger, context: context}
	return event
}
