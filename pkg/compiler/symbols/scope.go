package symbols

import (
	"strings"
)

type Scope struct {
	Name    string
	Symbols map[string]*Symbol
	Parent  *Scope
}

func NewScope(name string, parent *Scope) *Scope {
	return &Scope{Name: name, Symbols: make(map[string]*Symbol), Parent: parent}
}

func (s *Scope) Define(mygoName, goName string, kind SymbolKind, typeStr string) *Symbol {
	return s.DefineWithMeta(mygoName, goName, kind, typeStr, VisibilityPackage, "", "")
}

func (s *Scope) DefineWithMeta(mygoName, goName string, kind SymbolKind, typeStr string, visibility Visibility, packageName, filePath string) *Symbol {
	sym := &Symbol{
		MyGoName:             mygoName,
		GoName:               goName,
		Kind:                 kind,
		Type:                 typeStr,
		Visibility:           visibility,
		PackageName:          packageName,
		FilePath:             filePath,
		GenericParams:        []GenericParamMeta{},
		TraitMethods:         make(map[string]interface{}),
		AbstractTraitMethods: make(map[string]interface{}),
		ConcreteTraitMethods: make(map[string]interface{}),
		Methods:              make(map[string]interface{}),
		Variants:             make(map[string]*Symbol),
		Fields:               make(map[string]string),
		BoundTraits:          make(map[string]struct{}),
	}
	s.Symbols[mygoName] = sym
	return sym
}

// DefineSymbol adds a pre-constructed symbol to the scope
func (s *Scope) DefineSymbol(sym *Symbol) {
	if s.Symbols == nil {
		s.Symbols = make(map[string]*Symbol)
	}
	s.Symbols[sym.MyGoName] = sym
}

func (s *Scope) Resolve(name string) *Symbol {
	if sym, ok := s.Symbols[name]; ok {
		return sym
	}
	if s.Parent != nil {
		return s.Parent.Resolve(name)
	}
	return nil
}

func (s *Scope) ResolveByGoName(goName string) *Symbol {
	for _, sym := range s.Symbols {
		if sym.GoName == goName {
			return sym
		}
	}
	if s.Parent != nil {
		return s.Parent.ResolveByGoName(goName)
	}
	return nil
}

// ResolveQualified resolves a symbol by its qualified name (e.g., "pkg.Symbol").
// It supports simple names and one level of qualification (package access).
func (s *Scope) ResolveQualified(name string) *Symbol {
	// fmt.Printf("DEBUG: Scope %s resolving qualified %s\n", s.Name, name)
	parts := strings.Split(name, ".")
	if len(parts) == 1 {
		return s.Resolve(name)
	}

	// Resolve the package symbol
	pkgName := parts[0]
	pkgSym := s.Resolve(pkgName)
	if pkgSym == nil || pkgSym.Kind != KindPackage || pkgSym.ImportedScope == nil {
		return nil
	}

	// Resolve the member in the package scope
	memberName := parts[1]
	// TODO: Support nested packages if needed (parts[1:])
	return pkgSym.ImportedScope.Resolve(memberName)
}

// Merge merges symbols from another scope into this scope.
// It returns an error if a symbol with the same name already exists.
func (s *Scope) Merge(other *Scope) error {
	for name, sym := range other.Symbols {
		if _, exists := s.Symbols[name]; exists {
			return &SymbolConflictError{Name: name}
		}
		s.Symbols[name] = sym
	}
	return nil
}

type SymbolConflictError struct {
	Name string
}

func (e *SymbolConflictError) Error() string {
	return "symbol '" + e.Name + "' already exists"
}
