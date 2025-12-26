//go:build go1.25

package goroutinectx_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mpyw/goroutinectx"
)

func TestWaitgroup(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, goroutinectx.Analyzer, "waitgroup")
}

func TestWaitgroupDerive(t *testing.T) {
	testdata := analysistest.TestData()

	deriveFunc := "github.com/my-example-app/telemetry/apm.NewGoroutineContext"
	if err := goroutinectx.Analyzer.Flags.Set("goroutine-deriver", deriveFunc); err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = goroutinectx.Analyzer.Flags.Set("goroutine-deriver", "")
	}()

	analysistest.Run(t, testdata, goroutinectx.Analyzer, "waitgroupderive")
}
