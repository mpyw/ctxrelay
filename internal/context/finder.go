package context

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"
)

// FindFuncLitAssignment searches for the last func literal assigned to the variable
// before the variable's declaration position (for simple cases).
// For more accurate tracking with reassignments, use FindFuncLitAssignmentBefore.
func (c *CheckContext) FindFuncLitAssignment(v *types.Var) *ast.FuncLit {
	return c.FindFuncLitAssignmentBefore(v, token.NoPos)
}

// FindFuncLitAssignmentBefore searches for the last func literal assigned to the variable
// before the given position. If beforePos is token.NoPos, it finds assignments after the
// variable declaration. This handles variable reassignment correctly.
func (c *CheckContext) FindFuncLitAssignmentBefore(v *types.Var, beforePos token.Pos) *ast.FuncLit {
	var result *ast.FuncLit

	declPos := v.Pos()

	for _, f := range c.Pass.Files {
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
			if fl := c.findFuncLitInAssignment(assign, v); fl != nil {
				result = fl // Keep updating - we want the LAST assignment
			}

			return true
		})

		break
	}

	return result
}

// findFuncLitInAssignment checks if the assignment assigns a func literal to v.
func (c *CheckContext) findFuncLitInAssignment(assign *ast.AssignStmt, v *types.Var) *ast.FuncLit {
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

		if fl, ok := assign.Rhs[i].(*ast.FuncLit); ok {
			return fl
		}
	}

	return nil
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
