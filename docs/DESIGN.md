# Design Document

## Overview

ctxrelay is a static analysis tool for Go that enforces context propagation best practices.

## Problem Statement

In Go applications, `context.Context` is used for:
- Request cancellation
- Timeout propagation
- Trace/span correlation (APM, distributed tracing)
- Request-scoped values

However, developers often forget to pass context to downstream calls, breaking the propagation chain. This leads to:
- Incomplete traces in APM tools
- Uncancellable operations
- Lost request-scoped data

## Goals

1. **Detect missing context propagation** in common patterns
2. **Integration** with existing Go tooling (`go vet`, `golangci-lint`)
3. **Zero false positives** - prefer missing issues over false alarms
4. **Type-safe analysis** - use `go/types` for accurate detection

## Non-Goals

1. Auto-fixing (may be added later)
2. Runtime checking
3. Configuration files (v1.0 will have hardcoded rules)
4. Custom rule definition (future version)

---

## Architecture

### Directory Structure

```
pkg/analyzer/
├── analyzer.go              # Main entry point (orchestration)
├── analyzer_test.go         # Integration tests
└── checkers/
    ├── checker.go           # Checker interfaces & ContextScope
    ├── typeutil.go          # Type utilities (shared)
    ├── ignore.go            # Ignore directive handling
    ├── errgroupchecker/     # errgroup.Group.Go() checker
    │   └── checker.go
    ├── goroutinechecker/    # go statement checker
    │   └── checker.go
    ├── goroutinederivechecker/  # goroutine derive function checker
    │   └── checker.go
    ├── slogchecker/         # slog logging checker
    │   └── checker.go
    ├── waitgroupchecker/    # sync.WaitGroup.Go() checker
    │   └── checker.go
    └── zerologchecker/      # zerolog SSA-based checker
        ├── checker.go       # Entry point
        ├── tracer.go        # Strategy pattern tracers
        ├── trace.go         # SSA value tracing
        └── types.go         # Type checking
```

### Checker Interface

```go
// checker.go
package checkers

// Checker defines the interface for context propagation checks.
type Checker interface {
    Name() string
    CheckCall(pass *analysis.Pass, call *ast.CallExpr, scope *ContextScope)
    CheckGoStmt(pass *analysis.Pass, stmt *ast.GoStmt, scope *ContextScope)
}
```

### ContextScope

```go
// ContextScope tracks context availability in a scope.
type ContextScope struct {
    Var  *types.Var  // The context variable (from go/types)
    Name string      // Variable name (for error messages)
}

// UsesContext checks if the given AST node uses the context variable.
func (s *ContextScope) UsesContext(node ast.Node) bool
```

### Type Utilities

```go
// typeutil.go
package checkers

// IsNamedType checks if expr has the given named type (handles pointers).
func IsNamedType(pass *analysis.Pass, expr ast.Expr, pkgPath, typeName string) bool

// IsContextType checks if the type is context.Context.
func IsContextType(t types.Type) bool

// unwrapPointer returns the element type if t is a pointer, otherwise t.
func unwrapPointer(t types.Type) types.Type
```

---

## Key Design Decisions

### 1. `inspector.WithStack` for Nested Function Support

Using `inspector.WithStack` instead of `ast.Inspect` allows proper tracking of context through nested functions and closures:

```go
insp.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
    // Find nearest enclosing function with context
    for i := len(stack) - 1; i >= 0; i-- {
        if scope, ok := funcScopes[stack[i]]; ok {
            // Run checkers with this scope
            break
        }
    }
})
```

This correctly handles:
- Nested functions at any depth
- Closures capturing context
- Shadowed context parameters
- Context introduced in middle layers

### 2. Type-Safe Analysis

All checkers use `go/types` for accurate detection instead of name-based string matching:

```go
// Good: Type-safe checking
func IsNamedType(pass *analysis.Pass, expr ast.Expr, pkgPath, typeName string) bool {
    tv, ok := pass.TypesInfo.Types[expr]
    // ... check against actual type info
}

// Avoided: Name-based checking (error-prone)
// if sel.Sel.Name == "Info" { ... }
```

### 3. Checker Interface Pattern

Each API (zerolog, slog, etc.) has its own checker implementation in a dedicated package. This provides:
- Clear separation of concerns
- Easy addition of new checkers
- Testability

### 4. Package Structure and Method Design

**In split packages (e.g., `errgroupchecker`, `slogchecker`):**
- Methods MUST have meaningful receivers that are actually used
- Receiver-less struct methods are NOT allowed
- If a function doesn't need state, make it a package-level function
- This keeps the API honest - struct methods imply state dependency

**In non-split code (e.g., shared utilities in `checkers/`):**
- Receiver-less struct methods ARE allowed when needed to avoid name collisions
- Use judgment based on the specific situation

**Rationale:**
When code is properly separated into packages, naming conflicts are naturally resolved by the package namespace. The package name itself provides context, so internal names can be simpler (e.g., `checker` instead of `errgroupChecker`). Struct methods without receivers in split packages would be a code smell - they indicate either:
1. The function should be package-level, or
2. The struct should hold some state

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2024-12-15 | No config files for v1.0 | Keep it simple. Add later if needed |
| 2024-12-15 | Checker interface | Extensibility and testability |
| 2024-12-15 | Type info over names | Name-based matching is error-prone |
| 2024-12-15 | Exclude zap | Low priority. Add if needed |
| 2024-12-15 | Use `inspector.WithStack` | Accurate tracking of nested functions |

## References

- [go/analysis package](https://pkg.go.dev/golang.org/x/tools/go/analysis)
- [Writing Go Analysis Tools](https://arslan.io/2019/06/13/using-go-analysis-to-write-a-custom-linter/)
- [golangci-lint custom linters](https://golangci-lint.run/contributing/new-linters/)
