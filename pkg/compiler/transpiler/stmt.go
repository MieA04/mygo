package transpiler

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/compiler/symbols"
	"github.com/miea04/mygo/pkg/compiler/types"
)

func parseTryUnwrapMarker(code string) (kind string, exprCode string, innerType string, blockCode string, ok bool) {
	if strings.HasPrefix(code, tryUnwrapOptionPrefix) {
		parts := strings.Split(strings.TrimPrefix(code, tryUnwrapOptionPrefix), "__INNER__")
		if len(parts) != 2 {
			return "", "", "", "", false
		}
		tail := strings.Split(parts[1], "__BLOCK__")
		if len(tail) != 2 {
			return "", "", "", "", false
		}
		return "option", parts[0], tail[0], tail[1], true
	}
	if strings.HasPrefix(code, tryUnwrapPrefix) {
		parts := strings.Split(strings.TrimPrefix(code, tryUnwrapPrefix), "__BLOCK__")
		if len(parts) != 2 {
			return "", "", "", "", false
		}
		return "result", parts[0], "", parts[1], true
	}
	return "", "", "", "", false
}

func (v *MyGoTranspiler) VisitStatement(ctx *ast.StatementContext) interface{} {
	child := ctx.GetChild(0)
	if child == nil {
		return nil
	}

	// Explicit dispatch
	if s, ok := child.(*ast.StructDeclContext); ok {
		return v.VisitStructDecl(s)
	}
	if f, ok := child.(*ast.FnDeclContext); ok {
		return v.VisitFnDecl(f)
	}
	if e, ok := child.(*ast.EnumDeclContext); ok {
		return v.VisitEnumDecl(e)
	}
	if t, ok := child.(*ast.PureTraitDeclContext); ok {
		return v.VisitPureTraitDecl(t)
	}
	if t, ok := child.(*ast.BindTraitDeclContext); ok {
		return v.VisitBindTraitDecl(t)
	}

	if s, ok := child.(*ast.SingleLetDeclContext); ok {
		return v.VisitSingleLetDecl(s)
	}
	if s, ok := child.(*ast.TupleLetDeclContext); ok {
		return v.VisitTupleLetDecl(s)
	}
	if s, ok := child.(*ast.ConstDeclContext); ok {
		return v.VisitConstDecl(s)
	}

	if s, ok := child.(*ast.AssignmentStmtContext); ok {
		return v.VisitAssignmentStmt(s)
	}
	if s, ok := child.(*ast.ExprStmtContext); ok {
		return v.VisitExprStmt(s)
	}
	if s, ok := child.(*ast.IfStmtContext); ok {
		return v.VisitIfStmt(s)
	}
	if s, ok := child.(*ast.MatchStmtContext); ok {
		return v.VisitMatchStmt(s)
	}
	if s, ok := child.(*ast.WhileStmtContext); ok {
		return v.VisitWhileStmt(s)
	}
	if s, ok := child.(*ast.LoopStmtContext); ok {
		return v.VisitLoopStmt(s)
	}

	if s, ok := child.(*ast.RangeForStmtContext); ok {
		return v.VisitRangeForStmt(s)
	}
	if s, ok := child.(*ast.TraditionalForStmtContext); ok {
		return v.VisitTraditionalForStmt(s)
	}
	if s, ok := child.(*ast.IteratorForStmtContext); ok {
		return v.VisitIteratorForStmt(s)
	}

	if s, ok := child.(*ast.BreakStmtContext); ok {
		return v.VisitBreakStmt(s)
	}
	if s, ok := child.(*ast.ContinueStmtContext); ok {
		return v.VisitContinueStmt(s)
	}
	if s, ok := child.(*ast.ReturnStmtContext); ok {
		return v.VisitReturnStmt(s)
	}
	if s, ok := child.(*ast.DeferStmtContext); ok {
		return v.VisitDeferStmt(s)
	}
	if s, ok := child.(*ast.SpawnStmtContext); ok {
		return v.VisitSpawnStmt(s)
	}
	if s, ok := child.(*ast.SelectStmtContext); ok {
		return v.VisitSelectStmt(s)
	}

	// Fallback
	var res interface{}
	if tree, ok := child.(antlr.ParseTree); ok {
		res = tree.Accept(v)
	}

	if resStr, ok := res.(string); ok && v.CurrentFile != "" {
		line := ctx.GetStart().GetLine()
		path := filepath.ToSlash(v.CurrentFile)
		return fmt.Sprintf("//line %s:%d\n%s", path, line, resStr)
	}
	return res
}

func (v *MyGoTranspiler) VisitBlock(ctx *ast.BlockContext) interface{} {
	v.pushScope("block")
	var stmts []string
	for _, s := range ctx.AllStatement() {
		res := v.VisitStatement(s.(*ast.StatementContext))
		if res != nil {
			stmts = append(stmts, "\t"+res.(string))
		}
	}
	v.popScope()
	return "{\n" + strings.Join(stmts, "\n") + "\n}"
}

func (v *MyGoTranspiler) VisitSingleLetDecl(ctx *ast.SingleLetDeclContext) interface{} {
	name := ctx.ID().GetText()
	lhs := name
	rhs := ""
	rhsType := ""
	if ctx.Expr() != nil {
		rhs = ctx.Expr().Accept(v).(string)
		rhsType = types.InferExprType(ctx.Expr(), v.CurrentScope)
	}

	var typ string
	if ctx.TypeType() != nil {
		typ = v.resolveType(ctx.TypeType())
		v.CurrentScope.Define(name, name, symbols.KindVar, typ)
		if rhs != "" {
			if kind, exprCode, innerType, blockCode, ok := parseTryUnwrapMarker(rhs); ok {
				if kind == "result" {
					return fmt.Sprintf("var %s %s\n%s, err := %s\nif err != nil %s", lhs, typ, lhs, exprCode, blockCode)
				}
				return fmt.Sprintf("var %s %s\n__mygo_opt := %s\nswitch __mygo_opt_v := any(__mygo_opt).(type) {\ncase Option_Some[%s]:\n\t%s = __mygo_opt_v.Item1\ndefault:\n\t%s\n}", lhs, typ, exprCode, innerType, lhs, blockCode)
			}
			rhs = v.boxOptionExpr(rhs, rhsType, typ)
			return fmt.Sprintf("var %s %s = %s", lhs, typ, rhs)
		}
		return fmt.Sprintf("var %s %s", lhs, typ)
	}

	// Type inference
	if ctx.Expr() != nil {
		typ = rhsType
		if kind, _, innerType, _, ok := parseTryUnwrapMarker(rhs); ok && kind == "option" {
			typ = innerType
		}
	} else {
		typ = "interface{}"
	}
	v.CurrentScope.Define(name, name, symbols.KindVar, typ)

	if rhs != "" {
		if kind, exprCode, innerType, blockCode, ok := parseTryUnwrapMarker(rhs); ok {
			if kind == "result" {
				return fmt.Sprintf("%s, err := %s\nif err != nil %s", lhs, exprCode, blockCode)
			}
			return fmt.Sprintf("var %s %s\n__mygo_opt := %s\nswitch __mygo_opt_v := any(__mygo_opt).(type) {\ncase Option_Some[%s]:\n\t%s = __mygo_opt_v.Item1\ndefault:\n\t%s\n}", lhs, innerType, exprCode, innerType, lhs, blockCode)
		}
		// If inferred type is an Enum (Interface), we must explicitly declare it as interface
		// to allow re-assignment to other variants or type switching on the interface.
		baseType := types.SplitBaseType(typ)
		if sym := v.CurrentScope.Resolve(baseType); sym != nil && sym.Kind == symbols.KindEnum {
			goType := sym.GoName
			// Handle generics if necessary (simple for now)
			return fmt.Sprintf("var %s %s = %s", lhs, goType, rhs)
		}
		return fmt.Sprintf("%s := %s", lhs, rhs)
	}
	return fmt.Sprintf("var %s interface{}", lhs)
}

func (v *MyGoTranspiler) VisitTupleLetDecl(ctx *ast.TupleLetDeclContext) interface{} {
	var names []string
	for _, id := range ctx.AllID() {
		names = append(names, id.GetText())
	}
	rhs := ctx.Expr().Accept(v).(string)
	if kind, exprCode, _, blockCode, ok := parseTryUnwrapMarker(rhs); ok {
		if kind == "result" {
			lhs := strings.Join(names, ", ")
			return fmt.Sprintf("%s, err := %s\nif err != nil %s", lhs, exprCode, blockCode)
		}
		return rhs
	}

	// Define variables in scope with inferred type
	rhsType := types.InferExprType(ctx.Expr(), v.CurrentScope)

	// Handle tuple type string: "(T1, T2)"
	// Remove outer parens if present
	if strings.HasPrefix(rhsType, "(") && strings.HasSuffix(rhsType, ")") {
		inner := rhsType[1 : len(rhsType)-1]
		elemTypes := types.SplitTopLevelTypeArgs(inner)

		if len(elemTypes) == len(names) {
			for i, name := range names {
				v.CurrentScope.Define(name, name, symbols.KindVar, elemTypes[i])
			}
		} else {
			// Mismatch in count or parsing error, fallback to interface{}
			for _, name := range names {
				v.CurrentScope.Define(name, name, symbols.KindVar, "interface{}")
			}
		}
	} else {
		// Not a tuple type literal or single type (which shouldn't happen for tuple let unless it's a struct/array destructuring which we don't support here yet)
		// Or maybe InferExprType returned "unknown"
		for _, name := range names {
			v.CurrentScope.Define(name, name, symbols.KindVar, "interface{}")
		}
	}

	return fmt.Sprintf("%s := %s", strings.Join(names, ", "), rhs)
}

func (v *MyGoTranspiler) VisitConstDecl(ctx *ast.ConstDeclContext) interface{} {
	name := ctx.ID().GetText()
	rhs := ctx.Expr().Accept(v).(string)

	var typ string
	if ctx.TypeType() != nil {
		typ = v.resolveType(ctx.TypeType())
	} else {
		typ = types.InferExprType(ctx.Expr(), v.CurrentScope)
	}
	v.CurrentScope.Define(name, name, symbols.KindVar, typ)

	if ctx.TypeType() != nil {
		return fmt.Sprintf("const %s %s = %s", name, typ, rhs)
	}
	return fmt.Sprintf("const %s = %s", name, rhs)
}

func (v *MyGoTranspiler) VisitIteratorForStmt(ctx *ast.IteratorForStmtContext) interface{} {
	v.loopDepth++
	v.pushScope("for_iter")
	rawExprCtx := ctx.Expr()
	isIndexMacro := false
	if methodCall, ok := rawExprCtx.(*ast.MethodCallExprContext); ok {
		methodName := methodCall.ID().GetText()
		if methodName == "index" && methodCall.ExprList() == nil {
			isIndexMacro = true
			rawExprCtx = methodCall.Expr()
		} else if methodName == "item" && methodCall.ExprList() == nil {
			rawExprCtx = methodCall.Expr()
		}
	}
	colType := types.InferExprType(rawExprCtx, v.CurrentScope)
	kType, vType := types.ExtractCollectionTypes(colType, v.CurrentScope)

	var ids []string
	idCtxs := ctx.AllID()
	if len(idCtxs) == 1 {
		ids = append(ids, idCtxs[0].GetText())
		if isIndexMacro {
			v.CurrentScope.Define(ids[0], ids[0], symbols.KindVar, kType)
		} else {
			v.CurrentScope.Define(ids[0], ids[0], symbols.KindVar, vType)
		}
	} else if len(idCtxs) == 2 {
		ids = append(ids, idCtxs[0].GetText(), idCtxs[1].GetText())
		v.CurrentScope.Define(ids[0], ids[0], symbols.KindVar, kType)
		v.CurrentScope.Define(ids[1], ids[1], symbols.KindVar, vType)
	}

	baseExprStr := ctx.Expr().Accept(v).(string)
	if isIndexMacro || (!isIndexMacro && len(idCtxs) == 1 && strings.HasSuffix(baseExprStr, ".item()")) {
		baseExprStr = strings.TrimSuffix(baseExprStr, ".index()")
		baseExprStr = strings.TrimSuffix(baseExprStr, ".item()")
	}

	isNative := strings.HasPrefix(colType, "[]") || strings.HasSuffix(colType, "[]") || colType == "string" || strings.HasPrefix(colType, "map") || strings.HasPrefix(colType, "Map<")
	if colType == "unknown" {
		isNative = true
	}
	if sym := v.CurrentScope.Resolve(colType); sym != nil && sym.Kind == symbols.KindStruct {
		isNative = false
	}

	blockStr := ctx.Block().Accept(v).(string)
	if blockStr == "" {
		if b, ok := ctx.Block().(*ast.BlockContext); ok {
			res := v.VisitBlock(b)
			if s, ok := res.(string); ok {
				blockStr = s
			}
		}
	}

	blockBody := strings.TrimSuffix(strings.TrimPrefix(blockStr, "{\n"), "\n}")
	v.popScope()
	v.loopDepth--

	var result strings.Builder
	if isNative {
		leftSide := ""
		if len(ids) == 1 {
			if isIndexMacro {
				leftSide = ids[0]
			} else {
				leftSide = "_, " + ids[0]
			}
		} else {
			leftSide = ids[0] + ", " + ids[1]
		}
		result.WriteString(fmt.Sprintf("for %s := range %s {\n%s\n}", leftSide, baseExprStr, blockBody))
	} else {
		result.WriteString(fmt.Sprintf("for %s.has_next() {\n", baseExprStr))
		if len(ids) == 1 {
			if isIndexMacro {
				result.WriteString(fmt.Sprintf("\t%s, _ := %s.next()\n", ids[0], baseExprStr))
			} else {
				result.WriteString(fmt.Sprintf("\t_, %s := %s.next()\n", ids[0], baseExprStr))
			}
		} else {
			result.WriteString(fmt.Sprintf("\t%s, %s := %s.next()\n", ids[0], ids[1], baseExprStr))
		}
		result.WriteString(blockBody)
		result.WriteString("\n}")
	}
	return result.String()
}

func (v *MyGoTranspiler) VisitReturnStmt(ctx *ast.ReturnStmtContext) interface{} {
	if ctx.Expr() != nil {
		return fmt.Sprintf("return %s", ctx.Expr().Accept(v).(string))
	}
	return "return"
}

func (v *MyGoTranspiler) VisitAssignmentStmt(ctx *ast.AssignmentStmtContext) interface{} {
	lhs := ctx.Expr(0).Accept(v).(string)
	lhsType := types.InferExprType(ctx.Expr(0), v.CurrentScope)
	oldExpected := v.expectedType
	v.expectedType = v.toGoType(lhsType)
	rhs := ctx.Expr(1).Accept(v).(string)
	v.expectedType = oldExpected
	rhsType := types.InferExprType(ctx.Expr(1), v.CurrentScope)
	rhs = v.boxOptionExpr(rhs, rhsType, lhsType)
	rhs = v.applyImplicitPromotion(rhs, rhsType, lhsType)
	if kind, exprCode, innerType, blockCode, ok := parseTryUnwrapMarker(rhs); ok {
		if kind == "result" {
			// Use a block to avoid scope pollution and handle variable declaration
			return fmt.Sprintf("{\n\tvar _err error\n\t%s, _err = %s\n\tif _err != nil {\n\t\terr := _err\n\t\t_ = err\n\t\t%s\n\t}\n}", lhs, exprCode, blockCode)
		}
		return fmt.Sprintf("{\n\t__mygo_opt := %s\n\tswitch __mygo_opt_v := any(__mygo_opt).(type) {\n\tcase Option_Some[%s]:\n\t\t%s = __mygo_opt_v.Item1\n\tdefault:\n\t\t%s\n\t}\n}", exprCode, innerType, lhs, blockCode)
	}
	return fmt.Sprintf("%s = %s", lhs, rhs)
}

func (v *MyGoTranspiler) VisitExprStmt(ctx *ast.ExprStmtContext) interface{} {
	if methodCall, ok := ctx.Expr().(*ast.MethodCallExprContext); ok {
		methodName := methodCall.ID().GetText()
		objType := types.InferExprType(methodCall.Expr(), v.CurrentScope)
		if strings.HasSuffix(objType, "[]") || strings.HasPrefix(objType, "[]") {
			switch methodName {
			case "append", "insert", "remove_range":
				exprCode := ctx.Expr().Accept(v).(string)
				objCode := methodCall.Expr().Accept(v).(string)
				return fmt.Sprintf("%s = %s", objCode, exprCode)
			}
		}
	} else if funcCall, ok := ctx.Expr().(*ast.FuncCallExprContext); ok {
		callee := funcCall.QualifiedName().GetText()
		parts := strings.Split(callee, ".")
		if len(parts) > 1 {
			methodName := parts[len(parts)-1]
			objName := strings.Join(parts[:len(parts)-1], ".")
			sym := v.CurrentScope.ResolveQualified(objName)
			if sym == nil {
				sym = v.CurrentScope.Resolve(objName)
			}
			if sym != nil && sym.Kind == symbols.KindVar && (strings.HasSuffix(sym.Type, "[]") || strings.HasPrefix(sym.Type, "[]")) {
				switch methodName {
				case "append", "insert", "remove_range":
					exprCode := ctx.Expr().Accept(v).(string)
					return fmt.Sprintf("%s = %s", objName, exprCode)
				}
			}
		}
	}
	res := ctx.Expr().Accept(v)
	if resStr, ok := res.(string); ok {
		if kind, exprCode, innerType, blockCode, parsed := parseTryUnwrapMarker(resStr); parsed {
			if kind == "option" {
				return fmt.Sprintf("{\n\t__mygo_opt := %s\n\tswitch any(__mygo_opt).(type) {\n\tcase Option_Some[%s]:\n\tdefault:\n\t\t%s\n\t}\n}", exprCode, innerType, blockCode)
			}
			// Check if expression returns tuple (T, error) or just error
			// If type inference returns tuple/multiple values, handle it.
			// Otherwise assume single error return.
			typ := types.InferExprType(ctx.Expr(), v.CurrentScope)
			if strings.Contains(typ, ",") || strings.HasPrefix(typ, "(") {
				return fmt.Sprintf("if _, err := %s; err != nil %s", exprCode, blockCode)
			}
			return fmt.Sprintf("if err := %s; err != nil %s", exprCode, blockCode)
		}
	}
	return res
}

func (v *MyGoTranspiler) VisitDeferStmt(ctx *ast.DeferStmtContext) interface{} {
	if ctx.Block() != nil {
		blockCode := ""
		if b, ok := ctx.Block().(*ast.BlockContext); ok {
			res := v.VisitBlock(b)
			if s, ok := res.(string); ok {
				blockCode = s
			}
		} else {
			blockCode = ctx.Block().Accept(v).(string)
		}
		return fmt.Sprintf("defer func() %s()", blockCode)
	}
	exprStmt := ctx.ExprStmt().(*ast.ExprStmtContext)
	expr := exprStmt.Expr()
	if panicExpr, ok := expr.(*ast.PanicUnwrapExprContext); ok {
		innerExpr := panicExpr.Expr()
		innerCode := innerExpr.Accept(v).(string)
		return fmt.Sprintf("defer func() {\n\tif err := %s; err != nil {\n\t\tpanic(err)\n\t}\n}()", innerCode)
	}
	code := expr.Accept(v).(string)
	return fmt.Sprintf("defer %s", code)
}

func (v *MyGoTranspiler) VisitLoopStmt(ctx *ast.LoopStmtContext) interface{} {
	v.loopDepth++
	block := ""
	if b, ok := ctx.Block().(*ast.BlockContext); ok {
		res := v.VisitBlock(b)
		if s, ok := res.(string); ok {
			block = s
		}
	} else {
		block = ctx.Block().Accept(v).(string)
	}
	v.loopDepth--
	return fmt.Sprintf("for %s", block)
}

func (v *MyGoTranspiler) VisitWhileStmt(ctx *ast.WhileStmtContext) interface{} {
	v.loopDepth++
	cond := ctx.Expr().Accept(v).(string)
	block := ""
	if b, ok := ctx.Block().(*ast.BlockContext); ok {
		res := v.VisitBlock(b)
		if s, ok := res.(string); ok {
			block = s
		}
	} else {
		block = ctx.Block().Accept(v).(string)
	}
	v.loopDepth--
	return fmt.Sprintf("for %s %s", cond, block)
}

func (v *MyGoTranspiler) VisitTraditionalForStmt(ctx *ast.TraditionalForStmtContext) interface{} {
	v.loopDepth++
	v.pushScope("for_traditional")
	initStr, condStr, stepStr := "", "", ""
	if ctx.ForInit() != nil {
		if f, ok := ctx.ForInit().(*ast.ForInitContext); ok {
			initStr = v.VisitForInit(f).(string)
		} else {
			initStr = ctx.ForInit().Accept(v).(string)
		}
	}
	if ctx.GetCond() != nil {
		condStr = ctx.GetCond().Accept(v).(string)
	}
	if ctx.GetStep() != nil {
		stepStr = ctx.GetStep().Accept(v).(string)
	}
	block := ""
	if b, ok := ctx.Block().(*ast.BlockContext); ok {
		res := v.VisitBlock(b)
		if s, ok := res.(string); ok {
			block = s
		}
	} else {
		block = ctx.Block().Accept(v).(string)
	}
	v.popScope()
	v.loopDepth--
	return fmt.Sprintf("for %s; %s; %s %s", initStr, condStr, stepStr, block)
}

func (v *MyGoTranspiler) VisitRangeForStmt(ctx *ast.RangeForStmtContext) interface{} {
	v.loopDepth++
	v.pushScope("for_range")
	id := ctx.ID().GetText()
	v.CurrentScope.Define(id, id, symbols.KindVar, "int")

	block := ""
	if b, ok := ctx.Block().(*ast.BlockContext); ok {
		res := v.VisitBlock(b)
		if s, ok := res.(string); ok {
			block = s
		}
	} else {
		block = ctx.Block().Accept(v).(string)
	}
	v.popScope()
	v.loopDepth--
	return fmt.Sprintf("for %s := %s; %s <= %s; %s++ %s", id, ctx.Expr(0).Accept(v), id, ctx.Expr(1).Accept(v), id, block)
}

func (v *MyGoTranspiler) VisitBreakStmt(ctx *ast.BreakStmtContext) interface{} { return "break" }
func (v *MyGoTranspiler) VisitContinueStmt(ctx *ast.ContinueStmtContext) interface{} {
	return "continue"
}

func (v *MyGoTranspiler) VisitForInit(ctx *ast.ForInitContext) interface{} {
	if ctx.ID() != nil {
		id := ctx.ID().GetText()
		if ctx.Expr() != nil {
			inferredType := types.InferExprType(ctx.Expr(), v.CurrentScope)
			v.CurrentScope.Define(id, id, symbols.KindVar, inferredType)
			return fmt.Sprintf("%s := %s", id, ctx.Expr().Accept(v).(string))
		}
		v.CurrentScope.Define(id, id, symbols.KindVar, "unknown")
		return fmt.Sprintf("var %s", id)
	}
	return ctx.Expr().Accept(v).(string)
}

func (v *MyGoTranspiler) buildIfBranch(keyword string, expr ast.IExprContext, block ast.IBlockContext) string {
	v.pushScope("if_branch")

	// Type narrowing for "is" expression
	if isCtx, ok := expr.(*ast.IsExprContext); ok {
		varName := isCtx.Expr().GetText()
		if types.IsSimpleIdentifier(varName) {
			targetType := v.resolveType(isCtx.TypeType())
			sym := v.CurrentScope.Resolve(varName)
			goName := varName
			if sym != nil {
				goName = sym.GoName
			}
			v.CurrentScope.Define(varName, goName, symbols.KindVar, targetType)
		}
	}

	blockCode := ""
	if b, ok := block.(*ast.BlockContext); ok {
		res := v.VisitBlock(b)
		if s, ok := res.(string); ok {
			blockCode = s
		}
	} else {
		blockCode = block.Accept(v).(string)
	}

	v.popScope()

	if isCtx, ok := expr.(*ast.IsExprContext); ok {
		varName := isCtx.Expr().GetText()
		if types.IsSimpleIdentifier(varName) {
			if _, isCompileTime := v.compileTimeTraitCheck(isCtx.Expr(), isCtx.TypeType()); isCompileTime {
				return fmt.Sprintf("%s %s %s", keyword, expr.Accept(v), blockCode)
			}
			targetType := v.resolveType(isCtx.TypeType())
			if strings.HasPrefix(blockCode, "{\n") {
				blockCode = strings.Replace(blockCode, "{\n", fmt.Sprintf("{\n\t%s := _mygo_is_v\n", varName), 1)
			}
			return fmt.Sprintf("%s _mygo_is_v, _mygo_is_ok := any(%s).(%s); _mygo_is_ok %s", keyword, varName, targetType, blockCode)
		}
	}
	return fmt.Sprintf("%s %s %s", keyword, expr.Accept(v), blockCode)
}

func (v *MyGoTranspiler) VisitIfStmt(ctx *ast.IfStmtContext) interface{} {
	var sb strings.Builder
	sb.WriteString(v.buildIfBranch("if", ctx.Expr(0), ctx.Block(0)))

	exprIdx := 1
	blockIdx := 1

	totalExprs := len(ctx.AllExpr())
	totalBlocks := len(ctx.AllBlock())

	for exprIdx < totalExprs {
		sb.WriteString(v.buildIfBranch(" else if", ctx.Expr(exprIdx), ctx.Block(blockIdx)))
		exprIdx++
		blockIdx++
	}

	if blockIdx < totalBlocks {
		blockCode := ""
		if b, ok := ctx.Block(blockIdx).(*ast.BlockContext); ok {
			res := v.VisitBlock(b)
			if s, ok := res.(string); ok {
				blockCode = s
			}
		}
		sb.WriteString(" else " + blockCode)
	}

	return sb.String()
}

func (v *MyGoTranspiler) VisitMatchStmt(ctx *ast.MatchStmtContext) interface{} {
	exprCtx := ctx.Expr()
	exprStr := exprCtx.Accept(v).(string)
	matchExprType := types.InferExprType(exprCtx, v.CurrentScope)

	// Check if it's an Enum Match
	baseType := types.SplitBaseType(matchExprType)
	enumSym := v.CurrentScope.Resolve(baseType)
	if enumSym == nil {
		enumSym = v.CurrentScope.ResolveByGoName(baseType)
	}

	isEnumMatch := enumSym != nil && enumSym.Kind == symbols.KindEnum
	if isEnumMatch {
		return v.transpileEnumMatch(ctx, exprStr, matchExprType, enumSym)
	}

	// Normal Match (Switch)
	var cases []string
	isTypeSwitch := false

	// Check if any case is a type match
	for _, caseCtx := range ctx.AllMatchCase() {
		if _, ok := caseCtx.(*ast.TypeMatchCaseContext); ok {
			isTypeSwitch = true
			break
		}
	}

	if isTypeSwitch {
		exprStr = fmt.Sprintf("%s.(type)", exprStr)
	}

	for _, caseCtx := range ctx.AllMatchCase() {
		var caseHead string
		var blockCode string

		if typeCase, ok := caseCtx.(*ast.TypeMatchCaseContext); ok {
			typeStr := v.resolveType(typeCase.TypeType())
			caseHead = fmt.Sprintf("case %s:", typeStr)
			blockCode = v.visitBlockOrStmt(typeCase.Block(), typeCase.Statement())

		} else if valueCase, ok := caseCtx.(*ast.ValueMatchCaseContext); ok {
			// Check for variable binding pattern (single identifier, not a constant/variable in scope)
			isBinding := false
			bindingName := ""

			if len(valueCase.AllExpr()) == 1 {
				if idCtx, ok := valueCase.Expr(0).(*ast.IdentifierExprContext); ok {
					name := idCtx.QualifiedName().GetText()
					// If name is "_" or not resolvable in current scope, treat as binding/default
					// Note: "true" and "false" are usually keywords or constants
					if name != "true" && name != "false" {
						sym := v.CurrentScope.Resolve(name)
						if sym == nil {
							isBinding = true
							bindingName = name
						}
					}
				}
			}

			if isBinding {
				caseHead = "default:"
				blockCode = v.visitBlockOrStmt(valueCase.Block(), valueCase.Statement())
				if bindingName != "_" {
					// Inject variable binding: other := expr
					if strings.HasPrefix(blockCode, "{\n") {
						blockCode = strings.Replace(blockCode, "{\n", fmt.Sprintf("{\n\t%s := %s\n", bindingName, exprStr), 1)
					} else {
						blockCode = fmt.Sprintf("{\n\t%s := %s\n\t%s\n}", bindingName, exprStr, blockCode)
					}
				}
			} else {
				var exprs []string
				for _, e := range valueCase.AllExpr() {
					exprs = append(exprs, e.Accept(v).(string))
				}
				caseHead = fmt.Sprintf("case %s:", strings.Join(exprs, ", "))
				blockCode = v.visitBlockOrStmt(valueCase.Block(), valueCase.Statement())
			}

		} else if defCase, ok := caseCtx.(*ast.DefaultMatchCaseContext); ok {
			caseHead = "default:"
			blockCode = v.visitBlockOrStmt(defCase.Block(), defCase.Statement())
		}

		cases = append(cases, fmt.Sprintf("%s\n%s", caseHead, blockCode))
	}

	return fmt.Sprintf("switch %s {\n%s\n}", exprStr, strings.Join(cases, "\n"))
}

func (v *MyGoTranspiler) transpileEnumMatch(ctx *ast.MatchStmtContext, exprStr, matchType string, enumSym *symbols.Symbol) string {
	var sb strings.Builder
	matchVar := "_match_v" // TODO: Generate unique name if nested
	sb.WriteString(fmt.Sprintf("switch %s := (%s).(type) {\n", matchVar, exprStr))

	for _, caseCtx := range ctx.AllMatchCase() {
		if valCase, ok := caseCtx.(*ast.ValueMatchCaseContext); ok {
			// Expecting single pattern for destructuring, or multiple for simple matching (but type switch forbids multiple types in one case with var binding?)
			// Go: case T1, T2: v is interface{}
			// Go: case T1: v is T1.
			// So we must have separate cases if we want binding.
			// But MyGo syntax allows comma. We will handle first one if binding is needed.

			for _, patternExpr := range valCase.AllExpr() {
				variantName, args := v.parseEnumPattern(patternExpr, enumSym)
				if variantName != "" {
					fullVariantName := fmt.Sprintf("%s_%s", enumSym.GoName, variantName)
					// Handle generics: Result[int] -> Result_OK[int]
					if idx := strings.Index(matchType, "["); idx != -1 {
						typeArgs := matchType[idx:]
						fullVariantName += typeArgs
					}

					sb.WriteString(fmt.Sprintf("case %s:\n", fullVariantName))

					// Variable bindings
					// Only if arguments are provided and they are identifiers
					if len(args) > 0 {
						for i, arg := range args {
							// Bind: let arg = matchVar.Item{i+1}
							sb.WriteString(fmt.Sprintf("\t%s := %s.Item%d\n", arg, matchVar, i+1))
						}
					}
				} else {
					// Check if it is a binding variable (catch-all)
					// Only valid if it's the only expression in the case (to avoid ambiguity with bindings)
					if len(valCase.AllExpr()) == 1 {
						if idCtx, ok := patternExpr.(*ast.IdentifierExprContext); ok {
							name := idCtx.QualifiedName().GetText()
							if name != "true" && name != "false" {
								sym := v.CurrentScope.Resolve(name)
								if sym == nil {
									// It is a binding variable
									sb.WriteString("default:\n")
									if name != "_" {
										sb.WriteString(fmt.Sprintf("\t%s := %s\n", name, matchVar))
									}
									// We must break here because we've handled the case as default
									// The block will be appended after the loop over exprs, but wait...
									// The block is appended at line 679.
									// If we generate "default:\n binding := ...\n", then line 679 appends block.
									// That works.
								}
							}
						}
					}
				}
			}
			sb.WriteString(v.visitBlockOrStmt(valCase.Block(), valCase.Statement()) + "\n")

		} else if defCase, ok := caseCtx.(*ast.DefaultMatchCaseContext); ok {
			sb.WriteString("default:\n")
			sb.WriteString(v.visitBlockOrStmt(defCase.Block(), defCase.Statement()) + "\n")
		}
	}
	sb.WriteString("}")
	return sb.String()
}

func (v *MyGoTranspiler) parseEnumPattern(expr ast.IExprContext, enumSym *symbols.Symbol) (string, []string) {
	var variantName string
	var args []string

	if fc, ok := expr.(*ast.FuncCallExprContext); ok {
		// OK(val) or Result.OK(val)
		qName := fc.QualifiedName().GetText()
		if strings.Contains(qName, ".") {
			parts := strings.Split(qName, ".")
			// Check if prefix matches Enum name
			if parts[0] == enumSym.MyGoName || parts[0] == enumSym.GoName {
				variantName = parts[1]
			}
		} else {
			variantName = qName
		}

		if fc.ExprList() != nil {
			for _, arg := range fc.ExprList().(*ast.ExprListContext).AllExpr() {
				if id, ok := arg.(*ast.IdentifierExprContext); ok {
					args = append(args, id.GetText())
				}
			}
		}
	} else if ma, ok := expr.(*ast.MemberAccessExprContext); ok {
		// Result.OK
		// Check if LHS is Enum
		if id, ok := ma.Expr().(*ast.IdentifierExprContext); ok {
			name := id.GetText()
			if name == enumSym.MyGoName || name == enumSym.GoName {
				variantName = ma.ID().GetText()
			}
		}
	} else if id, ok := expr.(*ast.IdentifierExprContext); ok {
		// OK
		variantName = id.GetText()
		// Verify it's a variant
		if _, ok := enumSym.Variants[variantName]; !ok {
			// If not a variant, maybe it is a variable pattern?
			// But for now, we only support explicit variants
			return "", nil
		}
	}

	if variantName != "" {
		if _, ok := enumSym.Variants[variantName]; ok {
			return variantName, args
		}
	}
	return "", nil
}

func (v *MyGoTranspiler) visitBlockOrStmt(block ast.IBlockContext, stmt ast.IStatementContext) string {
	if block != nil {
		if b, ok := block.(*ast.BlockContext); ok {
			res := v.VisitBlock(b)
			if s, ok := res.(string); ok {
				// Strip braces for cleaner switch case body
				s = strings.TrimSpace(s)
				if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
					return s[1 : len(s)-1]
				}
				return s
			}
		}
	} else if stmt != nil {
		res := v.VisitStatement(stmt.(*ast.StatementContext))
		if s, ok := res.(string); ok {
			return s
		}
	}
	return ""
}

func (v *MyGoTranspiler) VisitSpawnStmt(ctx *ast.SpawnStmtContext) interface{} {
	block := ""
	if ctx.Block() != nil {
		if b, ok := ctx.Block().(*ast.BlockContext); ok {
			res := v.VisitBlock(b)
			if s, ok := res.(string); ok {
				block = s
			}
		}
	} else if ctx.ExprStmt() != nil {
		expr := ctx.ExprStmt().Expr().Accept(v).(string)
		block = fmt.Sprintf("{ %s }", expr)
	}
	return fmt.Sprintf("go func() %s()", block)
}

func (v *MyGoTranspiler) VisitSelectStmt(ctx *ast.SelectStmtContext) interface{} {
	var cases []string

	for _, branch := range ctx.AllSelectBranch() {
		var caseHead string
		var blockCode string

		v.pushScope("select_case")

		if readBranch, ok := branch.(*ast.SelectReadBranchContext); ok {
			read := readBranch.SelectRead()
			exprCtx := read.Expr()
			chanExpr := exprCtx.Accept(v).(string)

			var lhs string
			ids := read.AllID()
			if len(ids) > 0 {
				if len(ids) == 1 {
					name := ids[0].GetText()
					goName := types.FormatVisibility(name, nil)
					lhs = goName + " :="

					chanType := types.InferExprType(exprCtx, v.CurrentScope)
					elemType := "unknown"
					if strings.HasPrefix(chanType, "chan ") {
						elemType = strings.TrimPrefix(chanType, "chan ")
					}
					v.CurrentScope.Define(name, goName, symbols.KindVar, elemType)

				} else if len(ids) == 2 {
					name1 := ids[0].GetText()
					name2 := ids[1].GetText()
					goName1 := types.FormatVisibility(name1, nil)
					goName2 := types.FormatVisibility(name2, nil)

					lhs = fmt.Sprintf("%s, %s :=", goName1, goName2)

					chanType := types.InferExprType(exprCtx, v.CurrentScope)
					elemType := "unknown"
					if strings.HasPrefix(chanType, "chan ") {
						elemType = strings.TrimPrefix(chanType, "chan ")
					}

					v.CurrentScope.Define(name1, goName1, symbols.KindVar, elemType)
					v.CurrentScope.Define(name2, goName2, symbols.KindVar, "bool")
				}
			}

			if lhs != "" {
				caseHead = fmt.Sprintf("case %s <-%s:", lhs, chanExpr)
			} else {
				caseHead = fmt.Sprintf("case <-%s:", chanExpr)
			}

			if b, ok := readBranch.Block().(*ast.BlockContext); ok {
				res := v.VisitBlock(b)
				if s, ok := res.(string); ok {
					blockCode = s
				}
			}

		} else if writeBranch, ok := branch.(*ast.SelectWriteBranchContext); ok {
			write := writeBranch.SelectWrite()
			chanExpr := write.Expr(0).Accept(v).(string)
			valExpr := write.Expr(1).Accept(v).(string)
			caseHead = fmt.Sprintf("case %s <- %s:", chanExpr, valExpr)

			if b, ok := writeBranch.Block().(*ast.BlockContext); ok {
				res := v.VisitBlock(b)
				if s, ok := res.(string); ok {
					blockCode = s
				}
			}

		} else if otherBranch, ok := branch.(*ast.SelectOtherBranchContext); ok {
			caseHead = "default:"
			if b, ok := otherBranch.Block().(*ast.BlockContext); ok {
				res := v.VisitBlock(b)
				if s, ok := res.(string); ok {
					blockCode = s
				}
			}
		}

		v.popScope()
		cases = append(cases, fmt.Sprintf("%s\n%s", caseHead, blockCode))
	}

	return fmt.Sprintf("select {\n%s\n}", strings.Join(cases, "\n"))
}
