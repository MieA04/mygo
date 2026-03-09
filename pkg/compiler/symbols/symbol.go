package symbols

type SymbolKind string

const (
	KindEnum    SymbolKind = "enum"
	KindStruct  SymbolKind = "struct"
	KindFunc    SymbolKind = "func"
	KindVar     SymbolKind = "var"
	KindTrait   SymbolKind = "trait"
	KindPackage SymbolKind = "package"
)

type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityPackage Visibility = "package"
	VisibilityPrivate Visibility = "private"
)

type GenericParamMeta struct {
	Name           string
	ConstraintType string
	ConstraintMyGo string // Original MyGo type name for semantic checks
	DefaultType    string
	DefaultMyGo    string // Original MyGo type name for semantic checks
}

type Symbol struct {
	MyGoName             string
	GoName               string
	Kind                 SymbolKind
	Type                 string
	Visibility           Visibility
	PackageName          string
	FilePath             string
	GenericParams        []GenericParamMeta
	TraitMethods         map[string]interface{}
	AbstractTraitMethods map[string]interface{}
	ConcreteTraitMethods map[string]interface{}
	Methods              map[string]interface{}
	Variants             map[string]*Symbol // For Enums: Variant Name -> Variant Symbol
	Fields               map[string]string  // For Structs/Variants: Field Name -> Field Type
	BoundTraits          map[string]struct{}
	ImportedScope        *Scope // For KindPackage: The scope of the imported package
}
