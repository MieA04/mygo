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

func (v *MyGoTranspiler) VisitStatement(ctx *ast.StatementContext) interface{} {
	child := ctx.GetChild(0)
	if child == nil {
		return nil
	}

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
		res := s.Accept(v)
		if res != nil {
			stmts = append(stmts, "\t"+res.(string))
		}
	}
	v.popScope()
	return "{\n" + strings.Join(stmts, "\n") + "\n}"
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

	isNative := strings.HasPrefix(colType, "[]") || colType == "string" || strings.HasPrefix(colType, "map")
	if colType == "unknown" {
		isNative = true
	}
	if sym := v.CurrentScope.Resolve(colType); sym != nil && sym.Kind == symbols.KindStruct {
		isNative = false
	}

	blockStr := ctx.Block().Accept(v).(string)
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
	rhs := ctx.Expr(1).Accept(v).(string)
	lhsType := types.InferExprType(ctx.Expr(0), v.CurrentScope)
	rhsType := types.InferExprType(ctx.Expr(1), v.CurrentScope)
	rhs = v.applyImplicitPromotion(rhs, rhsType, lhsType)
	return fmt.Sprintf("%s = %s", lhs, rhs)
}

func (v *MyGoTranspiler) VisitExprStmt(ctx *ast.ExprStmtContext) interface{} {
	return ctx.Expr().Accept(v)
}

func (v *MyGoTranspiler) VisitDeferStmt(ctx *ast.DeferStmtContext) interface{} {
	if ctx.Block() != nil {
		// defer { ... } -> defer func() { ... }()
		blockCode := ctx.Block().Accept(v).(string)
		return fmt.Sprintf("defer func() %s()", blockCode)
	}

	// defer expr;
	exprStmt := ctx.ExprStmt().(*ast.ExprStmtContext)
	expr := exprStmt.Expr()

	// Check if it is defer expr ?!!
	if panicExpr, ok := expr.(*ast.PanicUnwrapExprContext); ok {
		// defer expr ?!! -> defer func() { if err := expr; err != nil { panic(err) } }()
		// We need to extract the inner expression of panicExpr
		innerExpr := panicExpr.Expr()
		innerCode := innerExpr.Accept(v).(string)
		return fmt.Sprintf("defer func() {\n\tif err := %s; err != nil {\n\t\tpanic(err)\n\t}\n}()", innerCode)
	}

	// Standard defer expr
	code := expr.Accept(v).(string)
	return fmt.Sprintf("defer %s", code)
}

func (v *MyGoTranspiler) VisitLoopStmt(ctx *ast.LoopStmtContext) interface{} {
	v.loopDepth++
	block := ctx.Block().Accept(v).(string)
	v.loopDepth--
	return fmt.Sprintf("for %s", block)
}

func (v *MyGoTranspiler) VisitWhileStmt(ctx *ast.WhileStmtContext) interface{} {
	v.loopDepth++
	cond := ctx.Expr().Accept(v).(string)
	block := ctx.Block().Accept(v).(string)
	v.loopDepth--
	return fmt.Sprintf("for %s %s", cond, block)
}

func (v *MyGoTranspiler) VisitTraditionalForStmt(ctx *ast.TraditionalForStmtContext) interface{} {
	v.loopDepth++
	initStr, condStr, stepStr := "", "", ""
	if ctx.ForInit() != nil {
		initStr = ctx.ForInit().Accept(v).(string)
	}
	if ctx.GetCond() != nil {
		condStr = ctx.GetCond().Accept(v).(string)
	}
	if ctx.GetStep() != nil {
		stepStr = ctx.GetStep().Accept(v).(string)
	}
	block := ctx.Block().Accept(v).(string)
	v.loopDepth--
	return fmt.Sprintf("for %s; %s; %s %s", initStr, condStr, stepStr, block)
}

func (v *MyGoTranspiler) VisitRangeForStmt(ctx *ast.RangeForStmtContext) interface{} {
	v.loopDepth++
	id := ctx.ID().GetText()
	block := ctx.Block().Accept(v).(string)
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
			return fmt.Sprintf("%s := %s", id, ctx.Expr().Accept(v).(string))
		}
		return fmt.Sprintf("var %s", id)
	}
	return ctx.Expr().Accept(v).(string)
}

func (v *MyGoTranspiler) buildIfBranch(keyword string, expr ast.IExprContext, block ast.IBlockContext) string {
	blockCode := block.Accept(v).(string)
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
	exprs := ctx.AllExpr()
	blocks := ctx.AllBlock()
	res := v.buildIfBranch("if", exprs[0], blocks[0])
	for i := 1; i < len(exprs); i++ {
		res += " " + v.buildIfBranch("else if", exprs[i], blocks[i])
	}
	if len(blocks) > len(exprs) {
		res += fmt.Sprintf(" else %s", blocks[len(blocks)-1].Accept(v))
	}
	return res
}

func (v *MyGoTranspiler) VisitMatchStmt(ctx *ast.MatchStmtContext) interface{} {
	matchExpr := ctx.Expr()
	matchExprType := types.InferExprType(matchExpr, v.CurrentScope)
	matchTarget := matchExpr.Accept(v).(string)
	hasTypeMatchCase := false
	hasValueMatchCase := false
	allTypeOrDefaultCase := true
	for _, c := range ctx.AllMatchCase() {
		if _, ok := c.(*ast.TypeMatchCaseContext); ok {
			hasTypeMatchCase = true
			continue
		}
		if _, ok := c.(*ast.DefaultMatchCaseContext); ok {
			continue
		}
		if _, ok := c.(*ast.ValueMatchCaseContext); ok {
			hasValueMatchCase = true
		}
		allTypeOrDefaultCase = false
	}

	matchTargetVarName := ""
	if idCtx, ok := matchExpr.(*ast.IdentifierExprContext); ok {
		matchTargetVarName = idCtx.QualifiedName().GetText()
	}

	baseType := matchExprType
	genericArgs := ""
	if idx := strings.Index(matchExprType, "["); idx != -1 {
		baseType = matchExprType[:idx]
		genericArgs = matchExprType[idx:]
	}

	sym := v.CurrentScope.Resolve(baseType)
	if sym == nil {
		sym = v.CurrentScope.ResolveByGoName(baseType)
	}
	isEnum := sym != nil && sym.Kind == symbols.KindEnum

	prevTarget := v.currentMatchTarget
	prevExpr := v.currentMatchExpr
	prevExprNode := v.currentMatchNode
	prevVar := v.currentMatchVar
	prevUseTypeMatch := v.useTypeMatch
	prevCompileTimeTypeMatch := v.compileTimeTypeMatch
	prevMixedBoolMatch := v.mixedBoolMatch
	prevIsEnum := v.isEnumMatch
	prevGenArgs := v.enumGenericArgs

	compileTimeTypeMatch := false
	if hasTypeMatchCase && allTypeOrDefaultCase {
		compileTimeTypeMatch = true
		for _, c := range ctx.AllMatchCase() {
			typeCase, ok := c.(*ast.TypeMatchCaseContext)
			if !ok {
				continue
			}
			if _, ok := v.compileTimeTypeMatchCheck(matchExpr, typeCase.TypeType()); !ok {
				compileTimeTypeMatch = false
				break
			}
		}
	}

	mixedBoolMatch := !isEnum && hasTypeMatchCase && hasValueMatchCase

	v.currentMatchTarget = "_match_v"
	v.currentMatchExpr = matchTarget
	v.currentMatchNode = matchExpr
	v.currentMatchVar = matchTargetVarName
	v.useTypeMatch = !compileTimeTypeMatch && !isEnum && hasTypeMatchCase
	v.compileTimeTypeMatch = compileTimeTypeMatch
	v.mixedBoolMatch = mixedBoolMatch
	v.isEnumMatch = isEnum
	v.enumGenericArgs = genericArgs

	var res strings.Builder
	if v.compileTimeTypeMatch || v.mixedBoolMatch {
		res.WriteString("switch true {\n")
	} else if isEnum {
		res.WriteString(fmt.Sprintf("switch %s := %s.(type) {\n", v.currentMatchTarget, matchTarget))
	} else if v.useTypeMatch {
		res.WriteString(fmt.Sprintf("switch %s := any(%s).(type) {\n", v.currentMatchTarget, matchTarget))
	} else {
		res.WriteString(fmt.Sprintf("switch %s {\n", matchTarget))
	}

	for _, c := range ctx.AllMatchCase() {
		if valCase, ok := c.(*ast.ValueMatchCaseContext); ok {
			res.WriteString(v.visitValueMatchCaseWithShadowing(valCase, matchTargetVarName))
		} else {
			res.WriteString(c.Accept(v).(string) + "\n")
		}
	}
	res.WriteString("}")

	v.currentMatchTarget = prevTarget
	v.currentMatchExpr = prevExpr
	v.currentMatchNode = prevExprNode
	v.currentMatchVar = prevVar
	v.useTypeMatch = prevUseTypeMatch
	v.compileTimeTypeMatch = prevCompileTimeTypeMatch
	v.mixedBoolMatch = prevMixedBoolMatch
	v.isEnumMatch = prevIsEnum
	v.enumGenericArgs = prevGenArgs

	return res.String()
}

func (v *MyGoTranspiler) visitValueMatchCaseWithShadowing(ctx *ast.ValueMatchCaseContext, shadowVarName string) string {
	var cases []string
	var bindings []string

	if shadowVarName != "" && v.isEnumMatch {
		bindings = append(bindings, fmt.Sprintf("%s := %s", shadowVarName, v.currentMatchTarget))
	}

	for _, e := range ctx.AllExpr() {
		if v.isEnumMatch {
			if mc, ok := e.(*ast.MethodCallExprContext); ok {
				methodName := mc.ID().GetText()
				enumName := ""
				if idExpr, ok := mc.Expr().(*ast.IdentifierExprContext); ok {
					enumName = idExpr.QualifiedName().GetText()
				}
				if sym := v.CurrentScope.Resolve(enumName); sym != nil {
					enumName = sym.GoName
				}

				fullVariantName := fmt.Sprintf("%s_%s%s", enumName, methodName, v.enumGenericArgs)
				cases = append(cases, fullVariantName)

			} else if ma, ok := e.(*ast.MemberAccessExprContext); ok {
				enumName := ""
				if idExpr, ok := ma.Expr().(*ast.IdentifierExprContext); ok {
					enumName = idExpr.QualifiedName().GetText()
				}
				if sym := v.CurrentScope.Resolve(enumName); sym != nil {
					enumName = sym.GoName
				}
				variantName := ma.ID().GetText()

				fullVariantName := fmt.Sprintf("%s_%s%s", enumName, variantName, v.enumGenericArgs)
				cases = append(cases, fullVariantName)
			} else {
				cases = append(cases, e.Accept(v).(string))
			}
		} else {
			valueCode := e.Accept(v).(string)
			if v.mixedBoolMatch {
				cases = append(cases, fmt.Sprintf("%s == %s", v.currentMatchExpr, valueCode))
			} else {
				cases = append(cases, valueCode)
			}
		}
	}

	bodyStr := ""
	v.pushScope("match_case")
	if ctx.Block() != nil {
		bodyStr = strings.TrimSuffix(strings.TrimPrefix(ctx.Block().Accept(v).(string), "{\n"), "\n}")
	} else if ctx.Statement() != nil {
		bodyStr = "\t" + ctx.Statement().Accept(v).(string) + "\n"
	}
	v.popScope()

	bindingCode := ""
	if len(bindings) > 0 {
		bindingCode = strings.Join(bindings, "\n\t") + "\n\t"
	}

	return fmt.Sprintf("case %s:\n\t%s%s", strings.Join(cases, ", "), bindingCode, bodyStr)
}

func (v *MyGoTranspiler) VisitValueMatchCase(ctx *ast.ValueMatchCaseContext) interface{} {
	return v.visitValueMatchCaseWithShadowing(ctx, "")
}

func (v *MyGoTranspiler) VisitTypeMatchCase(ctx *ast.TypeMatchCaseContext) interface{} {
	bodyStr := ""
	if ctx.Block() != nil {
		bodyStr = strings.TrimSuffix(strings.TrimPrefix(ctx.Block().Accept(v).(string), "{\n"), "\n}")
	} else if ctx.Statement() != nil {
		bodyStr = "\t" + ctx.Statement().Accept(v).(string) + "\n"
	}

	bindingCode := ""
	if v.useTypeMatch && v.currentMatchVar != "" {
		bindingCode = fmt.Sprintf("%s := %s\n\t", v.currentMatchVar, v.currentMatchTarget)
	}
	if v.compileTimeTypeMatch {
		if result, ok := v.compileTimeTypeMatchCheck(v.currentMatchNode, ctx.TypeType()); ok {
			return fmt.Sprintf("case %t:\n\t%s", result, bodyStr)
		}
	}
	if v.mixedBoolMatch {
		if result, ok := v.compileTimeTypeMatchCheck(v.currentMatchNode, ctx.TypeType()); ok {
			return fmt.Sprintf("case %t:\n\t%s", result, bodyStr)
		}
		targetType := v.resolveType(ctx.TypeType())
		bindingCode = ""
		if v.currentMatchVar != "" {
			bindingCode = fmt.Sprintf("%s := any(%s).(%s)\n\t", v.currentMatchVar, v.currentMatchExpr, targetType)
		}
		cond := fmt.Sprintf("func() bool { _, _ok := any(%s).(%s); return _ok }()", v.currentMatchExpr, targetType)
		return fmt.Sprintf("case %s:\n\t%s%s", cond, bindingCode, bodyStr)
	}
	return fmt.Sprintf("case %s:\n\t%s%s", v.resolveType(ctx.TypeType()), bindingCode, bodyStr)
}

func (v *MyGoTranspiler) VisitDefaultMatchCase(ctx *ast.DefaultMatchCaseContext) interface{} {
	bodyStr := ""
	if ctx.Block() != nil {
		bodyStr = strings.TrimSuffix(strings.TrimPrefix(ctx.Block().Accept(v).(string), "{\n"), "\n}")
	} else if ctx.Statement() != nil {
		bodyStr = "\t" + ctx.Statement().Accept(v).(string) + "\n"
	}
	return fmt.Sprintf("default:\n%s", bodyStr)
}

func (v *MyGoTranspiler) VisitSpawnStmt(ctx *ast.SpawnStmtContext) interface{} {
	if ctx.Block() != nil {
		blockCode := ctx.Block().Accept(v).(string)
		return fmt.Sprintf("go func() %s()", blockCode)
	}
	exprCode := ctx.ExprStmt().Accept(v).(string)
	return fmt.Sprintf("go %s", exprCode)
}

func (v *MyGoTranspiler) VisitSelectStmt(ctx *ast.SelectStmtContext) interface{} {
	var cases []string
	for _, branch := range ctx.AllSelectBranch() {
		cases = append(cases, branch.Accept(v).(string))
	}
	return fmt.Sprintf("select {\n%s\n}", strings.Join(cases, "\n"))
}

func (v *MyGoTranspiler) VisitSelectReadBranch(ctx *ast.SelectReadBranchContext) interface{} {
	readExpr := ctx.SelectRead()
	chExpr := readExpr.Expr().Accept(v).(string)

	blockCode := ""
	if ctx.Block() != nil {
		blockCode = strings.TrimSuffix(strings.TrimPrefix(ctx.Block().Accept(v).(string), "{\n"), "\n}")
	}

	allIDs := readExpr.AllID()
	// The last ID is the method name "read", so exclude it
	vars := allIDs[:len(allIDs)-1]

	if len(vars) > 0 {
		if len(vars) == 1 {
			// case x = <-ch:
			return fmt.Sprintf("case %s := <-%s:\n%s", vars[0].GetText(), chExpr, blockCode)
		} else {
			// case x, ok = <-ch:
			return fmt.Sprintf("case %s, %s := <-%s:\n%s", vars[0].GetText(), vars[1].GetText(), chExpr, blockCode)
		}
	}
	// case <-ch:
	return fmt.Sprintf("case <-%s:\n%s", chExpr, blockCode)
}

func (v *MyGoTranspiler) VisitSelectWriteBranch(ctx *ast.SelectWriteBranchContext) interface{} {
	writeExpr := ctx.SelectWrite()
	chExpr := writeExpr.Expr(0).Accept(v).(string)
	valExpr := writeExpr.Expr(1).Accept(v).(string)

	blockCode := ""
	if ctx.Block() != nil {
		blockCode = strings.TrimSuffix(strings.TrimPrefix(ctx.Block().Accept(v).(string), "{\n"), "\n}")
	}

	return fmt.Sprintf("case %s <- %s:\n%s", chExpr, valExpr, blockCode)
}

func (v *MyGoTranspiler) VisitSelectOtherBranch(ctx *ast.SelectOtherBranchContext) interface{} {
	blockCode := ""
	if ctx.Block() != nil {
		blockCode = strings.TrimSuffix(strings.TrimPrefix(ctx.Block().Accept(v).(string), "{\n"), "\n}")
	}
	return fmt.Sprintf("default:\n%s", blockCode)
}
