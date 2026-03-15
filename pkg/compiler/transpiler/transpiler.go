package transpiler

import (
	"fmt"
	"sort"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/compiler/interpreter"
	"github.com/miea04/mygo/pkg/compiler/symbols"
	"github.com/miea04/mygo/pkg/compiler/types"
)

// MyGoTranspiler converts MyGo AST to Go source code
type MyGoTranspiler struct {
	*ast.BaseMyGoVisitor
	Scope                *symbols.Scope
	CurrentScope         *symbols.Scope
	needsTernaryHelper   bool
	needsOptionHelper    bool
	loopDepth            int
	currentMatchTarget   string
	currentMatchExpr     string
	currentMatchNode     ast.IExprContext
	currentMatchVar      string
	useTypeMatch         bool
	compileTimeTypeMatch bool
	mixedBoolMatch       bool
	isEnumMatch          bool
	enumGenericArgs      string // e.g., "[int, string]" for Result<int, string>
	currentImplType      string
	currentImplSymbol    *symbols.Symbol // Add this field
	currentBindVar       string
	currentBindTypeDef   string
	expectedType         string
	CurrentFile          string // Path to the current MyGo source file
	UsedPackages         map[string]bool

	// RFC-007 Annotations
	CurrentStructAnnotations map[string][]string
	CurrentProcessingStruct  string
	InitFunctions            []string
	CurrentOriginalFnName    string // For #inner_call resolution
}

func NewMyGoTranspiler(global *symbols.Scope) *MyGoTranspiler {
	t := &MyGoTranspiler{
		BaseMyGoVisitor:          &ast.BaseMyGoVisitor{},
		Scope:                    global,
		CurrentScope:             global,
		UsedPackages:             make(map[string]bool),
		CurrentStructAnnotations: make(map[string][]string),
	}
	// Verify interface implementation
	var _ ast.MyGoVisitor = t
	return t
}

func (v *MyGoTranspiler) AddImport(pkg string) {
	v.UsedPackages[pkg] = true
}

func (v *MyGoTranspiler) SetCurrentFile(path string) {
	v.CurrentFile = path
}

func (v *MyGoTranspiler) VisitProgram(ctx *ast.ProgramContext) interface{} {
	var sb strings.Builder

	// 1. Package Declaration
	pkgName := "main"
	if ctx.PackageDecl() != nil {
		pkgName = ctx.PackageDecl().ID().GetText()
	}
	sb.WriteString(fmt.Sprintf("package %s\n\n", pkgName))

	// 2. Process Imports (First Pass)
	for _, imp := range ctx.AllImportStmt() {
		imp.Accept(v)
	}

	// 3. Process Statements & Annotations (First Pass to collect output and usage)
	var bodySb strings.Builder

	// Iterate over children to preserve order
	for _, child := range ctx.GetChildren() {
		// fmt.Printf("DEBUG: Visiting child %d type %T\n", i, child)
		if stmt, ok := child.(ast.IStatementContext); ok {
			res := stmt.Accept(v)
			// fmt.Printf("DEBUG: Child %d result type %T value: %v\n", i, res, res)
			if str, ok := res.(string); ok {
				bodySb.WriteString(str)
				bodySb.WriteString("\n")
			}
		} else if ann, ok := child.(ast.IAnnotationDeclContext); ok {
			ann.Accept(v)
		}
	}

	// 4. Generate Imports Block
	if len(v.UsedPackages) > 0 {
		sb.WriteString("import (\n")
		var imports []string
		for pkg := range v.UsedPackages {
			imports = append(imports, pkg)
		}
		sort.Strings(imports)
		for _, pkg := range imports {
			sb.WriteString(fmt.Sprintf("\t\"%s\"\n", pkg))
		}
		sb.WriteString(")\n\n")
	}

	// 5. Append Body
	sb.WriteString(bodySb.String())

	// 6. Helpers
	if v.needsTernaryHelper {
		sb.WriteString("\nfunc _mygo_ternary(cond bool, a, b interface{}) interface{} {\n")
		sb.WriteString("\tif cond { return a }\n")
		sb.WriteString("\treturn b\n")
		sb.WriteString("}\n")
	}
	if v.needsOptionHelper {
		sb.WriteString("\nfunc __MYGO_OPTION__Some[T any](v T) Option[T] {\n")
		sb.WriteString("\treturn Option_Some[T]{Item1: v}\n")
		sb.WriteString("}\n")
		sb.WriteString("\nfunc __MYGO_OPTION__None[T any]() Option[T] {\n")
		sb.WriteString("\treturn Option_None[T]{}\n")
		sb.WriteString("}\n")
	}

	return sb.String()
}

func (v *MyGoTranspiler) VisitBlockImport(ctx *ast.BlockImportContext) interface{} {
	for _, spec := range ctx.AllImportSpec() {
		v.processImportSpec(spec)
	}
	return ""
}

func (v *MyGoTranspiler) VisitSingleImport(ctx *ast.SingleImportContext) interface{} {
	v.processImportSpec(ctx.ImportSpec())
	return ""
}

func (v *MyGoTranspiler) processImportSpec(spec ast.IImportSpecContext) {
	path := spec.STRING().GetText()
	// Remove quotes
	path = strings.Trim(path, "\"")
	v.AddImport(path)
}

func (v *MyGoTranspiler) resolveType(ctx ast.ITypeTypeContext) string {
	if ctx == nil {
		return ""
	}
	return v.resolveTypeStr(ctx.GetText())
}

func (v *MyGoTranspiler) resolveTypeStr(rawText string) string {
	t := types.ResolveTypeWithScope(rawText, v.CurrentScope)
	return v.toGoType(t)
}

func (v *MyGoTranspiler) toGoType(t string) string {
	if types.IsOptionType(t) {
		inner, ok := types.OptionInnerType(t)
		if ok {
			return fmt.Sprintf("Option[%s]", v.resolveTypeStr(inner))
		}
	}
	if strings.HasPrefix(t, "Map<") && strings.HasSuffix(t, ">") {
		inner := t[4 : len(t)-1]
		parts := types.SplitTopLevelTypeArgs(inner)
		if len(parts) == 2 {
			kType := v.resolveTypeStr(parts[0])
			vType := v.resolveTypeStr(parts[1])
			return fmt.Sprintf("map[%s]%s", kType, vType)
		}
	}

	if strings.HasPrefix(t, "chan<") && strings.HasSuffix(t, ">") {
		elem := t[5 : len(t)-1]
		return "chan " + v.resolveTypeStr(elem)
	}
	// Handle function type fn(...) -> func(...)
	if strings.HasPrefix(t, "fn") {
		rest := t[2:]
		// Replace return type colon if present: fn(): int -> func() int
		rest = strings.Replace(rest, ":", " ", 1)
		return "func" + rest
	}

	// Handle slice T[] -> []T
	if strings.HasSuffix(t, "[]") {
		elem := t[:len(t)-2]
		return "[]" + v.resolveTypeStr(elem)
	}
	// Handle array T[N] -> [N]T
	if strings.HasSuffix(t, "]") {
		lastOpen := strings.LastIndex(t, "[")
		if lastOpen != -1 {
			sizeStr := t[lastOpen+1 : len(t)-1]
			isNum := true
			if len(sizeStr) == 0 {
				isNum = false
			}
			for _, r := range sizeStr {
				if r < '0' || r > '9' {
					isNum = false
					break
				}
			}
			if isNum {
				elem := t[:lastOpen]
				return "[" + sizeStr + "]" + v.resolveTypeStr(elem)
			}
		}
	}
	if t == "float" {
		return "float64"
	}

	// Try to resolve symbol to get Go name (case sensitivity)
	if sym := v.CurrentScope.Resolve(t); sym != nil {
		if sym.GoName != "" {
			return sym.GoName
		}
	}

	return t
}

func (v *MyGoTranspiler) boxOptionExpr(exprCode, valueType, targetType string) string {
	if !types.IsOptionType(targetType) {
		return exprCode
	}
	inner, ok := types.OptionInnerType(targetType)
	if !ok || inner == "" {
		return exprCode
	}
	if types.IsOptionType(valueType) {
		return exprCode
	}
	v.needsOptionHelper = true
	innerGoType := v.toGoType(inner)
	if types.NormalizeTypeName(valueType) == "nil" {
		return fmt.Sprintf("__MYGO_OPTION__None[%s]()", innerGoType)
	}
	if !types.IsTypeAssignable(inner, valueType, v.CurrentScope) {
		return exprCode
	}
	promoted := v.applyImplicitPromotion(exprCode, valueType, inner)
	return fmt.Sprintf("__MYGO_OPTION__Some[%s](%s)", innerGoType, promoted)
}

func isJSONStructTag(tag string) bool {
	trimmed := strings.TrimSpace(tag)
	trimmed = strings.Trim(trimmed, "\"`")
	return strings.Contains(trimmed, "json:")
}

func (v *MyGoTranspiler) specializedTaggedOptionalFieldGoType(fieldType, fieldTag string) (string, bool) {
	if !types.IsOptionType(fieldType) || !isJSONStructTag(fieldTag) {
		return "", false
	}
	inner, ok := types.OptionInnerType(fieldType)
	if !ok || inner == "" {
		return "", false
	}
	return "*" + v.toGoType(inner), true
}

func (v *MyGoTranspiler) boxTaggedOptionalFieldExpr(exprCode, exprType, fieldType, fieldTag string) string {
	if _, ok := v.specializedTaggedOptionalFieldGoType(fieldType, fieldTag); !ok {
		return exprCode
	}
	inner, ok := types.OptionInnerType(fieldType)
	if !ok || inner == "" {
		return exprCode
	}
	innerGoType := v.toGoType(inner)
	normalizedExprType := types.NormalizeTypeName(exprType)
	if normalizedExprType == "nil" {
		return "nil"
	}
	if types.IsOptionType(exprType) {
		return fmt.Sprintf("func() *%s { switch __mygo_opt := any(%s).(type) { case Option_Some[%s]: __mygo_v := __mygo_opt.Item1; return &__mygo_v; default: return nil } }()", innerGoType, exprCode, innerGoType)
	}
	if strings.HasPrefix(normalizedExprType, "*") && strings.TrimPrefix(normalizedExprType, "*") == types.NormalizeTypeName(innerGoType) {
		return exprCode
	}
	if !types.IsTypeAssignable(inner, exprType, v.CurrentScope) {
		return exprCode
	}
	promoted := v.applyImplicitPromotion(exprCode, exprType, inner)
	return fmt.Sprintf("func() *%s { __mygo_v := %s; return &__mygo_v }()", innerGoType, promoted)
}

func (v *MyGoTranspiler) resolveStructFieldTypeAndTag(sym *symbols.Symbol, typeArgs, fieldName string) (string, string, bool) {
	if sym == nil || sym.Kind != symbols.KindStruct {
		return "", "", false
	}
	fieldSym, ok := sym.FieldMap[fieldName]
	if !ok || fieldSym == nil {
		return "", "", false
	}
	fieldType := fieldSym.Type
	if len(sym.GenericParams) > 0 && strings.HasPrefix(typeArgs, "[") && strings.HasSuffix(typeArgs, "]") {
		inner := strings.TrimSpace(typeArgs[1 : len(typeArgs)-1])
		if inner != "" {
			args := types.SplitTopLevelTypeArgs(inner)
			if len(args) > 0 {
				fieldType = types.SubstituteTypeParams(fieldType, sym.GenericParams, args)
			}
		}
	}
	return fieldType, fieldSym.Tag, true
}

func (v *MyGoTranspiler) pushScope(name string) {
	v.CurrentScope = symbols.NewScope(name, v.CurrentScope)
}

func (v *MyGoTranspiler) executeMacro(annName string, targetNode interface{}) string {
	sym := v.CurrentScope.Resolve(annName)
	if sym == nil || sym.Kind != symbols.KindAnnotation {
		return "" // Not a macro or not found
	}

	decl, ok := sym.ASTNode.(*ast.AnnotationDeclContext)
	if !ok {
		return ""
	}

	interp := interpreter.NewInterpreter(v.Scope)
	meta := interpreter.MetaValue{Props: make(map[string]interpreter.Value)}

	// Try to get name from targetNode
	var targetName string
	var targetBody string
	if ctx, ok := targetNode.(antlr.ParserRuleContext); ok {
		switch n := ctx.(type) {
		case *ast.FnDeclContext:
			targetName = n.ID().GetText()
			// Extract body source text (excluding braces)
			if n.Block() != nil {
				start := n.Block().GetStart()
				stop := n.Block().GetStop()
				input := start.GetInputStream()

				// Use character indices to skip '{' and '}'
				startIndex := start.GetStop() + 1
				stopIndex := stop.GetStart() - 1

				if stopIndex > startIndex {
					interval := antlr.NewInterval(startIndex, stopIndex)
					targetBody = input.GetTextFromInterval(interval)
					// Trim trailing semicolon to avoid double semicolon when used as 'body;' in macro
					targetBody = strings.TrimSpace(targetBody)
					if strings.HasSuffix(targetBody, ";") {
						targetBody = targetBody[:len(targetBody)-1]
					}
				}
			}
		case *ast.StructDeclContext:
			targetName = n.ID().GetText()
			var fields []interpreter.Value
			for _, fieldDecl := range n.AllStructField() {
				fieldName := fieldDecl.ID().GetText()
				fieldType := fieldDecl.TypeType().GetText()
				fieldTag := ""
				if fieldDecl.STRING() != nil {
					fieldTag = fieldDecl.STRING().GetText()
					// Remove quotes
					if len(fieldTag) >= 2 && fieldTag[0] == '"' && fieldTag[len(fieldTag)-1] == '"' {
						fieldTag = fieldTag[1 : len(fieldTag)-1]
					}
				}

				fieldMeta := interpreter.MetaValue{Props: make(map[string]interpreter.Value)}
				fieldMeta.Props["name"] = interpreter.StringValue{Val: fieldName}
				fieldMeta.Props["type"] = interpreter.StringValue{Val: fieldType}
				fieldMeta.Props["tag"] = interpreter.StringValue{Val: fieldTag}

				fields = append(fields, fieldMeta)
			}
			meta.Props["fields"] = interpreter.ListValue{Val: fields}
		}
	}

	meta.Props["name"] = interpreter.StringValue{Val: targetName}
	if targetBody != "" {
		meta.Props["body"] = interpreter.StringValue{Val: targetBody}
	}
	interp.Env.Set("target", meta)
	interp.Env.Set("meta", meta) // Also set meta for backward compatibility or explicit access

	res := decl.Block().Accept(interp)
	if ret, ok := res.(interpreter.ReturnValue); ok {
		if str, ok := ret.Val.(interpreter.StringValue); ok {
			return str.Val
		}
	}
	return ""
}

// SyntaxErrorCounter counts syntax errors
type SyntaxErrorCounter struct {
	*antlr.DefaultErrorListener
	Errors int
}

func (c *SyntaxErrorCounter) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	c.Errors++
}

func (v *MyGoTranspiler) transpileMacroResult(code string) string {
	codeWithBraces := "{\n" + code + "\n}"
	input := antlr.NewInputStream(codeWithBraces)
	lexer := ast.NewMyGoLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := ast.NewMyGoParser(stream)

	// Suppress error output for cleaner logs, or keep for debugging
	parser.RemoveErrorListeners()
	counter := &SyntaxErrorCounter{DefaultErrorListener: antlr.NewDefaultErrorListener()}
	parser.AddErrorListener(counter)

	tree := parser.Block()

	// If syntax errors occur, assume it's raw Go code and return as is
	if counter.Errors > 0 {
		return code
	}

	res := v.VisitBlock(tree.(*ast.BlockContext))
	if s, ok := res.(string); ok {
		// Strip outer braces from result: "{\n ... \n}"
		// Find first { and last }
		first := strings.Index(s, "{")
		last := strings.LastIndex(s, "}")
		if first != -1 && last != -1 && last > first {
			return s[first+1 : last]
		}
		return s
	}
	return code // Fallback to raw code if transpilation fails
}

func (v *MyGoTranspiler) VisitInnerCallExpr(ctx *ast.InnerCallExprContext) interface{} {
	if v.CurrentOriginalFnName == "" {
		return "panic(\"inner_call used outside of macro context\")"
	}

	args := ""
	if ctx.ExprList() != nil {
		var argStrs []string
		for _, expr := range ctx.ExprList().(*ast.ExprListContext).AllExpr() {
			argStrs = append(argStrs, expr.Accept(v).(string))
		}
		args = strings.Join(argStrs, ", ")
	}

	return fmt.Sprintf("%s(%s)", v.CurrentOriginalFnName, args)
}

func (v *MyGoTranspiler) popScope() {
	if v.CurrentScope.Parent != nil {
		v.CurrentScope = v.CurrentScope.Parent
	}
}

func (v *MyGoTranspiler) NeedsTernaryHelper() bool {
	return v.needsTernaryHelper
}
