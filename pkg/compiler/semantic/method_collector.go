package semantic

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/compiler/symbols"
	"github.com/miea04/mygo/pkg/compiler/types"
)

// MethodCollector visits the AST and collects methods from `trait bind` blocks
// into the symbol table. It should run after DeclarationCollector.
type MethodCollector struct {
	*ast.BaseMyGoVisitor
	CurrentScope   *symbols.Scope
	currentPackage string
	currentFile    string
}

func NewMethodCollector(global *symbols.Scope) *MethodCollector {
	return &MethodCollector{BaseMyGoVisitor: &ast.BaseMyGoVisitor{}, CurrentScope: global}
}

func (d *MethodCollector) SetCompilationUnit(packageName, filePath string) {
	d.currentPackage = packageName
	d.currentFile = filePath
}

func (d *MethodCollector) VisitProgram(ctx *ast.ProgramContext) interface{} {
	for _, stmt := range ctx.AllStatement() {
		stmt.Accept(d)
	}
	return nil
}

func (d *MethodCollector) VisitStatement(ctx *ast.StatementContext) interface{} {
	child := ctx.GetChild(0)
	if child == nil {
		return nil
	}
	if tree, ok := child.(antlr.ParseTree); ok {
		return tree.Accept(d)
	}
	return nil
}

func (d *MethodCollector) VisitTraitDecl(ctx *ast.TraitDeclContext) interface{} {
	// Dispatch to specific trait declaration types
	if bindTrait, ok := ctx.GetChild(0).(*ast.BindTraitDeclContext); ok {
		return bindTrait.Accept(d)
	}
	return nil
}

func (d *MethodCollector) VisitBindTraitDecl(ctx *ast.BindTraitDeclContext) interface{} {
	// 1. Resolve target type
	for _, targetCtx := range ctx.AllBindTarget() {
		tCtx := targetCtx.TypeType()
		if tCtx == nil {
			continue
		}

		// Use simple name resolution for now.
		// types.ResolveTypeWithScope handles generic instantiation syntax,
		// but we need the base symbol.
		typeName := tCtx.GetText()
		baseName := types.SplitBaseType(typeName)

		sym := d.CurrentScope.Resolve(baseName)
		if sym == nil {
			sym = d.CurrentScope.ResolveByGoName(baseName)
		}

		if sym == nil {
			// Skip if target type not found (maybe error reporting later in semantic pass)
			continue
		}

		// Only attach methods to Structs or Enums
		if sym.Kind != symbols.KindStruct && sym.Kind != symbols.KindEnum {
			continue
		}

		if sym.Methods == nil {
			sym.Methods = make(map[string]interface{})
		}

		// 2. Collect methods
		for _, item := range ctx.AllTraitBodyItem() {
			if fnCtx := item.TraitFnDecl(); fnCtx != nil {
				fnDecl := fnCtx.(*ast.TraitFnDeclContext)
				name := fnDecl.ID().GetText()

				// Store the AST node as the method definition
				sym.Methods[name] = fnDecl
			}
		}
	}
	return nil
}

// Ignore other declarations
func (d *MethodCollector) VisitStructDecl(ctx *ast.StructDeclContext) interface{} { return nil }
func (d *MethodCollector) VisitEnumDecl(ctx *ast.EnumDeclContext) interface{}     { return nil }
func (d *MethodCollector) VisitFnDecl(ctx *ast.FnDeclContext) interface{}         { return nil }
func (d *MethodCollector) VisitVarDecl(ctx *ast.VarDeclContext) interface{}       { return nil }
