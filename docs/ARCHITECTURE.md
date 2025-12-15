# ctxrelay Architecture

## Design Principles

Based on analysis of successful Go linters (staticcheck, errcheck, contextcheck, nilaway, bodyclose, ineffassign), ctxrelay follows these design principles:

### 1. Single-Purpose Focus

Like `errcheck` (error handling only) and `bodyclose` (res.Body.Close only), ctxrelay focuses exclusively on **context propagation**. We don't try to be a general-purpose linter.

### 2. go/analysis Framework

All modern Go linters use `golang.org/x/tools/go/analysis`. Benefits:
- Integration with `go vet`
- Integration with `golangci-lint`
- Standard fact mechanism for cross-package analysis
- Built-in testing via `analysistest`

### 3. Library + CLI Duality

Following `gomodifytags` and `errcheck`:
- `pkg/analyzer/` - Core analysis logic as importable package
- `cmd/ctxrelay/` - Thin CLI wrapper using `singlechecker`

This enables both standalone usage and programmatic integration.

## Directory Structure

```
ctxrelay/
├── cmd/
│   └── ctxrelay/                  # CLI entry point (singlechecker)
│       └── main.go
├── pkg/
│   └── analyzer/
│       ├── analyzer.go            # Main analyzer (orchestration)
│       ├── analyzer_test.go       # Integration tests
│       ├── checkers/              # Checker implementations
│       │   ├── checker.go         # CheckContext, ContextScope, interfaces
│       │   ├── typeutil.go        # Type checking utilities
│       │   ├── ignore.go          # IgnoreMap for comment directives
│       │   ├── slogchecker/       # slog checker
│       │   ├── errgroupchecker/   # errgroup checker
│       │   ├── waitgroupchecker/  # waitgroup checker
│       │   ├── goroutinechecker/  # goroutine checker
│       │   ├── goroutinederivechecker/ # goroutine_derive checker (flag-activated)
│       │   └── zerologchecker/    # SSA-based zerolog analysis (subpackage)
│       │       ├── checker.go     # Entry point
│       │       ├── tracer.go      # Strategy Pattern tracers
│       │       ├── trace.go       # Value tracing, Phi handling
│       │       └── types.go       # Type checking, constants
│       └── testdata/              # Test fixtures for analysistest
│           └── src/
│               ├── zerolog/       # basic.go, edge_*.go
│               ├── slog/
│               ├── goroutine/
│               ├── errgroup/
│               ├── waitgroup/
│               ├── nested/
│               └── carrier/       # context-carriers tests
├── bin/                           # Build output (gitignored)
├── docs/
│   ├── ARCHITECTURE.md            # Technical specification (this file)
│   └── TUTORIAL.md                # Beginner-friendly learning guide
├── .golangci.yaml                 # golangci-lint configuration
├── go.mod
├── go.sum
├── CLAUDE.md                      # AI assistant guidance
├── TASKS.md                       # Session notes (gitignored)
└── README.md
```

## Core Components

### analyzer.go

Main entry point. Responsibilities:
1. Define flags (`-goroutine-deriver`, `-context-carriers`)
2. Use `inspector.WithStack` to traverse AST with stack context
3. Call `checkers.FindContextScope()` to identify functions with context parameters
4. Build `funcScopes` map (function node -> ContextScope)
5. For each node, find nearest enclosing function with context
6. Dispatch to appropriate checkers (CallChecker or GoStmtChecker)
7. Run SSA-based zerolog analysis separately

### checkers/checker.go

Core types and interfaces:

```go
// ContextScope tracks context variable in a function
type ContextScope struct {
    Var  *types.Var  // The context variable
    Name string      // Variable name (for error messages)
}

// CheckContext holds runtime context for checks
type CheckContext struct {
    Pass      *analysis.Pass
    Scope     *ContextScope
    IgnoreMap IgnoreMap
    Carriers  []ContextCarrier
}

// Separated interfaces (Interface Segregation Principle)
type CallChecker interface {
    CheckCall(cctx *CheckContext, call *ast.CallExpr)
}

type GoStmtChecker interface {
    CheckGoStmt(cctx *CheckContext, stmt *ast.GoStmt)
}
```

### checkers/registry.go

Configuration and checker instantiation:

```go
type Config struct {
    GoroutineDeriver string           // Flag value
    ContextCarriers  []ContextCarrier // Parsed carrier types
}

type Checkers struct {
    Call   []CallChecker
    GoStmt []GoStmtChecker
}

func NewCheckers(cfg Config) Checkers
```

### checkers/pkgpath.go

Centralized package path constants:

```go
const (
    contextPkgPath    = "context"
    slogPkgPath       = "log/slog"
    syncPkgPath       = "sync"
    errgroupPkgPath   = "golang.org/x/sync/errgroup"
    zerologPkgPath    = "github.com/rs/zerolog"
    zerologLogPkgPath = "github.com/rs/zerolog/log"
)
```

### checkers/typeutil.go

Type checking utilities:

```go
// ContextCarrier for custom context types (echo.Context, cli.Context)
type ContextCarrier struct {
    PkgPath  string
    TypeName string
}

func ParseContextCarriers(s string) []ContextCarrier
func IsContextType(t types.Type) bool
func IsContextOrCarrierType(t types.Type, carriers []ContextCarrier) bool
func FindContextScope(pass *analysis.Pass, fnType *ast.FuncType, carriers []ContextCarrier) *ContextScope
```

### checkers/ignore.go

Comment directive support:

```go
type IgnoreMap map[int]struct{}  // Line numbers with ignore comments

func BuildIgnoreMap(fset *token.FileSet, file *ast.File) IgnoreMap
func (m IgnoreMap) ShouldIgnore(line int) bool  // Checks same line and previous line
```

## Checker Implementations

| Checker | File | Interface | Analysis | Checks |
|---------|------|-----------|----------|--------|
| zerolog | zerolog_ssa.go | (direct) | SSA | `.Ctx(ctx)` in chains |
| slog | slog.go | CallChecker | AST | Use `*Context` variants |
| errgroup | errgroup.go | CallChecker | AST | Context in `g.Go()` |
| waitgroup | waitgroup.go | CallChecker | AST | Context in `wg.Go()` |
| goroutine | goroutine.go | GoStmtChecker | AST | Context in `go func()` |
| goroutine_derive | goroutine_derive.go | GoStmtChecker | AST | Specific function in `go func()` |

## Analysis Approaches

### AST-based Analysis

Used for: slog, errgroup, waitgroup, goroutine, goroutine_derive

```go
insp.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
    scope := findEnclosingScope(funcScopes, stack)
    if scope == nil {
        return true
    }

    switch node := n.(type) {
    case *ast.GoStmt:
        for _, checker := range cs.GoStmt {
            checker.CheckGoStmt(cctx, node)
        }
    case *ast.CallExpr:
        for _, checker := range cs.Call {
            checker.CheckCall(cctx, node)
        }
    }
    return true
})
```

### SSA-based Analysis

Used for: zerolog (in `pkg/analyzer/checkers/zerologchecker/` subpackage)

SSA (Static Single Assignment) enables tracking values through:
- Variable assignments
- Phi nodes (control flow merges)
- Closures (via Parent() and FreeVar traversal)
- Struct fields (via Store/Load tracking)
- Type assertions and conversions

#### File Structure

```
zerologchecker/
├── checker.go  # Entry point (RunSSA), ssaChecker, function context discovery
├── tracer.go   # Strategy Pattern: ssaTracer interface, eventTracer/loggerTracer/contextTracer
├── trace.go    # Value tracing: traceValue, tracePhi, traceFreeVar, Store tracking
└── types.go    # Type checking: isEvent, isLogger, isContext, method classification
```

#### Strategy Pattern Architecture

Three tracers handle zerolog's type hierarchy:

```
┌─────────────┐     ┌─────────────┐     ┌───────────────┐
│ eventTracer │────▶│loggerTracer │────▶│ contextTracer │
│  (Event)    │◀────│  (Logger)   │◀────│   (Context)   │
└─────────────┘     └─────────────┘     └───────────────┘
        │                   │                    │
        └───────────────────┴────────────────────┘
                            │
                     ┌──────▼──────┐
                     │ traceCommon │
                     └─────────────┘
```

Each tracer implements:
```go
type ssaTracer interface {
    hasContext(call *ssa.Call, callee *ssa.Function, recv *types.Var) (found bool, delegate ssaTracer, delegateVal ssa.Value)
    continueOnReceiverType(recv *types.Var) bool
}
```

#### Context Sources by Type

| Type | Context Sources |
|------|-----------------|
| Event | `Event.Ctx(ctx)`, `Context.Ctx(ctx)`, `zerolog.Ctx(ctx)`, inherits from Logger |
| Logger | `zerolog.Ctx(ctx)`, `Context.Logger()`, `Logger.With()` |
| Context | `Context.Ctx(ctx)`, `Logger.With()` |

#### SSA Value Types Handled

| SSA Type | Purpose | Handler |
|----------|---------|---------|
| `*ssa.Call` | Method/function calls | `traceValue` (main) |
| `*ssa.Phi` | Control flow merge | `tracePhi` (ALL branches must have ctx) |
| `*ssa.UnOp` | Pointer dereference | `traceUnOp` + Store tracking |
| `*ssa.FreeVar` | Closure captures | `traceFreeVar` via MakeClosure |
| `*ssa.FieldAddr` | Struct field access | `findStoredValue` |
| `*ssa.TypeAssert` | Type assertion | Direct delegation |
| Others | Various transformations | Direct delegation to inner value |

#### Analysis Flow

```go
func RunSSA(pass, ssaInfo, ignoreMaps, isContextType) {
    // 1. Build function context map
    funcCtx := buildFunctionContextMap(ssaInfo, isContextType)

    // 2. Propagate context to closures (iterate until stable)

    // 3. For each function with context
    for fn, info := range funcCtx {
        chk := newSSAChecker(pass, info.name, ignoreMap)
        chk.checkFunction(fn)  // Find terminators, trace back
    }
}
```

## How Context Tracking Works

1. **First pass**: Identify all functions with context parameters via `FindContextScope`
2. **Second pass**: Use `inspector.WithStack` to traverse nodes
3. **For each node**: Find nearest enclosing function with context in stack
4. **Run checkers**: Each checker examines the node with the context scope

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
- Shadowed context parameters (uses type identity, not name)
- Context introduced in middle layers

## Shadowing Support

`UsesContext` uses `pass.TypesInfo.ObjectOf()` to check type identity:

```go
func (s *ContextScope) UsesContext(pass *analysis.Pass, node ast.Node) bool {
    ast.Inspect(node, func(n ast.Node) bool {
        if ident, ok := n.(*ast.Ident); ok {
            if obj := pass.TypesInfo.ObjectOf(ident); obj == s.Var {
                found = true
            }
        }
        return true
    })
    return found
}
```

This ensures that a shadowed variable with the same name is not confused with the original context.

## context-carriers Support

Custom context types (e.g., `echo.Context`) can be treated as context carriers:

```bash
ctxrelay -context-carriers=github.com/labstack/echo/v4.Context ./...
```

This affects:
- `FindContextScope`: Recognizes carrier types as context parameters
- `hasContextOrCarrierParam`: Checks for carrier types in closures
- AST-based checkers: All use carriers from CheckContext

**Not affected**: zerolog SSA analysis (zerolog.Ctx only accepts context.Context)

## Comparison with contextcheck

| Aspect | contextcheck | ctxrelay |
|--------|-------------|----------|
| Focus | `context.Background()`/`context.TODO()` | Specific API usage |
| Detection | Missing context in call chains | Not using available context |
| APIs | General | zerolog, slog, errgroup, etc. |

They are **complementary tools**.

## Integration Points

1. **Standalone**: `ctxrelay ./...`
2. **go vet**: `go vet -vettool=$(which ctxrelay) ./...`
3. **golangci-lint**: Requires upstream acceptance (future)

## Design Decisions

### Interface Segregation

Instead of a single `Checker` interface with `BaseChecker` for no-ops:

```go
// Before
type Checker interface {
    CheckCall(cctx *CheckContext, call *ast.CallExpr)
    CheckGoStmt(cctx *CheckContext, stmt *ast.GoStmt)
}
type BaseChecker struct{} // no-op implementations

// After
type CallChecker interface { CheckCall(...) }
type GoStmtChecker interface { CheckGoStmt(...) }
```

Benefits:
- Each checker only implements what it needs
- No "fake" method implementations
- Follows Interface Segregation Principle
- Type system enforces correct usage

### Variable Naming Convention

- `ctx` reserved for `context.Context`
- `cctx` used for `*CheckContext`

This avoids confusion in code that handles both.

### Centralized Constants

All package paths in `pkgpath.go` rather than scattered across files:
- Single source of truth
- Easy to update
- Clear dependencies

## Known SSA Limitations

The zerolog SSA analyzer has inherent limitations due to the complexity of static analysis.

### False Positives (Reports when it shouldn't)

| Pattern | SSA Reason | Example |
|---------|------------|---------|
| Channel send/receive | Channels are opaque in SSA; can't track value identity through send/receive | `ch <- e.Ctx(ctx); x := <-ch; x.Msg()` |
| Embedded field promotion | Promoted method call creates different receiver binding than explicit field access | `type H struct{ *Event }; h.Msg()` vs `h.Event.Msg()` |
| Closure-modified capture | Can't track side effects of closure invocation on captured variables | `f := func() { e = logger.Ctx(ctx) }; f(); e.Msg()` |
| sync.Pool | Get/Put creates opaque value flow; returned value type is `interface{}` | `pool.Get().(*Event).Msg()` |
| Phi node with nil | SSA models all branches including unreachable; `var e; if true { e = ... }` has implicit else | `var e *Event; if cond { e = x }; e.Msg()` |

### False Negatives (Misses when it should report)

| Pattern | SSA Reason | Example |
|---------|------------|---------|
| Method values | Method value extraction creates closure with special capture pattern | `msg := e.Msg; msg("test")` |
| Cross-function ctx | No interprocedural analysis; can't see ctx added inside called functions | `e := helperWithCtx(ctx); e.Msg()` |
| IIFE returns | Function return values not tracked across call boundaries | `func() *Event { return e.Ctx(ctx) }().Msg()` |

### Design Philosophy

The analyzer follows **"false positives over false negatives"** principle:
- False positives are annoying but safe (user adds `//ctxrelay:ignore`)
- False negatives are dangerous (bugs slip through silently)

Implementing interprocedural analysis would significantly increase complexity and compilation time. The current intraprocedural approach provides good coverage for common patterns while maintaining fast analysis.

## References

### Linters Studied

| Linter | Focus | Key Insight |
|--------|-------|-------------|
| staticcheck | Multi-purpose | Modular rule categories |
| errcheck | Error handling | Single-purpose elegance |
| contextcheck | Context propagation | Similar domain |
| nilaway | Nil safety | Cross-package analysis |
| bodyclose | HTTP body close | Minimal scope wins |
| ineffassign | Dead assignments | pkg/ separation |
| gomodifytags | Struct tags | Library + CLI pattern |

### Best Practices Adopted

1. **checkers/ subpackage** - Separate checker implementations from orchestration
2. **Type-aware checking** - Use `go/types` for precision
3. **Minimal exports** - Only expose what's necessary
4. **Minimal false positives** - Better to miss than annoy
5. **analysistest** - Standard testing approach
6. **singlechecker** - Standard CLI wrapper
7. **inspector.WithStack** - Proper context tracking through nested functions
8. **Interface segregation** - Separate interfaces for different check types
9. **SSA for complex tracking** - Use SSA when AST isn't enough
