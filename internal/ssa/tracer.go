package ssa

import (
	"go/types"

	"golang.org/x/tools/go/ssa"

	"github.com/mpyw/goroutinectx/internal/directives/carrier"
	"github.com/mpyw/goroutinectx/internal/directives/deriver"
	"github.com/mpyw/goroutinectx/internal/funcspec"
	"github.com/mpyw/goroutinectx/internal/typeutil"
)

// Tracer provides SSA-based value tracing for goroutinectx.
type Tracer struct{}

// NewTracer creates a new SSA tracer.
func NewTracer() *Tracer {
	return &Tracer{}
}

// =============================================================================
// Closure Context Checking
// =============================================================================

// ClosureCapturesContext checks if a closure captures any context.Context variable
// or a configured carrier type. It includes nested closures due to SSA's FreeVars propagation.
func (t *Tracer) ClosureCapturesContext(closure *ssa.Function, carriers []carrier.Carrier) bool {
	if closure == nil {
		return false
	}

	// Check free variables (captured from enclosing scope)
	for _, fv := range closure.FreeVars {
		if typeutil.IsContextOrCarrierType(fv.Type(), carriers) {
			return true
		}
	}

	return false
}

// =============================================================================
// Deriver Function Detection
// =============================================================================

// DeriverResult represents the result of deriver function detection.
type DeriverResult struct {
	// FoundAtStart indicates the deriver is called at goroutine start (not in defer)
	FoundAtStart bool
	// FoundOnlyInDefer indicates the deriver is called, but only in defer statements
	FoundOnlyInDefer bool
}

// ClosureCallsDeriver checks if a closure calls any of the required deriver functions.
// It traverses into immediately-invoked function expressions (IIFE) but tracks
// whether calls are made in defer statements.
func (t *Tracer) ClosureCallsDeriver(closure *ssa.Function, matcher *deriver.Matcher) DeriverResult {
	if closure == nil || matcher == nil || matcher.IsEmpty() {
		return DeriverResult{FoundAtStart: true} // No check needed
	}

	// Collect all function calls with their defer status
	calls := t.collectDeriverCalls(closure, false, make(map[*ssa.Function]bool))

	// Check if any OR group is satisfied
	for _, andGroup := range matcher.OrGroups {
		startResult := t.checkAndGroup(calls, andGroup, false)
		if startResult {
			return DeriverResult{FoundAtStart: true}
		}
	}

	// Check if deriver is only in defer
	for _, andGroup := range matcher.OrGroups {
		deferResult := t.checkAndGroup(calls, andGroup, true)
		if deferResult {
			return DeriverResult{FoundOnlyInDefer: true}
		}
	}

	return DeriverResult{}
}

// deriverCall represents a function call with its context (defer or not).
type deriverCall struct {
	fn      *types.Func
	inDefer bool
}

// collectDeriverCalls collects all function calls in a closure, including IIFE.
func (t *Tracer) collectDeriverCalls(fn *ssa.Function, inDefer bool, visited map[*ssa.Function]bool) []deriverCall {
	if fn == nil || visited[fn] {
		return nil
	}
	visited[fn] = true

	var calls []deriverCall

	for _, block := range fn.Blocks {
		for _, instr := range block.Instrs {
			switch v := instr.(type) {
			case *ssa.Call:
				// Regular function call
				if calledFn := t.extractCalledFunc(v); calledFn != nil {
					calls = append(calls, deriverCall{fn: calledFn, inDefer: inDefer})
				}
				// Check for IIFE: call where the callee is a MakeClosure
				if iifeFn := t.extractIIFE(v); iifeFn != nil {
					// Traverse into the IIFE with the same defer status
					calls = append(calls, t.collectDeriverCalls(iifeFn, inDefer, visited)...)
				}

			case *ssa.Defer:
				// Deferred function call - mark as inDefer
				if calledFn := t.extractCalledFuncFromCallCommon(&v.Call); calledFn != nil {
					calls = append(calls, deriverCall{fn: calledFn, inDefer: true})
				}
				// Check for deferred IIFE
				if iifeFn := t.extractIIFEFromCallCommon(&v.Call); iifeFn != nil {
					// Traverse into the deferred IIFE with inDefer=true
					calls = append(calls, t.collectDeriverCalls(iifeFn, true, visited)...)
				}
			}
		}
	}

	return calls
}

// extractCalledFunc extracts the types.Func from a Call instruction.
func (t *Tracer) extractCalledFunc(call *ssa.Call) *types.Func {
	return t.extractCalledFuncFromCallCommon(&call.Call)
}

// extractCalledFuncFromCallCommon extracts the types.Func from a CallCommon.
func (t *Tracer) extractCalledFuncFromCallCommon(call *ssa.CallCommon) *types.Func {
	if call.IsInvoke() {
		// Interface method call
		return call.Method
	}

	// Static call
	if fn := call.StaticCallee(); fn != nil {
		if obj, ok := fn.Object().(*types.Func); ok {
			return obj
		}
	}

	return nil
}

// extractIIFE checks if a Call instruction is an IIFE (immediately invoked function expression).
// Returns the called function if it's an IIFE, nil otherwise.
func (t *Tracer) extractIIFE(call *ssa.Call) *ssa.Function {
	return t.extractIIFEFromCallCommon(&call.Call)
}

// extractIIFEFromCallCommon checks if a CallCommon is an IIFE.
func (t *Tracer) extractIIFEFromCallCommon(call *ssa.CallCommon) *ssa.Function {
	if call.IsInvoke() {
		return nil
	}

	// Check if the callee is a MakeClosure
	if mc, ok := call.Value.(*ssa.MakeClosure); ok {
		if fn, ok := mc.Fn.(*ssa.Function); ok {
			return fn
		}
	}

	// Check if the callee is a direct function reference
	if fn, ok := call.Value.(*ssa.Function); ok {
		// Only count as IIFE if it's an anonymous function (has no name in package scope)
		if fn.Parent() != nil {
			return fn
		}
	}

	return nil
}

// checkAndGroup checks if all specs in an AND group are satisfied.
// If includeDefer is false, only non-defer calls are considered.
// If includeDefer is true, all calls (including defer) are considered.
func (t *Tracer) checkAndGroup(calls []deriverCall, andGroup []funcspec.Spec, includeDefer bool) bool {
	for _, spec := range andGroup {
		found := false
		for _, call := range calls {
			if !includeDefer && call.inDefer {
				continue
			}
			if call.fn != nil && spec.Matches(call.fn) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
