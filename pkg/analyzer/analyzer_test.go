package analyzer_test

import (
	"testing"

	"github.com/mpyw/ctxrelay/pkg/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestZerolog(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzer.Analyzer, "zerolog")
}

func TestSlog(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzer.Analyzer, "slog")
}

func TestGoroutine(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzer.Analyzer, "goroutine")
}

func TestErrgroup(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzer.Analyzer, "errgroup")
}

func TestWaitgroup(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzer.Analyzer, "waitgroup")
}

func TestGoroutineDerive(t *testing.T) {
	testdata := analysistest.TestData()
	deriveFunc := "github.com/my-example-app/telemetry/apm.NewGoroutineContext"
	if err := analyzer.Analyzer.Flags.Set("goroutine-deriver", deriveFunc); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = analyzer.Analyzer.Flags.Set("goroutine-deriver", "")
	}()
	analysistest.Run(t, testdata, analyzer.Analyzer, "goroutinederive")
}

func TestGoroutineDeriveAnd(t *testing.T) {
	testdata := analysistest.TestData()
	// AND: all must be called (Transaction.NewGoroutine + NewContext)
	deriveFunc := "github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine+" +
		"github.com/newrelic/go-agent/v3/newrelic.NewContext"
	if err := analyzer.Analyzer.Flags.Set("goroutine-deriver", deriveFunc); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = analyzer.Analyzer.Flags.Set("goroutine-deriver", "")
	}()
	analysistest.Run(t, testdata, analyzer.Analyzer, "goroutinederiveand")
}

func TestGoroutineDeriveMixed(t *testing.T) {
	testdata := analysistest.TestData()
	// Mixed: (Transaction.NewGoroutine AND NewContext) OR apm.NewGoroutineContext
	deriveFunc := "github.com/newrelic/go-agent/v3/newrelic.Transaction.NewGoroutine+" +
		"github.com/newrelic/go-agent/v3/newrelic.NewContext," +
		"github.com/my-example-app/telemetry/apm.NewGoroutineContext"
	if err := analyzer.Analyzer.Flags.Set("goroutine-deriver", deriveFunc); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = analyzer.Analyzer.Flags.Set("goroutine-deriver", "")
	}()
	analysistest.Run(t, testdata, analyzer.Analyzer, "goroutinederivemixed")
}

func TestContextCarriers(t *testing.T) {
	testdata := analysistest.TestData()
	carriers := "github.com/labstack/echo/v4.Context"
	if err := analyzer.Analyzer.Flags.Set("context-carriers", carriers); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = analyzer.Analyzer.Flags.Set("context-carriers", "")
	}()
	analysistest.Run(t, testdata, analyzer.Analyzer, "carrier")
}

func TestGoroutineCreator(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzer.Analyzer, "goroutinecreator")
}

func TestGotask(t *testing.T) {
	testdata := analysistest.TestData()
	deriveFunc := "github.com/my-example-app/telemetry/apm.NewGoroutineContext"
	if err := analyzer.Analyzer.Flags.Set("goroutine-deriver", deriveFunc); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = analyzer.Analyzer.Flags.Set("goroutine-deriver", "")
	}()
	analysistest.Run(t, testdata, analyzer.Analyzer, "gotask")
}
