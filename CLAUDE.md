# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**ctxrelay** is a Go linter that enforces context propagation best practices. It detects cases where a `context.Context` is available in function parameters but not properly passed to downstream calls that should receive it.

### Supported Checkers

- **zerolog**: Detect missing `.Ctx(ctx)` in zerolog chains
- **slog**: Detect non-context variants (`Info` vs `InfoContext`)
- **goroutine**: Detect `go func()` that doesn't capture/use context
- **errgroup**: Detect `errgroup.Group.Go()` closures without context
- **waitgroup**: Detect `sync.WaitGroup.Go()` closures without context (Go 1.25+)
- **goroutine-creator**: Detect calls to functions marked with `//goroutinectx:goroutine_creator` that pass closures without context
- **goroutine-derive**: Detect goroutines that don't call a specified context-derivation function (e.g., `apm.NewGoroutineContext`)
  - Activated via flag: `-goroutine-deriver=pkg/path.Func` or `-goroutine-deriver=pkg/path.Type.Method`
  - OR (comma): `-goroutine-deriver=pkg1.Func1,pkg2.Func2` - at least one must be called
  - AND (plus): `-goroutine-deriver=pkg1.Func1+pkg2.Func2` - all must be called
  - Mixed: `-goroutine-deriver=pkg1.Func1+pkg2.Func2,pkg3.Func3` - (Func1 AND Func2) OR Func3
- **gotask**: Detect `github.com/siketyan/gotask` task functions without context derivation (requires `-goroutine-deriver`)
  - `Do*` functions: checks that task arguments call the deriver
  - `Task.DoAsync` / `CancelableTask.DoAsync`: checks that ctx argument is derived

### Directives

- `//goroutinectx:ignore` - Suppress warnings for the next line or same line
- `//goroutinectx:goroutine_creator` - Mark a function as spawning goroutines with its func arguments

## Architecture

```
ctxrelay/
├── cmd/
│   └── ctxrelay/              # CLI entry point (singlechecker)
│       └── main.go
├── pkg/
│   └── analyzer/
│       ├── analyzer.go        # Main analyzer (orchestration)
│       ├── analyzer_test.go   # Integration tests
│       ├── checkers/          # Individual checker implementations
│       │   ├── checker.go     # CheckContext, ContextScope, CallChecker, GoStmtChecker
│       │   ├── typeutil.go    # Type checking utilities
│       │   ├── ignore.go      # IgnoreMap for //goroutinectx:ignore
│       │   ├── directive.go   # GoroutineCreatorMap for //goroutinectx:goroutine_creator
│       │   ├── deriver.go     # DeriveMatcher for OR/AND deriver function matching
│       │   ├── slogchecker/           # slog checker
│       │   ├── errgroupchecker/       # errgroup checker
│       │   ├── waitgroupchecker/      # waitgroup checker
│       │   ├── goroutinechecker/      # goroutine checker
│       │   ├── goroutinecreatorchecker/ # goroutine_creator directive checker
│       │   ├── goroutinederivechecker/  # goroutine_derive checker
│       │   ├── gotaskchecker/         # gotask task function checker
│       │   └── zerologchecker/        # SSA-based zerolog analysis
│       │       ├── checker.go # Entry point
│       │       ├── tracer.go  # Strategy Pattern tracers
│       │       ├── trace.go   # Value tracing, Phi handling
│       │       └── types.go   # Type checking, constants
│       └── testdata/          # Test fixtures
│           └── src/
│               ├── zerolog/   # basic.go, edge_*.go
│               ├── slog/
│               ├── goroutine/
│               ├── errgroup/
│               ├── waitgroup/
│               ├── goroutinecreator/     # goroutine_creator directive tests
│               ├── goroutinederive/      # Single deriver tests
│               ├── goroutinederiveand/   # AND (all must be called) tests
│               ├── goroutinederivemixed/ # Mixed AND/OR tests
│               └── gotask/               # gotask checker tests
├── scripts/
│   └── verify-test-patterns.pl  # Test naming consistency checker
├── docs/
│   └── ARCHITECTURE.md        # Architecture overview
├── .golangci.yaml             # golangci-lint configuration
└── README.md
```

### Key Design Decisions

1. **Type-safe analysis**: Uses `go/types` for accurate detection (not just name-based)
2. **Nested function support**: Uses `inspector.WithStack` to track context through closures
3. **Shadowing support**: `UsesContext` uses type identity, not name matching
4. **Interface segregation**: `CallChecker` and `GoStmtChecker` interfaces (no BaseChecker)
5. **Minimal exports**: Only necessary types/functions are exported from `checkers` package
6. **Zero false positives**: Prefer missing issues over false alarms
7. **SSA for zerolog**: Uses SSA form to track Event values through assignments
8. **Multiple context tracking**: Tracks ALL context parameters, not just the first one. If ANY context variable is used, the check passes. Error messages report the first context name for consistency.

### Checker Interface Design

```go
// Separated interfaces - checkers implement only what they need
type CallChecker interface { CheckCall(cctx *CheckContext, call *ast.CallExpr) }
type GoStmtChecker interface { CheckGoStmt(cctx *CheckContext, stmt *ast.GoStmt) }

type Checkers struct {
    Call   []CallChecker   // slog, errgroup, waitgroup
    GoStmt []GoStmtChecker // goroutine, goroutine_derive
}

// Functional Option Pattern for flexible checker composition
type Option func(*Checkers)

// Usage:
cs := NewCheckers(
    WithSlog(),
    WithGoroutine(),
    WithErrgroup(),
    WithGoroutineDerive("pkg.NewGoroutine"),
)
```

**Why two interfaces?**
- `GoStmtChecker`: For `go` keyword statements (`go func() {}()`)
- `CallChecker`: For function calls (`g.Go()`, `wg.Go()`, `log.Info()`)

These are AST-level distinctions: `go` is a statement, function calls are expressions.

### Goroutine-Related Checkers

Four checkers handle goroutine context propagation:

| Checker | Target | AST Node | Higher-Order Support |
|---------|--------|----------|---------------------|
| goroutine | `go func(){}()` | `*ast.GoStmt` | Yes (`go fn()()`) |
| errgroup | `g.Go(func(){})` | `*ast.CallExpr` | Yes (`g.Go(fn)`, `g.Go(make())`) |
| waitgroup | `wg.Go(func(){})` | `*ast.CallExpr` | Yes (`wg.Go(fn)`, `wg.Go(make())`) |
| goroutine-creator | `//goroutinectx:goroutine_creator` marked funcs | `*ast.CallExpr` | Yes (func args checked) |

**Supported patterns:**
- Literal: `g.Go(func() { ... })`
- Variable: `g.Go(fn)` where `fn := func() { ... }`
- Call result: `g.Go(makeWorker())` where `makeWorker` returns a func
- Call with ctx: `g.Go(makeWorker(ctx))` - ctx passed to factory
- Directive: Functions marked with `//goroutinectx:goroutine_creator` check their func arguments

**goroutine_creator Directive:**
```go
//goroutinectx:goroutine_creator
func runWithGroup(g *errgroup.Group, fn func() error) {
    g.Go(fn)  // fn is spawned as goroutine
}

func caller(ctx context.Context) {
    runWithGroup(g, func() error {
        // Warning: should use ctx
        return nil
    })
}
```

**Known LIMITATIONs:**
- Channel receives - can't trace func from channel
- Nested closure ctx (e.g., `defer func() { _ = ctx }()`) - intentionally not counted
- `interface{}` type assertion - can't trace func through type assertion

### Derive Function Matching (DeriveMatcher)

`checkers/deriver.go` provides shared OR/AND logic for matching derive functions. Used by:
- `goroutinederivechecker`: checks `go func()` calls
- `gotaskchecker`: checks gotask task functions and DoAsync calls

```go
// DeriveMatcher supports OR (comma) and AND (plus) operators
type DeriveMatcher struct {
    OrGroups [][]DeriveFuncSpec  // Each group must have ALL specs satisfied
    Original string               // Original flag value for error messages
}

// SatisfiesAnyGroup checks if ANY OR group is fully satisfied
func (m *DeriveMatcher) SatisfiesAnyGroup(pass *analysis.Pass, node ast.Node) bool
```

**Parsing Logic:**
- `pkg.Func1,pkg.Func2` → OR: either Func1 or Func2
- `pkg.Func1+pkg.Func2` → AND: both Func1 and Func2
- `pkg.A+pkg.B,pkg.C` → Mixed: (A AND B) OR C

### gotask Checker

The gotask checker handles `github.com/siketyan/gotask` library:

| Target | Check |
|--------|-------|
| `gotask.Do*` functions | Task arguments (2nd+) must call deriver in their body |
| `Task.DoAsync` | Context argument (1st) must be derived |
| `CancelableTask.DoAsync` | Context argument (1st) must be derived |

**Key insight:** Since gotask tasks run as goroutines, they need to call the deriver function inside their body - there's no way to wrap the context at the call site.

**Known LIMITATIONs:**
- Variable references can't be traced (e.g., `task := NewTask(fn); DoAll(ctx, task)`)
- Nested function literals aren't traversed (e.g., deriver in `defer func(){}()` inside task)
- Higher-order function returns can't be traced (e.g., `DoAll(ctx, makeTask())`)

### Zerolog SSA Strategy Pattern

The zerolog checker uses SSA analysis with Strategy Pattern for tracing:

```
┌─────────────┐     ┌─────────────┐     ┌───────────────┐
│ eventTracer │────▶│loggerTracer │────▶│ contextTracer │
│  (Event)    │◀────│  (Logger)   │◀────│   (Context)   │
└─────────────┘     └─────────────┘     └───────────────┘
        │                   │                    │
        └───────────────────┴────────────────────┘
                            │
                     ┌──────▼──────┐
                     │ traceCommon │  (Phi, UnOp, FreeVar, etc.)
                     └─────────────┘
```

- `ssaTracer` interface: `hasContext()`, `continueOnReceiverType()`
- Each tracer knows its context sources and delegates across type boundaries
- Handles: variable assignments, conditionals (Phi), closures, struct fields, defer

**Why zerologchecker can't be split into subpackages:**
The tracing logic (trace.go) defines methods on the `checker` struct, which is defined in checker.go.
Go requires methods to be in the same package as the type, creating a bidirectional dependency:
- checker.go uses `tracer` interface and `newTracers()` from tracer.go
- trace.go defines methods on `checker` struct and uses `tracer` interface
This tight coupling is intentional for performance and simplicity.

## Development Commands

```bash
# Run tests
go test ./...

# Run tests with verbose output
go test ./pkg/analyzer/... -v

# Build CLI
go build -o bin/ctxrelay ./cmd/ctxrelay

# Run linter on itself
go vet -vettool=./bin/ctxrelay ./...

# Run golangci-lint
golangci-lint run ./...

# Format code
go fmt ./...

# Verify test pattern naming consistency
./scripts/verify-test-patterns.pl        # Check for inconsistencies
./scripts/verify-test-patterns.pl -v     # Verbose: show all patterns
./scripts/verify-test-patterns.pl -q     # Quiet: exit code only (for CI)

# Run test metadata validation (IMPORTANT: Must specify file path)
go test ./testdata/metatest/validation_test.go           # Run validation
go test -v ./testdata/metatest/validation_test.go        # Verbose output
go test -v -run TestStructureValidation/AllFunctionsAccountedFor ./testdata/metatest/validation_test.go
```

**IMPORTANT:** The validation test MUST be run with the file path `./testdata/metatest/validation_test.go`. Running `go test ./testdata/metatest` or `cd testdata/metatest && go test` will NOT execute the test due to its special structure.

## Adding a New Checker

1. Create `pkg/analyzer/checkers/<name>.go`:
```go
package checkers

type myChecker struct{}  // unexported, no base needed

// Implement CallChecker for call expression checks
func (c *myChecker) CheckCall(cctx *CheckContext, call *ast.CallExpr) {
    // Implementation using cctx.Pass, cctx.Scope, cctx.IgnoreMap
}

// OR implement GoStmtChecker for go statement checks
func (c *myChecker) CheckGoStmt(cctx *CheckContext, stmt *ast.GoStmt) {
    // Implementation
}
```

2. Register in `pkg/analyzer/checkers/registry.go` under `NewCheckers()`

3. Add test fixtures in `pkg/analyzer/testdata/src/<name>/`

4. Add test case in `pkg/analyzer/analyzer_test.go`

## Testing Strategy

- Use `analysistest` for all analyzer tests
- Test fixtures use `// want` comments for expected diagnostics
- Test structure per checker:
  - `===== SHOULD REPORT =====` - Cases that should trigger warnings
  - `===== NESTED FUNCTIONS - SHOULD REPORT =====` - Nested cases
  - `===== SHOULD NOT REPORT =====` - Negative cases
  - `===== SHADOWING TESTS =====` - Variable shadowing cases
  - `===== EDGE CASES =====` - Corner cases

## Code Style

- Follow standard Go conventions
- Use `go/analysis` framework
- Prefer `inspector.WithStack` over `ast.Inspect` for traversal
- Type utilities go in `checkers/typeutil.go` (unexported)
- Checker types are unexported; only interface and registry are public
- Prefix file-specific variables with checker name (e.g., `slogNonContextFuncs`)

### Comment Guidelines

**Comments should inform newcomers, not document history.**

- ❌ Bad: `// moved from evil.go - this is higher-order function`
- ❌ Bad: `// refactored in session 5`
- ✓ Good: `// tests basic go fn()() pattern`
- ✓ Good: `// LIMITATION: cross-function tracking not supported`

**When NOT to comment:**
- Refactoring moves (where something came from)
- Session/date information
- Obvious code behavior

**When to comment:**
- WHY something exists (design rationale)
- LIMITATION markers for known gaps
- Non-obvious behavior that would confuse readers

**Exception:** Major architectural changes that affect understanding may warrant brief explanation, but prefer updating documentation (CLAUDE.md, ARCHITECTURE.md) over inline comments.

## Documentation Strategy

| File | Purpose | Git Tracked |
|------|---------|-------------|
| `CLAUDE.md` | AI assistant guidance, architecture overview, coding conventions | Yes |
| `docs/` | Detailed design docs, API references, user-facing documentation | Yes |
| `TASKS.md` | Temporary session notes, in-progress work, handoff context | No (gitignored) |

**Principle**: Design decisions, architecture diagrams, and anything useful for future development goes in git-tracked files. `TASKS.md` is ephemeral scratch space for the current session only.

## File Organization

**Proactive file organization is mandatory.** When creating new files or adding symbols to existing files, always evaluate:

1. **Naming consistency**: Does the name follow existing conventions in the directory?
2. **Location appropriateness**: Is this the right directory/package for this content?
3. **File consolidation**: Should this be merged with an existing file?
4. **File splitting**: Is this file getting too large or handling too many concerns?

### Testdata Naming Conventions

Test files are organized by complexity and purpose:

```
pkg/analyzer/testdata/src/<checker>/
├── basic.go           # Core functionality - simple good/bad cases
├── advanced.go        # Complex patterns - higher-order functions, deep nesting
├── evil.go            # Evil edge cases - adversarial/unusual patterns
├── evil_<aspect>.go   # Evil cases for specific aspects (e.g., evil_ssa.go, evil_logger.go)
└── <feature>.go       # Feature-specific tests (e.g., with_logger.go)
```

**Example: zerolog testdata structure**
```
pkg/analyzer/testdata/src/zerolog/
├── basic.go           # Simple good/bad cases
├── evil.go            # General edge cases (nesting, closures, conditionals)
├── evil_ssa.go        # SSA-specific limitations (IIFE, Phi, channels)
├── evil_logger.go     # Logger transformation patterns (Level, Output, With)
└── with_logger.go     # WithLogger-specific tests
```

**File Classification Principle:**

Classification is based on **human intuition** - "would a developer write this daily?"

| File | Criterion | Content |
|------|-----------|---------|
| `basic.go` | Daily patterns | Patterns you write and see every day |
| `advanced.go` | Real-world but not daily | Production patterns that are common but not routine |
| `evil.go` | Adversarial | Unusual patterns that test analyzer limits |

**Classification Guidelines:**

1. **basic.go** - Daily patterns (1-level nesting max)
   - Simple good/bad cases (direct context use vs. no use)
   - 1-level goroutine (`go func() { ... }()`)
   - Variable shadowing
   - Ignore directives (`//goroutinectx:ignore`)
   - Multiple context parameters
   - Direct function calls (`go doSomething(ctx)`)

2. **advanced.go** - Real-world complex patterns (production code, but not daily)
   - Defer patterns (deferred cleanup, recovery)
   - Loop patterns (for, range with goroutines)
   - Channel operations (send/receive, select)
   - WaitGroup patterns
   - Method calls on captured objects
   - Control flow (switch, conditional goroutines)

3. **evil.go** - Adversarial patterns (tests analyzer limits)
   - 2+ level goroutine nesting
   - Higher-order functions (`go fn()()`, `go fn()()()`)
   - IIFE (Immediately Invoked Function Expression)
   - Interface method calls
   - LIMITATION cases documenting analyzer boundaries
   - Goroutines in expressions, deferred functions

**Decision Tree:**

```
Is it 1-level goroutine with straightforward code?
├─ Yes → basic.go
└─ No → Is it a production pattern (defer, loops, channels, WaitGroup)?
         ├─ Yes → advanced.go
         └─ No → evil.go (nesting 2+, go fn()(), IIFE, LIMITATION)
```

**LIMITATION Comments:**
Test cases that document known analyzer limitations should be prefixed with `limitation` in their function name and include a `// LIMITATION:` comment explaining the gap:

```go
// LIMITATION: Variable reassignment not tracked - uses first assignment only
func limitationReassignedFn(ctx context.Context) {
    fn := func() { doSomething(ctx) }
    fn = func() { doNothing() }  // Reassigned!
    go fn()()  // Currently passes - should fail
}
```

### Trigger Points for Reorganization
- **New file creation**: Consider if existing file should be renamed/split
- **Symbol addition**: Check if file is growing beyond single responsibility
- **Test addition**: Verify test file naming matches pattern
- **Phase 3 review**: Always include file organization in code style review

## Quality Improvement Cycle

When improving code quality, follow this iterative cycle:

### Phase 1: QA Engineer - Evil Edge Case Testing
- Add thorough, adversarial test cases that push the analyzer to its limits
- Cover edge cases: deep nesting, closures, loops, conditionals, type conversions
- Mark failing cases with `// LIMITATION:` comments explaining the gap
- Document what the ideal behavior should be vs current behavior

### Phase 2: Implementation Engineer - Address Limitations
- Review all `LIMITATION` comments and attempt to resolve them
- Prioritize fixes that improve real-world detection accuracy
- When a limitation is resolved, remove the comment and update the test expectation
- Document truly unfixable limitations (e.g., SSA optimization in test stubs)

### Phase 3: Code Style Engineer - Refactoring Review
- Review code for clarity, maintainability, and consistency
- Categorize each suggestion:
  - **Should not do**: Would harm readability or add unnecessary complexity
  - **Either way**: Neutral impact, matter of preference
  - **Should do**: Clear improvement to code quality
- Implement all "Should do" items first, then "Either way" items
- Only skip "Should not do" items

**Code Style Engineer Principles:**

1. **Namespace Pollution Intolerance**: The code style engineer strongly opposes "ad-hoc namespace pollution" common in Go's conventional compromises. When a package handles multiple concerns, generic names that only reflect one concern risk collisions. Solutions:
   - Use prefixes to disambiguate (e.g., `ssaTraceEvent`, `astVisitNode`)
   - Split into separate packages when concerns are distinct enough

2. **Design Pattern Advocate**: The code style engineer loves design patterns and actively proposes their application when encountering ad-hoc code. Particularly favors:
   - **Strategy Pattern** for AST/SSA traversal with pluggable behaviors
   - **Visitor Pattern** for tree-structured data processing
   - **Factory Pattern** for creating checker instances

3. **Responsibility Boundary Enforcement**: The code style engineer is strict about clear responsibility boundaries. Each function/type should have a single, well-defined purpose. Cross-cutting concerns should be handled through composition or dependency injection, not ad-hoc parameter passing.

4. **Export/Unexport Discipline**: The code style engineer is particular about visibility. Everything should be unexported by default; only export what is genuinely needed by external packages. Internal helpers, implementation details, and intermediate types must remain unexported to maintain encapsulation.

5. **Method vs Function Discipline**: For packages focused on a single concern, avoid unnecessary methods. Functions are preferred unless:
   - The method genuinely operates on the receiver's state
   - Grouping as methods provides clearer semantic organization
   - Namespace protection is needed (defensive programming)

   If a method doesn't use its receiver, either:
   - Convert it to a function if appropriate
   - Omit the receiver name (use `_`) to signal intentional non-use
   - Keep as method if semantic grouping justifies it (document why)

   **Function → Method Conversion Guideline:**
   When the first argument is clearly the "subject" of the operation, convert to a method:
   - `func doSomething(cctx *CheckContext, target *ast.Expr)` → `func (cctx *CheckContext) doSomething(target *ast.Expr)`
   - The "subject" is the entity performing the action, not just providing context

   **Keep as function** when:
   - There are 2+ arguments and it's unclear which is the "subject"
   - Example: `FindFuncLitAssignment(cctx, v)` - finding for `v` using `cctx` (ambiguous subject)

### Phase 4: Newbie - Naive Questions
Become a complete beginner who has never seen the code. Ask genuinely confused questions:
- "What is SSA? Why do we need it?"
- "Why are there three tracers? Can't we just use one?"
- "What does 'Phi node' mean? Why does it matter?"
- "I don't understand why this function exists"
- "What's the flow when I call the analyzer?"

The goal is to identify knowledge gaps and unclear abstractions. Don't pretend to understand - if something is confusing, it needs better documentation.

### Phase 5: Teacher Duo - Explanation & Documentation
The **Implementation Engineer** and **Design Pattern Advocate** collaborate to answer the Newbie's questions:
- Explain concepts step-by-step, building from fundamentals
- Use analogies and diagrams where helpful
- Identify which explanations belong in which document

**Documentation Outputs:**

| Document | Purpose | Style |
|----------|---------|-------|
| `docs/ARCHITECTURE.md` | Precise technical specification | Reference-oriented, complete |
| `docs/TUTORIAL.md` | Step-by-step learning guide | Beginner-friendly, progressive |

Both documents must be kept in sync - when code changes, update both:
- ARCHITECTURE.md: What it is (accurate specification)
- TUTORIAL.md: How to understand it (pedagogical progression)

### Repeat
Continue the cycle until:
- No new meaningful edge cases can be found
- All addressable limitations are resolved
- Code style meets quality standards
- Newbie questions are answered in documentation
- Both reference and tutorial docs are current

## Test Pattern Coverage Matrix

Test cases use 2-letter prefixes to identify checker groups:

**Goroutine Group** (context usage check):
- `GO` - goroutine checker (`go func(){}()`)
- `GE` - errgroup checker (`g.Go(func(){})`)
- `GW` - waitgroup checker (`wg.Go(func(){})`)

**Creator Group** (goroutine_creator directive):
- `GC` - goroutinecreator (`//goroutinectx:goroutine_creator` marked functions)

**Derive Group** (deriver function call check):
- `DD` - goroutinederive (single deriver)
- `DA` - goroutinederiveand (AND - all must be called)
- `DM` - goroutinederivemixed (Mixed AND/OR)

**Gotask Group** (gotask library deriver check):
- `GT` - gotask checker (basic.go patterns)
- `EV` - gotask evil patterns (evil.go patterns)

Goroutine group patterns should be consistent across GO/GE/GW.
Creator group (GC) patterns are directive-specific and standalone.
Derive group patterns intentionally diverge (DD/DA/DM test different semantics).

### Goroutine Group - Basic Patterns (01-19)

| # | Pattern | GO | GE | GW | Description |
|---|---------|----|----|----|----|
| 01 | Literal without ctx | GO01 | GE01 | GW01 | Basic bad case |
| 02 | Literal with ctx | GO02 | GE02 | GW02 | Basic good case |
| 03 | No ctx param | GO03 | GE03 | GW03 | Not checked |
| 04 | Shadow with non-ctx type | GO04 | GE04 | GW04 | Shadows ctx with different type |
| 05 | Uses ctx before shadow | GO05 | GE05 | GW05 | Valid usage before shadowing |
| 06 | Ignore directive (same line) | GO06 | GE06 | GW06 | `//goroutinectx:ignore` |
| 07 | Ignore directive (prev line) | GO07 | GE07 | GW07 | `//goroutinectx:ignore` |
| 08 | Multiple ctx params (bad) | GO08 | GE08 | GW08 | Reports first ctx when none used |
| 09 | Multiple ctx params (good) | GO09 | GE09 | GW09 | Uses one of multiple ctx params |
| 10 | Inner func has own ctx param | GO10 | GE10 | GW10 | Closure has own ctx param |
| 11 | Direct function call | GO11 | - | - | `go doSomething(ctx)` |
| 12 | Variable func | - | GE12 | GW12 | `g.Go(fn)` patterns |
| 13 | Higher-order func | - | GE13 | GW13 | `g.Go(makeWorker())` patterns |
| 14 | Ctx as non-first param | GO14 | GE14 | GW14 | Context not first param |
| 15 | Slice index | - | GE15 | GW15 | `g.Go(tasks[0])` |
| 16 | Map key | - | GE16 | GW16 | `g.Go(tasks["key"])` |
| 17 | Traditional WaitGroup | - | - | GW17 | Add/Done pattern (not checked) |
| 18 | Struct field | - | GE18 | GW18 | `g.Go(holder.task)` |

### Goroutine Group - Advanced Patterns (20-39)

| # | Pattern | GO | GE | GW | Description |
|---|---------|----|----|----|----|
| 20 | Defer without ctx | GO20 | - | - | Closure has defer but no ctx |
| 21 | Deferred nested closure (LIMIT) | GO21 | GE21 | GW21 | Ctx only in deferred closure |
| 22 | For loop | GO22 | GE22 | GW22 | Go in for loop |
| 23 | Range loop | GO23 | GE23 | GW23 | Go in range loop |
| 24 | Conditional | GO24 | GE24 | GW24 | Go in if/else branches |
| 25-32 | GO-specific patterns | GO25-32 | - | - | Channel, select, method call, etc. |
| 35 | Go call inside inner func | - | GE35 | GW35 | Nested IIFE patterns |

### Goroutine Group - Evil Patterns (40-90)

| # | Pattern | GO | GE | GW | Description |
|---|---------|----|----|----|----|
| 40-65 | Nesting & higher-order | GO40-65 | - | - | Nested goroutines, `go fn()()`, IIFE |
| 70-78 | Multiple context | GO70-78 | GE70-74 | GW70-74 | Multiple ctx params |
| 85-86 | Higher-order multiple ctx | - | GE85-86 | GW85-86 | Factory with multiple ctx |
| 90-92 | GO-specific LIMITATIONs | GO90-92 | - | - | Closure levels, deferred spawn |
| 100-108 | GE/GW LIMITATIONs | - | GE100-108 | GW100-108 | Interface, channel, chaotic |

### Creator Group Patterns (GC)

GC patterns test the `//goroutinectx:goroutine_creator` directive:

| # | Pattern | GC | Description |
|---|---------|----|----|
| 01 | Errgroup func without ctx | GC01 | Basic bad case with errgroup creator |
| 02 | WaitGroup func without ctx | GC02 | Basic bad case with waitgroup creator |
| 03 | Inline literal without ctx | GC03 | Direct func literal in call |
| 04 | Multiple func args - both bad | GC04 | Multiple func params, none use ctx |
| 05 | Multiple func args - first bad | GC05 | First doesn't use ctx, second does |
| 06 | Multiple func args - second bad | GC06 | First uses ctx, second doesn't |
| 10 | Errgroup func with ctx | GC10 | Basic good case with errgroup creator |
| 11 | WaitGroup func with ctx | GC11 | Basic good case with waitgroup creator |
| 12 | Inline literal with ctx | GC12 | Direct func literal uses ctx |
| 13 | Multiple func args - both good | GC13 | Both func params use ctx |
| 14 | No ctx param | GC14 | Function has no ctx param - not checked |
| 15 | Func has own ctx param | GC15 | Closure declares own ctx param |
| 20 | Non-creator function | GC20 | Call without directive - not checked |

### Derive Group Patterns

DD, DA, DM patterns intentionally diverge because they test different deriver logic:
- **DD** (single): Tests single deriver function call
- **DA** (AND): Tests that ALL specified derivers are called
- **DM** (Mixed): Tests AND groups with OR alternatives

### Gotask Group Patterns (GT)

GT patterns test gotask library checker (requires `-goroutine-deriver`):
- Prefix `GT` in basic.go and `EV` in evil.go

| # | Pattern | GT | Description |
|---|---------|----|----|
| 01 | DoAllFnsSettled without deriver | GT01 | Basic bad case |
| 02 | Multiple args - some without | GT02 | Partial deriver coverage |
| 03 | Deriver on parent ctx | GT03 | Bad - deriver must be in task body |
| 10-11 | DoAllSettled with NewTask | GT10-11 | NewTask wrapper patterns |
| 20-22 | DoAsync without deriver | GT20-22 | Task/CancelableTask.DoAsync |
| 30-32 | DoAllFnsSettled with deriver | GT30-32 | Good cases |
| 40 | DoAllSettled NewTask with deriver | GT40 | Good case |
| 50-51 | DoAsync with deriver | GT50-51 | Good Task/CancelableTask |
| 60-63 | Other Do* without deriver | GT60-63 | DoAll, DoAllFns, DoRace, DoRaceFns |
| 70-73 | Other Do* with deriver | GT70-73 | Good other Do* |
| 80-81 | Ignore directive | GT80-81 | `//goroutinectx:ignore` |
| 90 | No ctx param | GT90 | Not checked |

Evil patterns (EV01-EV110):
- Variable/variadic tasks (EV01, EV10-11)
- Nested closure (EV20)
- Method chaining (EV40-51)
- LIMITATIONs (EV100+)

Use `./scripts/verify-test-patterns.pl -v` to see all patterns by group.

### Maintaining This Matrix

**Numbering Rules:**
- Use `b`, `c`, `d` suffixes for variants (e.g., GO01b, GE02c)
- Basic patterns (01-19): Common across checkers
- Advanced patterns (20-39): Real-world complex patterns
- Evil patterns (40-99): Adversarial, checker-specific
- LIMITATIONs (90-99 for GO, 100+ for GE/GW): Known analyzer boundaries
- GC patterns (01-20): Creator directive-specific tests

**Verification:**
```bash
./scripts/verify-test-patterns.pl        # Check for inconsistencies
./scripts/verify-test-patterns.pl -v     # Show all patterns by group
```

Goroutine group (GO/GE/GW) patterns should be consistent.
Creator group (GC) patterns are standalone (directive-specific).
Derive group (DD/DA/DM) patterns are expected to diverge.

## Serena MCP Server Usage Guidelines

When using Serena for code analysis, avoid excessive parallel searches to prevent server freezing.

**Best Practices:**
- Use sequential symbol searches when analyzing broad code areas
- Start with `get_symbols_overview` before diving into `find_symbol` calls
- Prefer single `find_symbol` calls over parallel searches for the same file
- When exploring multiple checkers, analyze them one at a time

**Sequential Pattern (Recommended):**
```
1. get_symbols_overview for file A
2. find_symbol for specific symbol in A
3. get_symbols_overview for file B
4. find_symbol for specific symbol in B
```

**Avoid:**
- Launching multiple parallel `find_symbol` calls across many files
- Running broad searches (e.g., searching entire codebase) in parallel
- Using `search_for_pattern` with very broad patterns in parallel

**Throughput vs Latency:**
When search scope is large, prioritize reliability over speed by executing searches sequentially rather than in parallel.
