// Command goroutinectx is a linter that checks goroutine context propagation.
package main

import (
	"github.com/mpyw/goroutinectx"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(goroutinectx.Analyzer)
}
