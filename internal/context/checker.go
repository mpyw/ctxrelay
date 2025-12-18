package context

import (
	"go/ast"
	"go/types"
	"slices"
)

// CheckClosureUsesContext checks if a closure (func literal) uses the context from scope.
// Returns true if the closure properly uses context, false if it should be reported.
func (c *CheckContext) CheckClosureUsesContext(funcLit *ast.FuncLit) bool {
	// If the func literal has its own context parameter, it doesn't need to
	// capture the outer context - it will be checked with its own scope
	if HasContextOrCarrierParam(c.Pass, funcLit.Type, c.Carriers) {
		return true
	}

	// Check if context is captured in the closure
	return c.Scope.UsesContext(c.Pass, funcLit.Body)
}

// CheckFuncArgUsesContext checks if a function argument (to .Go() methods) uses context.
// Handles:
//   - func literal: g.Go(func() { ... })
//   - variable: g.Go(fn) where fn is assigned a func literal
//   - call result: g.Go(makeWorker()) where makeWorker returns a func
//   - struct field: g.Go(holder.task) where holder.task is a func literal
//   - slice index: g.Go(tasks[0]) where tasks[0] is a func literal
//   - map index: g.Go(tasks["key"]) where tasks["key"] is a func literal
//
// Returns true if context usage is found, false otherwise.
func (c *CheckContext) CheckFuncArgUsesContext(arg ast.Expr) bool {
	switch a := arg.(type) {
	case *ast.FuncLit:
		return c.CheckClosureUsesContext(a)

	case *ast.Ident:
		return c.checkIdentFuncUsesContext(a)

	case *ast.CallExpr:
		return c.checkCallResultFuncUsesContext(a)

	case *ast.SelectorExpr:
		return c.checkSelectorFuncUsesContext(a)

	case *ast.IndexExpr:
		return c.checkIndexFuncUsesContext(a)

	default:
		return false
	}
}

// FuncLitReturnUsesContext checks if any return statement in the func literal
// returns a func literal that uses context.
func (c *CheckContext) FuncLitReturnUsesContext(funcLit *ast.FuncLit) bool {
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

		if slices.ContainsFunc(ret.Results, c.ReturnedValueUsesContext) {
			usesContext = true

			return false
		}

		return true
	})

	return usesContext
}

// ReturnedValueUsesContext checks if a returned value is a func that uses context.
func (c *CheckContext) ReturnedValueUsesContext(result ast.Expr) bool {
	if innerFuncLit, ok := result.(*ast.FuncLit); ok {
		return c.CheckClosureUsesContext(innerFuncLit)
	}

	ident, ok := result.(*ast.Ident)
	if !ok {
		return false
	}

	obj := c.Pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return false
	}

	v, ok := obj.(*types.Var)
	if !ok {
		return false
	}

	innerFuncLit := c.FindFuncLitAssignment(v)
	if innerFuncLit == nil {
		return false
	}

	return c.CheckClosureUsesContext(innerFuncLit)
}

// checkIdentFuncUsesContext checks if a function stored in a variable uses context.
func (c *CheckContext) checkIdentFuncUsesContext(ident *ast.Ident) bool {
	obj := c.Pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return false
	}

	v, ok := obj.(*types.Var)
	if !ok {
		return false
	}

	funcLit := c.FindFuncLitAssignment(v)
	if funcLit == nil {
		return false
	}

	return c.CheckClosureUsesContext(funcLit)
}

// checkCallResultFuncUsesContext checks if a call returns a func that uses context.
// Handles patterns like:
//   - g.Go(makeWorker()) where makeWorker returns func() error - bad if no ctx
//   - g.Go(makeWorkerWithCtx(ctx)) where ctx is passed to factory - good
func (c *CheckContext) checkCallResultFuncUsesContext(call *ast.CallExpr) bool {
	// First, check if ctx is passed as an argument to the call.
	// If so, assume the returned func will use it (e.g., makeWorkerWithCtx(ctx)).
	for _, arg := range call.Args {
		if c.Scope.UsesContext(c.Pass, arg) {
			return true
		}
	}

	// Try to find the function being called and check if its return uses context
	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}

	obj := c.Pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return false
	}

	v, ok := obj.(*types.Var)
	if !ok {
		return false
	}

	funcLit := c.FindFuncLitAssignment(v)
	if funcLit == nil {
		return false
	}

	// Check if the returned func uses context
	return c.FuncLitReturnUsesContext(funcLit)
}

// checkSelectorFuncUsesContext checks if a struct field func uses context.
// Handles patterns like: g.Go(holder.task) where holder is a local struct.
func (c *CheckContext) checkSelectorFuncUsesContext(sel *ast.SelectorExpr) bool {
	// Get the receiver (e.g., "holder" in holder.task)
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	obj := c.Pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return false
	}

	v, ok := obj.(*types.Var)
	if !ok {
		return false
	}

	// Find the struct literal that initialized this variable
	fieldName := sel.Sel.Name

	funcLit := c.findStructFieldFuncLit(v, fieldName)
	if funcLit == nil {
		return false
	}

	return c.CheckClosureUsesContext(funcLit)
}

// checkIndexFuncUsesContext checks if a slice/map indexed func uses context.
// Handles patterns like: g.Go(tasks[0]) or g.Go(tasks["key"]).
func (c *CheckContext) checkIndexFuncUsesContext(idx *ast.IndexExpr) bool {
	// Get the array/slice/map variable (e.g., "tasks" in tasks[0])
	ident, ok := idx.X.(*ast.Ident)
	if !ok {
		return false
	}

	obj := c.Pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return false
	}

	v, ok := obj.(*types.Var)
	if !ok {
		return false
	}

	// Find the composite literal that initialized this variable
	funcLit := c.findIndexedFuncLit(v, idx.Index)
	if funcLit == nil {
		return false
	}

	return c.CheckClosureUsesContext(funcLit)
}
