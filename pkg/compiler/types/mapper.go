package types

import (
	"fmt"
	"go/types"
	"strings"
)

// MapGoTypeToMyGo converts a Go type (from go/types) to a MyGo type string.
func MapGoTypeToMyGo(t types.Type) string {
	switch t := t.(type) {
	case *types.Basic:
		return mapBasicType(t)
	case *types.Pointer:
		elemType := MapGoTypeToMyGo(t.Elem())
		return "*" + elemType
	case *types.Slice:
		elemType := MapGoTypeToMyGo(t.Elem())
		return elemType + "[]"
	case *types.Array:
		// MyGo doesn't fully distinguish arrays vs slices in syntax yet,
		// but let's map to T[] for now or fixed size syntax if we have one.
		// Let's assume T[] for compatibility.
		elemType := MapGoTypeToMyGo(t.Elem())
		return elemType + "[]"
	case *types.Map:
		keyType := MapGoTypeToMyGo(t.Key())
		valType := MapGoTypeToMyGo(t.Elem())
		return fmt.Sprintf("Map<%s, %s>", keyType, valType)
	case *types.Signature:
		return mapFunctionType(t)
	case *types.Named:
		// For named types (structs, interfaces, aliases), we need the qualified name.
		// e.g. "time.Time" or "fmt.Formatter"
		obj := t.Obj()
		if obj.Pkg() != nil {
			return obj.Pkg().Name() + "." + obj.Name()
		}
		return obj.Name()
	case *types.Interface:
		// Empty interface is "any"
		if t.NumMethods() == 0 {
			return "any"
		}
		// Non-empty interface: treat as "any" or "trait" depending on context.
		// For type checking, "any" is safest fallback for now unless we reconstruct the trait.
		return "any"
	default:
		return "any"
	}
}

func mapBasicType(t *types.Basic) string {
	switch t.Kind() {
	case types.Bool:
		return "bool"
	case types.Int, types.Int8, types.Int16, types.Int32, types.Int64:
		return "int" // MyGo 'int' is 64-bit usually, or platform dependent.
	case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64:
		if t.Kind() == types.Uint8 {
			return "byte"
		}
		return "int" // Simplify unsigned to int for now, or add uint support
	case types.Float32, types.Float64:
		return "float"
	case types.String:
		return "string"
	default:
		return "any"
	}
}

func mapFunctionType(t *types.Signature) string {
	var params []string
	if t.Params() != nil {
		for i := 0; i < t.Params().Len(); i++ {
			param := t.Params().At(i)
			pType := MapGoTypeToMyGo(param.Type())
			if t.Variadic() && i == t.Params().Len()-1 {
				// Variadic arg in Go: ...T
				// t.Params().At(i).Type() is actually []T for variadic args in go/types usually?
				// Actually for Variadic, the last param type is indeed a slice.
				// MyGo doesn't have variadic syntax yet? Or maybe we map to T[]
				// Let's keep it as T[] for now, but semantically it's variadic.
				// If we want to support calling it, we might need to know it's variadic.
				// But type string is just for checking.
			}
			params = append(params, pType)
		}
	}

	var results []string
	if t.Results() != nil {
		for i := 0; i < t.Results().Len(); i++ {
			res := t.Results().At(i)
			results = append(results, MapGoTypeToMyGo(res.Type()))
		}
	}

	// MyGo syntax: fn(A, B): C
	// If multiple returns: fn(A, B): (C, D)
	paramStr := strings.Join(params, ", ")

	if len(results) == 0 {
		return fmt.Sprintf("fn(%s)", paramStr)
	}
	if len(results) == 1 {
		return fmt.Sprintf("fn(%s): %s", paramStr, results[0])
	}
	return fmt.Sprintf("fn(%s): (%s)", paramStr, strings.Join(results, ", "))
}
