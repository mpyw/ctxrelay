package carrier

import (
	"github.com/labstack/echo/v4"
	"log/slog"
)

// Test that echo.Context is treated as context carrier when configured.

func badEchoHandler(c echo.Context) {
	slog.Info("missing context") // want `use slog.InfoContext with context "c" instead of slog.Info`
}

func goodEchoHandler(c echo.Context) {
	// When using echo.Context, the user would typically extract context.Context
	// and pass it. But the point is that ctxrelay recognizes `c` as a context carrier.
	_ = c // use context carrier
}

func badGoroutineInEchoHandler(c echo.Context) {
	go func() { // want `goroutine does not propagate context "c"`
		// Note: slog.Info inside is not reported because there's no context in goroutine scope
		println("in goroutine")
	}()
}

func goodGoroutineInEchoHandler(c echo.Context) {
	go func() {
		_ = c // captures echo.Context (note: still reports slog since echo.Context is captured)
		println("in goroutine")
	}()
}

// ===== MULTIPLE CONTEXT/CARRIER COMBINATIONS =====

// C01: Both context.Context and carrier - uses carrier (good)
func goodMixedCtxAndCarrierUsesCarrier(c echo.Context, prefix string) {
	go func() {
		_ = c // uses carrier
	}()
}

// C02: Both context.Context and carrier - uses neither (bad - reports carrier name)
func badMixedCtxAndCarrierUsesNeither(c echo.Context, prefix string) {
	go func() { // want `goroutine does not propagate context "c"`
		_ = prefix
	}()
}

// C03: Carrier as second param - uses it (good)
func goodCarrierAsSecondParam(prefix string, c echo.Context) {
	go func() {
		_ = c
	}()
}

// C04: Carrier as second param - doesn't use it (bad)
func badCarrierAsSecondParam(prefix string, c echo.Context) {
	go func() { // want `goroutine does not propagate context "c"`
		_ = prefix
	}()
}
