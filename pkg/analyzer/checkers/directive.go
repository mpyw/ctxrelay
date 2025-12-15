package checkers

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// GoroutineCreatorMap tracks functions marked with //ctxrelay:goroutine_creator.
// These functions are expected to spawn goroutines with their func arguments.
type GoroutineCreatorMap map[*types.Func]struct{}

// BuildGoroutineCreatorMap scans files for functions marked with the directive.
func BuildGoroutineCreatorMap(pass *analysis.Pass) GoroutineCreatorMap {
	m := make(GoroutineCreatorMap)

	for _, file := range pass.Files {
		buildGoroutineCreatorsForFile(pass, file, m)
	}

	return m
}

// buildGoroutineCreatorsForFile scans a single file for goroutine creator directives.
func buildGoroutineCreatorsForFile(pass *analysis.Pass, file *ast.File, m GoroutineCreatorMap) {
	// Build a map of line -> comment for quick lookup
	lineComments := make(map[int]string)
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if isGoroutineCreatorComment(c.Text) {
				line := pass.Fset.Position(c.Pos()).Line
				lineComments[line] = c.Text
			}
		}
	}

	// Find function declarations that have the directive on the previous line
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		funcLine := pass.Fset.Position(funcDecl.Pos()).Line

		// Check if directive is on previous line
		if _, hasDirective := lineComments[funcLine-1]; !hasDirective {
			continue
		}

		// Get the types.Func for this declaration
		obj := pass.TypesInfo.ObjectOf(funcDecl.Name)
		if obj == nil {
			continue
		}
		fn, ok := obj.(*types.Func)
		if !ok {
			continue
		}

		m[fn] = struct{}{}
	}
}

// isGoroutineCreatorComment checks if a comment is a goroutine_creator directive.
func isGoroutineCreatorComment(text string) bool {
	text = strings.TrimPrefix(text, "//")
	text = strings.TrimSpace(text)
	return strings.HasPrefix(text, "ctxrelay:goroutine_creator")
}

// IsGoroutineCreator checks if a function is marked as a goroutine creator.
func (m GoroutineCreatorMap) IsGoroutineCreator(fn *types.Func) bool {
	_, ok := m[fn]
	return ok
}

// GetFuncFromCall extracts the *types.Func from a call expression if possible.
// Returns nil if the callee cannot be determined statically.
func GetFuncFromCall(pass *analysis.Pass, call *ast.CallExpr) *types.Func {
	var ident *ast.Ident

	switch fun := call.Fun.(type) {
	case *ast.Ident:
		ident = fun
	case *ast.SelectorExpr:
		ident = fun.Sel
	default:
		return nil
	}

	obj := pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return nil
	}

	fn, ok := obj.(*types.Func)
	if !ok {
		return nil
	}

	return fn
}

// FindFuncArgs finds all arguments in a call that are func types.
// Returns the indices and the arguments themselves.
func FindFuncArgs(pass *analysis.Pass, call *ast.CallExpr) []ast.Expr {
	var funcArgs []ast.Expr

	for _, arg := range call.Args {
		tv, ok := pass.TypesInfo.Types[arg]
		if !ok {
			continue
		}

		// Check if argument is a function type
		if _, isFunc := tv.Type.Underlying().(*types.Signature); isFunc {
			funcArgs = append(funcArgs, arg)
		}
	}

	return funcArgs
}

// FuncArgPosition returns the position of a func argument for error messages.
func FuncArgPosition(pass *analysis.Pass, call *ast.CallExpr) token.Pos {
	// Return position of first func argument, or call position if none
	for _, arg := range call.Args {
		tv, ok := pass.TypesInfo.Types[arg]
		if !ok {
			continue
		}
		if _, isFunc := tv.Type.Underlying().(*types.Signature); isFunc {
			return arg.Pos()
		}
	}
	return call.Pos()
}
