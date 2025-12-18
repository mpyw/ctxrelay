// Package ignore handles //goroutinectx:ignore directives.
package ignore

import (
	"go/ast"
	"go/token"
	"strings"
)

// Map tracks line numbers that have ignore comments.
type Map map[int]struct{}

// Build scans a file for ignore comments and returns a map.
func Build(fset *token.FileSet, file *ast.File) Map {
	m := make(Map)

	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if isIgnoreComment(c.Text) {
				line := fset.Position(c.Pos()).Line
				m[line] = struct{}{}
			}
		}
	}

	return m
}

// isIgnoreComment checks if a comment is an ignore directive.
// Supports both "//goroutinectx:ignore" and "// goroutinectx:ignore".
func isIgnoreComment(text string) bool {
	text = strings.TrimPrefix(text, "//")
	text = strings.TrimSpace(text)

	return strings.HasPrefix(text, "goroutinectx:ignore")
}

// ShouldIgnore returns true if the given line should be ignored.
// It checks if the same line or the previous line has an ignore comment.
func (m Map) ShouldIgnore(line int) bool {
	_, onSameLine := m[line]
	_, onPrevLine := m[line-1]

	return onSameLine || onPrevLine
}
