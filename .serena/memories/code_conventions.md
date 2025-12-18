# goroutinectx - Code Conventions

## Variable Naming
- `ctx` : reserved for `context.Context`
- `cctx` : used for `*CheckContext`
- `pass` : `*analysis.Pass` from go/analysis

## Package Structure
- `internal/` : 内部実装（外部からimport不可）
- `testdata/src/` : analysistest用のテストフィクスチャ

## Test Patterns
- `// want "..."` : 期待される診断メッセージを指定
- `//goroutinectx:ignore` : 警告を抑制するディレクティブ

## Interface Design
- Interface Segregation Principle を適用
- `CallChecker` と `GoStmtChecker` を分離
- 各checkerは必要なメソッドのみ実装

## Error Reporting
```go
cctx.Reportf(node, "message with %s", args)
```

## Type Checking
- `go/types` を使用してtype-safeな解析
- `pass.TypesInfo.ObjectOf()` でshadowingを正しく処理
