package transpiler

import (
	"fmt"
	"os"
	"strings"

	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/compiler/symbols"
	"github.com/miea04/mygo/pkg/compiler/types"
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

	// If there's a block or statement, we generate an inline function or similar structure.
	// But TryUnwrapExpr is an expression, it must return a value.
	// The syntax `val ?! return` implies `val` is returned if success, else `return` executes.
	// This is tricky in Go because we can't easily embed statement in expression.
	// Go approach:
	// func() T { v, err := expr; if err != nil { return/panic }; return v }()
	// But `return` inside closure only returns from closure.
	// So `?! return` is control flow syntax, not just expression.
	// However, in MyGo parser, it is under `expr`.
	// If it's used as a statement: `let x = call() ?! return;`
	// This is actually handled in VarDecl or Assignment if we want to be smart.
	// But if we want it to work as expression:
	// `call() ?! panic("err")` -> `_mygo_unwrap_or_panic(call(), func(err error){ panic("err") })`
	// But `return` cannot be wrapped in func.

	// RFC-002:
	// let file = os.Open(...) ?! return;
	// Transpiled:
	// file, err := os.Open(...)
	// if err != nil { return err }

	// This requires context awareness. If `?!` is used in a Let/Assignment, we can split it.
	// If it's used in nested expression `process(open() ?! return)`, we can't easily transpile to Go without major restructuring (lifting).
	// For MVP Phase 3, maybe we restrict `?!` to be top-level in Let/Assignment or ExprStmt.

	// Current implementation of `PanicUnwrapExpr` (?! in old grammar, ?!! in new) uses helper.
	// New `?!` (TryUnwrap) allows custom block.

	// Since we can't easily lift statements in this visitor pattern without rewrite,
	// we might implement a limited version or use a special marker that `VisitSingleLetDecl` looks for.
	// OR, we assume `?!` is only allowed where we can handle it.

	// For now, let's implement `?!` (PanicUnwrap) style for `?!!` (already done above as VisitPanicUnwrapExpr).
	// For `?! block`, we need to implement `VisitTryUnwrapExpr`.

	// Wait, I am editing `expr.go` but `VisitTryUnwrapExpr` is not defined yet?
	// It was added to grammar. I need to implement it.

	// Implementation of TryUnwrapExpr logic:
	// Since TryUnwrapExpr is an expression, we need to return something that represents value.
	// But it has side effects (block/stmt execution).
	// If this expression is part of a larger expression, we can't easily emit statement.
	// We only support top-level usage in LetDecl or ExprStmt for now,
	// OR we return a placeholder that will be processed by the parent statement visitor.
	// However, the visitor pattern returns string.

	// Strategy: Return a special marker string that parent (VisitSingleLetDecl) can parse?
	// That's fragile.
	// Better Strategy:
	// Check if `ctx.Block()` or `ctx.Statement()` exists.
	// If yes, generate the custom error handling code.
	// If no, generate `return err` (default behavior).

	// But `exprStr` is the call itself.
	// If we are in `let x = foo() ?! { ... }`
	// We want to generate:
	// x, err := foo()
	// if err != nil { ... }

	// The `exprStr` returned here will be put into the `let` generation string.
	// If we return `foo() ?! { ... }` (raw), the parent can parse it?
	// Or we return a structured object? But interface{} is usually string.

	// Let's use a special prefix marker for now, as seen in `VisitSingleLetDecl` checking for `?!`.
	// `VisitSingleLetDecl` currently does: `if strings.Contains(exprGoCode, "?!") ...`
	// This is very hacky and only supports the suffix `?!`.
	// Now we have `?! block`.

	// Let's construct a marker string that encodes the block content.
	// Marker: `__MYGO_TRY_UNWRAP__<BaseExpr>__BLOCK__<BlockCode>__`
	// This is ugly but fits the current string-based transpiler architecture without major refactor.

	blockCode := ""
	if ctx.Block() != nil {
		blockCode = ctx.Block().Accept(v).(string)
	} else if ctx.Statement() != nil {
		blockCode = ctx.Statement().Accept(v).(string)
	} else {
		// Default: return err
		blockCode = "return err"
	}

	// Clean up block code (remove braces if needed or ensure it's a block)
	// Actually `VisitBlock` returns `{ ... }`. `VisitStatement` returns code.
	// We want `if err != nil { <blockCode> }`

	if ctx.Statement() != nil {
		blockCode = "{\n\t" + blockCode + "\n}"
	}

	return fmt.Sprintf("__MYGO_TRY_UNWRAP__%s__BLOCK__%s", exprStr, blockCode)
}

func (v *MyGoTranspiler) VisitPanicUnwrapExpr(ctx *ast.PanicUnwrapExprContext) interface{} {
	exprStr := ctx.Expr().Accept(v).(string)
	typ := types.InferExprType(ctx.Expr(), v.CurrentScope)

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
			fields = append(fields, fmt.Sprintf("%s: %s", fieldName, exprCode))
		}
	}

	return fmt.Sprintf("%s%s{%s}", structName, typeArgs, strings.Join(fields, ", "))
}

func (v *MyGoTranspiler) VisitFuncCallExpr(ctx *ast.FuncCallExprContext) interface{} {
	callee := ctx.QualifiedName().GetText()
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

	// Check if it's an enum variant constructor: Enum.Variant
	parts := strings.Split(callee, ".")
	// Check if it is a channel method call: ch.write(v) or ch.read() or ch.close()
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
			}
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

	// Channel operations
	objType := types.InferExprType(ctx.Expr(), v.CurrentScope)
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
			argType := types.InferExprType(ctx.ExprList().(*ast.ExprListContext).Expr(0), v.CurrentScope)
			if argType != "unknown" {
				typeArgs = "[" + argType + "]"
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
		return "{}"
	}
	var elems []string
	for _, eCtx := range ctx.ExprList().(*ast.ExprListContext).AllExpr() {
		elems = append(elems, eCtx.Accept(v).(string))
	}
	return "{" + strings.Join(elems, ", ") + "}"
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
	return ctx.QualifiedName().GetText()
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
