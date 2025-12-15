// Package analyzer provides a go/analysis based analyzer for detecting
// missing context propagation in Go code.
package analyzer

import (
	"flag"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/mpyw/ctxrelay/pkg/analyzer/checkers"
	"github.com/mpyw/ctxrelay/pkg/analyzer/checkers/errgroupchecker"
	"github.com/mpyw/ctxrelay/pkg/analyzer/checkers/goroutinechecker"
	"github.com/mpyw/ctxrelay/pkg/analyzer/checkers/goroutinecreatorchecker"
	"github.com/mpyw/ctxrelay/pkg/analyzer/checkers/goroutinederivechecker"
	"github.com/mpyw/ctxrelay/pkg/analyzer/checkers/gotaskchecker"
	"github.com/mpyw/ctxrelay/pkg/analyzer/checkers/slogchecker"
	"github.com/mpyw/ctxrelay/pkg/analyzer/checkers/waitgroupchecker"
	"github.com/mpyw/ctxrelay/pkg/analyzer/checkers/zerologchecker"
)

// Flags for the analyzer.
var (
	goroutineDeriver string
	contextCarriers  string

	// Checker enable/disable flags (all enabled by default).
	enableSlog             bool
	enableErrgroup         bool
	enableWaitgroup        bool
	enableGoroutine        bool
	enableGoroutineCreator bool
	enableGotask           bool
	enableZerolog          bool
)

func init() {
	Analyzer.Flags.StringVar(&goroutineDeriver, "goroutine-deriver", "",
		"require goroutines to call this function to derive context (e.g., pkg.Func or pkg.Type.Method)")
	Analyzer.Flags.StringVar(&contextCarriers, "context-carriers", "",
		"comma-separated list of types to treat as context carriers (e.g., github.com/labstack/echo/v4.Context)")

	// Checker flags (default: all enabled)
	Analyzer.Flags.BoolVar(&enableSlog, "slog", true, "enable slog checker")
	Analyzer.Flags.BoolVar(&enableErrgroup, "errgroup", true, "enable errgroup checker")
	Analyzer.Flags.BoolVar(&enableWaitgroup, "waitgroup", true, "enable waitgroup checker")
	Analyzer.Flags.BoolVar(&enableGoroutine, "goroutine", true, "enable goroutine checker")
	Analyzer.Flags.BoolVar(&enableGoroutineCreator, "goroutine-creator", true, "enable goroutine-creator checker")
	Analyzer.Flags.BoolVar(&enableGotask, "gotask", true, "enable gotask checker (requires -goroutine-deriver)")
	Analyzer.Flags.BoolVar(&enableZerolog, "zerolog", true, "enable zerolog checker")
}

// Analyzer is the main analyzer for ctxrelay.
var Analyzer = &analysis.Analyzer{
	Name:     "ctxrelay",
	Doc:      "checks that context.Context is properly propagated to downstream calls",
	Requires: []*analysis.Analyzer{inspect.Analyzer, buildssa.Analyzer},
	Run:      run,
	Flags:    flag.FlagSet{},
}

func run(pass *analysis.Pass) (any, error) {
	insp, _ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	ssaInfo, _ := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)

	// Parse configuration
	carriers := checkers.ParseContextCarriers(contextCarriers)

	// Build ignore maps for each file
	ignoreMaps := buildIgnoreMaps(pass)

	// Build goroutine creator map from //ctxrelay:goroutine_creator directives
	goroutineCreators := checkers.BuildGoroutineCreatorMap(pass)

	// Run AST-based checks (slog, goroutine, errgroup, waitgroup)
	runASTChecks(pass, insp, ignoreMaps, carriers, goroutineCreators)

	// Run SSA-based zerolog analysis for variable tracking
	if enableZerolog {
		// Convert checkers.IgnoreMap to zerologchecker.IgnoreMap interface
		zerologIgnoreMaps := make(map[string]zerologchecker.IgnoreMap, len(ignoreMaps))
		for k, v := range ignoreMaps {
			zerologIgnoreMaps[k] = v
		}
		zerologchecker.RunSSA(pass, ssaInfo, zerologIgnoreMaps, checkers.IsContextType)
	}

	return nil, nil
}

// buildIgnoreMaps creates ignore maps for each file in the pass.
func buildIgnoreMaps(pass *analysis.Pass) map[string]checkers.IgnoreMap {
	ignoreMaps := make(map[string]checkers.IgnoreMap)
	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		ignoreMaps[filename] = checkers.BuildIgnoreMap(pass.Fset, file)
	}
	return ignoreMaps
}

// runASTChecks runs AST-based checkers on the pass.
func runASTChecks(
	pass *analysis.Pass,
	insp *inspector.Inspector,
	ignoreMaps map[string]checkers.IgnoreMap,
	carriers []checkers.ContextCarrier,
	goroutineCreators checkers.GoroutineCreatorMap,
) {
	// Build context scopes for functions with context parameters
	funcScopes := buildFuncScopes(pass, insp, carriers)

	// Build checkers based on flags
	var callCheckers []checkers.CallChecker
	var goStmtCheckers []checkers.GoStmtChecker

	if enableSlog {
		callCheckers = append(callCheckers, slogchecker.New())
	}
	if enableErrgroup {
		callCheckers = append(callCheckers, errgroupchecker.New())
	}
	if enableWaitgroup {
		callCheckers = append(callCheckers, waitgroupchecker.New())
	}

	// Add goroutine creator checker if enabled and any functions are marked
	if enableGoroutineCreator && len(goroutineCreators) > 0 {
		callCheckers = append(callCheckers, goroutinecreatorchecker.New(goroutineCreators))
	}

	// When goroutine-deriver is set, it replaces the base goroutine checker.
	// The derive checker is a more specific version that checks for deriver function calls.
	if goroutineDeriver != "" {
		goStmtCheckers = append(goStmtCheckers, goroutinederivechecker.New(goroutineDeriver))
		// gotask checker also requires goroutine-deriver to be set
		if enableGotask {
			callCheckers = append(callCheckers, gotaskchecker.New(goroutineDeriver))
		}
	} else if enableGoroutine {
		goStmtCheckers = append(goStmtCheckers, goroutinechecker.New())
	}

	// Node types we're interested in
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
		(*ast.GoStmt)(nil),
		(*ast.CallExpr)(nil),
	}

	// Check nodes within context-aware functions
	insp.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return true
		}

		scope := findEnclosingScope(funcScopes, stack)
		if scope == nil {
			return true // No context in scope
		}

		filename := pass.Fset.Position(n.Pos()).Filename
		cctx := &checkers.CheckContext{
			Pass:      pass,
			Scope:     scope,
			IgnoreMap: ignoreMaps[filename],
			Carriers:  carriers,
		}

		switch node := n.(type) {
		case *ast.GoStmt:
			for _, checker := range goStmtCheckers {
				checker.CheckGoStmt(cctx, node)
			}
		case *ast.CallExpr:
			for _, checker := range callCheckers {
				checker.CheckCall(cctx, node)
			}
		}

		return true
	})
}

// buildFuncScopes identifies functions with context parameters.
func buildFuncScopes(
	pass *analysis.Pass,
	insp *inspector.Inspector,
	carriers []checkers.ContextCarrier,
) map[ast.Node]*checkers.ContextScope {
	funcScopes := make(map[ast.Node]*checkers.ContextScope)
	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil), (*ast.FuncLit)(nil)}, func(n ast.Node) {
		var fnType *ast.FuncType
		switch fn := n.(type) {
		case *ast.FuncDecl:
			fnType = fn.Type
		case *ast.FuncLit:
			fnType = fn.Type
		}

		if scope := checkers.FindContextScope(pass, fnType, carriers); scope != nil {
			funcScopes[n] = scope
		}
	})
	return funcScopes
}

// findEnclosingScope finds the closest enclosing function with a context parameter.
func findEnclosingScope(funcScopes map[ast.Node]*checkers.ContextScope, stack []ast.Node) *checkers.ContextScope {
	for i := len(stack) - 1; i >= 0; i-- {
		if scope, ok := funcScopes[stack[i]]; ok {
			return scope
		}
	}
	return nil
}
