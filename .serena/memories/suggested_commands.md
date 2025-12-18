# goroutinectx - Suggested Commands

## Testing
```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test pattern
go test -run TestAnalyzer ./...
```

## Building
```bash
# Build the linter
go build -o bin/goroutinectx ./cmd/goroutinectx

# Install globally
go install ./cmd/goroutinectx
```

## Using the Linter
```bash
# Standalone usage
goroutinectx ./...

# With go vet
go vet -vettool=$(which goroutinectx) ./...

# With custom flags
goroutinectx -goroutine-deriver=pkg.Func -context-carriers=pkg.Type ./...
```

## Development
```bash
# Format code
gofmt -w .

# Lint (if golangci-lint is installed)
golangci-lint run

# Show module dependencies
go mod graph
```

## Test Pattern Verification
```bash
# Verify test patterns (Perl script)
perl scripts/verify-test-patterns.pl
```
