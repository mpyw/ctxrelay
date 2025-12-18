// Package carrier contains test fixtures for context carrier recognition.
// Tests that echo.Context is treated as context carrier when configured.
package carrier

import (
	"github.com/labstack/echo/v4"
	"log/slog"
)

// Echo handler without context usage
// echo.Context should be treated as context carrier
func badEchoHandler(c echo.Context) {
	// Note: slog checker has been removed (delegated to sloglint)
	// This test now only verifies echo.Context is recognized as a context carrier
	slog.Info("missing context")
}

// Echo handler with context usage
// echo.Context used properly
func goodEchoHandler(c echo.Context) {
	// When using echo.Context, the user would typically extract context.Context
	// and pass it. But the point is that goroutinectx recognizes `c` as a context carrier.
	_ = c // use context carrier
}

// Goroutine in echo handler without carrier usage
// Goroutine does not capture echo.Context carrier
func badGoroutineInEchoHandler(c echo.Context) {
	go func() { // want `goroutine does not propagate context "c"`
		// Note: slog checker has been removed (delegated to sloglint)
		println("in goroutine")
	}()
}

// Goroutine in echo handler with carrier usage
// Goroutine properly captures echo.Context carrier
func goodGoroutineInEchoHandler(c echo.Context) {
	go func() {
		_ = c // captures echo.Context
		println("in goroutine")
	}()
}

// ===== MULTIPLE CONTEXT/CARRIER COMBINATIONS =====

// Mixed ctx and carrier - uses carrier
// Function has both context.Context and carrier, uses carrier
func goodMixedCtxAndCarrierUsesCarrier(c echo.Context, prefix string) {
	go func() {
		_ = c // uses carrier
	}()
}

// Mixed ctx and carrier - uses neither
// Function has both context.Context and carrier, uses neither
func badMixedCtxAndCarrierUsesNeither(c echo.Context, prefix string) {
	go func() { // want `goroutine does not propagate context "c"`
		_ = prefix
	}()
}

// Carrier as second param - uses it
// Carrier is second parameter and is properly used
func goodCarrierAsSecondParam(prefix string, c echo.Context) {
	go func() {
		_ = c
	}()
}

// Carrier as second param - doesn't use it
// Carrier is second parameter but not used
func badCarrierAsSecondParam(prefix string, c echo.Context) {
	go func() { // want `goroutine does not propagate context "c"`
		_ = prefix
	}()
}
