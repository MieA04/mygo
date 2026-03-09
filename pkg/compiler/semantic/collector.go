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
	return &DeclarationCollector{BaseMyGoVisitor: &ast.BaseMyGoVisitor{}, CurrentScope: global}
}

func (d *DeclarationCollector) SetCompilationUnit(packageName, filePath string) {
	d.currentPackage = packageName
	d.currentFile = filePath
}

func (d *DeclarationCollector) defineSymbol(mygoName, goName string, kind symbols.SymbolKind, typeStr string, modCtx ast.IModifierContext) *symbols.Symbol {
	return d.CurrentScope.DefineWithMeta(
		mygoName,
		goName,
		kind,
		typeStr,
		types.ParseVisibility(modCtx),
		d.currentPackage,
		d.currentFile,
	)
}

func (d *DeclarationCollector) VisitProgram(ctx *ast.ProgramContext) interface{} {
	for _, stmt := range ctx.AllStatement() {
		stmt.Accept(d)
	}
	return nil
}

func (d *DeclarationCollector) VisitStatement(ctx *ast.StatementContext) interface{} {
	child := ctx.GetChild(0)
	if child == nil {
		return nil
	}
	if tree, ok := child.(antlr.ParseTree); ok {
		return tree.Accept(d)
	}
	return nil
}

func (d *DeclarationCollector) VisitStructDecl(ctx *ast.StructDeclContext) interface{} {
	sym := d.defineSymbol(ctx.ID().GetText(), types.FormatVisibility(ctx.ID().GetText(), ctx.Modifier()), symbols.KindStruct, "struct", ctx.Modifier())
	baseMeta := types.ExtractGenericParamMeta(ctx.TypeParams(), d.CurrentScope)
	whereMeta := types.ExtractWhereConstraintMeta(ctx.WhereClause(), d.CurrentScope)
	mergedMeta, issues := types.MergeGenericConstraints(baseMeta, whereMeta)
	for _, issue := range issues {
		fmt.Printf("Type Error: %s\n", issue.Error())
	}
	sym.GenericParams = mergedMeta
	sym.Fields = make(map[string]string)

	for _, field := range ctx.AllStructField() {
		fieldName := field.ID().GetText()
		fieldType := types.ResolveTypeWithScope(field.TypeType().GetText(), d.CurrentScope)
		sym.Fields[fieldName] = fieldType
	}
	return nil
}

func (d *DeclarationCollector) VisitEnumDecl(ctx *ast.EnumDeclContext) interface{} {
	enumName := ctx.ID().GetText()
	sym := d.defineSymbol(enumName, types.FormatVisibility(enumName, ctx.Modifier()), symbols.KindEnum, "enum", ctx.Modifier())
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
			Fields:        make(map[string]string),
		}

		// Add fields if present (e.g. Some(T))
		if member.TypeList() != nil {
			for i, typeCtx := range member.TypeList().(*ast.TypeListContext).AllTypeType() {
				fieldName := fmt.Sprintf("Item%d", i+1)
				fieldType := types.ResolveTypeWithScope(typeCtx.GetText(), d.CurrentScope)
				variantSym.Fields[fieldName] = fieldType
			}
		}

		sym.Variants[variantName] = variantSym
	}
	return nil
}

func (d *DeclarationCollector) VisitPureTraitDecl(ctx *ast.PureTraitDeclContext) interface{} {
	traitName := ctx.ID().GetText()
	sym := d.defineSymbol(traitName, types.FormatVisibility(traitName, ctx.Modifier()), symbols.KindTrait, "trait", ctx.Modifier())
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
	sym := d.defineSymbol(fnName, types.FormatVisibility(fnName, ctx.Modifier()), symbols.KindFunc, returnType, ctx.Modifier())
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
