// Command ctxrelay is a linter that checks for proper context propagation in Go code.
package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/mpyw/ctxrelay/pkg/analyzer"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
