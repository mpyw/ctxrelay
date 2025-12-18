// Package goroutinecreator checks calls to functions marked with
// //goroutinectx:goroutine_creator for context propagation.
package goroutinecreator

import (
	"go/ast"

	"github.com/mpyw/goroutinectx/internal/context"
	"github.com/mpyw/goroutinectx/internal/directives/creator"
)

// Checker checks calls to goroutine creator functions.
type Checker struct {
	creators creator.Map
}

// New creates a new goroutine creator checker.
func New(creators creator.Map) *Checker {
	return &Checker{creators: creators}
}

// CheckCall implements checkers.CallChecker.
func (c *Checker) CheckCall(cctx *context.CheckContext, call *ast.CallExpr) {
	if len(c.creators) == 0 {
		return
	}

	// Get the function being called
	fn := creator.GetFuncFromCall(cctx.Pass, call)
	if fn == nil {
		return
	}

	// Check if it's a goroutine creator
	if !c.creators.IsGoroutineCreator(fn) {
		return
	}

	// Find func arguments and check each one for context usage
	funcArgs := creator.FindFuncArgs(cctx.Pass, call)
	for _, arg := range funcArgs {
		if !cctx.CheckFuncArgUsesContext(arg) {
			cctx.Reportf(arg.Pos(), "%s() func argument should use context %q", fn.Name(), cctx.Scope.Name)
		}
	}
}
