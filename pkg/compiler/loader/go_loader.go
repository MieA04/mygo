package loader

import (
	"fmt"
	"go/types"

	"github.com/miea04/mygo/pkg/compiler/core"
	"github.com/miea04/mygo/pkg/compiler/symbols"
	mygotypes "github.com/miea04/mygo/pkg/compiler/types"
	"golang.org/x/tools/go/packages"
)

// LoadGoPackage loads a standard Go package using golang.org/x/tools/go/packages
func LoadGoPackage(importPath string, dir string) (*core.Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:  dir,
	}
	pkgs, err := packages.Load(cfg, importPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load go package %s: %w", importPath, err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("package %s not found", importPath)
	}
	if len(pkgs[0].Errors) > 0 {
		return nil, fmt.Errorf("package %s has errors: %v", importPath, pkgs[0].Errors)
	}

	goPkg := pkgs[0]
	scope := symbols.NewScope(goPkg.Name, nil)

	// Iterate over exported symbols in the Go package
	goScope := goPkg.Types.Scope()
	for _, name := range goScope.Names() {
		obj := goScope.Lookup(name)
		if !obj.Exported() {
			continue
		}

		sym := convertGoSymbol(obj)
		if sym != nil {
			// In MyGo, imported Go symbols are Public
			sym.Visibility = symbols.VisibilityPublic
			sym.PackageName = goPkg.Name
			scope.DefineSymbol(sym)
		}
	}

	return &core.Package{
		Name:       goPkg.Name,
		ImportPath: importPath,
		Scope:      scope,
		// Go packages don't have source files we need to parse directly
		Files:       nil,
		IsGoPackage: true,
	}, nil
}

func convertGoSymbol(obj types.Object) *symbols.Symbol {
	var kind symbols.SymbolKind
	var typeName string

	switch obj.(type) {
	case *types.Func:
		kind = symbols.KindFunc
		typeName = mygotypes.MapGoTypeToMyGo(obj.Type())
	case *types.Var:
		kind = symbols.KindVar
		typeName = mygotypes.MapGoTypeToMyGo(obj.Type())
	case *types.Const:
		kind = symbols.KindVar // Treat const as var for now
		typeName = mygotypes.MapGoTypeToMyGo(obj.Type())
	case *types.TypeName:
		if _, ok := obj.Type().Underlying().(*types.Struct); ok {
			kind = symbols.KindStruct
		} else if _, ok := obj.Type().Underlying().(*types.Interface); ok {
			kind = symbols.KindTrait
		} else {
			kind = symbols.KindStruct // Default to struct for other types like named basic types
		}
		typeName = obj.Name()
	default:
		return nil
	}

	sym := &symbols.Symbol{
		MyGoName:    obj.Name(),
		GoName:      obj.Name(),
		Kind:        kind,
		Type:        typeName,
		Visibility:  symbols.VisibilityPublic,
		PackageName: obj.Pkg().Name(),
	}

	// For Interfaces, extract methods
	if kind == symbols.KindTrait {
		if iface, ok := obj.Type().Underlying().(*types.Interface); ok {
			sym.TraitMethods = make(map[string]interface{})
			for i := 0; i < iface.NumMethods(); i++ {
				m := iface.Method(i)
				// Store the MyGo function type signature string (e.g. "fn(int): string")
				sym.TraitMethods[m.Name()] = mygotypes.MapGoTypeToMyGo(m.Type())
			}
		}
	}

	return sym
}
