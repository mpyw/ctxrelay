// Package checkers provides context propagation checkers for various Go APIs.
package checkers

import (
	"go/ast"

	"github.com/mpyw/goroutinectx/internal/context"
)

// CallChecker checks call expressions for context propagation issues.
type CallChecker interface {
	CheckCall(cctx *context.CheckContext, call *ast.CallExpr)
}

// GoStmtChecker checks go statements for context propagation issues.
type GoStmtChecker interface {
	CheckGoStmt(cctx *context.CheckContext, stmt *ast.GoStmt)
}
