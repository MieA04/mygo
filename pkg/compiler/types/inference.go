package types

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/compiler/symbols"
)

// ResolveGenericArgs resolves generic arguments for a symbol
func ResolveGenericArgs(sym *symbols.Symbol, ctxArgs ast.ITypeArgsContext, resolver func(string) string, useMyGoDefaults bool) ([]string, error) {
	if sym == nil {
		return nil, fmt.Errorf("symbol is nil")
	}

	provided := []string{}
	if ctxArgs != nil {
		raw := ctxArgs.GetText() // <T, U>
		if len(raw) > 2 {
			inner := raw[1 : len(raw)-1] // T, U
			provided = SplitTopLevelTypeArgs(inner)
			for i, p := range provided {
				provided[i] = resolver(p)
			}
		}
	}

	required := sym.GenericParams
	if len(provided) > len(required) {
		return nil, fmt.Errorf("泛型参数过多: 期望 %d, 实际 %d", len(required), len(provided))
	}

	result := make([]string, len(required))
	copy(result, provided)

	for i := len(provided); i < len(required); i++ {
		defaultVal := required[i].DefaultType
		if useMyGoDefaults && required[i].DefaultMyGo != "" {
			defaultVal = required[i].DefaultMyGo
		}
		if defaultVal == "" {
			return nil, fmt.Errorf("缺失泛型参数: %s", required[i].Name)
		}
		result[i] = resolver(defaultVal)
	}

	return result, nil
}

// SubstituteTypeParams replaces generic parameters with actual arguments
func SubstituteTypeParams(typeStr string, params []symbols.GenericParamMeta, args []string) string {
	res := typeStr
	for i, param := range params {
		if i >= len(args) {
			break
		}
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(param.Name) + `\b`)
		res = re.ReplaceAllString(res, args[i])
	}
	return res
}

func CheckTypeConstraint(argType, constraint string, scope *symbols.Scope) bool {
	if constraint == "any" || constraint == "" {
		return true
	}

	if argType == constraint {
		return true
	}

	constraintBase := SplitBaseType(constraint)
	constraintSym := ResolveTypeSymbol(constraintBase, scope)

	if constraintSym == nil {
		if IsBuiltinConcreteType(constraint) {
			return argType == constraint
		}
		return false
	}

	if constraintSym.Kind == symbols.KindTrait {
		argBase := SplitBaseType(argType)
		argSym := ResolveTypeSymbol(argBase, scope)
		if argSym == nil {
			return false
		}

		for methodName, constraintMethod := range constraintSym.TraitMethods {
			implMethod, ok := argSym.Methods[methodName]
			if !ok {
				return false
			}

			// Check signature match
			constraintSig := GetMethodSignature(constraintMethod, scope)
			implSig := GetMethodSignature(implMethod, scope)

			if NormalizeSignature(constraintSig) != NormalizeSignature(implSig) {
				return false
			}
		}
		return true
	}
	return argType == constraint
}

func GetMethodSignature(method interface{}, scope *symbols.Scope) string {
	switch m := method.(type) {
	case string:
		return m // Already formatted string (from Go interface)
	case *ast.TraitFnDeclContext:
		// Extract signature from Trait definition or Bind implementation
		var params []string
		if m.ParamList() != nil {
			for _, pCtx := range m.ParamList().(*ast.ParamListContext).AllParam() {
				p := pCtx.(*ast.ParamContext)
				pType := ResolveTypeWithScope(p.TypeType().GetText(), scope)
				params = append(params, pType)
			}
		}

		paramStr := strings.Join(params, ", ")

		if m.TypeType() != nil {
			// Handle tuple return types properly if they exist in AST
			// But TypeType usually returns one type unless it's a tuple type "(A, B)"
			retType := ResolveTypeWithScope(m.TypeType().GetText(), scope)
			return fmt.Sprintf("fn(%s): %s", paramStr, retType)
		}

		return fmt.Sprintf("fn(%s)", paramStr)
	default:
		return ""
	}
}

func NormalizeSignature(sig string) string {
	// Remove spaces around commas, colons, parens for consistent comparison
	s := strings.ReplaceAll(sig, " ", "")
	return s
}

func IsBuiltinConcreteType(t string) bool {
	switch t {
	case "int", "string", "bool", "float64", "byte", "rune":
		return true
	}
	return false
}

func ParseTypeArgs(ctx ast.ITypeArgsContext) string {
	if ctx == nil {
		return ""
	}
	return ctx.GetText()
}

func SplitTopLevelTypeArgs(s string) []string {
	var args []string
	start := 0
	depth := 0
	for i, r := range s {
		switch r {
		case '<', '[', '(':
			depth++
		case '>', ']', ')':
			depth--
		case ',':
			if depth == 0 {
				args = append(args, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	args = append(args, strings.TrimSpace(s[start:]))
	return args
}

func NormalizeTypeName(typeName string) string {
	t := strings.TrimSpace(typeName)
	t = strings.ReplaceAll(t, " ", "")
	if t == "byte[]" {
		return "[]byte"
	}
	if t == "rune[]" {
		return "[]rune"
	}
	return t
}

func IsNumericType(typeName string) bool {
	switch NormalizeTypeName(typeName) {
	case "int", "float64", "byte", "rune":
		return true
	}
	return false
}

func IsStringType(typeName string) bool {
	return NormalizeTypeName(typeName) == "string"
}

func IsBoolType(typeName string) bool {
	return NormalizeTypeName(typeName) == "bool"
}

func IsPointerType(typeName string) bool {
	return strings.HasPrefix(NormalizeTypeName(typeName), "*")
}

func MakePointerType(typeName string) string {
	return "*" + typeName
}

func ElemTypeOfPointer(typeName string) string {
	return strings.TrimPrefix(NormalizeTypeName(typeName), "*")
}

func IsByteSliceType(typeName string) bool {
	return NormalizeTypeName(typeName) == "[]byte"
}

func IsRuneSliceType(typeName string) bool {
	return NormalizeTypeName(typeName) == "[]rune"
}

func IsNilCompatible(typeName string, scope *symbols.Scope) bool {
	t := NormalizeTypeName(typeName)
	if strings.HasPrefix(t, "*") || strings.HasPrefix(t, "[]") || strings.HasPrefix(t, "map[") || strings.HasPrefix(t, "func") {
		return true
	}
	// Check if it's a trait
	sym := ResolveTypeSymbol(t, scope)
	if sym != nil && sym.Kind == symbols.KindTrait {
		return true // Trait can be nil? In Go interfaces can be nil.
	}
	return false
}

func CanImplicitPromote(from, to string) bool {
	if from == "int" && to == "float64" {
		return true
	}
	return false
}

func IsTypeAssignable(target, value string, scope *symbols.Scope) bool {
	if target == value {
		return true
	}
	if target == "any" {
		return true
	}
	if value == "nil" {
		return IsNilCompatible(target, scope)
	}
	if CanImplicitPromote(value, target) {
		return true
	}
	// Trait check
	targetSym := ResolveTypeSymbol(target, scope)
	if targetSym != nil && targetSym.Kind == symbols.KindTrait {
		valueSym := ResolveTypeSymbol(value, scope)
		return HasTraitRelationBySymbol(valueSym, targetSym, scope)
	}
	return false
}

func CommonNumericType(t1, t2 string) (string, bool) {
	if !IsNumericType(t1) || !IsNumericType(t2) {
		return "", false
	}
	if t1 == t2 {
		return t1, true
	}
	if t1 == "float64" || t2 == "float64" {
		return "float64", true
	}
	return "", false
}

func ResolveBinaryOp(op, lType, rType string, scope *symbols.Scope) (kind, resultType, method string, negate bool, ok bool) {
	// 1. Builtin Numeric
	if common, isNum := CommonNumericType(lType, rType); isNum {
		switch op {
		case "+", "-", "*", "/":
			return "builtin_numeric", common, "", false, true
		case ">", "<", ">=", "<=", "==", "!=":
			return "builtin_numeric", "bool", "", false, true
		}
	}
	// 2. String concat
	if IsStringType(lType) && IsStringType(rType) {
		if op == "+" {
			return "builtin_string", "string", "", false, true
		}
		if op == "==" || op == "!=" || op == ">" || op == "<" || op == ">=" || op == "<=" {
			return "builtin_string", "bool", "", false, true
		}
	}
	// 3. Bool logic
	if IsBoolType(lType) && IsBoolType(rType) {
		if op == "&&" || op == "||" {
			return "builtin_bool", "bool", "", false, true
		}
		if op == "==" || op == "!=" {
			return "builtin_bool", "bool", "", false, true
		}
	}

	// 4. Operator Overloading via Traits
	methodMap := map[string]string{
		"+": "add", "-": "sub", "*": "mul", "/": "div",
		"==": "eq", "!=": "eq",
		"<": "lt", "<=": "le", ">": "gt", ">=": "ge",
	}
	if methodName, exists := methodMap[op]; exists {
		lSym := ResolveTypeSymbol(lType, scope)
		if lSym != nil {
			// Check methods on lSym
			if methodCtx, has := lSym.Methods[methodName]; has {
				// Check return type
				retType, _, _ := ExtractMethodInfo(methodCtx, scope)
				return "overload_method", retType, methodName, op == "!=", true
			}
			// Check bound traits
			for traitName := range lSym.BoundTraits {
				traitSym := ResolveTypeSymbol(traitName, scope)
				if traitSym != nil {
					if methodCtx, has := traitSym.AbstractTraitMethods[methodName]; has {
						retType, _, _ := ExtractMethodInfo(methodCtx, scope)
						return "overload_method", retType, methodName, op == "!=", true
					}
					if methodCtx, has := traitSym.ConcreteTraitMethods[methodName]; has {
						retType, _, _ := ExtractMethodInfo(methodCtx, scope)
						return "overload_method", retType, methodName, op == "!=", true
					}
				}
			}
		}
	}

	return "", "", "", false, false
}

func SplitBaseType(typeName string) string {
	typeName = NormalizeTypeName(typeName)
	for strings.HasPrefix(typeName, "*") {
		typeName = strings.TrimPrefix(typeName, "*")
	}
	for i, ch := range typeName {
		if ch == '[' || ch == '<' {
			return typeName[:i]
		}
	}
	return typeName
}

func ResolveTypeSymbol(typeName string, scope *symbols.Scope) *symbols.Symbol {
	if scope == nil {
		return nil
	}
	baseType := SplitBaseType(typeName)
	// Handle package.Type syntax
	if strings.Contains(baseType, ".") {
		parts := strings.SplitN(baseType, ".", 2)
		pkgName := parts[0]
		typeName := parts[1]

		pkgSym := scope.Resolve(pkgName)
		if pkgSym != nil {
			if pkgSym.Kind == symbols.KindPackage && pkgSym.ImportedScope != nil {
				return pkgSym.ImportedScope.Resolve(typeName)
			} else {
				fmt.Fprintf(os.Stderr, "DEBUG: pkgSym %s is not a valid package (Kind=%s, Scope=%v)\n", pkgName, pkgSym.Kind, pkgSym.ImportedScope)
			}
		} else {
			fmt.Fprintf(os.Stderr, "DEBUG: pkgSym %s not found in scope %s\n", pkgName, scope.Name)
		}
	}

	sym := scope.Resolve(baseType)
	if sym == nil {
		sym = scope.ResolveByGoName(baseType)
	}
	return sym
}

func SymbolName(sym *symbols.Symbol) string {
	if sym == nil {
		return ""
	}
	if sym.MyGoName != "" {
		return sym.MyGoName
	}
	return sym.GoName
}

func AddBoundTrait(sym *symbols.Symbol, traitSym *symbols.Symbol) {
	if sym == nil || traitSym == nil || traitSym.Kind != symbols.KindTrait {
		return
	}
	if sym.BoundTraits == nil {
		sym.BoundTraits = make(map[string]struct{})
	}
	name := SymbolName(traitSym)
	if name == "" {
		return
	}
	sym.BoundTraits[name] = struct{}{}
}

func HasTraitRelationBySymbol(sourceSym *symbols.Symbol, targetTraitSym *symbols.Symbol, scope *symbols.Scope) bool {
	if sourceSym == nil || targetTraitSym == nil || targetTraitSym.Kind != symbols.KindTrait {
		return false
	}
	targetName := SymbolName(targetTraitSym)
	if targetName == "" {
		return false
	}
	if sourceSym.Kind == symbols.KindTrait && SymbolName(sourceSym) == targetName {
		return true
	}
	visited := make(map[string]struct{})
	stack := make([]*symbols.Symbol, 0, len(sourceSym.BoundTraits))
	for name := range sourceSym.BoundTraits {
		if traitSym := ResolveTypeSymbol(name, scope); traitSym != nil && traitSym.Kind == symbols.KindTrait {
			stack = append(stack, traitSym)
		}
	}
	for len(stack) > 0 {
		n := len(stack) - 1
		current := stack[n]
		stack = stack[:n]
		currentName := SymbolName(current)
		if currentName == "" {
			continue
		}
		if currentName == targetName {
			return true
		}
		if _, ok := visited[currentName]; ok {
			continue
		}
		visited[currentName] = struct{}{}
		for name := range current.BoundTraits {
			if nextTrait := ResolveTypeSymbol(name, scope); nextTrait != nil && nextTrait.Kind == symbols.KindTrait {
				stack = append(stack, nextTrait)
			}
		}
	}
	return false
}

func ParseVisibility(modCtx ast.IModifierContext) symbols.Visibility {
	if modCtx == nil {
		return symbols.VisibilityPackage
	}
	switch modCtx.GetText() {
	case "pub":
		return symbols.VisibilityPublic
	case "pri":
		return symbols.VisibilityPrivate
	default:
		return symbols.VisibilityPackage
	}
}

func FormatVisibility(name string, modCtx ast.IModifierContext) string {
	if modCtx != nil && modCtx.GetText() == "pub" {
		return strings.ToUpper(name[:1]) + name[1:]
	}
	return strings.ToLower(name[:1]) + name[1:]
}

func SniffType(expr antlr.ParseTree) string {
	arrCtx, ok := expr.(*ast.ArrayLiteralExprContext)
	if !ok || arrCtx.ExprList() == nil {
		return "[]interface{}"
	}
	switch arrCtx.ExprList().(*ast.ExprListContext).Expr(0).(type) {
	case *ast.IntExprContext:
		return "[]int"
	case *ast.StringExprContext:
		return "[]string"
	case *ast.FloatExprContext:
		return "[]float64"
	default:
		return "[]interface{}"
	}
}

func ExtractCollectionTypes(colType string, scope *symbols.Scope) (kType string, vType string) {
	kType, vType = "unknown", "unknown"
	if strings.HasPrefix(colType, "[]") {
		kType = "int"
		vType = strings.TrimPrefix(colType, "[]")
		return
	} else if strings.HasPrefix(colType, "map[") {
		idx := strings.Index(colType, "]")
		if idx != -1 {
			kType = colType[4:idx]
			vType = colType[idx+1:]
			return
		}
	} else if colType == "string" {
		kType = "int"
		vType = "byte"
		return
	}
	return
}

func ExtractMethodInfo(methodCtx interface{}, scope *symbols.Scope) (string, []symbols.GenericParamMeta, bool) {
	switch m := methodCtx.(type) {
	case *ast.FnDeclContext:
		retType := "void"
		if m.TypeType() != nil {
			retType = ResolveTypeWithScope(m.TypeType().GetText(), scope)
		}
		return retType, ExtractGenericParamMeta(m.TypeParams(), scope), true
	case *ast.TraitFnDeclContext:
		retType := "void"
		if m.TypeType() != nil {
			retType = ResolveTypeWithScope(m.TypeType().GetText(), scope)
		}
		return retType, ExtractGenericParamMeta(m.TypeParams(), scope), true
	default:
		return "", nil, false
	}
}

func ExtractGenericParamMeta(typeParams ast.ITypeParamsContext, scope *symbols.Scope) []symbols.GenericParamMeta {
	if typeParams == nil {
		return nil
	}
	var metas []symbols.GenericParamMeta
	for _, tp := range typeParams.AllTypeParam() {
		name := tp.ID().GetText()
		var constraint string
		if tp.TypeType(0) != nil {
			constraint = ResolveTypeWithScope(tp.TypeType(0).GetText(), scope)
		}
		var defaultType string
		if tp.TypeType(1) != nil {
			defaultType = ResolveTypeWithScope(tp.TypeType(1).GetText(), scope)
		} else if tp.TypeType(0) != nil && strings.Contains(tp.GetText(), "=") {
			// Handle T = Default (no constraint) vs T : Constraint
			// The parser rule is ID (':' typeType)? ('=' typeType)?
			// If TypeType(0) exists, check if it's constraint or default
			// Actually antlr context accessors are by index
			// We need to check children or rule context structure
			// Simplified: assume 0 is constraint if ':' is present
		}
		metas = append(metas, symbols.GenericParamMeta{
			Name:           name,
			ConstraintMyGo: constraint,
			DefaultMyGo:    defaultType,
		})
	}
	return metas
}

func ExtractWhereConstraintMeta(whereClause ast.IWhereClauseContext, scope *symbols.Scope) map[string]symbols.GenericParamMeta {
	// Placeholder implementation
	return nil
}

func MergeGenericConstraints(base []symbols.GenericParamMeta, where map[string]symbols.GenericParamMeta) ([]symbols.GenericParamMeta, []error) {
	return base, nil
}

func ExtractMethodParamTypes(methodCtx interface{}, scope *symbols.Scope) ([]string, bool) {
	switch m := methodCtx.(type) {
	case *ast.FnDeclContext:
		if m.ParamList() == nil {
			return nil, true
		}
		var paramTypes []string
		for _, pCtx := range m.ParamList().(*ast.ParamListContext).AllParam() {
			p := pCtx.(*ast.ParamContext)
			paramTypes = append(paramTypes, ResolveTypeWithScope(p.TypeType().GetText(), scope))
		}
		return paramTypes, true
	case *ast.TraitFnDeclContext:
		if m.ParamList() == nil {
			return nil, true
		}
		var paramTypes []string
		for _, pCtx := range m.ParamList().(*ast.ParamListContext).AllParam() {
			p := pCtx.(*ast.ParamContext)
			paramTypes = append(paramTypes, ResolveTypeWithScope(p.TypeType().GetText(), scope))
		}
		return paramTypes, true
	}
	return nil, false
}

func ResolveInstanceTypeArgs(objType string, sym *symbols.Symbol, resolver func(string) string, useDefaults bool) []string {
	// Extract args from objType e.g. List[int] -> [int]
	if idx := strings.Index(objType, "["); idx != -1 {
		inner := objType[idx+1 : len(objType)-1]
		return SplitTopLevelTypeArgs(inner)
	}
	return nil
}

func ReplaceTypeParam(typeStr, param, arg string) string {
	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(param) + `\b`)
	return re.ReplaceAllString(typeStr, arg)
}

func IsAnyOrTraitType(t string, scope *symbols.Scope) bool {
	if t == "any" {
		return true
	}
	sym := ResolveTypeSymbol(t, scope)
	return sym != nil && sym.Kind == symbols.KindTrait
}

func TranspileType(t string) string {
	return t // Placeholder
}

func InferExprType(expr antlr.ParseTree, scope *symbols.Scope) string {
	if expr == nil {
		return "unknown"
	}
	switch e := expr.(type) {
	case *ast.IntExprContext:
		return "int"
	case *ast.StringExprContext:
		return "string"
	case *ast.FloatExprContext:
		return "float64"
	case *ast.ArrayLiteralExprContext:
		return SniffType(e)
	case *ast.LambdaExprContext:
		return "func"
	case *ast.StructLiteralExprContext:
		structName := e.QualifiedName().GetText()
		sym := scope.ResolveQualified(structName)
		if sym == nil && !strings.Contains(structName, ".") {
			sym = scope.ResolveByGoName(structName)
		}

		if sym != nil && sym.Kind == symbols.KindStruct && len(sym.GenericParams) > 0 {
			if e.TypeArgs() != nil {
				return structName + ParseTypeArgs(e.TypeArgs())
			}
			inferredMap := make(map[string]string)
			ids := e.AllID()
			exprs := e.AllExpr()

			for i := 0; i < len(ids); i++ {
				fieldName := ids[i].GetText()
				if fieldType, ok := sym.Fields[fieldName]; ok {
					for _, gp := range sym.GenericParams {
						if fieldType == gp.Name {
							if i < len(exprs) {
								actualType := InferExprType(exprs[i], scope)
								if actualType != "unknown" {
									inferredMap[gp.Name] = actualType
								}
							}
						}
					}
				}
			}

			var args []string
			for _, gp := range sym.GenericParams {
				if val, ok := inferredMap[gp.Name]; ok {
					args = append(args, val)
				} else {
					if gp.DefaultMyGo != "" {
						args = append(args, gp.DefaultMyGo)
					} else {
						args = append(args, "any")
					}
				}
			}
			return fmt.Sprintf("%s[%s]", structName, strings.Join(args, ", "))
		}
		return structName
	case *ast.IdentifierExprContext:
		idName := e.QualifiedName().GetText()
		if idName == "true" || idName == "false" {
			return "bool"
		}
		if idName == "nil" {
			return "nil"
		}
		if sym := scope.ResolveQualified(idName); sym != nil {
			if sym.Kind == symbols.KindStruct || sym.Kind == symbols.KindEnum || sym.Kind == symbols.KindTrait || sym.Kind == symbols.KindPackage {
				return idName
			}
			return sym.Type
		}
	case *ast.NilExprContext:
		return "nil"
	case *ast.AddrOfExprContext:
		inner := InferExprType(e.Expr(), scope)
		if inner == "unknown" || inner == "nil" {
			return "unknown"
		}
		return MakePointerType(inner)
	case *ast.DerefExprContext:
		inner := InferExprType(e.Expr(), scope)
		if IsPointerType(inner) {
			return ElemTypeOfPointer(inner)
		}
		return "unknown"
	case *ast.ThisExprContext:
		if sym := scope.Resolve("this"); sym != nil {
			return sym.Type
		}
		return "unknown"
	case *ast.BinaryCompareExprContext, *ast.IsExprContext, *ast.NotIsExprContext, *ast.LogicalAndExprContext, *ast.LogicalOrExprContext:
		return "bool"
	case *ast.NotExprContext:
		if t := InferExprType(e.Expr(), scope); IsBoolType(t) || t == "unknown" {
			return "bool"
		}
		return "unknown"
	case *ast.AddSubExprContext:
		lType := InferExprType(e.Expr(0), scope)
		rType := InferExprType(e.Expr(1), scope)
		if _, resultType, _, _, ok := ResolveBinaryOp(e.GetOp().GetText(), lType, rType, scope); ok {
			return resultType
		}
		return "unknown"
	case *ast.MulDivExprContext:
		lType := InferExprType(e.Expr(0), scope)
		rType := InferExprType(e.Expr(1), scope)
		if _, resultType, _, _, ok := ResolveBinaryOp(e.GetOp().GetText(), lType, rType, scope); ok {
			return resultType
		}
		return "unknown"
	case *ast.CastExprContext:
		return ResolveTypeWithScope(e.TypeType().GetText(), scope)
	case *ast.FuncCallExprContext:
		callee := e.QualifiedName().GetText()
		if callee == "chan" {
			if e.TypeArgs() != nil {
				args := SplitTopLevelTypeArgs(e.TypeArgs().GetText()[1 : len(e.TypeArgs().GetText())-1])
				if len(args) > 0 {
					return fmt.Sprintf("chan<%s>", ResolveTypeWithScope(args[0], scope))
				}
			}
			return "chan<any>"
		}
		if sym := scope.ResolveQualified(callee); sym != nil && sym.Kind == symbols.KindFunc {
			baseType := sym.Type
			if len(sym.GenericParams) > 0 {
				if args, err := ResolveGenericArgs(sym, e.TypeArgs(), func(t string) string {
					return ResolveTypeWithScope(t, scope)
				}, true); err == nil {
					return SubstituteTypeParams(baseType, sym.GenericParams, args)
				}
			}
			return baseType
		}
		// Handle enum variant constructor Enum.Variant
		parts := strings.Split(callee, ".")
		if len(parts) == 2 {
			enumName := parts[0]
			if enumSym := scope.ResolveQualified(enumName); enumSym != nil && enumSym.Kind == symbols.KindEnum {
				if e.TypeArgs() != nil {
					return enumName + ParseTypeArgs(e.TypeArgs())
				}
				// Try to infer from arguments
				if len(enumSym.GenericParams) > 0 && e.ExprList() != nil {
					argExprs := e.ExprList().(*ast.ExprListContext).AllExpr()
					if len(argExprs) > 0 {
						firstArgType := InferExprType(argExprs[0], scope)
						if firstArgType != "unknown" {
							return fmt.Sprintf("%s[%s]", enumName, firstArgType)
						}
					}
				}
				return enumName
			}
		}
	case *ast.ParenExprContext:
		return InferExprType(e.Expr(), scope)
	case *ast.TupleExprContext:
		var types []string
		for _, subExpr := range e.AllExpr() {
			types = append(types, InferExprType(subExpr, scope))
		}
		return "(" + strings.Join(types, ", ") + ")"
	case *ast.MethodCallExprContext:
		objName := ""
		if id, ok := e.Expr().(*ast.IdentifierExprContext); ok {
			objName = id.QualifiedName().GetText()
		}
		if objName != "" {
			if sym := scope.ResolveQualified(objName); sym != nil && sym.Kind == symbols.KindEnum {
				// Enum variant constructor
				if e.TypeArgs() != nil {
					return objName + ParseTypeArgs(e.TypeArgs())
				}
				// Try to infer type args from arguments
				if len(sym.GenericParams) > 0 && e.ExprList() != nil {
					// Simplified inference: use first argument's type for first generic param
					argExprs := e.ExprList().(*ast.ExprListContext).AllExpr()
					if len(argExprs) > 0 {
						firstArgType := InferExprType(argExprs[0], scope)
						if firstArgType != "unknown" {
							return fmt.Sprintf("%s[%s]", objName, firstArgType)
						}
					}
				}
				return objName
			}
		}
		// General method call
		objType := InferExprType(e.Expr(), scope)

		if IsChannelType(objType) {
			methodName := e.ID().GetText()
			if methodName == "read" {
				return GetChannelElementType(objType)
			} else if methodName == "write" || methodName == "close" {
				return "void"
			}
			return "unknown"
		}

		baseType := SplitBaseType(objType)
		methodName := e.ID().GetText()
		if sym := ResolveTypeSymbol(baseType, scope); sym != nil {
			var methodRetType string
			if methodCtx, ok := sym.Methods[methodName]; ok {
				methodRetType, _, _ = ExtractMethodInfo(methodCtx, scope)
			} else if traitMethodCtx, ok := sym.TraitMethods[methodName]; ok {
				methodRetType, _, _ = ExtractMethodInfo(traitMethodCtx, scope)
			} else {
				return "unknown"
			}
			// Substitute generics
			if len(sym.GenericParams) > 0 {
				if args := ResolveInstanceTypeArgs(objType, sym, func(t string) string {
					return ResolveTypeWithScope(t, scope)
				}, true); len(args) > 0 {
					methodRetType = SubstituteTypeParams(methodRetType, sym.GenericParams, args)
				}
			}
			return methodRetType
		}

	case *ast.MemberAccessExprContext:
		objName := ""
		if id, ok := e.Expr().(*ast.IdentifierExprContext); ok {
			objName = id.QualifiedName().GetText()
		}
		if objName != "" {
			if sym := scope.ResolveQualified(objName); sym != nil && sym.Kind == symbols.KindEnum {
				return objName
			}
		}

		objType := InferExprType(e.Expr(), scope)
		if objType == "unknown" {
			return "unknown"
		}
		fieldName := e.ID().GetText()

		fmt.Fprintf(os.Stderr, "DEBUG: MemberAccess %s.%s (objType=%s)\n", e.Expr().GetText(), fieldName, objType)

		baseType := SplitBaseType(objType)
		sym := ResolveTypeSymbol(baseType, scope)
		if sym != nil {
			if sym.Kind == symbols.KindPackage {
				if sym.ImportedScope != nil {
					if member := sym.ImportedScope.Resolve(fieldName); member != nil {
						return member.Type
					}
				}
			} else if sym.Kind == symbols.KindStruct {
				if fieldType, ok := sym.Fields[fieldName]; ok {
					if len(sym.GenericParams) > 0 {
						if args := ResolveInstanceTypeArgs(objType, sym, func(t string) string {
							return ResolveTypeWithScope(t, scope)
						}, true); len(args) > 0 {
							return SubstituteTypeParams(fieldType, sym.GenericParams, args)
						}
					}
					return fieldType
				}
			}
		}
	}
	return "unknown"
}

func ResolveTypeWithScope(rawText string, scope *symbols.Scope) string {
	if rawText == "" {
		return ""
	}
	rawText = NormalizeTypeName(rawText)
	if strings.HasPrefix(rawText, "*") {
		return "*" + ResolveTypeWithScope(strings.TrimSpace(rawText[1:]), scope)
	}

	baseName := SplitBaseType(rawText)
	sym := ResolveTypeSymbol(baseName, scope) // Changed from scope.Resolve(baseName) to ResolveTypeSymbol to handle pkg.Type
	if sym != nil && (sym.Kind == symbols.KindStruct || sym.Kind == symbols.KindEnum || sym.Kind == symbols.KindTrait) {
		targetName := sym.GoName
		if sym.PackageName != "" { // Prepend package name if needed (Wait, how to check current package?)
			// For now just return ColumnName. Transpiler will handle imports.
			// Actually, ResolveTypeWithScope is used for string representation.
			// If it's pkg.Type, baseName is pkg.Type.
			// If we return just Config, we lose package info.
			// But ResolveTypeSymbol uses baseName.

			// If rawText was pkg.Config, baseName is pkg.Config.
			// sym is Config symbol.
			// targetName is Config.
			// We should probably return the qualified name if it was qualified.
			if strings.Contains(baseName, ".") {
				targetName = baseName
			}
		}
		if targetName == "" {
			targetName = baseName
		}

		if len(sym.GenericParams) > 0 {
			var provided []string
			start := strings.IndexAny(rawText, "[<")
			if start != -1 {
				end := strings.LastIndexAny(rawText, "]>")
				if end > start {
					inner := rawText[start+1 : end]
					provided = SplitTopLevelTypeArgs(inner)
				}
			}

			required := sym.GenericParams
			if len(provided) > len(required) {
				panic(fmt.Sprintf("泛型参数过多: %s, 期望 %d, 实际 %d", baseName, len(required), len(provided)))
			}

			finalArgs := make([]string, len(required))
			copy(finalArgs, provided)

			for i := len(provided); i < len(required); i++ {
				if required[i].DefaultType == "" {
					panic(fmt.Sprintf("类型引用 %s 缺失泛型参数 %s 且无默认值", baseName, required[i].Name))
				}
				finalArgs[i] = required[i].DefaultType
			}

			for i, arg := range finalArgs {
				finalArgs[i] = ResolveTypeWithScope(arg, scope)
			}

			return fmt.Sprintf("%s[%s]", targetName, strings.Join(finalArgs, ", "))
		}

		if rawText == baseName {
			return targetName
		}
	}

	return TranspileType(rawText)
}

func ParseTypeParams(ctx ast.ITypeParamsContext, scope *symbols.Scope) (string, string) {
	if ctx == nil {
		return "", ""
	}
	metas := ExtractGenericParamMeta(ctx, scope)
	return ParseGenericParamMeta(metas)
}

func ParseGenericParamMeta(metas []symbols.GenericParamMeta) (string, string) {
	if len(metas) == 0 {
		return "", ""
	}
	var defs, usages []string
	for _, meta := range metas {
		constraint := "any"
		if meta.ConstraintMyGo != "" {
			constraint = meta.ConstraintMyGo
		}
		defs = append(defs, fmt.Sprintf("%s %s", meta.Name, constraint))
		usages = append(usages, meta.Name)
	}
	return fmt.Sprintf("[%s]", strings.Join(defs, ", ")), fmt.Sprintf("[%s]", strings.Join(usages, ", "))
}

func IsSimpleIdentifier(expr interface{}) bool {
	if s, ok := expr.(string); ok {
		return !strings.Contains(s, ".")
	}
	if pt, ok := expr.(antlr.ParseTree); ok {
		_, ok := pt.(*ast.IdentifierExprContext)
		return ok
	}
	return false
}

func IsChannelType(typeName string) bool {
	t := NormalizeTypeName(typeName)
	return strings.HasPrefix(t, "chan<") && strings.HasSuffix(t, ">")
}

func GetChannelElementType(typeName string) string {
	t := NormalizeTypeName(typeName)
	if IsChannelType(t) {
		return t[5 : len(t)-1]
	}
	return ""
}

func InferImplTypeParamDefs(typeName string, scope *symbols.Scope) (string, string) {
	return "", ""
}
