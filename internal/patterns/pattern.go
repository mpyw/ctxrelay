// Package patterns defines pattern interfaces and types for goroutinectx.
package patterns

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ssa"

	internalssa "github.com/mpyw/goroutinectx/internal/ssa"
)

// CheckContext provides context for pattern checking.
type CheckContext struct {
	Pass    *analysis.Pass
	Tracer  *internalssa.Tracer
	SSAProg *internalssa.Program
}

// Report reports a diagnostic at the given position.
func (c *CheckContext) Report(pos token.Pos, msg string) {
	c.Pass.Reportf(pos, "%s", msg)
}

// findSSAValue finds the SSA value corresponding to an AST expression.
// Handles: FuncLit, Ident, CallExpr, SelectorExpr, IndexExpr
func (c *CheckContext) findSSAValue(fn *ssa.Function, expr ast.Expr) ssa.Value {
	if fn == nil || fn.Blocks == nil {
		return nil
	}

	switch e := expr.(type) {
	case *ast.FuncLit:
		return c.findFuncLitValue(fn, e)
	case *ast.Ident:
		return c.findIdentValue(fn, e)
	case *ast.CallExpr:
		return c.findCallValue(fn, e)
	case *ast.SelectorExpr:
		return c.findSelectorValue(fn, e)
	case *ast.IndexExpr:
		return c.findIndexValue(fn, e)
	default:
		return c.findValueByPos(fn, expr.Pos())
	}
}

// findFuncLitValue finds the MakeClosure instruction for a function literal.
func (c *CheckContext) findFuncLitValue(fn *ssa.Function, lit *ast.FuncLit) ssa.Value {
	pos := lit.Pos()
	for _, block := range fn.Blocks {
		for _, instr := range block.Instrs {
			if mc, ok := instr.(*ssa.MakeClosure); ok {
				if mc.Pos() == pos {
					return mc
				}
			}
		}
	}
	return nil
}

// findIdentValue finds the SSA value for an identifier (variable reference).
// For `g.Go(fn)` where `fn := func() {}`, finds the MakeClosure assigned to fn.
func (c *CheckContext) findIdentValue(fn *ssa.Function, ident *ast.Ident) ssa.Value {
	obj := c.Pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return nil
	}

	v, ok := obj.(*types.Var)
	if !ok {
		return nil
	}

	// Search for the value assigned to this variable
	// In SSA, we look for instructions that define a value at the variable's declaration position
	declPos := v.Pos()

	for _, block := range fn.Blocks {
		for _, instr := range block.Instrs {
			// Check MakeClosure at declaration position
			if mc, ok := instr.(*ssa.MakeClosure); ok {
				if mc.Pos() == declPos {
					return mc
				}
			}
			// Check Call result at declaration position
			if call, ok := instr.(*ssa.Call); ok {
				if call.Pos() == declPos {
					return call
				}
			}
			// Check any value-producing instruction at declaration position
			if val, ok := instr.(ssa.Value); ok {
				if instr.Pos() == declPos {
					return val
				}
			}
		}
	}

	return nil
}

// findCallValue finds the SSA Call instruction for a call expression.
func (c *CheckContext) findCallValue(fn *ssa.Function, call *ast.CallExpr) ssa.Value {
	pos := call.Pos()
	// Try Fun position for method calls
	funPos := call.Fun.Pos()

	for _, block := range fn.Blocks {
		for _, instr := range block.Instrs {
			if ssaCall, ok := instr.(*ssa.Call); ok {
				if ssaCall.Pos() == pos || ssaCall.Pos() == funPos {
					return ssaCall
				}
			}
		}
	}
	return nil
}

// findSelectorValue finds the SSA value for a selector expression (field access).
func (c *CheckContext) findSelectorValue(fn *ssa.Function, sel *ast.SelectorExpr) ssa.Value {
	pos := sel.Pos()
	selPos := sel.Sel.Pos()

	for _, block := range fn.Blocks {
		for _, instr := range block.Instrs {
			// Check FieldAddr
			if fa, ok := instr.(*ssa.FieldAddr); ok {
				if fa.Pos() == pos || fa.Pos() == selPos {
					return fa
				}
			}
			// Check Field (for value types)
			if f, ok := instr.(*ssa.Field); ok {
				if f.Pos() == pos || f.Pos() == selPos {
					return f
				}
			}
		}
	}
	return nil
}

// findIndexValue finds the SSA value for an index expression.
func (c *CheckContext) findIndexValue(fn *ssa.Function, idx *ast.IndexExpr) ssa.Value {
	pos := idx.Pos()

	for _, block := range fn.Blocks {
		for _, instr := range block.Instrs {
			// Check IndexAddr
			if ia, ok := instr.(*ssa.IndexAddr); ok {
				if ia.Pos() == pos {
					return ia
				}
			}
			// Check Index (for value types)
			if i, ok := instr.(*ssa.Index); ok {
				if i.Pos() == pos {
					return i
				}
			}
			// Check Lookup (for maps)
			if l, ok := instr.(*ssa.Lookup); ok {
				if l.Pos() == pos {
					return l
				}
			}
		}
	}
	return nil
}

// findValueByPos finds any SSA value at the given position (fallback).
func (c *CheckContext) findValueByPos(fn *ssa.Function, pos token.Pos) ssa.Value {
	for _, block := range fn.Blocks {
		for _, instr := range block.Instrs {
			if instr.Pos() == pos {
				if val, ok := instr.(ssa.Value); ok {
					return val
				}
			}
		}
	}

	// Check parameters
	for _, param := range fn.Params {
		if param.Pos() == pos {
			return param
		}
	}

	return nil
}

// isContextType checks if a type is context.Context.
func isContextType(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Pkg() != nil && obj.Pkg().Path() == "context" && obj.Name() == "Context"
}

// Pattern defines the interface for context propagation patterns.
type Pattern interface {
	// Name returns a human-readable name for the pattern.
	Name() string

	// Check checks if the pattern is satisfied for the given call.
	// Returns true if the pattern is satisfied (no error).
	Check(cctx *CheckContext, call *ast.CallExpr, callbackArg ast.Expr) bool

	// Message returns the diagnostic message when the pattern is violated.
	Message(apiName string, ctxName string) string
}
