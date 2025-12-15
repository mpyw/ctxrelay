// Package errgroupchecker checks errgroup.Group.Go() calls for context propagation.
package errgroupchecker

import (
	"go/ast"

	"github.com/mpyw/ctxrelay/pkg/analyzer/checkers"
)

const pkgPath = "golang.org/x/sync/errgroup"

// Checker checks errgroup.Group.Go() calls for context propagation.
type Checker struct{}

// New creates a new errgroup checker.
func New() *Checker {
	return &Checker{}
}

// CheckCall implements checkers.CallChecker.
func (c *Checker) CheckCall(cctx *checkers.CheckContext, call *ast.CallExpr) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	// Check for .Go() or .TryGo() method
	methodName := sel.Sel.Name
	if methodName != "Go" && methodName != "TryGo" {
		return
	}

	// Check if receiver is errgroup.Group
	if !checkers.IsNamedType(cctx.Pass, sel.X, pkgPath, "Group") {
		return
	}

	// errgroup.Group.Go() takes a func() error
	if len(call.Args) != 1 {
		return
	}

	if !cctx.CheckFuncArgUsesContext(call.Args[0]) {
		cctx.Reportf(call.Pos(), "errgroup.Group.%s() closure should use context %q", methodName, cctx.Scope.Name)
	}
}
