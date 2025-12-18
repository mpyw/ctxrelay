// Package carrier provides context carrier type parsing.
package carrier

import (
	"strings"
)

// Carrier represents a type that can carry context.
// Format: "pkg/path.TypeName" (e.g., "github.com/labstack/echo/v4.Context").
type Carrier struct {
	PkgPath  string
	TypeName string
}

// Parse parses a comma-separated list of context carriers.
func Parse(s string) []Carrier {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")

	carriers := make([]Carrier, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		lastDot := strings.LastIndex(part, ".")
		if lastDot == -1 {
			continue // Invalid format
		}

		carriers = append(carriers, Carrier{
			PkgPath:  part[:lastDot],
			TypeName: part[lastDot+1:],
		})
	}

	return carriers
}
