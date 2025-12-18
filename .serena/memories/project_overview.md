# goroutinectx Project Overview

## Purpose
Go言語用のLinterで、goroutineにおけるcontext伝播を検査する。
`context.Context` が関数パラメータで利用可能な場合に、それが適切に下流の呼び出しに渡されているかをチェックする。

## Tech Stack
- Go 1.24.0
- golang.org/x/tools (go/analysis framework)
- SSA (Static Single Assignment) for complex tracing

## Checker Types
| Checker | Analysis | Description |
|---------|----------|-------------|
| goroutine | AST | `go func()` でのcontext使用をチェック |
| errgroup | AST | `errgroup.Group.Go()` closureでのcontext使用 |
| waitgroup | AST | `sync.WaitGroup.Go()` closureでのcontext使用 (Go 1.25+) |
| gotask | AST | gotaskでのderiver関数呼び出しをチェック |
| goroutine-creator | AST | カスタムgoroutine生成関数のチェック |

## Architecture
```
goroutinectx/
├── analyzer.go           # Main analyzer (orchestration)
├── internal/
│   ├── checkers/         # Checker implementations
│   │   ├── checker.go    # Interfaces (CallChecker, GoStmtChecker)
│   │   ├── errgroup/
│   │   ├── goroutine/
│   │   ├── waitgroup/
│   │   └── gotask/
│   ├── context/          # Context tracking utilities
│   └── directives/       # Comment directive handling
└── testdata/             # Test fixtures
```

## Key Interfaces
```go
type CallChecker interface {
    CheckCall(cctx *CheckContext, call *ast.CallExpr)
}

type GoStmtChecker interface {
    CheckGoStmt(cctx *CheckContext, stmt *ast.GoStmt)
}
```

## Design Principles
1. Zero false positives - 誤検知を避ける
2. Type-safe analysis - go/types を使用した正確な検出
3. Nested function support - クロージャを通したcontext追跡
