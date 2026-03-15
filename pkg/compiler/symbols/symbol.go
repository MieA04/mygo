package symbols

type SymbolKind string

const (
	KindEnum       SymbolKind = "enum"
	KindStruct     SymbolKind = "struct"
	KindFunc       SymbolKind = "func"
	KindVar        SymbolKind = "var"
	KindTrait      SymbolKind = "trait"
	KindPackage    SymbolKind = "package"
	KindAnnotation SymbolKind = "annotation"
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

type FieldSymbol struct {
	Name   string
	Type   string
	Tag    string // Raw tag string, e.g., "json:\"id\""
	Line   int    // 0-based
	Column int    // 0-based
}

type Symbol struct {
	MyGoName             string
	GoName               string
	Kind                 SymbolKind
	Type                 string
	Visibility           Visibility
	PackageName          string
	FilePath             string
	Line                 int // 0-based
	Column               int // 0-based
	GenericParams        []GenericParamMeta
	TraitMethods         map[string]interface{}
	AbstractTraitMethods map[string]interface{}
	ConcreteTraitMethods map[string]interface{}
	Methods              map[string]interface{}
	Variants             map[string]*Symbol      // For Enums: Variant Name -> Variant Symbol
	Fields               []FieldSymbol           // Ordered list of fields
	FieldMap             map[string]*FieldSymbol // Helper map for fast lookups
	BoundTraits          map[string]struct{}
	ImportedScope        *Scope        // For KindPackage: The scope of the imported package
	ASTNode              interface{}   // For Macros: Store the AST node
	Annotations          []*Annotation // List of annotations applied to this symbol
}

type Annotation struct {
	Name string
	Args []string // Store raw string representation or simplified values of arguments
}
