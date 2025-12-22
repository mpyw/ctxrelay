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
func (c *CheckContext) findSSAValue(fn *ssa.Function, expr ast.Expr) ssa.Value {
	if fn == nil || fn.Blocks == nil {
		return nil
	}

	pos := expr.Pos()

	// Search through all instructions in the function
	for _, block := range fn.Blocks {
		for _, instr := range block.Instrs {
			// Check if this instruction's position matches
			if instr.Pos() == pos {
				if val, ok := instr.(ssa.Value); ok {
					return val
				}
			}

			// For MakeClosure, check the Fn position
			if mc, ok := instr.(*ssa.MakeClosure); ok {
				if mc.Pos() == pos {
					return mc
				}
			}
		}
	}

	// Also check parameters and free variables
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
