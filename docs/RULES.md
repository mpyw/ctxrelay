# Built-in Rules

## zerolog-ctx

Detects zerolog logging chains that don't include `.Ctx(ctx)` when context is available.

### Bad

```go
func HandleRequest(ctx context.Context, req *Request) {
    log.Info().Msg("handling request")  // Missing .Ctx(ctx)
}
```

### Good

```go
func HandleRequest(ctx context.Context, req *Request) {
    log.Ctx(ctx).Info().Msg("handling request")  // OK
}
```

### Configuration

```yaml
rules:
  zerolog-ctx:
    enabled: true
    # Packages to check (default: github.com/rs/zerolog, github.com/rs/zerolog/log)
    packages:
      - github.com/rs/zerolog
      - github.com/rs/zerolog/log
```

---

## goroutine-ctx

Detects `go` statements that don't properly propagate context.

### Bad

```go
func Process(ctx context.Context) {
    go func() {
        doWork(ctx)  // ctx may be cancelled after parent returns
    }()
}
```

### Good

```go
func Process(ctx context.Context) {
    go func(ctx context.Context) {
        doWork(ctx)
    }(apm.NewGoroutineContext(ctx))  // Properly detached context
}
```

### Configuration

```yaml
rules:
  goroutine-ctx:
    enabled: true
    # Function that creates goroutine-safe context
    context_wrapper: "apm.NewGoroutineContext"
```

---

## errgroup-ctx

Detects `errgroup.Group.Go()` and `TryGo()` calls that don't properly propagate context.

### Bad

```go
func Process(ctx context.Context) error {
    g, _ := errgroup.WithContext(ctx)
    g.Go(func() error {
        return doWork(ctx)  // Should use goroutine-safe context
    })
    return g.Wait()
}
```

### Good

```go
func Process(ctx context.Context) error {
    g, _ := errgroup.WithContext(ctx)
    g.Go(func() error {
        return doWork(apm.NewGoroutineContext(ctx))
    })
    return g.Wait()
}
```

### Configuration

```yaml
rules:
  errgroup-ctx:
    enabled: true
    context_wrapper: "apm.NewGoroutineContext"
```

---

## zap-ctx (Planned)

Detects zap logger usage that doesn't include context.

### Bad

```go
func HandleRequest(ctx context.Context, req *Request) {
    zap.L().Info("handling request")  // Missing context
}
```

### Good

```go
func HandleRequest(ctx context.Context, req *Request) {
    logger := ctxzap.Extract(ctx)
    logger.Info("handling request")
}
```

---

## slog-ctx (Planned)

Detects slog usage that doesn't include context.

### Bad

```go
func HandleRequest(ctx context.Context, req *Request) {
    slog.Info("handling request")  // Missing context
}
```

### Good

```go
func HandleRequest(ctx context.Context, req *Request) {
    slog.InfoContext(ctx, "handling request")
}
```
