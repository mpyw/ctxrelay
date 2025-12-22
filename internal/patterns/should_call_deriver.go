package patterns

import (
	"fmt"
	"go/ast"

	"github.com/mpyw/goroutinectx/internal/directives/deriver"
)

// ShouldCallDeriver checks that a callback/goroutine body calls a deriver function.
// Used by: goroutine-derive, gotask task functions, etc.
type ShouldCallDeriver struct {
	// Matcher is the deriver function matcher (OR/AND semantics).
	Matcher *deriver.Matcher
}

func (*ShouldCallDeriver) Name() string {
	return "ShouldCallDeriver"
}

func (p *ShouldCallDeriver) Check(cctx *CheckContext, call *ast.CallExpr, callbackArg ast.Expr) bool {
	if p.Matcher == nil || p.Matcher.IsEmpty() {
		return true // No deriver configured, nothing to check
	}

	// For function literals, check the body directly
	if lit, ok := callbackArg.(*ast.FuncLit); ok {
		return p.Matcher.SatisfiesAnyGroup(cctx.Pass, lit.Body)
	}

	// For identifiers, try to find the function declaration
	if ident, ok := callbackArg.(*ast.Ident); ok {
		return p.checkIdentifier(cctx, ident)
	}

	// Can't analyze, assume OK
	return true
}

// checkIdentifier checks if the function referenced by identifier calls deriver.
func (p *ShouldCallDeriver) checkIdentifier(cctx *CheckContext, ident *ast.Ident) bool {
	obj := cctx.Pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return true
	}

	// Try to find the function literal or declaration
	// For local variables, we need SSA tracing which is done elsewhere
	return true // Can't trace without SSA, assume OK
}

func (p *ShouldCallDeriver) Message(apiName string, _ string) string {
	return fmt.Sprintf("%s callback should call %s", apiName, p.Matcher.Original)
}
