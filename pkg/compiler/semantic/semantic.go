package semantic

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/compiler/symbols"
	"github.com/miea04/mygo/pkg/compiler/types"
)

type SemanticAnalyzer struct {
	*ast.BaseMyGoVisitor
	CurrentScope    *symbols.Scope
	GlobalScope     *symbols.Scope
	currentPackage  string
	currentFile     string
	errors          []string
	warnings        []string
	currentImplType string
	nonNilGuards    map[string]int
	nilGuards       map[string]int
	inDeferBlock    bool
	inSpawnBlock    bool
	PackageResolver func(name string) *symbols.Scope
	Diagnostics     []Diagnostic
}

type Diagnostic struct {
	Line    int
	Column  int
	Message string
	Code    string
	Type    string // "error" or "warning"
}

func NewSemanticAnalyzer(global *symbols.Scope) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		BaseMyGoVisitor: &ast.BaseMyGoVisitor{},
		CurrentScope:    global,
		GlobalScope:     global,
		nonNilGuards:    make(map[string]int),
		nilGuards:       make(map[string]int),
		Diagnostics:     []Diagnostic{},
	}
}

func (a *SemanticAnalyzer) SetCompilationUnit(packageName, filePath string) {
	a.currentPackage = packageName
	a.currentFile = filePath
}

func (a *SemanticAnalyzer) GetErrors() []string {
	return a.errors
}

func (a *SemanticAnalyzer) GetDiagnostics() []Diagnostic {
	return a.Diagnostics
}

func (a *SemanticAnalyzer) GetWarnings() []string {
	return a.warnings
}

func (a *SemanticAnalyzer) reportTypeError(ctx antlr.ParserRuleContext, code, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	line := 0
	column := 0
	if ctx != nil {
		line = ctx.GetStart().GetLine()
		column = ctx.GetStart().GetColumn()
	}
	a.errors = append(a.errors, fmt.Sprintf("[%s] line %d:%d %s", code, line, column, msg))
	a.Diagnostics = append(a.Diagnostics, Diagnostic{
		Line:    line,
		Column:  column,
		Message: msg,
		Code:    code,
		Type:    "error",
	})
}

func (a *SemanticAnalyzer) reportTypeWarning(ctx antlr.ParserRuleContext, code, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	line := 0
	column := 0
	if ctx != nil {
		line = ctx.GetStart().GetLine()
		column = ctx.GetStart().GetColumn()
	}
	a.warnings = append(a.warnings, fmt.Sprintf("[%s] line %d:%d %s", code, line, column, msg))
	a.Diagnostics = append(a.Diagnostics, Diagnostic{
		Line:    line,
		Column:  column,
		Message: msg,
		Code:    code,
		Type:    "warning",
	})
}

func (a *SemanticAnalyzer) pushScope(name string) {
	a.CurrentScope = symbols.NewScope(name, a.CurrentScope)
}

func (a *SemanticAnalyzer) popScope() {
	if a.CurrentScope.Parent != nil {
		a.CurrentScope = a.CurrentScope.Parent
	}
}

func (a *SemanticAnalyzer) isTypeAssignable(targetType, valueType string) bool {
	return types.IsTypeAssignable(targetType, valueType, a.CurrentScope)
}

func (a *SemanticAnalyzer) isCastAllowed(fromType, toType string) bool {
	from := types.NormalizeTypeName(fromType)
	to := types.NormalizeTypeName(toType)
	if from == "" || to == "" || from == "unknown" || to == "unknown" {
		return true
	}
	if from == to {
		return true
	}
	if types.IsNumericType(from) && types.IsNumericType(to) {
		return true
	}
	if types.IsStringType(from) && (types.IsByteSliceType(to) || types.IsRuneSliceType(to)) {
		return true
	}
	if types.IsStringType(to) && (types.IsByteSliceType(from) || types.IsRuneSliceType(from)) {
		return true
	}
	if types.IsAnyOrTraitType(to, a.CurrentScope) {
		return true
	}
	if types.IsPointerType(to) && types.IsAnyOrTraitType(from, a.CurrentScope) {
		return true
	}
	if types.IsPointerType(from) && types.IsAnyOrTraitType(to, a.CurrentScope) {
		return true
	}
	return false
}

func (a *SemanticAnalyzer) isAddressableExpr(expr ast.IExprContext) bool {
	switch e := expr.(type) {
	case *ast.IdentifierExprContext:
		return e.QualifiedName().GetText() != "nil"
	case *ast.MemberAccessExprContext, *ast.ArrayIndexExprContext, *ast.ThisExprContext:
		return true
	case *ast.ParenExprContext:
		return a.isAddressableExpr(e.Expr())
	default:
		return false
	}
}

func (a *SemanticAnalyzer) isNilLiteralExpr(expr ast.IExprContext) bool {
	switch e := expr.(type) {
	case *ast.NilExprContext:
		return true
	case *ast.IdentifierExprContext:
		return e.QualifiedName().GetText() == "nil"
	case *ast.ParenExprContext:
		return a.isNilLiteralExpr(e.Expr())
	default:
		return false
	}
}

func (a *SemanticAnalyzer) identifierName(expr ast.IExprContext) string {
	switch e := expr.(type) {
	case *ast.IdentifierExprContext:
		name := e.QualifiedName().GetText()
		if name == "nil" {
			return ""
		}
		return name
	case *ast.ParenExprContext:
		return a.identifierName(e.Expr())
	default:
		return ""
	}
}

func (a *SemanticAnalyzer) withNonNilGuard(varName string, run func()) {
	if varName == "" {
		run()
		return
	}
	if a.nonNilGuards == nil {
		a.nonNilGuards = make(map[string]int)
	}
	a.nonNilGuards[varName]++
	run()
	a.nonNilGuards[varName]--
	if a.nonNilGuards[varName] <= 0 {
		delete(a.nonNilGuards, varName)
	}
}

func (a *SemanticAnalyzer) withNilGuard(varName string, run func()) {
	if varName == "" {
		run()
		return
	}
	if a.nilGuards == nil {
		a.nilGuards = make(map[string]int)
	}
	a.nilGuards[varName]++
	run()
	a.nilGuards[varName]--
	if a.nilGuards[varName] <= 0 {
		delete(a.nilGuards, varName)
	}
}

func (a *SemanticAnalyzer) withNonNilGuards(names []string, run func()) {
	if len(names) == 0 {
		run()
		return
	}
	name := names[0]
	a.withNonNilGuard(name, func() {
		a.withNonNilGuards(names[1:], run)
	})
}

func (a *SemanticAnalyzer) withNilGuards(names []string, run func()) {
	if len(names) == 0 {
		run()
		return
	}
	name := names[0]
	a.withNilGuard(name, func() {
		a.withNilGuards(names[1:], run)
	})
}

func (a *SemanticAnalyzer) withGuardFacts(nonNilGuards, nilGuards []string, run func()) {
	a.withNonNilGuards(nonNilGuards, func() {
		a.withNilGuards(nilGuards, run)
	})
}

func (a *SemanticAnalyzer) isGuardedNonNilExpr(expr ast.IExprContext) bool {
	name := a.identifierName(expr)
	if name == "" || a.nonNilGuards == nil {
		return false
	}
	return a.nonNilGuards[name] > 0
}

func (a *SemanticAnalyzer) isGuardedNilExpr(expr ast.IExprContext) bool {
	name := a.identifierName(expr)
	if name == "" || a.nilGuards == nil {
		return false
	}
	return a.nilGuards[name] > 0
}

func (a *SemanticAnalyzer) validateTypeCheckType(typeCtx ast.ITypeTypeContext) string {
	if typeCtx == nil {
		return ""
	}
	typeName := types.ResolveTypeWithScope(typeCtx.GetText(), a.CurrentScope)
	baseType := types.SplitBaseType(typeName)

	// Built-in types check
	switch baseType {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "string", "bool",
		"byte", "rune", "any", "error":
		return typeName
	}

	sym := a.CurrentScope.Resolve(baseType)
	if sym == nil {
		sym = a.CurrentScope.ResolveByGoName(baseType)
	}
	if sym == nil {
		a.reportTypeError(typeCtx, "E_IS_TYPE_UNDEFINED", "undefined type '%s' in type check", typeCtx.GetText())
		return typeName
	}
	if sym.Kind != symbols.KindStruct && sym.Kind != symbols.KindTrait {
		a.reportTypeError(typeCtx, "E_IS_TYPE_INVALID_KIND", "'is' keyword only supports Struct or Trait types, got '%s'", sym.Kind)
	}
	return typeName
}

func (a *SemanticAnalyzer) resolveTypeIdentifierSymbol(expr ast.IExprContext) *symbols.Symbol {
	idExpr, ok := expr.(*ast.IdentifierExprContext)
	if !ok {
		return nil
	}
	name := idExpr.QualifiedName().GetText()
	sym := a.CurrentScope.Resolve(name)
	if sym == nil {
		sym = a.CurrentScope.ResolveByGoName(name)
	}
	if sym == nil {
		return nil
	}
	if sym.Kind != symbols.KindStruct && sym.Kind != symbols.KindTrait {
		return nil
	}
	return sym
}

func (a *SemanticAnalyzer) validateTraitTypeCheck(expr ast.IExprContext, targetType string, isNegated bool) {
	targetSym := types.ResolveTypeSymbol(targetType, a.CurrentScope)
	if targetSym == nil || targetSym.Kind != symbols.KindTrait {
		return
	}
	sourceSym := a.resolveTypeIdentifierSymbol(expr)
	if sourceSym == nil {
		return
	}
	rel := types.HasTraitRelationBySymbol(sourceSym, targetSym, a.CurrentScope)
	if sourceSym.Kind == symbols.KindTrait && sourceSym.MyGoName == targetSym.MyGoName {
		rel = true
	}
	if isNegated && !rel {
		return
	}
	if !isNegated && rel {
		return
	}
}

func (a *SemanticAnalyzer) checkAccess(ctx antlr.ParserRuleContext, sym *symbols.Symbol, name string) {
	if sym == nil {
		return
	}
	// Debug
	// fmt.Printf("Checking access for %s: vis=%s, symFile=%s, currFile=%s\n", name, sym.Visibility, sym.FilePath, a.currentFile)

	// 1. Cross-package check
	if sym.PackageName != "" && sym.PackageName != a.currentPackage {
		if sym.Visibility != symbols.VisibilityPublic {
			a.reportTypeError(ctx, "E_VISIBILITY_PUBLIC", "symbol '%s' is not public (package '%s')", name, sym.PackageName)
		}
		return
	}
	// 2. File-private check
	if sym.Visibility == symbols.VisibilityPrivate {
		if sym.FilePath != "" && a.currentFile != "" && sym.FilePath != a.currentFile {
			a.reportTypeError(ctx, "E_VISIBILITY_PRIVATE_FILE", "symbol '%s' is private to file '%s'", name, sym.FilePath)
		}
	}
}

func (a *SemanticAnalyzer) VisitProgram(ctx *ast.ProgramContext) interface{} {
	for _, imp := range ctx.AllImportStmt() {
		imp.Accept(a)
	}
	for _, stmt := range ctx.AllStatement() {
		stmt.Accept(a)
	}
	return nil
}

func (a *SemanticAnalyzer) VisitImportSpec(ctx *ast.ImportSpecContext) interface{} {
	path := strings.Trim(ctx.STRING().GetText(), "\"")
	var alias string
	if ctx.ID() != nil {
		alias = ctx.ID().GetText()
	}

	if a.PackageResolver == nil {
		return nil
	}

	scope := a.PackageResolver(path)
	if scope == nil {
		// Only report error if we expected to resolve it
		// For now, let's report warning or error?
		// If it's a standard library, it should resolve.
		// If it's a local package, maybe resolver handles it.
		a.reportTypeError(ctx, "E_IMPORT_FAILED", "failed to import package '%s'", path)
		return nil
	}

	if alias == "" {
		alias = scope.Name
	}

	sym := a.CurrentScope.Define(alias, alias, symbols.KindPackage, "package")
	sym.ImportedScope = scope
	// Imports are file-scoped usually, but MyGo might treat them as package-scoped if in package declaration?
	// For now, add to current scope (which is file/global scope).

	return nil
}

func (a *SemanticAnalyzer) VisitLambdaExpr(ctx *ast.LambdaExprContext) interface{} {
	oldInDefer := a.inDeferBlock
	a.inDeferBlock = false
	defer func() { a.inDeferBlock = oldInDefer }()

	a.pushScope("lambda")
	defer a.popScope()

	if ctx.ParamList() != nil {
		for _, pCtx := range ctx.ParamList().(*ast.ParamListContext).AllParam() {
			p := pCtx.(*ast.ParamContext)
			pName := p.ID().GetText()
			pType := types.ResolveTypeWithScope(p.TypeType().GetText(), a.CurrentScope)
			a.CurrentScope.Define(pName, pName, symbols.KindVar, pType)
		}
	}
	ctx.Block().Accept(a)
	return nil
}

func (a *SemanticAnalyzer) VisitStructDecl(ctx *ast.StructDeclContext) interface{} { return nil }
func (a *SemanticAnalyzer) VisitEnumDecl(ctx *ast.EnumDeclContext) interface{}     { return nil }

func (a *SemanticAnalyzer) VisitFnDecl(ctx *ast.FnDeclContext) interface{} {
	oldInDefer := a.inDeferBlock
	a.inDeferBlock = false
	defer func() { a.inDeferBlock = oldInDefer }()

	fnName := ctx.ID().GetText()
	a.pushScope("fn_" + fnName)
	defer a.popScope()

	if ctx.ParamList() != nil {
		for _, pCtx := range ctx.ParamList().(*ast.ParamListContext).AllParam() {
			p := pCtx.(*ast.ParamContext)
			pName := p.ID().GetText()
			pType := types.ResolveTypeWithScope(p.TypeType().GetText(), a.CurrentScope)
			a.CurrentScope.Define(pName, pName, symbols.KindVar, pType)
		}
	}
	if ctx.TypeType() != nil {
		retText := types.NormalizeTypeName(ctx.TypeType().GetText())
		if strings.HasSuffix(retText, "?") {
			a.reportTypeError(ctx, "E_OPTION_RETURN_UNSUPPORTED", "optional return type is not supported in current phase")
		}
	}

	if ctx.Block() != nil {
		for _, stmt := range ctx.Block().(*ast.BlockContext).AllStatement() {
			stmt.Accept(a)
		}
	}
	return nil
}

func (a *SemanticAnalyzer) VisitBlock(ctx *ast.BlockContext) interface{} {
	a.pushScope("block")
	defer a.popScope()
	for _, stmt := range ctx.AllStatement() {
		stmt.Accept(a)
	}
	return nil
}

func (a *SemanticAnalyzer) VisitStatement(ctx *ast.StatementContext) interface{} {
	if ctx.GetChild(0) != nil {
		return ctx.GetChild(0).(antlr.ParseTree).Accept(a)
	}
	return nil
}

func (a *SemanticAnalyzer) VisitVarDecl(ctx *ast.VarDeclContext) interface{} {
	if ctx.GetChild(0) != nil {
		if tree, ok := ctx.GetChild(0).(antlr.ParseTree); ok {
			return tree.Accept(a)
		}
	}
	return nil
}

func (a *SemanticAnalyzer) VisitSingleLetDecl(ctx *ast.SingleLetDeclContext) interface{} {
	name := ctx.ID().GetText()
	var inferredType string
	if ctx.TypeType() != nil {
		inferredType = types.ResolveTypeWithScope(ctx.TypeType().GetText(), a.CurrentScope)
	}
	if ctx.Expr() != nil {
		ctx.Expr().Accept(a)
		exprType := types.InferExprType(ctx.Expr(), a.CurrentScope)
		if inferredType == "" {
			inferredType = exprType
		} else {
			if !a.isTypeAssignable(inferredType, exprType) {
				a.reportTypeError(ctx, "E_TYPE_MISMATCH", "cannot assign type '%s' to variable of type '%s'", exprType, inferredType)
			}
		}
	}
	if inferredType == "" {
		inferredType = "unknown"
	}
	a.CurrentScope.Define(name, name, symbols.KindVar, inferredType)
	return nil
}

func (a *SemanticAnalyzer) VisitAssignmentStmt(ctx *ast.AssignmentStmtContext) interface{} {
	if ctx == nil || len(ctx.AllExpr()) < 2 {
		return nil
	}
	lhs := ctx.Expr(0)
	rhs := ctx.Expr(1)
	lhs.Accept(a)
	rhs.Accept(a)
	if ctx.GetOp() == nil || ctx.GetOp().GetText() != "=" {
		return nil
	}
	lhsType := types.InferExprType(lhs, a.CurrentScope)
	rhsType := types.InferExprType(rhs, a.CurrentScope)
	if !a.isTypeAssignable(lhsType, rhsType) {
		a.reportTypeError(ctx, "E_ASSIGNMENT_TYPE_MISMATCH", "cannot assign type '%s' to target of type '%s'", rhsType, lhsType)
	}
	return nil
}

func (a *SemanticAnalyzer) VisitReturnStmt(ctx *ast.ReturnStmtContext) interface{} {
	if a.inDeferBlock {
		a.reportTypeError(ctx, "E_DEFER_RETURN", "Cannot use 'return' inside a defer block")
	}
	if a.inSpawnBlock && ctx.Expr() != nil {
		a.reportTypeError(ctx, "E_SPAWN_RETURN_VALUE", "Cannot return a value from a spawn block")
	}
	if ctx.Expr() != nil {
		ctx.Expr().Accept(a)
	}
	return nil
}

func (a *SemanticAnalyzer) VisitSpawnStmt(ctx *ast.SpawnStmtContext) interface{} {
	oldInSpawn := a.inSpawnBlock
	a.inSpawnBlock = true
	defer func() { a.inSpawnBlock = oldInSpawn }()

	if ctx.Block() != nil {
		ctx.Block().Accept(a)
	} else if ctx.ExprStmt() != nil {
		ctx.ExprStmt().Accept(a)
	}
	return nil
}

func (a *SemanticAnalyzer) VisitSelectStmt(ctx *ast.SelectStmtContext) interface{} {
	for _, child := range ctx.GetChildren() {
		child.(antlr.ParseTree).Accept(a)
	}
	return nil
}

func (a *SemanticAnalyzer) VisitSelectReadBranch(ctx *ast.SelectReadBranchContext) interface{} {
	a.pushScope("select_case")
	defer a.popScope()

	// let x = <-ch or <-ch
	readCtx := ctx.SelectRead()
	chExpr := readCtx.Expr()
	chType := types.InferExprType(chExpr, a.CurrentScope)

	// Check method name
	if readCtx.GetMethod().GetText() != "read" {
		a.reportTypeError(readCtx, "E_SELECT_READ_METHOD", "Select read branch must use .read() method")
	}

	if !types.IsChannelType(chType) {
		a.reportTypeError(chExpr, "E_SELECT_READ_NOT_CHAN", "Select case must read from a channel, got '%s'", chType)
	} else {
		elemType := types.GetChannelElementType(chType)
		// If let vars are present
		// AllID includes the method name ID, so we exclude the last one
		allIDs := readCtx.AllID()
		varIDs := allIDs[:len(allIDs)-1]

		if len(varIDs) > 0 {
			// let x
			// or let x, ok
			if len(varIDs) == 1 {
				name := varIDs[0].GetText()
				a.CurrentScope.Define(name, name, symbols.KindVar, elemType)
			} else if len(varIDs) == 2 {
				name := varIDs[0].GetText()
				okName := varIDs[1].GetText()
				a.CurrentScope.Define(name, name, symbols.KindVar, elemType)
				a.CurrentScope.Define(okName, okName, symbols.KindVar, "bool")
			} else {
				a.reportTypeError(readCtx, "E_SELECT_READ_VARS", "Select read expects 1 or 2 variables, got %d", len(varIDs))
			}
		}
	}

	chExpr.Accept(a)
	if ctx.Block() != nil {
		// Manually visit statements to avoid double scoping
		for _, stmt := range ctx.Block().AllStatement() {
			stmt.Accept(a)
		}
	}
	return nil
}

func (a *SemanticAnalyzer) VisitSelectWriteBranch(ctx *ast.SelectWriteBranchContext) interface{} {
	a.pushScope("select_case")
	defer a.popScope()

	// case ch <- val
	writeCtx := ctx.SelectWrite()

	if writeCtx.GetMethod().GetText() != "write" {
		a.reportTypeError(writeCtx, "E_SELECT_WRITE_METHOD", "Select write branch must use .write() method")
	}

	chExpr := writeCtx.Expr(0)
	valExpr := writeCtx.Expr(1)

	chType := types.InferExprType(chExpr, a.CurrentScope)
	valType := types.InferExprType(valExpr, a.CurrentScope)

	if !types.IsChannelType(chType) {
		a.reportTypeError(chExpr, "E_SELECT_WRITE_NOT_CHAN", "Select case must write to a channel, got '%s'", chType)
	} else {
		elemType := types.GetChannelElementType(chType)
		if !a.isTypeAssignable(elemType, valType) {
			a.reportTypeError(valExpr, "E_SELECT_WRITE_TYPE", "Cannot write type '%s' to channel of type '%s'", valType, elemType)
		}
	}

	chExpr.Accept(a)
	valExpr.Accept(a)
	if ctx.Block() != nil {
		for _, stmt := range ctx.Block().AllStatement() {
			stmt.Accept(a)
		}
	}
	return nil
}

func (a *SemanticAnalyzer) VisitSelectOtherBranch(ctx *ast.SelectOtherBranchContext) interface{} {
	a.pushScope("select_case")
	defer a.popScope()

	if ctx.Block() != nil {
		for _, stmt := range ctx.Block().AllStatement() {
			stmt.Accept(a)
		}
	}
	return nil
}

func (a *SemanticAnalyzer) VisitExprStmt(ctx *ast.ExprStmtContext) interface{} {
	return ctx.Expr().Accept(a)
}

func (a *SemanticAnalyzer) VisitDeferStmt(ctx *ast.DeferStmtContext) interface{} {
	oldInDefer := a.inDeferBlock
	a.inDeferBlock = true
	defer func() { a.inDeferBlock = oldInDefer }()

	if ctx.Block() != nil {
		ctx.Block().Accept(a)
	} else if ctx.ExprStmt() != nil {
		// defer expr;
		exprStmt := ctx.ExprStmt().(*ast.ExprStmtContext)
		expr := exprStmt.Expr()

		isValid := false
		if _, ok := expr.(*ast.CallExprContext); ok {
			isValid = true
		} else if _, ok := expr.(*ast.MethodCallExprContext); ok {
			isValid = true
		} else if _, ok := expr.(*ast.FuncCallExprContext); ok {
			isValid = true
		} else if _, ok := expr.(*ast.PanicUnwrapExprContext); ok {
			isValid = true
		}

		if !isValid {
			a.reportTypeError(expr, "E_DEFER_INVALID_EXPR", "defer requires function call or block")
		}

		expr.Accept(a)
	}
	return nil
}

func (a *SemanticAnalyzer) VisitMatchStmt(ctx *ast.MatchStmtContext) interface{} {
	matchExpr := ctx.Expr()
	matchExprType := types.InferExprType(matchExpr, a.CurrentScope)

	baseType := types.SplitBaseType(matchExprType)
	enumSym := a.CurrentScope.Resolve(baseType)
	if enumSym == nil {
		enumSym = a.CurrentScope.ResolveByGoName(baseType)
	}

	isEnumMatch := enumSym != nil && enumSym.Kind == symbols.KindEnum

	for _, caseCtx := range ctx.AllMatchCase() {
		a.pushScope("case_block")

		var blockCtx ast.IBlockContext
		var stmtCtx ast.IStatementContext

		if valCase, ok := caseCtx.(*ast.ValueMatchCaseContext); ok {
			blockCtx = valCase.Block()
			stmtCtx = valCase.Statement()

			for _, patternExpr := range valCase.AllExpr() {
				if isEnumMatch {
					// Enum Pattern Matching
					variantName, args := a.parseEnumPattern(patternExpr, enumSym)
					if variantName != "" {
						// It is a valid variant pattern
						variantSym, ok := enumSym.Variants[variantName]
						if !ok {
							a.reportTypeError(patternExpr, "E_ENUM_VARIANT_NOT_FOUND", "Enum '%s' has no variant '%s'", enumSym.MyGoName, variantName)
							continue
						}

						// Validate arg count
						// We need to know expected types from variantSym
						// variantSym.Type stores the variant struct type? Or we look at fields?
						// In MyGo Enums, variants are structs. Item1, Item2...
						// We need to know how many items.
						// variantSym.FieldMap should contain Item1, Item2 etc.

						expectedCount := 0
						// Count "ItemX" fields
						for fName := range variantSym.FieldMap {
							if strings.HasPrefix(fName, "Item") {
								expectedCount++
							}
						}
						// This simple count might be wrong if there are other fields, but currently Enum variants only have ItemX.
						// Better: The grammar `enumVariant: ID ('(' typeList ')')?`
						// The symbol table should store the types of the variant fields in order.
						// For now, let's assume strict positional mapping if args are provided.

						if len(args) > 0 {
							// Check count
							// If we can't easily get the count from symbol (if not stored explicitly), we might skip count check or improve symbol table.
							// But for binding, we definitely need to define them.

							// Define variables
							// We need types.
							// For Result.OK(int), Item1 is int.
							// We need to resolve the type of Item{i+1} from variantSym.

							for i, argName := range args {
								itemField := fmt.Sprintf("Item%d", i+1)
								if fieldSym, ok := variantSym.FieldMap[itemField]; ok {
									// Handle generics substitution if needed
									fieldType := fieldSym.Type
									if idx := strings.Index(matchExprType, "["); idx != -1 {
										// This is a rough substitution. Ideally use a proper substitution helper.
										// If enum is Result<T>, and matchExprType is Result<int>.
										// fieldType might be "T". We need to map T -> int.
										// For now, let's define as "unknown" or try to infer?
										// Or just use the fieldType. If it's a generic param T, it might be resolved if we are inside a generic function?
										// But here we are instantiating.

										// TODO: Proper generic substitution.
										// For now, we define them.
									}
									a.CurrentScope.Define(argName, argName, symbols.KindVar, fieldType)
								} else {
									a.reportTypeError(patternExpr, "E_ENUM_PATTERN_ARG_COUNT", "Variant '%s' has no field for argument %d", variantName, i+1)
								}
							}
						}
					} else {
						// Not a variant pattern?
						// Maybe a catch-all variable?
						// But in Enum match, we usually expect variants.
						// If it is a simple identifier, it might be a catch-all binding if it doesn't match a variant?
						// Rust allows `match x { y => ... }` where y binds to x.
						if idCtx, ok := patternExpr.(*ast.IdentifierExprContext); ok {
							name := idCtx.QualifiedName().GetText()
							// If name is not a variant (checked inside parseEnumPattern), treat as variable binding
							if name != "_" && name != "true" && name != "false" {
								// Ensure it's not a constant or variable in scope
								if a.CurrentScope.Resolve(name) == nil {
									a.CurrentScope.Define(name, name, symbols.KindVar, matchExprType)
								}
							}
						}
					}
				} else {
					// Normal Match (Switch-like)
					// If pattern is a simple identifier and not a constant, treat as binding?
					if idCtx, ok := patternExpr.(*ast.IdentifierExprContext); ok {
						name := idCtx.QualifiedName().GetText()
						// Check if it resolves to a constant or variable
						sym := a.CurrentScope.Resolve(name)
						if sym == nil && name != "_" && name != "true" && name != "false" {
							// Treat as new variable binding (catch-all)
							a.CurrentScope.Define(name, name, symbols.KindVar, matchExprType)
						}
					}
				}
			}

		} else if typeCase, ok := caseCtx.(*ast.TypeMatchCaseContext); ok {
			blockCtx = typeCase.Block()
			stmtCtx = typeCase.Statement()
			// 'is' Type => block
			// Transpiler handles 'is' type narrowing.
			// Semantic analyzer should verify type exists.
			typeType := typeCase.TypeType()
			_ = a.validateTypeCheckType(typeType)

		} else if defCase, ok := caseCtx.(*ast.DefaultMatchCaseContext); ok {
			blockCtx = defCase.Block()
			stmtCtx = defCase.Statement()
		}

		if blockCtx != nil {
			// Visit Block
			if b, ok := blockCtx.(*ast.BlockContext); ok {
				// We manually visit statements to avoid creating another scope layer,
				// or we just let VisitBlock create one.
				// Current implementation of VisitBlock creates a scope "block".
				// We already pushed "case_block".
				// So variables defined in "case_block" are visible in "block".
				// That works.
				b.Accept(a)
			}
		} else if stmtCtx != nil {
			stmtCtx.Accept(a)
		}

		a.popScope()
	}
	return nil
}

func (a *SemanticAnalyzer) parseEnumPattern(expr ast.IExprContext, enumSym *symbols.Symbol) (string, []string) {
	if fc, ok := expr.(*ast.FuncCallExprContext); ok {
		// OK(val) or Result.OK(val)
		qName := fc.QualifiedName().GetText()
		var variantName string

		if strings.Contains(qName, ".") {
			parts := strings.Split(qName, ".")
			if parts[0] == enumSym.MyGoName || parts[0] == enumSym.GoName {
				variantName = parts[1]
			}
		} else {
			variantName = qName
		}

		if variantName != "" {
			var args []string
			if fc.ExprList() != nil {
				for _, arg := range fc.ExprList().(*ast.ExprListContext).AllExpr() {
					if id, ok := arg.(*ast.IdentifierExprContext); ok {
						args = append(args, id.GetText())
					}
				}
			}
			return variantName, args
		}
	} else if ma, ok := expr.(*ast.MemberAccessExprContext); ok {
		// Result.OK
		if id, ok := ma.Expr().(*ast.IdentifierExprContext); ok {
			name := id.GetText()
			if name == enumSym.MyGoName || name == enumSym.GoName {
				return ma.ID().GetText(), nil
			}
		}
	} else if id, ok := expr.(*ast.IdentifierExprContext); ok {
		// OK
		name := id.GetText()
		if _, ok := enumSym.Variants[name]; ok {
			return name, nil
		}
	}
	return "", nil
}

func (a *SemanticAnalyzer) VisitMemberAccessExpr(ctx *ast.MemberAccessExprContext) interface{} {
	objExpr := ctx.Expr()
	objType := types.InferExprType(objExpr, a.CurrentScope)
	fieldName := ctx.ID().GetText()

	objExpr.Accept(a)

	if sym := a.CurrentScope.Resolve(objType); sym != nil && sym.Kind == symbols.KindPackage {
		if sym.ImportedScope == nil {
			a.reportTypeError(ctx, "E_PKG_SCOPE_MISSING", "package '%s' has no scope", objType)
			return nil
		}
		memberSym := sym.ImportedScope.Resolve(fieldName)
		if memberSym == nil {
			a.reportTypeError(ctx, "E_PKG_MEMBER_NOT_FOUND", "package '%s' has no member '%s'", objType, fieldName)
			return nil
		}
		a.checkAccess(ctx, memberSym, fieldName)
		return nil
	}

	baseType := types.SplitBaseType(objType)
	parts := strings.Split(baseType, ".")
	if len(parts) == 2 {
		enumName := parts[0]
		variantName := parts[1]

		enumSym := a.CurrentScope.Resolve(enumName)
		if enumSym == nil {
			enumSym = a.CurrentScope.ResolveByGoName(enumName)
		}

		if enumSym != nil && enumSym.Kind == symbols.KindEnum {
			if variantSym, ok := enumSym.Variants[variantName]; ok {
				if _, ok := variantSym.FieldMap[fieldName]; ok {
					return nil
				}
			}
		}
	}
	return nil
}

func (a *SemanticAnalyzer) VisitTraitDecl(ctx *ast.TraitDeclContext) interface{} {
	if bindDecl, ok := ctx.GetChild(0).(*ast.BindTraitDeclContext); ok {
		return bindDecl.Accept(a)
	}
	return nil
}

func (a *SemanticAnalyzer) VisitBindTraitDecl(ctx *ast.BindTraitDeclContext) interface{} {
	// Simplified implementation
	// Collect directives
	for _, item := range ctx.AllTraitBodyItem() {
		if item.CompositionDirective() != nil {
			_ = item.CompositionDirective().Accept(a)
		}
	}
	return nil
}

func (a *SemanticAnalyzer) VisitBanDirective(ctx *ast.BanDirectiveContext) interface{} {
	var methods []string
	for _, id := range ctx.AllID() {
		methods = append(methods, id.GetText())
	}
	return methods
}

type FlipBanItem struct {
	Method string
	Trait  string
}

func (a *SemanticAnalyzer) VisitFlipBanDirective(ctx *ast.FlipBanDirectiveContext) interface{} {
	var items []FlipBanItem
	for _, itemCtx := range ctx.AllFlipBanItem() {
		if item, ok := itemCtx.Accept(a).(FlipBanItem); ok {
			items = append(items, item)
		}
	}
	return items
}

func (a *SemanticAnalyzer) VisitFlipBanItem(ctx *ast.FlipBanItemContext) interface{} {
	ids := ctx.AllID()
	if len(ids) == 2 {
		return FlipBanItem{
			Method: ids[0].GetText(),
			Trait:  ids[1].GetText(),
		}
	}
	return nil
}

func (a *SemanticAnalyzer) VisitStructLiteralExpr(ctx *ast.StructLiteralExprContext) interface{} {
	structName := ctx.QualifiedName().GetText()
	sym := a.CurrentScope.ResolveQualified(structName)
	if sym == nil && !strings.Contains(structName, ".") {
		sym = a.CurrentScope.ResolveByGoName(structName)
	}

	if sym != nil {
		a.checkAccess(ctx, sym, structName)
	}

	if sym != nil && sym.Kind == symbols.KindTrait {
		a.reportTypeError(ctx, "E_TRAIT_INSTANTIATION_FORBIDDEN", "cannot instantiate trait '%s'; use a bound struct or concrete value", structName)
		return nil
	}
	if sym != nil && sym.Kind == symbols.KindStruct {
		var args []string
		var err error

		if ctx.TypeArgs() != nil {
			args, err = types.ResolveGenericArgs(sym, ctx.TypeArgs(), func(t string) string {
				return t
			}, true)
		} else if len(sym.GenericParams) > 0 {
			inferredType := types.InferExprType(ctx, a.CurrentScope)
			if idx := strings.Index(inferredType, "["); idx != -1 {
				args = types.SplitTopLevelTypeArgs(inferredType[idx+1 : len(inferredType)-1])
			} else {
				args, err = types.ResolveGenericArgs(sym, nil, func(t string) string {
					return t
				}, true)
			}
		}

		if err != nil {
			a.reportTypeError(ctx, "E_GENERIC_ARGS_STRUCT_RESOLVE", "Error in struct instantiation %s: %v", structName, err)
			return nil
		}

		if len(args) > 0 {
			for i, arg := range args {
				if i < len(sym.GenericParams) {
					constraint := sym.GenericParams[i].ConstraintMyGo
					if !types.CheckTypeConstraint(arg, constraint, a.CurrentScope) {
						a.reportTypeError(ctx, "E_GENERIC_CONSTRAINT_STRUCT", "Generic argument '%s' for '%s' does not satisfy constraint '%s'", arg, sym.GenericParams[i].Name, constraint)
					}
				}
			}
		}
	}
	return nil
}

func (a *SemanticAnalyzer) VisitFuncCallExpr(ctx *ast.FuncCallExprContext) interface{} {
	callee := ctx.QualifiedName().GetText()
	sym := a.CurrentScope.ResolveQualified(callee)
	if sym == nil && !strings.Contains(callee, ".") {
		sym = a.CurrentScope.ResolveByGoName(callee)
	}

	if sym != nil {
		a.checkAccess(ctx, sym, callee)
	}

	var genericArgs []string

	if sym != nil && sym.Kind == symbols.KindFunc && len(sym.GenericParams) > 0 {
		var err error
		genericArgs, err = types.ResolveGenericArgs(sym, ctx.TypeArgs(), func(t string) string {
			return t
		}, true)
		if err != nil {
			a.reportTypeError(ctx, "E_GENERIC_ARGS_FUNC_RESOLVE", "Error in function call %s: %v", callee, err)
			return nil
		}
		for i, arg := range genericArgs {
			constraint := sym.GenericParams[i].ConstraintMyGo
			if !types.CheckTypeConstraint(arg, constraint, a.CurrentScope) {
				a.reportTypeError(ctx, "E_GENERIC_CONSTRAINT_FUNC", "Generic argument '%s' for '%s' does not satisfy constraint '%s'", arg, sym.GenericParams[i].Name, constraint)
			}
		}
	}
	return nil
}

func extractFnParamTypes(fnDecl *ast.FnDeclContext, scope *symbols.Scope, genericParams []symbols.GenericParamMeta, args []string) []string {
	// Helper function used in snippet but not defined in snippet.
	// I'll skip implementation or add dummy.
	return nil
}

func (a *SemanticAnalyzer) VisitMethodCallExpr(ctx *ast.MethodCallExprContext) interface{} {
	objExpr := ctx.Expr()
	objType := types.InferExprType(objExpr, a.CurrentScope)
	methodName := ctx.ID().GetText()

	objExpr.Accept(a)
	if ctx.ExprList() != nil {
		for _, e := range ctx.ExprList().(*ast.ExprListContext).AllExpr() {
			e.Accept(a)
		}
	}

	baseType := types.SplitBaseType(objType)
	sym := types.ResolveTypeSymbol(baseType, a.CurrentScope)

	if types.IsChannelType(objType) {
		if methodName == "read" {
			// .read() -> T
			// Check args count = 0
			if ctx.ExprList() != nil && len(ctx.ExprList().(*ast.ExprListContext).AllExpr()) > 0 {
				a.reportTypeError(ctx, "E_CHAN_READ_ARGS", "Channel.read() takes no arguments")
			}
			return nil
		} else if methodName == "write" {
			// .write(val) -> void
			// Check args count = 1
			if ctx.ExprList() == nil || len(ctx.ExprList().(*ast.ExprListContext).AllExpr()) != 1 {
				a.reportTypeError(ctx, "E_CHAN_WRITE_ARGS", "Channel.write(val) takes exactly 1 argument")
			} else {
				argExpr := ctx.ExprList().(*ast.ExprListContext).Expr(0)
				argType := types.InferExprType(argExpr, a.CurrentScope)
				elemType := types.GetChannelElementType(objType)
				if !a.isTypeAssignable(elemType, argType) {
					a.reportTypeError(ctx, "E_CHAN_WRITE_TYPE", "Cannot write type '%s' to channel of type '%s'", argType, elemType)
				}
			}
			return nil
		} else if methodName == "close" {
			// .close() -> void
			if ctx.ExprList() != nil && len(ctx.ExprList().(*ast.ExprListContext).AllExpr()) > 0 {
				a.reportTypeError(ctx, "E_CHAN_CLOSE_ARGS", "Channel.close() takes no arguments")
			}
			return nil
		} else {
			a.reportTypeError(ctx, "E_CHAN_METHOD", "Channel only supports read, write, close methods, got '%s'", methodName)
			return nil
		}
	}

	if sym == nil {
		if id, ok := objExpr.(*ast.IdentifierExprContext); ok {
			name := id.QualifiedName().GetText()
			if s := types.ResolveTypeSymbol(name, a.CurrentScope); s != nil && (s.Kind == symbols.KindEnum || s.Kind == symbols.KindStruct || s.Kind == symbols.KindTrait) {
				sym = s
			}
		}
	}

	if sym != nil {
		if sym.Kind == symbols.KindEnum {
			if len(sym.GenericParams) > 0 {
				args, err := types.ResolveGenericArgs(sym, ctx.TypeArgs(), func(t string) string {
					return t
				}, true)
				if err != nil {
				} else {
					for i, arg := range args {
						constraint := sym.GenericParams[i].ConstraintMyGo
						if !types.CheckTypeConstraint(arg, constraint, a.CurrentScope) {
							a.reportTypeError(ctx, "E_GENERIC_CONSTRAINT_ENUM", "Generic argument '%s' for Enum '%s' does not satisfy constraint '%s'", arg, sym.GenericParams[i].Name, constraint)
						}
					}
				}
			}
			return nil
		}

		var methodCtx interface{}
		ok := false
		if methodCtx, ok = sym.Methods[methodName]; !ok {
			methodCtx, ok = sym.TraitMethods[methodName]
		}
		if ok {
			_, methodGenericParams, ok := types.ExtractMethodInfo(methodCtx, a.CurrentScope)
			if !ok {
				methodGenericParams = nil
			}
			if len(methodGenericParams) > 0 {
				dummySym := &symbols.Symbol{GenericParams: methodGenericParams}
				args, err := types.ResolveGenericArgs(dummySym, ctx.TypeArgs(), func(t string) string {
					return t
				}, true)
				if err != nil {
					a.reportTypeError(ctx, "E_GENERIC_ARGS_METHOD_RESOLVE", "Error in method call %s.%s: %v", baseType, methodName, err)
					return nil
				}
				for i, arg := range args {
					constraint := methodGenericParams[i].ConstraintMyGo
					if !types.CheckTypeConstraint(arg, constraint, a.CurrentScope) {
						a.reportTypeError(ctx, "E_GENERIC_CONSTRAINT_METHOD", "Generic argument '%s' for method '%s' does not satisfy constraint '%s'", arg, methodGenericParams[i].Name, constraint)
					}
				}
			}
		}
	}
	return nil
}

func (a *SemanticAnalyzer) VisitDerefExpr(ctx *ast.DerefExprContext) interface{} {
	if a.isNilLiteralExpr(ctx.Expr()) {
		a.reportTypeError(ctx, "E_PTR_DEREF_NIL", "cannot dereference nil; assign a non-nil pointer before dereference")
		ctx.Expr().Accept(a)
		return nil
	}
	if a.isGuardedNilExpr(ctx.Expr()) {
		a.reportTypeError(ctx, "E_PTR_DEREF_GUARDED_NIL", "cannot dereference pointer proven nil in current branch")
		ctx.Expr().Accept(a)
		return nil
	}
	exprType := types.InferExprType(ctx.Expr(), a.CurrentScope)
	if !types.IsPointerType(exprType) {
		a.reportTypeError(ctx, "E_PTR_DEREF_NON_POINTER", "cannot dereference non-pointer type '%s'", exprType)
	} else if !a.isGuardedNonNilExpr(ctx.Expr()) {
		a.reportTypeWarning(ctx, "W_PTR_DEREF_NIL_POSSIBLE", "possible nil dereference on pointer type '%s'", exprType)
	}
	ctx.Expr().Accept(a)
	return nil
}

func (a *SemanticAnalyzer) VisitNilExpr(ctx *ast.NilExprContext) interface{} {
	return nil
}

func (a *SemanticAnalyzer) VisitIsExpr(ctx *ast.IsExprContext) interface{} {
	targetType := a.validateTypeCheckType(ctx.TypeType())
	a.validateTraitTypeCheck(ctx.Expr(), targetType, false)
	ctx.Expr().Accept(a)
	return nil
}

func (a *SemanticAnalyzer) VisitNotIsExpr(ctx *ast.NotIsExprContext) interface{} {
	targetType := a.validateTypeCheckType(ctx.TypeType())
	a.validateTraitTypeCheck(ctx.Expr(), targetType, true)
	ctx.Expr().Accept(a)
	return nil
}

func (a *SemanticAnalyzer) VisitParenExpr(ctx *ast.ParenExprContext) interface{} {
	return ctx.Expr().Accept(a)
}

func (a *SemanticAnalyzer) VisitTupleExpr(ctx *ast.TupleExprContext) interface{} {
	for _, e := range ctx.AllExpr() {
		e.Accept(a)
	}
	return nil
}

func (a *SemanticAnalyzer) VisitIdentifierExpr(ctx *ast.IdentifierExprContext) interface{} {
	name := ctx.QualifiedName().GetText()
	if name == "nil" {
		return nil
	}
	sym := a.CurrentScope.Resolve(name)
	if sym == nil {
		sym = a.CurrentScope.ResolveByGoName(name)
	}
	if sym != nil {
		a.checkAccess(ctx, sym, name)
	}
	return nil
}
func (a *SemanticAnalyzer) VisitStringExpr(ctx *ast.StringExprContext) interface{} { return nil }
func (a *SemanticAnalyzer) VisitIntExpr(ctx *ast.IntExprContext) interface{}       { return nil }
func (a *SemanticAnalyzer) VisitFloatExpr(ctx *ast.FloatExprContext) interface{}   { return nil }
func (a *SemanticAnalyzer) VisitThisExpr(ctx *ast.ThisExprContext) interface{}     { return nil }
func (a *SemanticAnalyzer) VisitAddrOfExpr(ctx *ast.AddrOfExprContext) interface{} {
	ctx.Expr().Accept(a)
	return nil
}
func (a *SemanticAnalyzer) VisitArrayLiteralExpr(ctx *ast.ArrayLiteralExprContext) interface{} {
	return nil
}
func (a *SemanticAnalyzer) VisitBinaryCompareExpr(ctx *ast.BinaryCompareExprContext) interface{} {
	ctx.Expr(0).Accept(a)
	ctx.Expr(1).Accept(a)
	return nil
}
func (a *SemanticAnalyzer) VisitLogicalAndExpr(ctx *ast.LogicalAndExprContext) interface{} {
	ctx.Expr(0).Accept(a)
	ctx.Expr(1).Accept(a)
	return nil
}
func (a *SemanticAnalyzer) VisitLogicalOrExpr(ctx *ast.LogicalOrExprContext) interface{} {
	ctx.Expr(0).Accept(a)
	ctx.Expr(1).Accept(a)
	return nil
}
func (a *SemanticAnalyzer) VisitAddSubExpr(ctx *ast.AddSubExprContext) interface{} {
	ctx.Expr(0).Accept(a)
	ctx.Expr(1).Accept(a)
	return nil
}
func (a *SemanticAnalyzer) VisitMulDivExpr(ctx *ast.MulDivExprContext) interface{} {
	ctx.Expr(0).Accept(a)
	ctx.Expr(1).Accept(a)
	return nil
}
func (a *SemanticAnalyzer) VisitNotExpr(ctx *ast.NotExprContext) interface{} {
	ctx.Expr().Accept(a)
	return nil
}
func (a *SemanticAnalyzer) VisitCastExpr(ctx *ast.CastExprContext) interface{} {
	ctx.Expr().Accept(a)
	return nil
}
func (a *SemanticAnalyzer) VisitArrayIndexExpr(ctx *ast.ArrayIndexExprContext) interface{} {
	ctx.Expr(0).Accept(a)
	ctx.Expr(1).Accept(a)
	return nil
}

func (a *SemanticAnalyzer) VisitTryUnwrapExpr(ctx *ast.TryUnwrapExprContext) interface{} {
	if a.inDeferBlock {
		a.reportTypeError(ctx, "E_DEFER_TRY_UNWRAP", "Cannot use '?!' inside a defer block")
	}
	if a.inSpawnBlock {
		a.reportTypeError(ctx, "E_SPAWN_TRY_UNWRAP", "Cannot use '?!' inside a spawn block")
	}
	ctx.Expr().Accept(a)
	return nil
}

func (a *SemanticAnalyzer) VisitPanicUnwrapExpr(ctx *ast.PanicUnwrapExprContext) interface{} {
	return a.VisitChildren(ctx)
}
