// Package goroutinecreatorchecker checks calls to functions marked with
// //ctxrelay:goroutine_creator for context propagation.
package goroutinecreatorchecker

import (
	"go/ast"

	"github.com/mpyw/ctxrelay/pkg/analyzer/checkers"
)

// Checker checks calls to goroutine creator functions.
type Checker struct {
	creators checkers.GoroutineCreatorMap
}

// New creates a new goroutine creator checker.
func New(creators checkers.GoroutineCreatorMap) *Checker {
	return &Checker{creators: creators}
}

// CheckCall implements checkers.CallChecker.
func (c *Checker) CheckCall(cctx *checkers.CheckContext, call *ast.CallExpr) {
	if len(c.creators) == 0 {
		return
	}

	// Get the function being called
	fn := checkers.GetFuncFromCall(cctx.Pass, call)
	if fn == nil {
		return
	}

	// Check if it's a goroutine creator
	if !c.creators.IsGoroutineCreator(fn) {
		return
	}

	// Find func arguments and check each one for context usage
	funcArgs := checkers.FindFuncArgs(cctx.Pass, call)
	for _, arg := range funcArgs {
		if !cctx.CheckFuncArgUsesContext(arg) {
			cctx.Reportf(arg.Pos(), "%s() func argument should use context %q", fn.Name(), cctx.Scope.Name)
		}
	}
}
