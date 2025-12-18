// Package context provides context-related types and utilities for the analyzer.
package context

import (
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/mpyw/goroutinectx/internal/directives/carrier"
	"github.com/mpyw/goroutinectx/internal/directives/ignore"
)

// CheckContext holds the context for running checks.
type CheckContext struct {
	Pass      *analysis.Pass
	Scope     *Scope
	IgnoreMap ignore.Map
	Carriers  []carrier.Carrier
}

// Reportf reports a diagnostic if the position is not ignored.
func (c *CheckContext) Reportf(pos token.Pos, format string, args ...any) {
	line := c.Pass.Fset.Position(pos).Line
	if c.IgnoreMap.ShouldIgnore(line) {
		return
	}

	c.Pass.Reportf(pos, format, args...)
}
