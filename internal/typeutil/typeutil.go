package typeutil

import (
	"go/types"
)

const contextPkgPath = "context"

// IsContextType checks if the type is context.Context.
func IsContextType(t types.Type) bool {
	t = UnwrapPointer(t)

	named, ok := t.(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}

	return obj.Pkg().Path() == contextPkgPath && obj.Name() == "Context"
}

// MatchPkgPath checks if pkgPath matches targetPkg, allowing version suffixes.
// For example, "github.com/pkg/v2" matches "github.com/pkg".
func MatchPkgPath(pkgPath, targetPkg string) bool {
	if pkgPath == targetPkg {
		return true
	}
	prefix := targetPkg + "/v"
	if len(pkgPath) <= len(prefix) {
		return false
	}
	if pkgPath[:len(prefix)] != prefix {
		return false
	}
	rest := pkgPath[len(prefix):]
	return len(rest) > 0 && rest[0] >= '0' && rest[0] <= '9'
}

// UnwrapPointer recursively unwraps all pointer layers.
//
// This is critical for SSA-based carrier type matching. When a closure captures
// a pointer variable, SSA represents it with an additional level of indirection:
//
//	func handler(ctx *CarrierType) {
//	    go func() {
//	        _ = ctx  // SSA FreeVars: **CarrierType (not *CarrierType)
//	    }()
//	}
//
// Therefore, we must unwrap ALL pointer layers to match against the registered
// carrier type (CarrierType, no pointer). Single-layer unwrapping would leave
// *CarrierType, which wouldn't match.
func UnwrapPointer(t types.Type) types.Type {
	for {
		ptr, ok := t.(*types.Pointer)
		if !ok {
			return t
		}
		t = ptr.Elem()
	}
}
