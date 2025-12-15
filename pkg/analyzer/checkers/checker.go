// Package checkers provides context propagation checkers for various Go APIs.
package checkers

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ContextScope tracks context availability in a function scope.
// It tracks all context parameters, not just the first one.
type ContextScope struct {
	// Vars contains all context variables (from go/types).
	// Multiple contexts can be present (e.g., func(ctx1, ctx2 context.Context)).
	Vars []*types.Var

	// Name is the first variable name (for error messages).
	// Using the first name provides consistent, predictable error messages.
	Name string
}

// FindContextScope finds all context parameters in a function and creates a ContextScope.
// Returns nil if no context parameter is found.
// If carriers is non-nil, it also considers those types as context carriers.
func FindContextScope(pass *analysis.Pass, fnType *ast.FuncType, carriers []ContextCarrier) *ContextScope {
	if fnType == nil || fnType.Params == nil {
		return nil
	}

	var vars []*types.Var
	var firstName string

	for _, field := range fnType.Params.List {
		tv, ok := pass.TypesInfo.Types[field.Type]
		if !ok {
			continue
		}

		if !IsContextOrCarrierType(tv.Type, carriers) {
			continue
		}

		// Found context parameter(s) - collect all names in this field
		for _, name := range field.Names {
			obj := pass.TypesInfo.ObjectOf(name)
			if v, ok := obj.(*types.Var); ok {
				vars = append(vars, v)
				if firstName == "" {
					firstName = name.Name
				}
			}
		}
	}

	if len(vars) == 0 {
		return nil
	}

	return &ContextScope{
		Vars: vars,
		Name: firstName,
	}
}

// IsContextVar checks if obj matches any of the tracked context variables.
func (s *ContextScope) IsContextVar(obj types.Object) bool {
	for _, v := range s.Vars {
		if obj == v {
			return true
		}
	}
	return false
}

// UsesContext checks if the given AST node uses any of the context variables.
// It uses type information to correctly handle shadowing.
// It does NOT descend into nested function literals (closures) - each closure
// should be checked separately with its own scope.
// Returns true if ANY of the tracked context variables is used.
func (s *ContextScope) UsesContext(pass *analysis.Pass, node ast.Node) bool {
	if s == nil || len(s.Vars) == 0 {
		return false
	}

	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		if found {
			return false
		}
		// Skip nested function literals - they will be checked separately
		if _, ok := n.(*ast.FuncLit); ok && n != node {
			return false
		}
		if ident, ok := n.(*ast.Ident); ok {
			if s.IsContextVar(pass.TypesInfo.ObjectOf(ident)) {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

// CheckContext holds the context for running checks.
type CheckContext struct {
	Pass      *analysis.Pass
	Scope     *ContextScope
	IgnoreMap IgnoreMap
	Carriers  []ContextCarrier
}

// Reportf reports a diagnostic if the position is not ignored.
func (c *CheckContext) Reportf(pos token.Pos, format string, args ...any) {
	line := c.Pass.Fset.Position(pos).Line
	if c.IgnoreMap.ShouldIgnore(line) {
		return
	}
	c.Pass.Reportf(pos, format, args...)
}

// CallChecker checks call expressions for context propagation issues.
type CallChecker interface {
	CheckCall(cctx *CheckContext, call *ast.CallExpr)
}

// GoStmtChecker checks go statements for context propagation issues.
type GoStmtChecker interface {
	CheckGoStmt(cctx *CheckContext, stmt *ast.GoStmt)
}

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

	funcLit := FindFuncLitAssignment(c, v)
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

	funcLit := FindFuncLitAssignment(c, v)
	if funcLit == nil {
		return false
	}

	// Check if the returned func uses context
	return c.funcLitReturnUsesContext(funcLit)
}

// FindFuncLitAssignment searches for the last func literal assigned to the variable
// before the variable's declaration position (for simple cases).
// For more accurate tracking with reassignments, use FindFuncLitAssignmentBefore.
func FindFuncLitAssignment(cctx *CheckContext, v *types.Var) *ast.FuncLit {
	return FindFuncLitAssignmentBefore(cctx, v, token.NoPos)
}

// FindFuncLitAssignmentBefore searches for the last func literal assigned to the variable
// before the given position. If beforePos is token.NoPos, it finds assignments after the
// variable declaration. This handles variable reassignment correctly.
func FindFuncLitAssignmentBefore(cctx *CheckContext, v *types.Var, beforePos token.Pos) *ast.FuncLit {
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
			// Skip assignments after the usage point
			if beforePos != token.NoPos && assign.Pos() >= beforePos {
				return true
			}
			// Check if this assignment is to our variable
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

// funcLitReturnUsesContext checks if any return statement in the func literal
// returns a func literal that uses context.
func (c *CheckContext) funcLitReturnUsesContext(funcLit *ast.FuncLit) bool {
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
		if slices.ContainsFunc(ret.Results, c.returnedValueUsesContext) {
			usesContext = true
			return false
		}
		return true
	})

	return usesContext
}

// returnedValueUsesContext checks if a returned value is a func that uses context.
func (c *CheckContext) returnedValueUsesContext(result ast.Expr) bool {
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

	innerFuncLit := FindFuncLitAssignment(c, v)
	if innerFuncLit == nil {
		return false
	}

	return c.CheckClosureUsesContext(innerFuncLit)
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

// findStructFieldFuncLit finds a func literal assigned to a struct field.
func (c *CheckContext) findStructFieldFuncLit(v *types.Var, fieldName string) *ast.FuncLit {
	var result *ast.FuncLit
	pos := v.Pos()

	for _, f := range c.Pass.Files {
		if f.Pos() > pos || pos >= f.End() {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if result != nil {
				return false
			}
			assign, ok := n.(*ast.AssignStmt)
			if !ok {
				return true
			}
			result = c.findFieldInAssignment(assign, v, fieldName)
			return result == nil
		})
		break
	}

	return result
}

// findFieldInAssignment looks for a func literal in a struct field assignment.
func (c *CheckContext) findFieldInAssignment(assign *ast.AssignStmt, v *types.Var, fieldName string) *ast.FuncLit {
	for i, lhs := range assign.Lhs {
		ident, ok := lhs.(*ast.Ident)
		if !ok {
			continue
		}
		if c.Pass.TypesInfo.ObjectOf(ident) != v {
			continue
		}
		if i >= len(assign.Rhs) {
			continue
		}
		// Check if RHS is a composite literal (struct)
		compLit, ok := assign.Rhs[i].(*ast.CompositeLit)
		if !ok {
			continue
		}
		// Find the field in the struct literal
		for _, elt := range compLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != fieldName {
				continue
			}
			if fl, ok := kv.Value.(*ast.FuncLit); ok {
				return fl
			}
		}
	}
	return nil
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

// findIndexedFuncLit finds a func literal at a specific index in a composite literal.
func (c *CheckContext) findIndexedFuncLit(v *types.Var, indexExpr ast.Expr) *ast.FuncLit {
	var result *ast.FuncLit
	pos := v.Pos()

	for _, f := range c.Pass.Files {
		if f.Pos() > pos || pos >= f.End() {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if result != nil {
				return false
			}
			assign, ok := n.(*ast.AssignStmt)
			if !ok {
				return true
			}
			result = c.findFuncLitAtIndex(assign, v, indexExpr)
			return result == nil
		})
		break
	}

	return result
}

// findFuncLitAtIndex looks for a func literal at a specific index in a composite literal.
func (c *CheckContext) findFuncLitAtIndex(assign *ast.AssignStmt, v *types.Var, indexExpr ast.Expr) *ast.FuncLit {
	for i, lhs := range assign.Lhs {
		ident, ok := lhs.(*ast.Ident)
		if !ok {
			continue
		}
		if c.Pass.TypesInfo.ObjectOf(ident) != v {
			continue
		}
		if i >= len(assign.Rhs) {
			continue
		}
		// Check if RHS is a composite literal (slice/map)
		compLit, ok := assign.Rhs[i].(*ast.CompositeLit)
		if !ok {
			continue
		}

		// Handle based on index type
		if lit, ok := indexExpr.(*ast.BasicLit); ok {
			return c.findFuncLitByLiteral(compLit, lit)
		}
	}
	return nil
}

// findFuncLitByLiteral finds func literal by literal index/key.
func (c *CheckContext) findFuncLitByLiteral(compLit *ast.CompositeLit, lit *ast.BasicLit) *ast.FuncLit {
	switch lit.Kind {
	case token.INT:
		// Slice/array index
		index := 0
		if _, err := fmt.Sscanf(lit.Value, "%d", &index); err != nil {
			return nil
		}
		if index < 0 || index >= len(compLit.Elts) {
			return nil
		}
		if fl, ok := compLit.Elts[index].(*ast.FuncLit); ok {
			return fl
		}

	case token.STRING:
		// Map key - strip quotes
		key := strings.Trim(lit.Value, `"`)
		for _, elt := range compLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			keyLit, ok := kv.Key.(*ast.BasicLit)
			if !ok {
				continue
			}
			if strings.Trim(keyLit.Value, `"`) == key {
				if fl, ok := kv.Value.(*ast.FuncLit); ok {
					return fl
				}
			}
		}

	default:
		// Other token kinds (FLOAT, IMAG, CHAR, etc.) are not valid indices
		return nil
	}
	return nil
}
