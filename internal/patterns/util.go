package patterns

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/mpyw/goroutinectx/internal/typeutil"
)

// funcLitHasContextParam checks if a function literal has a context.Context parameter.
func funcLitHasContextParam(cctx *CheckContext, lit *ast.FuncLit) bool {
	if lit.Type == nil || lit.Type.Params == nil {
		return false
	}
	for _, field := range lit.Type.Params.List {
		typ := cctx.Pass.TypesInfo.TypeOf(field.Type)
		if typ == nil {
			continue
		}
		if isContextType(typ) {
			return true
		}
	}
	return false
}

// funcLitUsesContext checks if a function literal references any context variable.
// It does NOT descend into nested func literals - they have their own scope and
// will be checked separately.
func funcLitUsesContext(cctx *CheckContext, lit *ast.FuncLit) bool {
	usesCtx := false
	ast.Inspect(lit.Body, func(n ast.Node) bool {
		if usesCtx {
			return false
		}
		// Skip nested function literals - they will be checked separately
		if nested, ok := n.(*ast.FuncLit); ok && nested != lit {
			return false
		}
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		obj := cctx.Pass.TypesInfo.ObjectOf(ident)
		if obj == nil {
			return true
		}
		if typeutil.IsContextOrCarrierType(obj.Type(), cctx.Carriers) {
			usesCtx = true
			return false
		}
		return true
	})
	return usesCtx
}

// extractCallFunc extracts the types.Func from a call expression.
func extractCallFunc(pass *analysis.Pass, call *ast.CallExpr) *types.Func {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		if f, ok := pass.TypesInfo.ObjectOf(fun).(*types.Func); ok {
			return f
		}

	case *ast.SelectorExpr:
		sel := pass.TypesInfo.Selections[fun]
		if sel != nil {
			if f, ok := sel.Obj().(*types.Func); ok {
				return f
			}
		} else {
			if f, ok := pass.TypesInfo.ObjectOf(fun.Sel).(*types.Func); ok {
				return f
			}
		}
	}

	return nil
}

// argUsesContext checks if an expression references a context variable.
func argUsesContext(cctx *CheckContext, expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found {
			return false
		}
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		obj := cctx.Pass.TypesInfo.ObjectOf(ident)
		if obj == nil {
			return true
		}
		if typeutil.IsContextOrCarrierType(obj.Type(), cctx.Carriers) {
			found = true
			return false
		}
		return true
	})
	return found
}

// findFuncLitAssignment searches for the func literal assigned to the variable.
// Returns the LAST assignment found (for cases where variables are reassigned).
func findFuncLitAssignment(cctx *CheckContext, v *types.Var) *ast.FuncLit {
	var result *ast.FuncLit
	declPos := v.Pos()

	for _, f := range cctx.Pass.Files {
		if f.Pos() > declPos || declPos >= f.End() {
			continue
		}

		ast.Inspect(f, func(n ast.Node) bool {
			assign, ok := n.(*ast.AssignStmt)
			if !ok {
				return true
			}
			if fl := findFuncLitInAssignment(cctx, assign, v); fl != nil {
				result = fl // Keep updating - we want the LAST assignment
			}
			return true
		})
		break
	}

	return result
}

// findFuncLitInAssignment checks if the assignment assigns a func literal to v.
func findFuncLitInAssignment(cctx *CheckContext, assign *ast.AssignStmt, v *types.Var) *ast.FuncLit {
	for i, lhs := range assign.Lhs {
		ident, ok := lhs.(*ast.Ident)
		if !ok {
			continue
		}
		if cctx.Pass.TypesInfo.ObjectOf(ident) != v {
			continue
		}
		if i >= len(assign.Rhs) {
			continue
		}
		if fl, ok := assign.Rhs[i].(*ast.FuncLit); ok {
			return fl
		}
	}
	return nil
}

// factoryReturnsContextUsingFunc checks if a factory function's return statements
// return functions that use context. Recursively checks nested func literals.
func factoryReturnsContextUsingFunc(cctx *CheckContext, factory *ast.FuncLit) bool {
	usesContext := false

	ast.Inspect(factory.Body, func(n ast.Node) bool {
		if usesContext {
			return false
		}
		// For nested func literals, check both direct usage and returned values
		if fl, ok := n.(*ast.FuncLit); ok && fl != factory {
			// Check if this nested func lit uses context directly
			if funcLitUsesContext(cctx, fl) {
				usesContext = true
				return false
			}
			// Recursively check if it returns functions that use context
			if factoryReturnsContextUsingFunc(cctx, fl) {
				usesContext = true
				return false
			}
			return false // Don't descend into nested func literals
		}

		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}

		for _, result := range ret.Results {
			if returnedValueUsesContext(cctx, result) {
				usesContext = true
				return false
			}
		}
		return true
	})

	return usesContext
}

// returnedValueUsesContext checks if a returned value is a func that uses context.
func returnedValueUsesContext(cctx *CheckContext, result ast.Expr) bool {
	// If it's a func literal, check directly
	if innerFuncLit, ok := result.(*ast.FuncLit); ok {
		return funcLitUsesContext(cctx, innerFuncLit)
	}

	// If it's an identifier (variable), find its assignment
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

	innerFuncLit := findFuncLitAssignment(cctx, v)
	if innerFuncLit == nil {
		return false
	}

	return funcLitUsesContext(cctx, innerFuncLit)
}
