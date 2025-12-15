# ctxrelay Tutorial: Understanding the Zerolog SSA Analyzer

This guide explains how the zerolog SSA analyzer works, starting from the basics and building up to the full picture. It's designed for developers who are new to static analysis or SSA.

## Table of Contents

1. [What Problem Are We Solving?](#what-problem-are-we-solving)
2. [Why SSA? (And What Is It?)](#why-ssa-and-what-is-it)
3. [The Big Picture: Analysis Flow](#the-big-picture-analysis-flow)
4. [Understanding the Three Tracers](#understanding-the-three-tracers)
5. [What Is a Phi Node?](#what-is-a-phi-node)
6. [Handling Closures and Captured Variables](#handling-closures-and-captured-variables)
7. [Store Tracking: Following Values Through Memory](#store-tracking-following-values-through-memory)
8. [The Strategy Pattern: Why Three Tracers?](#the-strategy-pattern-why-three-tracers)

---

## What Problem Are We Solving?

When using zerolog, you should pass context to your log chains:

```go
// BAD - context is available but not used
func handleRequest(ctx context.Context, logger zerolog.Logger) {
    logger.Info().Msg("handling request")  // Missing .Ctx(ctx)!
}

// GOOD - context is propagated
func handleRequest(ctx context.Context, logger zerolog.Logger) {
    logger.Info().Ctx(ctx).Msg("handling request")  // Correct!
}
```

The analyzer detects the "BAD" case and reports it.

### Why Is This Hard?

If zerolog only allowed direct chains, detecting missing `.Ctx()` would be easy:

```go
logger.Info().Msg("test")  // Easy to check: no .Ctx() in chain
```

But real code is messy:

```go
// Variable assignment
e := logger.Info()
e = e.Str("key", "value")
e.Msg("test")  // Where did 'e' come from? Did it ever have .Ctx()?

// Conditionals
var e *zerolog.Event
if condition {
    e = logger.Info().Ctx(ctx)
} else {
    e = logger.Warn()  // Oops, no .Ctx() here!
}
e.Msg("test")  // Is this OK? Depends on which branch!

// Closures
e := logger.Info()
go func() {
    e.Msg("async")  // 'e' is captured - did it have .Ctx()?
}()
```

We need to **trace values backward** through all these transformations.

---

## Why SSA? (And What Is It?)

### The Problem with AST

AST (Abstract Syntax Tree) represents code **as written**:

```go
e := logger.Info()
e = e.Str("key", "value")
e.Msg("test")
```

In AST, `e` appears 4 times. Which `e` is which? AST doesn't track this - it just sees the name "e".

### What SSA Gives Us

SSA (Static Single Assignment) transforms code so each variable is assigned exactly once:

```
t1 = logger.Info()           // Original: e := logger.Info()
t2 = t1.Str("key", "value")  // Original: e = e.Str(...)
t3 = t2.Msg("test")          // Original: e.Msg("test")
```

Now there's no confusion: `t3` comes from `t2`, which comes from `t1`. We can trace backward unambiguously.

### An Analogy

Think of it like tracking a package through a shipping system:
- **AST**: "The package is at the warehouse" (but which package?)
- **SSA**: "Package #47291 moved from Station A → Station B → Station C"

SSA gives every value a unique identity we can follow.

---

## The Big Picture: Analysis Flow

Here's what happens when the analyzer runs:

```
┌──────────────────────────────────────────────────────────────┐
│ 1. Find functions with context.Context parameter             │
│    func handleRequest(ctx context.Context, ...) { ... }      │
│                           ↓                                  │
├──────────────────────────────────────────────────────────────┤
│ 2. Propagate context info to nested closures                 │
│    func handleRequest(ctx context.Context, ...) {            │
│        go func() { ... }  ← This closure also "has" ctx      │
│    }                                                         │
│                           ↓                                  │
├──────────────────────────────────────────────────────────────┤
│ 3. Find all terminator calls (Msg, Msgf, Send)               │
│    e.Msg("test")  ← Is this OK?                              │
│                           ↓                                  │
├──────────────────────────────────────────────────────────────┤
│ 4. Trace backward to find if .Ctx() was called               │
│    e.Msg("test")                                             │
│         ↑ trace back                                         │
│    e = logger.Info().Str("k","v").Ctx(ctx)  ← Found it!      │
│                           ↓                                  │
├──────────────────────────────────────────────────────────────┤
│ 5. Report if .Ctx() is missing                               │
│    "zerolog call chain missing .Ctx(ctx)"                    │
└──────────────────────────────────────────────────────────────┘
```

Step 4 is where most of the complexity lives.

---

## Understanding the Three Tracers

### Zerolog's Type Hierarchy

Zerolog has three main types that interact:

```go
// Logger - the main logger instance
var logger zerolog.Logger

// Event - what you build up with .Str(), .Int(), etc.
e := logger.Info()  // Returns *zerolog.Event
e.Str("key", "value").Msg("test")

// Context - a builder for derived loggers
c := logger.With()  // Returns zerolog.Context
c.Str("key", "value").Logger()  // Returns zerolog.Logger
```

### Why Three Tracers?

Each type has **different ways** to get context:

| Type | How it gets context |
|------|---------------------|
| Event | `.Ctx(ctx)` directly, or inherits from Logger |
| Logger | `zerolog.Ctx(ctx)`, or from `Context.Logger()` |
| Context | `.Ctx(ctx)`, or inherits from parent Logger |

When we trace an Event, we might find it came from a Logger:
```go
logger.Info().Msg("test")
//     ↑
// This is a Logger - switch to loggerTracer!
```

So we need tracers that can **delegate to each other**:

```
┌─────────────┐     ┌─────────────┐     ┌───────────────┐
│ eventTracer │────▶│loggerTracer │────▶│ contextTracer │
│  (Event)    │◀────│  (Logger)   │◀────│   (Context)   │
└─────────────┘     └─────────────┘     └───────────────┘
```

### Example Trace

```go
zerolog.Ctx(ctx).Info().Str("k","v").Msg("test")
```

1. Start at `Msg("test")` with `eventTracer`
2. Trace back to `.Str("k","v")` - still Event, continue
3. Trace back to `.Info()` - this returns Event from Logger
4. Switch to `loggerTracer`, trace back to `zerolog.Ctx(ctx)`
5. Found context source! Return `true`.

---

## What Is a Phi Node?

### The Problem: Conditional Assignment

```go
var e *zerolog.Event
if condition {
    e = logger.Info().Ctx(ctx)
} else {
    e = logger.Warn()
}
e.Msg("test")  // Which 'e' is this?
```

In SSA, this becomes:

```
block0:
    if condition goto block1 else block2

block1:
    t1 = logger.Info().Ctx(ctx)
    goto block3

block2:
    t2 = logger.Warn()
    goto block3

block3:
    t3 = φ(t1, t2)    ← Phi node!
    t3.Msg("test")
```

### What Phi Means

Phi (φ) is like a "choose one" node. It says:
> "My value is t1 if we came from block1, or t2 if we came from block2"

### How We Handle It

For context checking, we require **ALL branches** to have context:

```go
if condition {
    e = logger.Info().Ctx(ctx)  // Has context
} else {
    e = logger.Warn()           // NO context!
}
e.Msg("test")  // REPORT: not all branches have .Ctx()
```

The analyzer reports this because:
- Branch 1: has `.Ctx(ctx)` ✓
- Branch 2: missing `.Ctx()` ✗
- Result: not safe, report error

---

## Handling Closures and Captured Variables

### The Challenge

```go
func handleRequest(ctx context.Context, logger zerolog.Logger) {
    e := logger.Info()
    go func() {
        e.Msg("async")  // 'e' is captured from outer function
    }()
}
```

In SSA, the closure captures `e` as a "FreeVar" (free variable).

### How We Trace FreeVars

1. Find the FreeVar's index in the closure's parameter list
2. Find the MakeClosure instruction that created this closure
3. Look at the binding at that index - that's the original value
4. Continue tracing from there

```
handleRequest:
    t1 = logger.Info()
    t2 = make closure func$1 [t1]  ← t1 is bound to FreeVar[0]
    go t2

func$1:
    t3 = FreeVar[0]  ← This is t1 from above
    t3.Msg("async")
```

When we trace `t3`, we:
1. See it's FreeVar[0]
2. Find the MakeClosure that created func$1
3. See that Bindings[0] = t1
4. Continue tracing from t1

---

## Store Tracking: Following Values Through Memory

### The Challenge

```go
type holder struct {
    event *zerolog.Event
}

func example(ctx context.Context, logger zerolog.Logger) {
    h := holder{event: logger.Info().Ctx(ctx)}
    h.event.Msg("test")  // Is h.event the one with .Ctx()?
}
```

In SSA, struct field access creates separate load/store operations:

```
t1 = logger.Info().Ctx(ctx)
t2 = &h.event          // FieldAddr
*t2 = t1               // Store: write t1 to h.event
...
t3 = &h.event          // FieldAddr (same field)
t4 = *t3               // UnOp: read from h.event
t4.Msg("test")
```

### How Store Tracking Works

When we see `*t3` (a dereference), we:
1. Look for Store instructions that wrote to a matching address
2. `t2` and `t3` both point to `h.event` (same field)
3. Find that `*t2 = t1` stored `t1` at that address
4. Continue tracing from `t1`

This is what `findStoredValue()` and `addressesMatch()` do.

---

## The Strategy Pattern: Why Three Tracers?

### Without Strategy Pattern

We could have written one big function:

```go
func trace(v ssa.Value, expectedType string) bool {
    if expectedType == "Event" {
        // Check Event-specific context sources
        // Then call trace(x, "Logger") or trace(x, "Context")
    } else if expectedType == "Logger" {
        // Check Logger-specific context sources
    } else if expectedType == "Context" {
        // Check Context-specific context sources
    }
    // Handle Phi, FreeVar, etc.
}
```

This is messy - the type-specific logic is all mixed together.

### With Strategy Pattern

Each tracer is a separate struct with focused responsibility:

```go
type ssaTracer interface {
    hasContext(call, callee, recv) (found, delegate, delegateVal)
    continueOnReceiverType(recv) bool
}

type eventTracer struct { ... }   // Knows Event's context sources
type loggerTracer struct { ... }  // Knows Logger's context sources
type contextTracer struct { ... } // Knows Context's context sources
```

Benefits:
1. **Clear separation**: Each tracer only knows about its type
2. **Easy to extend**: Add a new type? Add a new tracer
3. **Testable**: Can test each tracer in isolation
4. **Self-documenting**: Code structure matches problem structure

### How Delegation Works

```go
func (t *eventTracer) hasContext(call, callee, recv) (bool, ssaTracer, ssa.Value) {
    // If we see Logger.Info(), delegate to loggerTracer
    if isLogLevelMethod(callee.Name()) && isLogger(recv.Type()) {
        return false, t.logger, call.Call.Args[0]
        //            ^^^^^^^^ Switch to logger tracer
        //                     ^^^^^^^^^^^^^^^^^ With this value
    }
}
```

The main `traceValue` function handles the delegation:

```go
func traceValue(v ssa.Value, tracer ssaTracer, visited) bool {
    found, delegate, delegateVal := tracer.hasContext(call, callee, recv)
    if delegate != nil {
        return traceValue(delegateVal, delegate, visited)  // Switch tracer!
    }
    // ... continue with current tracer
}
```

---

## Understanding Analyzer Limitations

### False Positives vs False Negatives

In linter terminology:
- **False Positive**: The analyzer reports an error when there isn't one (annoying but safe)
- **False Negative**: The analyzer misses a real error (dangerous - bugs slip through)

Example false positive:
```go
ch <- logger.Info().Ctx(ctx)  // Context IS here
e := <-ch
e.Msg("test")  // But analyzer reports "missing Ctx" - it lost track through the channel
```

Example false negative:
```go
msg := e.Msg      // Extract method
msg("test")       // Analyzer doesn't catch this - should report but doesn't
```

### Why Channels Break Tracking

SSA models channels as "black boxes" - values go in, values come out, but there's no guaranteed relationship:

```
ch <- value1      // Store into channel
ch <- value2      // Another store
x := <-ch         // Which value? Could be either!
```

The analyzer can't prove which specific value `x` receives, so it conservatively reports an error. This is a deliberate design choice - false positives are annoying but safe.

### Embedded Struct Field Promotion

Go's embedded fields create "promoted" methods. When you call `h.Msg()` on:

```go
type embeddedHolder struct {
    *zerolog.Event  // Embedded field
}
```

Go compiles this to something like:
```go
h.Event.Msg()  // Implicit field access
```

In SSA, the implicit field access creates a different value chain than explicit access. The analyzer can trace `h.Event.Msg()` but not `h.Msg()` because the method receiver binding works differently.

### Why Some Cross-Function Patterns Work

The key is **where the terminator call happens**:

**Works** (analyzer catches):
```go
func helper() *Event { return logger.Info() }
func main(ctx context.Context) {
    e := helper()
    e.Msg("test")  // Terminator is HERE, in main() which has ctx
}
```
→ Analyzer sees `Msg` in a function with context, traces back, finds no `.Ctx()`.

**Doesn't work** (false negative with ctx, false positive without):
```go
func helper(ctx context.Context) *Event {
    return logger.Info().Ctx(ctx)  // Context added HERE
}
func main(ctx context.Context) {
    e := helper(ctx)
    e.Msg("test")  // Analyzer can't see ctx was added inside helper
}
```
→ Cross-function tracking would need interprocedural analysis (expensive, not implemented).

### Method Values vs Method Calls

In Go, you can extract a method as a value:

```go
e := logger.Info()
msg := e.Msg        // Method value - msg is type func(string)
msg("test")         // Calling the function
```

This is **very different** in SSA from:
```go
e := logger.Info()
e.Msg("test")       // Direct method call
```

The method value `msg` is a closure that captures `e` in a special way. The analyzer's tracing doesn't follow this particular capture pattern, so it loses track.

---

## Summary

1. **SSA** gives each value a unique identity, enabling backward tracing
2. **Three tracers** handle zerolog's type hierarchy (Event, Logger, Context)
3. **Phi nodes** represent conditional assignments - we require ALL branches to have context
4. **FreeVars** are closure captures - we trace through MakeClosure bindings
5. **Store tracking** follows values through struct field assignments
6. **Strategy Pattern** keeps the code organized by separating type-specific logic

### Known Limitations

| Pattern | Type | Why |
|---------|------|-----|
| Channel send/receive | False Positive | Channels are "black boxes" in SSA |
| Embedded field promotion | False Positive | Implicit field access creates different value chain |
| Closure-modified captured var | False Positive | Can't track writes in called closures |
| sync.Pool | False Positive | Get/Put creates opaque value flow |
| Phi node with nil | False Positive | SSA models all branches even if unreachable |
| Method values | False Negative | Method value capture not traced |
| Cross-function ctx | False Negative | No interprocedural analysis |

---

## FAQ: Common Questions

### Why does sync.Pool cause false positives?

```go
pool := &sync.Pool{New: func() interface{} { return logger.Info().Ctx(ctx) }}
e := pool.Get().(*zerolog.Event)
e.Msg("test")  // False positive!
```

`sync.Pool.Get()` returns `interface{}`. Even though our `New` function creates Events with context, SSA can't prove what `Get()` returns. The pool might return a previously-`Put()` value from a different goroutine.

### Why does `if true` still create a Phi node problem?

```go
var e *zerolog.Event
if true {
    e = logger.Info().Ctx(ctx)
}
e.Msg("test")  // False positive!
```

Go's SSA doesn't do constant folding for control flow. Even `if true` creates two branches:
- Branch 1: `e = logger.Info().Ctx(ctx)`
- Branch 2: `e = nil` (implicit zero value)

The Phi node merges both, and since branch 2 has no context, the analyzer reports.

### Why aren't struct fields tracked as context sources?

```go
type Handler struct {
    ctx    context.Context
    logger zerolog.Logger
}
func (h *Handler) Log() {
    h.logger.Info().Msg("test")  // Not reported!
}
```

This is intentional. The analyzer only tracks **function parameters** as context sources, not struct fields. Reasons:

1. **Scope clarity**: Parameters are clearly "available" in the function
2. **Avoid false positives**: `h.ctx` might be stale or from a different request
3. **Encourage explicit passing**: `func (h *Handler) Log(ctx context.Context)` is clearer

### Why do function literal arguments work?

```go
doSomething(func() {
    logger.Info().Msg("test")  // Correctly reported!
})
```

The function literal is **defined** inside a function that has `ctx`. Even though it's passed as an argument, SSA sees the closure captures context from its enclosing scope. The analyzer correctly identifies this pattern.

### Why does the goroutine checker require `_ = ctx`?

```go
func handler(ctx context.Context) {
    // Bad: outer goroutine doesn't reference ctx directly
    go func() {
        go func() {
            doSomething(ctx)  // inner uses ctx
        }()
    }()

    // Good: outer goroutine explicitly acknowledges ctx
    go func() {
        _ = ctx  // "I know ctx exists and I'm propagating it"
        go func() {
            doSomething(ctx)
        }()
    }()
}
```

This is **intentional design**. Each goroutine should explicitly acknowledge context propagation:

1. **Visibility**: When reading code, `_ = ctx` makes it immediately clear that the goroutine is context-aware
2. **Intentionality**: It shows the programmer consciously considered context propagation
3. **Consistency**: Every level of goroutine nesting has the same rule
4. **Refactoring safety**: If the inner goroutine is later removed, the outer still properly propagates

Think of `_ = ctx` as a "context checkpoint" - a declaration that "yes, context flows through here."

When you see an error like:
```
zerolog call chain missing .Ctx(ctx)
```

The analyzer found a path from your `.Msg()` call back to where the Event was created, and that path never included a `.Ctx()` call. But remember - if your code uses patterns from the "False Positive" column above, the error might be spurious!
