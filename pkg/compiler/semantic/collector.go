package semantic

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/compiler/symbols"
	"github.com/miea04/mygo/pkg/compiler/types"
)

// DeclarationCollector visits the AST and collects top-level declarations
// (structs, enums, traits, functions) into the symbol table.
// It skips function bodies and variable initializers.
type DeclarationCollector struct {
	*ast.BaseMyGoVisitor
	CurrentScope   *symbols.Scope
	currentPackage string
	currentFile    string
}

func NewDeclarationCollector(global *symbols.Scope) *DeclarationCollector {
	d := &DeclarationCollector{BaseMyGoVisitor: &ast.BaseMyGoVisitor{}, CurrentScope: global}
	var _ ast.MyGoVisitor = d
	return d
}

func (d *DeclarationCollector) SetCompilationUnit(packageName, filePath string) {
	d.currentPackage = packageName
	d.currentFile = filePath
}

func (d *DeclarationCollector) collectAnnotations(ctxs []ast.IAnnotationUsageContext) []*symbols.Annotation {
	var anns []*symbols.Annotation
	for _, ctx := range ctxs {
		if c, ok := ctx.(*ast.AnnotationUsageContext); ok {
			annName := c.ID().GetText()
			// fmt.Printf("DEBUG: Found annotation usage: %s\n", annName)
			ann := &symbols.Annotation{
				Name: annName,
			}
			if c.ExprList() != nil {
				for _, expr := range c.ExprList().AllExpr() {
					ann.Args = append(ann.Args, expr.GetText())
				}
			}
			anns = append(anns, ann)
		}
	}
	return anns
}

func (d *DeclarationCollector) defineSymbol(mygoName, goName string, kind symbols.SymbolKind, typeStr string, modCtx ast.IModifierContext, annotations []*symbols.Annotation) *symbols.Symbol {
	sym := d.CurrentScope.DefineWithMeta(
		mygoName,
		goName,
		kind,
		typeStr,
		types.ParseVisibility(modCtx),
		d.currentPackage,
		d.currentFile,
	)
	sym.Annotations = annotations
	for _, ann := range annotations {
		d.CurrentScope.AddAnnotation(ann.Name, sym)
	}
	return sym
}

func (d *DeclarationCollector) VisitProgram(ctx *ast.ProgramContext) interface{} {
	for _, child := range ctx.GetChildren() {
		if tree, ok := child.(antlr.ParseTree); ok {
			tree.Accept(d)
		}
	}
	return nil
}

func (d *DeclarationCollector) VisitAnnotationDecl(ctx *ast.AnnotationDeclContext) interface{} {
	annName := ctx.ID().GetText()
	fmt.Printf("DEBUG: DeclarationCollector visiting annotation: %s\n", annName)
	// Annotations are currently package-private by default as per grammar
	sym := d.defineSymbol(annName, annName, symbols.KindAnnotation, "annotation", nil, nil)
	sym.ASTNode = ctx
	// Simplified rule has no target type
	sym.Type = "fn" // Default to fn for testing
	return nil
}

func (d *DeclarationCollector) VisitStatement(ctx *ast.StatementContext) interface{} {
	child := ctx.GetChild(0)
	if child == nil {
		return nil
	}

	if s, ok := child.(*ast.StructDeclContext); ok {
		return d.VisitStructDecl(s)
	}
	if e, ok := child.(*ast.EnumDeclContext); ok {
		return d.VisitEnumDecl(e)
	}
	if f, ok := child.(*ast.FnDeclContext); ok {
		return d.VisitFnDecl(f)
	}
	if t, ok := child.(*ast.TraitDeclContext); ok {
		return d.VisitTraitDecl(t)
	}

	if tree, ok := child.(antlr.ParseTree); ok {
		return tree.Accept(d)
	}
	return nil
}

func (d *DeclarationCollector) VisitStructDecl(ctx *ast.StructDeclContext) interface{} {
	anns := d.collectAnnotations(ctx.AllAnnotationUsage())
	sym := d.defineSymbol(ctx.ID().GetText(), types.FormatVisibility(ctx.ID().GetText(), ctx.Modifier()), symbols.KindStruct, "struct", ctx.Modifier(), anns)
	baseMeta := types.ExtractGenericParamMeta(ctx.TypeParams(), d.CurrentScope)
	whereMeta := types.ExtractWhereConstraintMeta(ctx.WhereClause(), d.CurrentScope)
	mergedMeta, issues := types.MergeGenericConstraints(baseMeta, whereMeta)
	for _, issue := range issues {
		fmt.Printf("Type Error: %s\n", issue.Error())
	}
	sym.GenericParams = mergedMeta
	sym.Fields = []symbols.FieldSymbol{}
	sym.FieldMap = make(map[string]*symbols.FieldSymbol)

	for _, field := range ctx.AllStructField() {
		fieldName := field.ID().GetText()
		fieldType := types.ResolveTypeWithScope(field.TypeType().GetText(), d.CurrentScope)

		tag := ""
		if field.STRING() != nil {
			tag = field.STRING().GetText()
		}

		fieldSym := symbols.FieldSymbol{
			Name: fieldName,
			Type: fieldType,
			Tag:  tag,
		}
		sym.Fields = append(sym.Fields, fieldSym)
		sym.FieldMap[fieldName] = &sym.Fields[len(sym.Fields)-1]
	}
	return nil
}

func (d *DeclarationCollector) VisitEnumDecl(ctx *ast.EnumDeclContext) interface{} {
	enumName := ctx.ID().GetText()
	sym := d.defineSymbol(enumName, types.FormatVisibility(enumName, ctx.Modifier()), symbols.KindEnum, "enum", ctx.Modifier(), nil)
	baseMeta := types.ExtractGenericParamMeta(ctx.TypeParams(), d.CurrentScope)
	whereMeta := types.ExtractWhereConstraintMeta(ctx.WhereClause(), d.CurrentScope)
	mergedMeta, issues := types.MergeGenericConstraints(baseMeta, whereMeta)
	for _, issue := range issues {
		fmt.Printf("Type Error: %s\n", issue.Error())
	}
	sym.GenericParams = mergedMeta

	// Process Enum Variants
	for _, member := range ctx.AllEnumVariant() {
		variantName := member.ID().GetText()
		variantSym := &symbols.Symbol{
			MyGoName:      variantName,
			GoName:        variantName,        // Will be prefixed during generation
			Kind:          symbols.KindStruct, // Treat variant as a struct-like entity
			Type:          enumName,           // It belongs to this Enum type
			GenericParams: sym.GenericParams,  // Inherit generics
			Fields:        []symbols.FieldSymbol{},
			FieldMap:      make(map[string]*symbols.FieldSymbol),
		}

		// Add fields if present (e.g. Some(T))
		if member.TypeList() != nil {
			for i, typeCtx := range member.TypeList().(*ast.TypeListContext).AllTypeType() {
				fieldName := fmt.Sprintf("Item%d", i+1)
				fieldType := types.ResolveTypeWithScope(typeCtx.GetText(), d.CurrentScope)

				fieldSym := symbols.FieldSymbol{
					Name: fieldName,
					Type: fieldType,
					Tag:  "",
				}
				variantSym.Fields = append(variantSym.Fields, fieldSym)
				variantSym.FieldMap[fieldName] = &variantSym.Fields[len(variantSym.Fields)-1]
			}
		}

		sym.Variants[variantName] = variantSym
	}
	return nil
}

func (d *DeclarationCollector) VisitPureTraitDecl(ctx *ast.PureTraitDeclContext) interface{} {
	traitName := ctx.ID().GetText()
	sym := d.defineSymbol(traitName, types.FormatVisibility(traitName, ctx.Modifier()), symbols.KindTrait, "trait", ctx.Modifier(), nil)
	baseMeta := types.ExtractGenericParamMeta(ctx.TypeParams(), d.CurrentScope)
	whereMeta := types.ExtractWhereConstraintMeta(ctx.WhereClause(), d.CurrentScope)
	mergedMeta, issues := types.MergeGenericConstraints(baseMeta, whereMeta)
	for _, issue := range issues {
		fmt.Printf("Type Error: %s\n", issue.Error())
	}
	sym.GenericParams = mergedMeta
	for _, fnCtx := range ctx.AllTraitFnDecl() {
		fnDecl := fnCtx.(*ast.TraitFnDeclContext)
		fnName := fnDecl.ID().GetText()
		sym.TraitMethods[fnName] = fnDecl
		if fnDecl.Block() == nil {
			sym.AbstractTraitMethods[fnName] = fnDecl
		} else {
			sym.ConcreteTraitMethods[fnName] = fnDecl
		}
	}
	return nil
}

func (d *DeclarationCollector) VisitFnDecl(ctx *ast.FnDeclContext) interface{} {
	fnName := ctx.ID().GetText()
	returnType := "void"
	if ctx.TypeType() != nil {
		returnType = types.ResolveTypeWithScope(ctx.TypeType().GetText(), d.CurrentScope)
	}
	anns := d.collectAnnotations(ctx.AllAnnotationUsage())
	sym := d.defineSymbol(fnName, types.FormatVisibility(fnName, ctx.Modifier()), symbols.KindFunc, returnType, ctx.Modifier(), anns)
	baseMeta := types.ExtractGenericParamMeta(ctx.TypeParams(), d.CurrentScope)
	whereMeta := types.ExtractWhereConstraintMeta(ctx.WhereClause(), d.CurrentScope)
	mergedMeta, issues := types.MergeGenericConstraints(baseMeta, whereMeta)
	for _, issue := range issues {
		fmt.Printf("Type Error: %s\n", issue.Error())
	}
	sym.GenericParams = mergedMeta
	sym.Methods["__fn_decl"] = ctx
	// We do NOT visit the block (body) here. That's for the second pass.
	return nil
}

// Ignore other statements for the collector pass
func (d *DeclarationCollector) VisitVarDecl(ctx *ast.VarDeclContext) interface{} { return nil }
func (d *DeclarationCollector) VisitAssignmentStmt(ctx *ast.AssignmentStmtContext) interface{} {
	return nil
}
func (d *DeclarationCollector) VisitExprStmt(ctx *ast.ExprStmtContext) interface{}   { return nil }
func (d *DeclarationCollector) VisitIfStmt(ctx *ast.IfStmtContext) interface{}       { return nil }
func (d *DeclarationCollector) VisitMatchStmt(ctx *ast.MatchStmtContext) interface{} { return nil }
func (d *DeclarationCollector) VisitWhileStmt(ctx *ast.WhileStmtContext) interface{} { return nil }
func (d *DeclarationCollector) VisitLoopStmt(ctx *ast.LoopStmtContext) interface{}   { return nil }
func (d *DeclarationCollector) VisitForStmt(ctx *ast.ForStmtContext) interface{}     { return nil }
func (d *DeclarationCollector) VisitBreakStmt(ctx *ast.BreakStmtContext) interface{} { return nil }
func (d *DeclarationCollector) VisitContinueStmt(ctx *ast.ContinueStmtContext) interface{} {
	return nil
}
func (d *DeclarationCollector) VisitReturnStmt(ctx *ast.ReturnStmtContext) interface{} { return nil }
func (d *DeclarationCollector) VisitTraitDecl(ctx *ast.TraitDeclContext) interface{} {
	// Dispatch to specific trait declaration types
	if pureTrait, ok := ctx.GetChild(0).(*ast.PureTraitDeclContext); ok {
		return pureTrait.Accept(d)
	}
	// BindTraitDecl (impl blocks) might be needed if they define methods on types?
	// Actually, in MyGo/Go, methods are attached to types.
	// If `trait bind` is implementing a trait for a type, we might need to know about it?
	// For now, let's assume struct methods are handled via `fn (recv) ...` which MyGo doesn't support directly yet?
	// Wait, MyGo `trait bind` is for implementing traits.
	// We might need to process them to check for method conflicts, but maybe not for basic symbol resolution of the types themselves.
	// However, if we support methods on structs (not just trait impls), we would need to collect them.
	// Current MyGo grammar seems to use `trait bind` for extensions.
	// Let's defer BindTraitDecl for now or include it if it defines new symbols.
	// The `BindTraitDecl` doesn't define a *new* type symbol, it implements methods.
	// But `SemanticAnalyzer` does checking inside it.
	return nil
}
