package checkers

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const contextPkgPath = "context"

// ContextCarrier represents a type that can carry context.
// Format: "pkg/path.TypeName" (e.g., "github.com/labstack/echo/v4.Context").
type ContextCarrier struct {
	PkgPath  string
	TypeName string
}

// ParseContextCarriers parses a comma-separated list of context carriers.
func ParseContextCarriers(s string) []ContextCarrier {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	carriers := make([]ContextCarrier, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		lastDot := strings.LastIndex(part, ".")
		if lastDot == -1 {
			continue // Invalid format
		}
		carriers = append(carriers, ContextCarrier{
			PkgPath:  part[:lastDot],
			TypeName: part[lastDot+1:],
		})
	}
	return carriers
}

// IsNamedType checks if the expression has the given named type.
// It handles pointer types automatically.
func IsNamedType(pass *analysis.Pass, expr ast.Expr, pkgPath, typeName string) bool {
	tv, ok := pass.TypesInfo.Types[expr]
	if !ok {
		return false
	}
	return isNamedTypeFromType(tv.Type, pkgPath, typeName)
}

// isNamedTypeFromType checks if the type matches the given package path and type name.
func isNamedTypeFromType(t types.Type, pkgPath, typeName string) bool {
	t = unwrapPointer(t)

	named, ok := t.(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}

	return obj.Pkg().Path() == pkgPath && obj.Name() == typeName
}

// unwrapPointer returns the element type if t is a pointer, otherwise returns t.
func unwrapPointer(t types.Type) types.Type {
	if ptr, ok := t.(*types.Pointer); ok {
		return ptr.Elem()
	}
	return t
}

// IsContextType checks if the type is context.Context.
func IsContextType(t types.Type) bool {
	return isNamedTypeFromType(t, contextPkgPath, "Context")
}

// IsContextOrCarrierType checks if the type is context.Context or a configured carrier type.
func IsContextOrCarrierType(t types.Type, carriers []ContextCarrier) bool {
	if IsContextType(t) {
		return true
	}
	for _, c := range carriers {
		if isNamedTypeFromType(t, c.PkgPath, c.TypeName) {
			return true
		}
	}
	return false
}

// HasContextOrCarrierParam checks if the function type has a context.Context
// or a context carrier parameter.
func HasContextOrCarrierParam(pass *analysis.Pass, fnType *ast.FuncType, carriers []ContextCarrier) bool {
	if fnType == nil || fnType.Params == nil {
		return false
	}
	for _, field := range fnType.Params.List {
		tv, ok := pass.TypesInfo.Types[field.Type]
		if !ok {
			continue
		}
		if IsContextOrCarrierType(tv.Type, carriers) {
			return true
		}
	}
	return false
}
