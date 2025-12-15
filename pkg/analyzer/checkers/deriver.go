package checkers

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// DeriveFuncSpec holds parsed components of a derive function specification.
type DeriveFuncSpec struct {
	PkgPath  string
	TypeName string // empty for package-level functions
	FuncName string
}

// DeriveMatcher provides OR/AND matching for derive function specifications.
// The check passes if ANY group is fully satisfied (OR semantics).
// A group is satisfied if ALL functions in that group are called (AND semantics).
type DeriveMatcher struct {
	OrGroups [][]DeriveFuncSpec
	Original string
}

// NewDeriveMatcher creates a DeriveMatcher from a derive function string.
// The deriveFuncsStr supports OR (comma) and AND (plus) operators.
// Format: "pkg/path.Func" or "pkg/path.Type.Method".
func NewDeriveMatcher(deriveFuncsStr string) *DeriveMatcher {
	m := &DeriveMatcher{
		Original: deriveFuncsStr,
	}

	// Split by comma first (OR groups)
	for orPart := range strings.SplitSeq(deriveFuncsStr, ",") {
		orPart = strings.TrimSpace(orPart)
		if orPart == "" {
			continue
		}

		// Split by plus (AND within group)
		var andGroup []DeriveFuncSpec
		for andPart := range strings.SplitSeq(orPart, "+") {
			andPart = strings.TrimSpace(andPart)
			if andPart == "" {
				continue
			}
			spec := ParseDeriveFunc(andPart)
			andGroup = append(andGroup, spec)
		}

		if len(andGroup) > 0 {
			m.OrGroups = append(m.OrGroups, andGroup)
		}
	}

	return m
}

// ParseDeriveFunc parses a single derive function string into components.
// Format: "pkg/path.Func" or "pkg/path.Type.Method".
func ParseDeriveFunc(s string) DeriveFuncSpec {
	spec := DeriveFuncSpec{}

	lastDot := strings.LastIndex(s, ".")
	if lastDot == -1 {
		spec.FuncName = s
		return spec
	}

	spec.FuncName = s[lastDot+1:]
	prefix := s[:lastDot]

	// Check if there's another dot (indicating Type.Method)
	// Type names start with uppercase in Go.
	secondLastDot := strings.LastIndex(prefix, ".")
	if secondLastDot != -1 {
		potentialTypeName := prefix[secondLastDot+1:]
		if len(potentialTypeName) > 0 && potentialTypeName[0] >= 'A' && potentialTypeName[0] <= 'Z' {
			spec.PkgPath = prefix[:secondLastDot]
			spec.TypeName = potentialTypeName
		} else {
			spec.PkgPath = prefix
			spec.TypeName = ""
		}
	} else {
		spec.PkgPath = prefix
		spec.TypeName = ""
	}

	return spec
}

// SatisfiesAnyGroup checks if the AST node satisfies ANY of the OR groups.
func (m *DeriveMatcher) SatisfiesAnyGroup(pass *analysis.Pass, node ast.Node) bool {
	calledFuncs := CollectCalledFuncs(pass, node)

	for _, andGroup := range m.OrGroups {
		if groupSatisfied(calledFuncs, andGroup) {
			return true
		}
	}
	return false
}

// IsEmpty returns true if no derive functions are configured.
func (m *DeriveMatcher) IsEmpty() bool {
	return len(m.OrGroups) == 0
}

// CollectCalledFuncs collects all types.Func that are called within the node.
// It does NOT traverse into nested function literals.
func CollectCalledFuncs(pass *analysis.Pass, node ast.Node) []*types.Func {
	var funcs []*types.Func

	ast.Inspect(node, func(n ast.Node) bool {
		// Don't traverse into nested function literals
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if fn := ExtractFunc(pass, call); fn != nil {
			funcs = append(funcs, fn)
		}
		return true
	})

	return funcs
}

// ExtractFunc extracts the types.Func from a call expression.
func ExtractFunc(pass *analysis.Pass, call *ast.CallExpr) *types.Func {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		obj := pass.TypesInfo.ObjectOf(fun)
		if f, ok := obj.(*types.Func); ok {
			return f
		}

	case *ast.SelectorExpr:
		sel := pass.TypesInfo.Selections[fun]
		if sel != nil {
			if f, ok := sel.Obj().(*types.Func); ok {
				return f
			}
		} else {
			obj := pass.TypesInfo.ObjectOf(fun.Sel)
			if f, ok := obj.(*types.Func); ok {
				return f
			}
		}
	}
	return nil
}

// groupSatisfied checks if ALL specs in the AND group are satisfied.
func groupSatisfied(calledFuncs []*types.Func, andGroup []DeriveFuncSpec) bool {
	for _, spec := range andGroup {
		if !specSatisfied(calledFuncs, spec) {
			return false
		}
	}
	return true
}

// specSatisfied checks if the spec is satisfied by any of the called functions.
func specSatisfied(calledFuncs []*types.Func, spec DeriveFuncSpec) bool {
	for _, fn := range calledFuncs {
		if MatchesSpec(fn, spec) {
			return true
		}
	}
	return false
}

// MatchesSpec checks if a types.Func matches the given derive function spec.
func MatchesSpec(fn *types.Func, spec DeriveFuncSpec) bool {
	if fn.Name() != spec.FuncName {
		return false
	}

	if fn.Pkg() == nil || fn.Pkg().Path() != spec.PkgPath {
		return false
	}

	if spec.TypeName != "" {
		return matchesMethod(fn, spec.TypeName)
	}

	return true
}

// matchesMethod checks if a types.Func is a method on the expected type.
func matchesMethod(fn *types.Func, typeName string) bool {
	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		return false
	}
	recv := sig.Recv()
	if recv == nil {
		return false
	}
	recvType := recv.Type()
	if ptr, ok := recvType.(*types.Pointer); ok {
		recvType = ptr.Elem()
	}
	named, ok := recvType.(*types.Named)
	if !ok {
		return false
	}
	return named.Obj().Name() == typeName
}
