// Package slogchecker checks slog logging calls for context propagation.
package slogchecker

import (
	"go/ast"
	"go/types"

	"github.com/mpyw/ctxrelay/pkg/analyzer/checkers"
)

const pkgPath = "log/slog"

// nonContextFuncs are slog functions that don't accept context.
var nonContextFuncs = map[string]bool{
	"Info":  true,
	"Debug": true,
	"Warn":  true,
	"Error": true,
	"Log":   true,
}

// Checker checks slog logging calls for context propagation.
type Checker struct{}

// New creates a new slog checker.
func New() *Checker {
	return &Checker{}
}

// CheckCall implements checkers.CallChecker.
func (c *Checker) CheckCall(cctx *checkers.CheckContext, call *ast.CallExpr) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	funcName := sel.Sel.Name
	if !nonContextFuncs[funcName] {
		return
	}

	if !isSlogCall(cctx, sel) {
		return
	}

	contextFunc := funcName + "Context"
	cctx.Reportf(call.Pos(), "use slog.%s with context %q instead of slog.%s", contextFunc, cctx.Scope.Name, funcName)
}

// isSlogCall checks if the selector is a slog package function or method.
func isSlogCall(cctx *checkers.CheckContext, sel *ast.SelectorExpr) bool {
	// Check for slog.Info(), slog.Debug(), etc. (package-level function)
	if ident, ok := sel.X.(*ast.Ident); ok {
		obj := cctx.Pass.TypesInfo.ObjectOf(ident)
		if pkgName, ok := obj.(*types.PkgName); ok {
			return pkgName.Imported().Path() == pkgPath
		}
	}

	// Check for *slog.Logger method call
	return checkers.IsNamedType(cctx.Pass, sel.X, pkgPath, "Logger")
}
