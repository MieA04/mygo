package transpiler

import (
	"fmt"
	"os"
	"strings"

	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/compiler/symbols"
	"github.com/miea04/mygo/pkg/compiler/types"
)

const (
	tryUnwrapPrefix       = "__MYGO_TRY_UNWRAP__"
	tryUnwrapOptionPrefix = "__MYGO_TRY_UNWRAP_OPTION__"
)

func (v *MyGoTranspiler) applyImplicitPromotion(exprCode, fromType, toType string) string {
	from := types.NormalizeTypeName(fromType)
	to := types.NormalizeTypeName(toType)
	if from == "" || to == "" || from == "unknown" || to == "unknown" {
		return exprCode
	}
	if from == to {
		return exprCode
	}
	if types.CanImplicitPromote(from, to) {
		return fmt.Sprintf("%s(%s)", to, exprCode)
	}
	return exprCode
}

func (v *MyGoTranspiler) VisitParenExpr(ctx *ast.ParenExprContext) interface{} {
	return "(" + ctx.Expr().Accept(v).(string) + ")"
}

func (v *MyGoTranspiler) VisitAddrOfExpr(ctx *ast.AddrOfExprContext) interface{} {
	return "&" + ctx.Expr().Accept(v).(string)
}

func (v *MyGoTranspiler) VisitDerefExpr(ctx *ast.DerefExprContext) interface{} {
	return "*" + ctx.Expr().Accept(v).(string)
}

func (v *MyGoTranspiler) VisitTryUnwrapExpr(ctx *ast.TryUnwrapExprContext) interface{} {
	exprStr := ctx.Expr().Accept(v).(string)
	exprType := types.InferExprType(ctx.Expr(), v.CurrentScope)
	if code, ok := v.buildOptionTryUnwrap(exprStr, exprType); ok {
		return code
	}

	// RFC-011: ?! is now error propagation operator (like Rust's ?)
	// It assumes the enclosing function returns a Result or error.
	// We generate a block that returns the error, wrapped in Result.Err if needed.
	// For Go compatibility (phase 1), we assume Result_Err is available or just return err.
	// The stmt.go visitor will wrap this in `if err != nil { ... }`.

	// Default propagation block
	blockCode := "return Result_Err(err)"

	return fmt.Sprintf("%s%s__BLOCK__%s", tryUnwrapPrefix, exprStr, blockCode)
}

func (v *MyGoTranspiler) VisitPanicUnwrapExpr(ctx *ast.PanicUnwrapExprContext) interface{} {
	exprStr := ctx.Expr().Accept(v).(string)
	typ := types.InferExprType(ctx.Expr(), v.CurrentScope)
	if code, ok := v.buildOptionPanicUnwrap(exprStr, typ); ok {
		return code
	}

	baseType := types.SplitBaseType(typ)
	sym := v.CurrentScope.ResolveQualified(baseType)
	if sym == nil && !strings.Contains(baseType, ".") {
		sym = v.CurrentScope.ResolveByGoName(baseType)
	}

	if sym != nil && sym.Kind == symbols.KindEnum {
		typeArgs := ""
		if idx := strings.Index(typ, "["); idx != -1 {
			typeArgs = typ[idx:]
		}
		return fmt.Sprintf("%s_unwrap%s(%s)", sym.GoName, typeArgs, exprStr)
	}
	return fmt.Sprintf("_mygo_must(%s)", exprStr)
}

func (v *MyGoTranspiler) buildOptionTryUnwrap(exprStr, exprType string) (string, bool) {
	if !types.IsOptionType(exprType) {
		return "", false
	}
	inner, ok := types.OptionInnerType(exprType)
	if !ok || inner == "" {
		return "", false
	}
	innerGoType := v.toGoType(inner)
	blockCode := fmt.Sprintf("return Option_None[%s]{}", innerGoType)
	return fmt.Sprintf("%s%s__INNER__%s__BLOCK__%s", tryUnwrapOptionPrefix, exprStr, innerGoType, blockCode), true
}

func (v *MyGoTranspiler) buildOptionPanicUnwrap(exprStr, exprType string) (string, bool) {
	if !types.IsOptionType(exprType) {
		return "", false
	}
	inner, ok := types.OptionInnerType(exprType)
	if !ok || inner == "" {
		return "", false
	}
	innerGoType := v.toGoType(inner)
	return fmt.Sprintf("func() %s { switch __mygo_opt := any(%s).(type) { case Option_Some[%s]: return __mygo_opt.Item1; default: panic(\"panic unwrap on None\") } }()", innerGoType, exprStr, innerGoType), true
}

func (v *MyGoTranspiler) VisitTupleExpr(ctx *ast.TupleExprContext) interface{} {
	var exprs []string
	for _, e := range ctx.AllExpr() {
		exprs = append(exprs, e.Accept(v).(string))
	}
	return strings.Join(exprs, ", ")
}

func (v *MyGoTranspiler) VisitStructLiteralExpr(ctx *ast.StructLiteralExprContext) interface{} {
	structName := ctx.QualifiedName().GetText()
	sym := v.CurrentScope.Resolve(structName)
	if sym != nil {
		structName = sym.GoName
	}
	typeArgs := ""
	if sym != nil && sym.Kind == symbols.KindStruct {
		if ctx.TypeArgs() == nil && len(sym.GenericParams) > 0 {
			inferredType := types.InferExprType(ctx, v.CurrentScope)
			if idx := strings.Index(inferredType, "["); idx != -1 {
				typeArgs = inferredType[idx:]
			} else {
				resolvedArgs, err := types.ResolveGenericArgs(sym, ctx.TypeArgs(), v.resolveTypeStr, false)
				if err == nil && len(resolvedArgs) > 0 {
					typeArgs = "[" + strings.Join(resolvedArgs, ", ") + "]"
				}
			}
		} else {
			resolvedArgs, err := types.ResolveGenericArgs(sym, ctx.TypeArgs(), v.resolveTypeStr, false)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Transpilation Error: %v\n", err)
				panic(err)
			}
			if len(resolvedArgs) > 0 {
				typeArgs = "[" + strings.Join(resolvedArgs, ", ") + "]"
			}
		}
	} else {
		if ctx.TypeArgs() != nil {
			typeArgs = types.ParseTypeArgs(ctx.TypeArgs())
		}
	}

	var fields []string
	allIDs := ctx.AllID()
	allExprs := ctx.AllExpr()

	for i := 0; i < len(allIDs); i++ {
		fieldName := allIDs[i].GetText()
		if i < len(allExprs) {
			exprCode := allExprs[i].Accept(v).(string)
			if fieldType, fieldTag, ok := v.resolveStructFieldTypeAndTag(sym, typeArgs, fieldName); ok {
				exprType := types.InferExprType(allExprs[i], v.CurrentScope)
				exprCode = v.boxTaggedOptionalFieldExpr(exprCode, exprType, fieldType, fieldTag)
			}
			fields = append(fields, fmt.Sprintf("%s: %s", fieldName, exprCode))
		}
	}

	return fmt.Sprintf("%s%s{%s}", structName, typeArgs, strings.Join(fields, ", "))
}

func (v *MyGoTranspiler) VisitFuncCallExpr(ctx *ast.FuncCallExprContext) interface{} {
	callee := ctx.QualifiedName().GetText()
	// fmt.Fprintf(os.Stderr, "DEBUG: VisitFuncCallExpr callee=%s\n", callee)

	// Check if it's a macro invocation
	if sym := v.CurrentScope.Resolve(callee); sym != nil && sym.Kind == symbols.KindAnnotation {
		fmt.Fprintf(os.Stderr, "DEBUG: Executing macro %s\n", callee)
		// Execute macro without target node (standalone invocation)
		macroResult := v.executeMacro(callee, nil)
		fmt.Fprintf(os.Stderr, "DEBUG: Macro result: %s\n", macroResult)
		return v.transpileMacroResult(macroResult)
	}

	if callee == "print" {
		callee = "fmt.Println"
	}

	// Channel creation: chan<T>(buffer) or chan<T>()
	if callee == "chan" {
		typeArgsStr := types.ParseTypeArgs(ctx.TypeArgs())
		elemType := "interface{}"
		if typeArgsStr != "" {
			if strings.HasPrefix(typeArgsStr, "<") && strings.HasSuffix(typeArgsStr, ">") {
				elemType = typeArgsStr[1 : len(typeArgsStr)-1]
			} else if strings.HasPrefix(typeArgsStr, "[") && strings.HasSuffix(typeArgsStr, "]") {
				elemType = typeArgsStr[1 : len(typeArgsStr)-1]
			}
		}

		// We need to resolve the element type to Go type
		elemType = v.toGoType(elemType)

		chanType := "chan " + elemType

		var args []string
		if ctx.ExprList() != nil {
			for _, eCtx := range ctx.ExprList().(*ast.ExprListContext).AllExpr() {
				args = append(args, eCtx.Accept(v).(string))
			}
		}

		if len(args) > 0 {
			return fmt.Sprintf("make(%s, %s)", chanType, args[0])
		}
		return fmt.Sprintf("make(%s)", chanType)
	}

	// Slice creation: slice<T>(len, cap) or slice<T>(len)
	if callee == "slice" {
		typeArgsStr := types.ParseTypeArgs(ctx.TypeArgs())
		elemType := "interface{}"
		if typeArgsStr != "" {
			if strings.HasPrefix(typeArgsStr, "<") && strings.HasSuffix(typeArgsStr, ">") {
				elemType = typeArgsStr[1 : len(typeArgsStr)-1]
			} else if strings.HasPrefix(typeArgsStr, "[") && strings.HasSuffix(typeArgsStr, "]") {
				elemType = typeArgsStr[1 : len(typeArgsStr)-1]
			}
		}

		// We need to resolve the element type to Go type
		elemType = v.toGoType(elemType)

		sliceType := "[]" + elemType

		var args []string
		if ctx.ExprList() != nil {
			for _, eCtx := range ctx.ExprList().(*ast.ExprListContext).AllExpr() {
				args = append(args, eCtx.Accept(v).(string))
			}
		}

		if len(args) == 1 {
			return fmt.Sprintf("make(%s, %s)", sliceType, args[0])
		} else if len(args) == 2 {
			return fmt.Sprintf("make(%s, %s, %s)", sliceType, args[0], args[1])
		}
		// Default to len 0 if no args provided
		return fmt.Sprintf("make(%s, 0)", sliceType)
	}

	// Map creation: Map<K, V>(cap) or Map<K, V>()
	if callee == "Map" {
		typeArgsStr := types.ParseTypeArgs(ctx.TypeArgs())
		// typeArgsStr usually comes as "[K, V]" from ParseTypeArgs if it was parsed as such
		// But here it might be "<K, V>" from the source text if not normalized yet?
		// types.ParseTypeArgs usually returns string representation.

		// We need to parse K and V.
		// If typeArgsStr is empty, it's an error (Maps need types).
		if typeArgsStr == "" {
			// Fallback or error?
			// For now, assume map[string]interface{} or similar? No, strict.
			// But maybe we can infer? For now let's just default to map[string]any if missing?
			// No, safer to require explicit types for now.
		}

		// Helper to extract K, V from "[K, V]" or "<K, V>"
		extractKV := func(s string) (string, string) {
			s = strings.TrimPrefix(s, "<")
			s = strings.TrimPrefix(s, "[")
			s = strings.TrimSuffix(s, ">")
			s = strings.TrimSuffix(s, "]")
			parts := types.SplitTopLevelTypeArgs(s)
			if len(parts) == 2 {
				return parts[0], parts[1]
			}
			return "string", "interface{}" // Default
		}

		kType, vType := extractKV(typeArgsStr)
		kType = v.toGoType(kType)
		vType = v.toGoType(vType)

		mapType := fmt.Sprintf("map[%s]%s", kType, vType)

		var args []string
		if ctx.ExprList() != nil {
			for _, eCtx := range ctx.ExprList().(*ast.ExprListContext).AllExpr() {
				args = append(args, eCtx.Accept(v).(string))
			}
		}

		if len(args) > 0 {
			return fmt.Sprintf("make(%s, %s)", mapType, args[0])
		}
		return fmt.Sprintf("make(%s)", mapType)
	}

	// Check if it's an enum variant constructor: Enum.Variant
	parts := strings.Split(callee, ".")
	// Check if it is a channel method call or slice method call
	if len(parts) > 1 {
		objName := strings.Join(parts[:len(parts)-1], ".")
		method := parts[len(parts)-1]

		sym := v.CurrentScope.ResolveQualified(objName)
		if sym == nil {
			sym = v.CurrentScope.Resolve(objName)
		}

		if sym != nil && sym.Kind == symbols.KindVar {
			if types.IsChannelType(sym.Type) {
				if method == "read" {
					return fmt.Sprintf("(<-%s)", objName)
				} else if method == "write" {
					if ctx.ExprList() != nil && len(ctx.ExprList().(*ast.ExprListContext).AllExpr()) > 0 {
						val := ctx.ExprList().(*ast.ExprListContext).Expr(0).Accept(v).(string)
						return fmt.Sprintf("%s <- %s", objName, val)
					}
				} else if method == "close" {
					return fmt.Sprintf("close(%s)", objName)
				}
			} else if strings.HasSuffix(sym.Type, "[]") || strings.HasPrefix(sym.Type, "[]") {
				// Slice method handling
				var sliceArgs []string
				if ctx.ExprList() != nil {
					for _, eCtx := range ctx.ExprList().(*ast.ExprListContext).AllExpr() {
						sliceArgs = append(sliceArgs, eCtx.Accept(v).(string))
					}
				}
				if res, ok := v.transpileSliceMethod(objName, sym.Type, method, sliceArgs); ok {
					return res
				}
			}
		} else {
			// Failed to resolve symbol
		}
	}

	if len(parts) == 2 {
		enumName := parts[0]
		variantName := parts[1]
		enumSym := v.CurrentScope.Resolve(enumName)
		if enumSym == nil {
			enumSym = v.CurrentScope.ResolveByGoName(enumName)
		}
		if enumSym != nil && enumSym.Kind == symbols.KindEnum {
			var args []string
			if ctx.ExprList() != nil {
				for _, eCtx := range ctx.ExprList().(*ast.ExprListContext).AllExpr() {
					args = append(args, eCtx.Accept(v).(string))
				}
			}
			structFields := make([]string, len(args))
			for i, arg := range args {
				structFields[i] = fmt.Sprintf("Item%d: %s", i+1, arg)
			}
			typeArgs := types.ParseTypeArgs(ctx.TypeArgs())
			// Basic type inference for Result
			if typeArgs == "" && (enumSym.GoName == "result" || enumSym.GoName == "Result") {
				// Check if Result is generic
				if len(enumSym.GenericParams) > 0 {
					if variantName == "Ok" && len(args) > 0 {
						arg0 := ctx.ExprList().(*ast.ExprListContext).Expr(0)
						argType := types.InferExprType(arg0, v.CurrentScope)
						if argType != "unknown" {
							typeArgs = "[" + argType + "]"
						}
					} else if v.expectedType != "" && (strings.HasPrefix(v.expectedType, "result[") || strings.HasPrefix(v.expectedType, "Result[")) {
						// Get from expected type
						idx := strings.Index(v.expectedType, "[")
						typeArgs = v.expectedType[idx:]
					}
				}
			}
			return fmt.Sprintf("%s_%s%s{%s}", enumSym.GoName, variantName, typeArgs, strings.Join(structFields, ", "))
		}
	}

	typeArgs := types.ParseTypeArgs(ctx.TypeArgs())

	if sym := v.CurrentScope.Resolve(callee); sym != nil && sym.Kind == symbols.KindFunc && len(sym.GenericParams) > 0 {
		if args, err := types.ResolveGenericArgs(sym, ctx.TypeArgs(), v.resolveTypeStr, false); err == nil {
			typeArgs = "[" + strings.Join(args, ", ") + "]"
		}
	}

	var args []string
	if ctx.ExprList() != nil {
		argExprs := ctx.ExprList().(*ast.ExprListContext).AllExpr()
		paramTypes := make([]string, 0)
		if sym := v.CurrentScope.Resolve(callee); sym != nil {
			if fnCtxRaw, ok := sym.Methods["__fn_decl"]; ok {
				if fnDecl, ok := fnCtxRaw.(*ast.FnDeclContext); ok && fnDecl.ParamList() != nil {
					for _, pCtx := range fnDecl.ParamList().(*ast.ParamListContext).AllParam() {
						p := pCtx.(*ast.ParamContext)
						paramTypes = append(paramTypes, v.resolveType(p.TypeType()))
					}
				}
			}
		}
		for i, eCtx := range argExprs {
			argCode := eCtx.Accept(v).(string)
			if i < len(paramTypes) {
				argType := types.InferExprType(eCtx, v.CurrentScope)
				argCode = v.applyImplicitPromotion(argCode, argType, paramTypes[i])
			}
			args = append(args, argCode)
		}
	}
	return fmt.Sprintf("%s%s(%s)", callee, typeArgs, strings.Join(args, ", "))
}

func (v *MyGoTranspiler) VisitMethodCallExpr(ctx *ast.MethodCallExprContext) interface{} {
	obj := ctx.Expr().Accept(v).(string)
	method := ctx.ID().GetText()
	typeArgs := types.ParseTypeArgs(ctx.TypeArgs())

	// Slice operations (RFC-006)
	objType := types.InferExprType(ctx.Expr(), v.CurrentScope)
	fmt.Fprintf(os.Stderr, "DEBUG: VisitMethodCallExpr obj=%s method=%s objType=%s\n", obj, method, objType)

	var sliceArgs []string
	if ctx.ExprList() != nil {
		for _, eCtx := range ctx.ExprList().(*ast.ExprListContext).AllExpr() {
			sliceArgs = append(sliceArgs, eCtx.Accept(v).(string))
		}
	}

	if res, ok := v.transpileSliceMethod(obj, objType, method, sliceArgs); ok {
		return res
	}

	if res, ok := v.transpileMapMethod(obj, objType, method, sliceArgs); ok {
		return res
	}

	// Channel operations
	if types.IsChannelType(objType) {
		if method == "read" {
			return fmt.Sprintf("(<-%s)", obj)
		} else if method == "write" {
			if ctx.ExprList() != nil && len(ctx.ExprList().(*ast.ExprListContext).AllExpr()) > 0 {
				val := ctx.ExprList().(*ast.ExprListContext).Expr(0).Accept(v).(string)
				// In Go, send is a statement, not an expression.
				// However, if we are in an expression context, we can't easily convert.
				// We assume semantic analysis ensures this is used as a statement.
				return fmt.Sprintf("%s <- %s", obj, val)
			}
		} else if method == "close" {
			return fmt.Sprintf("close(%s)", obj)
		}
	}

	// Check if obj is an enum
	var enumSym *symbols.Symbol
	if id, ok := ctx.Expr().(*ast.IdentifierExprContext); ok {
		name := id.QualifiedName().GetText()
		enumSym = v.CurrentScope.ResolveQualified(name)
		if enumSym == nil {
			enumSym = v.CurrentScope.Resolve(name)
		}
		if enumSym == nil {
			enumSym = v.CurrentScope.ResolveByGoName(name)
		}
		if enumSym != nil {
			// os.Stderr.WriteString("DEBUG: Found enumSym for " + name + " Kind=" + fmt.Sprintf("%v", enumSym.Kind) + "\n")
		} else {
			// os.Stderr.WriteString("DEBUG: FAILED to find enumSym for " + name + "\n")
		}
	}
	if enumSym != nil && enumSym.Kind == symbols.KindEnum {
		// It's an enum variant constructor
		var args []string
		if ctx.ExprList() != nil {
			for _, eCtx := range ctx.ExprList().(*ast.ExprListContext).AllExpr() {
				args = append(args, eCtx.Accept(v).(string))
			}
		}
		structFields := make([]string, len(args))
		for i, arg := range args {
			structFields[i] = fmt.Sprintf("Item%d: %s", i+1, arg)
		}
		// Try to infer type arguments if missing
		if typeArgs == "" && v.expectedType != "" {
			if strings.HasPrefix(v.expectedType, enumSym.GoName+"[") {
				typeArgs = v.expectedType[len(enumSym.GoName):]
			}
		}
		// If still missing and it's Result, try to infer from args
		if typeArgs == "" && enumSym.GoName == "Result" && len(args) > 0 {
			// Check if Result is generic
			if len(enumSym.GenericParams) > 0 {
				argType := types.InferExprType(ctx.ExprList().(*ast.ExprListContext).Expr(0), v.CurrentScope)
				if argType != "unknown" {
					typeArgs = "[" + argType + "]"
				}
			}
		}

		return fmt.Sprintf("%s_%s%s{%s}", enumSym.GoName, method, typeArgs, strings.Join(structFields, ", "))
	}

	var args []string
	if ctx.ExprList() != nil {
		for _, eCtx := range ctx.ExprList().(*ast.ExprListContext).AllExpr() {
			args = append(args, eCtx.Accept(v).(string))
		}
	}
	return fmt.Sprintf("%s.%s%s(%s)", obj, method, typeArgs, strings.Join(args, ", "))
}

func (v *MyGoTranspiler) VisitCallExpr(ctx *ast.CallExprContext) interface{} {
	callee := ctx.Expr().Accept(v).(string)
	// os.Stderr.WriteString("DEBUG: VisitCallExpr callee=" + callee + "\n")
	if callee == "print" {
		callee = "fmt.Println"
	}
	var args []string
	if ctx.ExprList() != nil {
		for _, eCtx := range ctx.ExprList().(*ast.ExprListContext).AllExpr() {
			args = append(args, eCtx.Accept(v).(string))
		}
	}
	return fmt.Sprintf("%s(%s)", callee, strings.Join(args, ", "))
}

func (v *MyGoTranspiler) typeIdentifierSymbol(expr ast.IExprContext) *symbols.Symbol {
	idExpr, ok := expr.(*ast.IdentifierExprContext)
	if !ok {
		return nil
	}
	name := idExpr.QualifiedName().GetText()
	sym := v.CurrentScope.Resolve(name)
	if sym == nil {
		sym = v.CurrentScope.ResolveByGoName(name)
	}
	if sym == nil {
		return nil
	}
	if sym.Kind != symbols.KindStruct && sym.Kind != symbols.KindTrait {
		return nil
	}
	return sym
}

func (v *MyGoTranspiler) compileTimeTraitCheck(expr ast.IExprContext, targetTypeCtx ast.ITypeTypeContext) (bool, bool) {
	targetType := types.ResolveTypeWithScope(targetTypeCtx.GetText(), v.CurrentScope)
	targetSym := types.ResolveTypeSymbol(targetType, v.CurrentScope)
	if targetSym == nil || targetSym.Kind != symbols.KindTrait {
		return false, false
	}
	sourceSym := v.typeIdentifierSymbol(expr)
	if sourceSym == nil {
		return false, false
	}
	return types.HasTraitRelationBySymbol(sourceSym, targetSym, v.CurrentScope), true
}

func (v *MyGoTranspiler) compileTimeTypeMatchCheck(expr ast.IExprContext, targetTypeCtx ast.ITypeTypeContext) (bool, bool) {
	sourceSym := v.typeIdentifierSymbol(expr)
	if sourceSym == nil {
		return false, false
	}
	targetType := types.ResolveTypeWithScope(targetTypeCtx.GetText(), v.CurrentScope)
	targetSym := types.ResolveTypeSymbol(targetType, v.CurrentScope)
	if targetSym == nil {
		return false, false
	}
	switch targetSym.Kind {
	case symbols.KindTrait:
		return types.HasTraitRelationBySymbol(sourceSym, targetSym, v.CurrentScope), true
	case symbols.KindStruct:
		return types.SymbolName(sourceSym) == types.SymbolName(targetSym), true
	default:
		return false, false
	}
}

func (v *MyGoTranspiler) VisitIsExpr(ctx *ast.IsExprContext) interface{} {
	if result, ok := v.compileTimeTraitCheck(ctx.Expr(), ctx.TypeType()); ok {
		if result {
			return "true"
		}
		return "false"
	}
	exprCode := ctx.Expr().Accept(v).(string)
	targetType := v.resolveType(ctx.TypeType())
	return fmt.Sprintf("func() bool { _, _ok := any(%s).(%s); return _ok }()", exprCode, targetType)
}

func (v *MyGoTranspiler) VisitNotIsExpr(ctx *ast.NotIsExprContext) interface{} {
	if result, ok := v.compileTimeTraitCheck(ctx.Expr(), ctx.TypeType()); ok {
		if result {
			return "false"
		}
		return "true"
	}
	exprCode := ctx.Expr().Accept(v).(string)
	targetType := v.resolveType(ctx.TypeType())
	return fmt.Sprintf("func() bool { _, _ok := any(%s).(%s); return !_ok }()", exprCode, targetType)
}

func (v *MyGoTranspiler) VisitTernaryExpr(ctx *ast.TernaryExprContext) interface{} {
	v.needsTernaryHelper = true
	return fmt.Sprintf("_mygo_ternary(%s, %s, %s)", ctx.Expr(0).Accept(v), ctx.Expr(1).Accept(v), ctx.Expr(2).Accept(v))
}

func (v *MyGoTranspiler) VisitPostfixExpr(ctx *ast.PostfixExprContext) interface{} {
	return fmt.Sprintf("%s%s", ctx.Expr().Accept(v).(string), ctx.GetOp().GetText())
}

func (v *MyGoTranspiler) VisitMulDivExpr(ctx *ast.MulDivExprContext) interface{} {
	leftType := types.InferExprType(ctx.Expr(0), v.CurrentScope)
	rightType := types.InferExprType(ctx.Expr(1), v.CurrentScope)
	leftCode := ctx.Expr(0).Accept(v).(string)
	rightCode := ctx.Expr(1).Accept(v).(string)
	op := ctx.GetOp().GetText()
	kind, _, methodName, negate, ok := types.ResolveBinaryOp(op, leftType, rightType, v.CurrentScope)
	if ok && kind == "overload_method" {
		call := fmt.Sprintf("%s.%s(%s)", leftCode, methodName, rightCode)
		if negate {
			return fmt.Sprintf("!(%s)", call)
		}
		return call
	}
	if commonType, ok := types.CommonNumericType(leftType, rightType); ok {
		leftCode = v.applyImplicitPromotion(leftCode, leftType, commonType)
		rightCode = v.applyImplicitPromotion(rightCode, rightType, commonType)
	}
	return fmt.Sprintf("%s %s %s", leftCode, op, rightCode)
}

func (v *MyGoTranspiler) VisitAddSubExpr(ctx *ast.AddSubExprContext) interface{} {
	leftType := types.InferExprType(ctx.Expr(0), v.CurrentScope)
	rightType := types.InferExprType(ctx.Expr(1), v.CurrentScope)
	leftCode := ctx.Expr(0).Accept(v).(string)
	rightCode := ctx.Expr(1).Accept(v).(string)
	op := ctx.GetOp().GetText()
	kind, _, methodName, negate, ok := types.ResolveBinaryOp(op, leftType, rightType, v.CurrentScope)
	if ok && kind == "overload_method" {
		call := fmt.Sprintf("%s.%s(%s)", leftCode, methodName, rightCode)
		if negate {
			return fmt.Sprintf("!(%s)", call)
		}
		return call
	}
	if commonType, ok := types.CommonNumericType(leftType, rightType); ok {
		leftCode = v.applyImplicitPromotion(leftCode, leftType, commonType)
		rightCode = v.applyImplicitPromotion(rightCode, rightType, commonType)
	}
	return fmt.Sprintf("%s %s %s", leftCode, op, rightCode)
}

func (v *MyGoTranspiler) VisitBinaryCompareExpr(ctx *ast.BinaryCompareExprContext) interface{} {
	leftType := types.InferExprType(ctx.Expr(0), v.CurrentScope)
	rightType := types.InferExprType(ctx.Expr(1), v.CurrentScope)
	leftCode := ctx.Expr(0).Accept(v).(string)
	rightCode := ctx.Expr(1).Accept(v).(string)
	op := ctx.GetOp().GetText()
	kind, _, methodName, negate, ok := types.ResolveBinaryOp(op, leftType, rightType, v.CurrentScope)
	if ok && kind == "overload_method" {
		call := fmt.Sprintf("%s.%s(%s)", leftCode, methodName, rightCode)
		if negate {
			return fmt.Sprintf("!(%s)", call)
		}
		return call
	}
	if commonType, isNumeric := types.CommonNumericType(leftType, rightType); isNumeric {
		leftCode = v.applyImplicitPromotion(leftCode, leftType, commonType)
		rightCode = v.applyImplicitPromotion(rightCode, rightType, commonType)
	}
	return fmt.Sprintf("%s %s %s", leftCode, op, rightCode)
}

func (v *MyGoTranspiler) VisitLogicalAndExpr(ctx *ast.LogicalAndExprContext) interface{} {
	return fmt.Sprintf("%s && %s", ctx.Expr(0).Accept(v), ctx.Expr(1).Accept(v))
}

func (v *MyGoTranspiler) VisitLogicalOrExpr(ctx *ast.LogicalOrExprContext) interface{} {
	return fmt.Sprintf("%s || %s", ctx.Expr(0).Accept(v), ctx.Expr(1).Accept(v))
}

func (v *MyGoTranspiler) VisitNotExpr(ctx *ast.NotExprContext) interface{} {
	return fmt.Sprintf("!(%s)", ctx.Expr().Accept(v))
}

func (v *MyGoTranspiler) VisitCastExpr(ctx *ast.CastExprContext) interface{} {
	targetType := v.resolveType(ctx.TypeType())
	oldExpected := v.expectedType
	v.expectedType = targetType
	exprCode := ctx.Expr().Accept(v).(string)
	v.expectedType = oldExpected

	exprType := types.InferExprType(ctx.Expr(), v.CurrentScope)
	normalizedTarget := types.NormalizeTypeName(targetType)
	if normalizedTarget == "any" || normalizedTarget == "interface{}" {
		return fmt.Sprintf("any(%s)", exprCode)
	}
	if types.IsAnyOrTraitType(targetType, v.CurrentScope) {
		return fmt.Sprintf("any(%s).(%s)", exprCode, targetType)
	}
	if types.IsPointerType(targetType) && types.IsAnyOrTraitType(exprType, v.CurrentScope) {
		return fmt.Sprintf("any(%s).(%s)", exprCode, targetType)
	}
	return fmt.Sprintf("%s(%s)", targetType, exprCode)
}

func (v *MyGoTranspiler) VisitArrayLiteralExpr(ctx *ast.ArrayLiteralExprContext) interface{} {
	if ctx.ExprList() == nil {
		if v.expectedType != "" {
			return v.expectedType + "{}"
		}
		// Empty array literal with no context? defaulting to []interface{}
		return "[]interface{}{}"
	}
	var elems []string
	for _, eCtx := range ctx.ExprList().(*ast.ExprListContext).AllExpr() {
		elems = append(elems, eCtx.Accept(v).(string))
	}

	content := "{" + strings.Join(elems, ", ") + "}"

	if v.expectedType != "" {
		return v.expectedType + content
	}

	// Fallback: infer from elements
	myGoType := types.SniffArrayType(ctx, v.CurrentScope)
	goType := v.toGoType(myGoType)
	return goType + content
}

func (v *MyGoTranspiler) VisitArrayIndexExpr(ctx *ast.ArrayIndexExprContext) interface{} {
	return fmt.Sprintf("%s[%s]", ctx.Expr(0).Accept(v), ctx.Expr(1).Accept(v))
}

func (v *MyGoTranspiler) VisitMemberAccessExpr(ctx *ast.MemberAccessExprContext) interface{} {
	obj := ctx.Expr().Accept(v).(string)
	if sym := v.CurrentScope.Resolve(obj); sym != nil && sym.Kind == symbols.KindEnum {
		typeArgs := ""
		if v.expectedType != "" {
			if strings.HasPrefix(v.expectedType, sym.GoName+"[") {
				typeArgs = v.expectedType[len(sym.GoName):]
			}
		}
		return fmt.Sprintf("%s_%s%s{}", sym.GoName, ctx.ID().GetText(), typeArgs)
	}
	return fmt.Sprintf("%s.%s", obj, ctx.ID().GetText())
}

func (v *MyGoTranspiler) VisitIntExpr(ctx *ast.IntExprContext) interface{} {
	return ctx.INT().GetText()
}
func (v *MyGoTranspiler) VisitStringExpr(ctx *ast.StringExprContext) interface{} {
	return ctx.STRING().GetText()
}
func (v *MyGoTranspiler) VisitFloatExpr(ctx *ast.FloatExprContext) interface{} {
	return ctx.FLOAT().GetText()
}
func (v *MyGoTranspiler) VisitIdentifierExpr(ctx *ast.IdentifierExprContext) interface{} {
	qn := ctx.QualifiedName()
	ids := qn.AllID()

	if len(ids) == 2 {
		first := ids[0].GetText()
		second := ids[1].GetText()
		sym := v.CurrentScope.Resolve(first)
		if sym != nil && sym.Kind == symbols.KindEnum {
			return fmt.Sprintf("%s_%s{}", sym.GoName, second)
		}
	}

	return qn.GetText()
}

func (v *MyGoTranspiler) VisitNilExpr(ctx *ast.NilExprContext) interface{} {
	return "nil"
}

func (v *MyGoTranspiler) VisitThisExpr(ctx *ast.ThisExprContext) interface{} {
	if v.currentBindVar != "" {
		return v.currentBindVar
	}
	return "this"
}

func (v *MyGoTranspiler) VisitLambdaExpr(ctx *ast.LambdaExprContext) interface{} {
	v.pushScope("lambda")
	var params []string
	if ctx.ParamList() != nil {
		for _, pCtx := range ctx.ParamList().(*ast.ParamListContext).AllParam() {
			p := pCtx.(*ast.ParamContext)
			pName := p.ID().GetText()
			pType := v.resolveType(p.TypeType())
			params = append(params, fmt.Sprintf("%s %s", pName, pType))
			v.CurrentScope.Define(pName, pName, symbols.KindVar, pType)
		}
	}
	returnTypeStr := ""
	if ctx.TypeType() != nil {
		returnTypeStr = " " + v.resolveType(ctx.TypeType())
	}
	blockStr := ctx.Block().Accept(v).(string)
	v.popScope()
	return fmt.Sprintf("func(%s)%s %s", strings.Join(params, ", "), returnTypeStr, blockStr)
}

func (v *MyGoTranspiler) transpileSliceMethod(obj, objType, method string, sliceArgs []string) (string, bool) {
	var elemType string
	if strings.HasSuffix(objType, "[]") {
		elemType = strings.TrimSuffix(objType, "[]")
	} else if strings.HasPrefix(objType, "[]") {
		elemType = strings.TrimPrefix(objType, "[]")
	} else {
		return "", false
	}

	switch method {
	case "len":
		return fmt.Sprintf("len(%s)", obj), true
	case "cap":
		return fmt.Sprintf("cap(%s)", obj), true
	case "is_empty":
		return fmt.Sprintf("(len(%s) == 0)", obj), true
	case "clear":
		// clear() empties the slice (len=0).
		// Use IIFE to mutate in place.
		goElemType := v.toGoType(elemType)
		goSliceType := "[]" + goElemType
		return fmt.Sprintf("func(s *%s) { *s = (*s)[:0] }(&%s)", goSliceType, obj), true
	case "append":
		return fmt.Sprintf("append(%s, %s)", obj, strings.Join(sliceArgs, ", ")), true
	case "insert":
		v.AddImport("slices")
		if len(sliceArgs) >= 2 {
			return fmt.Sprintf("slices.Insert(%s, %s, %s)", obj, sliceArgs[0], strings.Join(sliceArgs[1:], ", ")), true
		}
	case "remove":
		v.AddImport("slices")
		if len(sliceArgs) >= 1 {
			// remove(i) removes and returns the element.
			// We use an IIFE to mutate the slice in place and return the element.
			// This requires the slice object to be addressable.
			goElemType := v.toGoType(elemType)
			goSliceType := "[]" + goElemType
			idx := sliceArgs[0]
			return fmt.Sprintf("func(s *%s, i int) %s { v := (*s)[i]; *s = slices.Delete(*s, i, i+1); return v }(&%s, %s)", goSliceType, goElemType, obj, idx), true
		}
	case "pop":
		v.AddImport("slices")
		// pop() removes and returns the last element.
		// We use an IIFE to mutate the slice in place and return the element.
		goElemType := v.toGoType(elemType)
		goSliceType := "[]" + goElemType
		return fmt.Sprintf("func(s *%s) %s { i := len(*s)-1; v := (*s)[i]; *s = (*s)[:i]; return v }(&%s)", goSliceType, goElemType, obj), true
	case "remove_range":
		v.AddImport("slices")
		if len(sliceArgs) >= 2 {
			return fmt.Sprintf("slices.Delete(%s, %s, %s)", obj, sliceArgs[0], sliceArgs[1]), true
		}
	case "reverse":
		v.AddImport("slices")
		return fmt.Sprintf("slices.Reverse(%s)", obj), true
	case "sort":
		v.AddImport("slices")
		return fmt.Sprintf("slices.Sort(%s)", obj), true
	case "contains":
		v.AddImport("slices")
		if len(sliceArgs) >= 1 {
			return fmt.Sprintf("slices.Contains(%s, %s)", obj, sliceArgs[0]), true
		}
	case "index_of":
		v.AddImport("slices")
		if len(sliceArgs) >= 1 {
			return fmt.Sprintf("slices.Index(%s, %s)", obj, sliceArgs[0]), true
		}
	case "max":
		v.AddImport("slices")
		return fmt.Sprintf("slices.Max(%s)", obj), true
	case "min":
		v.AddImport("slices")
		return fmt.Sprintf("slices.Min(%s)", obj), true
	case "clone":
		v.AddImport("slices")
		return fmt.Sprintf("slices.Clone(%s)", obj), true
	}
	return "", false
}

func (v *MyGoTranspiler) transpileMapMethod(obj, objType, method string, args []string) (string, bool) {
	if !strings.HasPrefix(objType, "Map<") && !strings.HasPrefix(objType, "map[") {
		return "", false
	}

	kType, vType := types.ExtractCollectionTypes(objType, v.CurrentScope)
	goKType := v.toGoType(kType)
	goVType := v.toGoType(vType)

	switch method {
	case "keys":
		// keys() -> func(m map[K]V) []K { keys := make([]K, 0, len(m)); for k := range m { keys = append(keys, k) }; return keys }(obj)
		return fmt.Sprintf("func(m map[%s]%s) []%s { keys := make([]%s, 0, len(m)); for k := range m { keys = append(keys, k) }; return keys }(%s)", goKType, goVType, goKType, goKType, obj), true
	case "values":
		// values() -> func(m map[K]V) []V { values := make([]V, 0, len(m)); for _, v := range m { values = append(values, v) }; return values }(obj)
		return fmt.Sprintf("func(m map[%s]%s) []%s { values := make([]%s, 0, len(m)); for _, v := range m { values = append(values, v) }; return values }(%s)", goKType, goVType, goVType, goVType, obj), true
	case "has":
		if len(args) == 1 {
			// has(key) -> func(m map[K]V, k K) bool { _, ok := m[k]; return ok }(obj, key)
			return fmt.Sprintf("func(m map[%s]%s, k %s) bool { _, ok := m[k]; return ok }(%s, %s)", goKType, goVType, goKType, obj, args[0]), true
		}
	}
	return "", false
}
