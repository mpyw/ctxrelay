// Package waitgroupchecker checks sync.WaitGroup.Go() calls for context propagation.
// Note: sync.WaitGroup.Go() was added in Go 1.25.
package waitgroupchecker

import (
	"go/ast"

	"github.com/mpyw/ctxrelay/pkg/analyzer/checkers"
)

const pkgPath = "sync"

// Checker checks sync.WaitGroup.Go() calls for context propagation.
type Checker struct{}

// New creates a new waitgroup checker.
func New() *Checker {
	return &Checker{}
}

// CheckCall implements checkers.CallChecker.
func (c *Checker) CheckCall(cctx *checkers.CheckContext, call *ast.CallExpr) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	// Check for .Go() method (Go 1.25+)
	if sel.Sel.Name != "Go" {
		return
	}

	// Check if receiver is sync.WaitGroup
	if !checkers.IsNamedType(cctx.Pass, sel.X, pkgPath, "WaitGroup") {
		return
	}

	// sync.WaitGroup.Go() takes a func()
	if len(call.Args) != 1 {
		return
	}

	if !cctx.CheckFuncArgUsesContext(call.Args[0]) {
		cctx.Reportf(call.Pos(), "sync.WaitGroup.Go() closure should use context %q", cctx.Scope.Name)
	}
}
