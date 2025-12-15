// Package goroutinechecker checks go statements for context propagation.
package goroutinechecker

import (
	"go/ast"
	"go/types"

	"github.com/mpyw/ctxrelay/pkg/analyzer/checkers"
)

// Checker checks go statements for context propagation.
type Checker struct{}

// New creates a new goroutine checker.
func New() *Checker {
	return &Checker{}
}

// CheckGoStmt implements checkers.GoStmtChecker.
func (c *Checker) CheckGoStmt(cctx *checkers.CheckContext, goStmt *ast.GoStmt) {
	call := goStmt.Call

	// Check if context is used in the goroutine call chain
	if !callUsesContext(cctx, call) {
		cctx.Reportf(goStmt.Pos(), "goroutine does not propagate context %q", cctx.Scope.Name)
	}
}

// callUsesContext checks if a call expression (or its chain) uses context.
// Handles patterns like:
//   - go func() { ... }()           -> check func literal body
//   - go fn()                        -> check arguments + returned func
//   - go fn()()                      -> check all levels recursively
//   - go fn(ctx)()                   -> ctx used in inner call
//   - go fn()(ctx)                   -> ctx used in outer call
func callUsesContext(cctx *checkers.CheckContext, call *ast.CallExpr) bool {
	// Check if any argument in this call uses context
	for _, arg := range call.Args {
		if cctx.Scope.UsesContext(cctx.Pass, arg) {
			return true
		}
	}

	// Check the function being called
	switch fun := call.Fun.(type) {
	case *ast.FuncLit:
		// go func() { ... }() - check the func literal body
		return cctx.CheckClosureUsesContext(fun)

	case *ast.CallExpr:
		// go fn()() - check both the inner call and what it returns
		// First check if inner call uses context in its arguments
		if callUsesContext(cctx, fun) {
			return true
		}
		// Then check if the returned function (if we can find it) uses context
		return returnedFuncUsesContext(cctx, fun)

	case *ast.Ident:
		// go fn() where fn is a variable holding a func
		// Check if the func literal assigned to fn uses context
		return identFuncUsesContext(cctx, fun)

	default:
		// go obj.Method() - already checked args above
		return false
	}
}

// identFuncUsesContext checks if a function stored in a variable uses context.
func identFuncUsesContext(cctx *checkers.CheckContext, ident *ast.Ident) bool {
	obj := cctx.Pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return false
	}

	v, ok := obj.(*types.Var)
	if !ok {
		return false
	}

	funcLit := checkers.FindFuncLitAssignment(cctx, v)
	if funcLit == nil {
		return false
	}

	return funcLitUsesContextOrReturnsCtxFunc(cctx, funcLit)
}

// returnedFuncUsesContext tries to find the function literal returned by a call
// and checks if it uses context.
//

func returnedFuncUsesContext(cctx *checkers.CheckContext, call *ast.CallExpr) bool {
	// First, check if ctx is passed as an argument to the call
	for _, arg := range call.Args {
		if cctx.Scope.UsesContext(cctx.Pass, arg) {
			return true
		}
	}

	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}

	obj := cctx.Pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return false
	}

	v, ok := obj.(*types.Var)
	if !ok {
		return false
	}

	funcLit := checkers.FindFuncLitAssignment(cctx, v)
	if funcLit == nil {
		return false
	}

	return funcLitReturnUsesContext(cctx, funcLit)
}

// funcLitReturnUsesContext checks if any return statement in the func literal
// returns a func literal that uses context.
func funcLitReturnUsesContext(cctx *checkers.CheckContext, funcLit *ast.FuncLit) bool {
	var usesContext bool

	ast.Inspect(funcLit.Body, func(n ast.Node) bool {
		if usesContext {
			return false
		}
		// Skip nested func literals (they have their own returns)
		if fl, ok := n.(*ast.FuncLit); ok && fl != funcLit {
			return false
		}
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		usesContext = returnStmtUsesContext(cctx, ret)
		return !usesContext
	})

	return usesContext
}

// returnStmtUsesContext checks if a return statement returns a func that uses context.
func returnStmtUsesContext(cctx *checkers.CheckContext, ret *ast.ReturnStmt) bool {
	for _, result := range ret.Results {
		if returnedValueUsesContext(cctx, result) {
			return true
		}
	}
	return false
}

// returnedValueUsesContext checks if a returned value is a func that uses context.
func returnedValueUsesContext(cctx *checkers.CheckContext, result ast.Expr) bool {
	// Check if returning a func literal directly
	if innerFuncLit, ok := result.(*ast.FuncLit); ok {
		return funcLitUsesContextOrReturnsCtxFunc(cctx, innerFuncLit)
	}

	// Check if returning an identifier that references a func using ctx
	ident, ok := result.(*ast.Ident)
	if !ok {
		return false
	}

	obj := cctx.Pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return false
	}

	v, ok := obj.(*types.Var)
	if !ok {
		return false
	}

	innerFuncLit := checkers.FindFuncLitAssignment(cctx, v)
	if innerFuncLit == nil {
		return false
	}

	return funcLitUsesContextOrReturnsCtxFunc(cctx, innerFuncLit)
}

// funcLitUsesContextOrReturnsCtxFunc checks if a func literal either:
// 1. Directly uses context in its body (including nested closures), OR
// 2. Returns another func that (recursively) uses context.
func funcLitUsesContextOrReturnsCtxFunc(cctx *checkers.CheckContext, funcLit *ast.FuncLit) bool {
	// Check if this func's body uses context (including nested closures)
	if usesContextDeep(cctx, funcLit.Body) {
		return true
	}

	// Check if this func returns another func that uses context
	return funcLitReturnUsesContext(cctx, funcLit)
}

// usesContextDeep checks if the given AST node uses any context variable,
// INCLUDING nested function literals.
func usesContextDeep(cctx *checkers.CheckContext, node ast.Node) bool {
	if cctx.Scope == nil || len(cctx.Scope.Vars) == 0 {
		return false
	}

	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		if found {
			return false
		}
		// DO NOT skip nested function literals - we want to trace ctx
		// through closures like: captured := ctx; return func() { use(captured) }
		if ident, ok := n.(*ast.Ident); ok {
			if cctx.Scope.IsContextVar(cctx.Pass.TypesInfo.ObjectOf(ident)) {
				found = true
				return false
			}
		}
		return true
	})
	return found
}
