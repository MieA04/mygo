package semantic

import (
	"fmt"

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

// TraitCompositionContext holds the state for a single bind operation
type TraitCompositionContext struct {
	TargetType     *symbols.Symbol
	BannedMethods  map[string]bool   // MethodName -> isBanned
	FlippedMethods map[string]string // MethodName -> TraitName (keep this trait's method)
	LocalMethods   map[string]*ast.TraitFnDeclContext
}

func (d *MethodCollector) VisitBindTraitDecl(ctx *ast.BindTraitDeclContext) interface{} {
	// 1. Collect Directives & Local Methods (parse once, apply to all targets)
	compCtx := &TraitCompositionContext{
		BannedMethods:  make(map[string]bool),
		FlippedMethods: make(map[string]string),
		LocalMethods:   make(map[string]*ast.TraitFnDeclContext),
	}
	d.collectDirectivesAndMethods(ctx, compCtx)

	// 2. Iterate over all bind targets
	for _, targetCtx := range ctx.AllBindTarget() {
		tCtx := targetCtx.TypeType()
		if tCtx == nil {
			continue
		}

		// Resolve target type
		typeName := tCtx.GetText()
		baseName := types.SplitBaseType(typeName)

		sym := d.CurrentScope.Resolve(baseName)
		if sym == nil {
			sym = d.CurrentScope.ResolveByGoName(baseName)
		}

		if sym == nil {
			fmt.Printf("Semantic Error: Target type '%s' not found in bind declaration.\n", baseName)
			continue
		}

		// Only attach methods to Structs or Enums
		if sym.Kind != symbols.KindStruct && sym.Kind != symbols.KindEnum {
			fmt.Printf("Semantic Error: Target '%s' is not a struct or enum.\n", baseName)
			continue
		}

		if sym.Methods == nil {
			sym.Methods = make(map[string]interface{})
		}
		if sym.BoundTraits == nil {
			sym.BoundTraits = make(map[string]struct{})
		}

		compCtx.TargetType = sym

		// 3. Process Composed Traits
		if ctx.COMBS() != nil {
			// CombsClause structure: 'combs' '(' ID (',' ID)* ')'
			// We need to iterate over IDs.
			// The generated AST for CombsClause should have AllID()
			for _, traitID := range ctx.AllID() {
				traitName := traitID.GetText()
				traitSym := d.CurrentScope.Resolve(traitName)
				if traitSym == nil {
					fmt.Printf("Semantic Error: Trait '%s' not found in combs clause.\n", traitName)
					continue
				}
				if traitSym.Kind != symbols.KindTrait {
					fmt.Printf("Semantic Error: '%s' is not a trait.\n", traitName)
					continue
				}

				d.mergeTraitMethods(compCtx, traitSym)
			}
		}

		// 4. Attach Local Methods (Overriding logic)
		for name, method := range compCtx.LocalMethods {
			sym.Methods[name] = method
		}
	}
	return nil
}

func (d *MethodCollector) collectDirectivesAndMethods(ctx *ast.BindTraitDeclContext, compCtx *TraitCompositionContext) {
	for _, item := range ctx.AllTraitBodyItem() {
		// Collect Local Methods
		if fnCtx := item.TraitFnDecl(); fnCtx != nil {
			fnDecl := fnCtx.(*ast.TraitFnDeclContext)
			name := fnDecl.ID().GetText()
			compCtx.LocalMethods[name] = fnDecl
		}

		// Collect Directives
		if dirCtx := item.CompositionDirective(); dirCtx != nil {
			// Ban Directive
			if banCtx, ok := dirCtx.(*ast.BanDirectiveContext); ok {
				for _, id := range banCtx.AllID() {
					compCtx.BannedMethods[id.GetText()] = true
				}
			}
			// Flip Ban Directive
			if flipCtx, ok := dirCtx.(*ast.FlipBanDirectiveContext); ok {
				for _, item := range flipCtx.AllFlipBanItem() {
					// item text is "method : Trait"
					// We need to parse children manually or use visitor if convenient.
					// FlipBanItem : ID ':' ID ;
					ids := item.AllID()
					if len(ids) == 2 {
						methodName := ids[0].GetText()
						traitName := ids[1].GetText()
						compCtx.FlippedMethods[methodName] = traitName
					}
				}
			}
		}
	}
}

func (d *MethodCollector) mergeTraitMethods(compCtx *TraitCompositionContext, traitSym *symbols.Symbol) {
	targetSym := compCtx.TargetType

	// Record that this trait is bound
	targetSym.BoundTraits[traitSym.MyGoName] = struct{}{}

	// Merge methods from trait
	// Trait methods are populated in TraitMethods by DeclarationCollector.
	// We iterate over these methods to merge them into the target symbol.

	// Helper to process a map of methods
	processMethods := func(sourceMethods map[string]interface{}) {
		for methodName, methodNode := range sourceMethods {
			// 1. Check Ban
			if compCtx.BannedMethods[methodName] {
				continue
			}

			// 2. Check Local Override
			if _, ok := compCtx.LocalMethods[methodName]; ok {
				// Local method will overwrite at the end, so we can skip or add.
				// If we skip, we don't detect conflict, which is correct (override).
				continue
			}

			// 3. Check Conflict with existing methods in Target
			if _, exists := targetSym.Methods[methodName]; exists {
				// Conflict exists. Check Flip Ban.
				keptTrait, hasFlip := compCtx.FlippedMethods[methodName]
				if hasFlip {
					if keptTrait == traitSym.MyGoName {
						// Keep THIS trait's method (Overwrite existing)
						targetSym.Methods[methodName] = methodNode
					} else {
						// Keep OTHER trait's method (Skip this one)
						// We assume the existing one is from the 'keptTrait' or will be overwritten by it later.
						// If the existing one is NOT from 'keptTrait' and 'keptTrait' hasn't been processed yet,
						// we will overwrite it when we process 'keptTrait'.
						// If 'keptTrait' was already processed, it's already there.
						continue
					}
				} else {
					// No Flip Ban directive for this conflict -> Error
					fmt.Printf("Semantic Error: Duplicate method '%s' in type '%s' (conflict with trait '%s'). Use 'ban' or 'flip ban' to resolve.\n", methodName, targetSym.MyGoName, traitSym.MyGoName)
				}
			} else {
				// No conflict, just add
				targetSym.Methods[methodName] = methodNode
			}
		}
	}

	processMethods(traitSym.TraitMethods)
	// TODO: Support trait inheritance. Currently assuming flat traits.
}

// Ignore other declarations
func (d *MethodCollector) VisitStructDecl(ctx *ast.StructDeclContext) interface{} { return nil }
func (d *MethodCollector) VisitEnumDecl(ctx *ast.EnumDeclContext) interface{}     { return nil }
func (d *MethodCollector) VisitFnDecl(ctx *ast.FnDeclContext) interface{}         { return nil }
func (d *MethodCollector) VisitVarDecl(ctx *ast.VarDeclContext) interface{}       { return nil }
