// Code generated from MyGo.g4 by ANTLR 4.13.1. DO NOT EDIT.

package ast // MyGo
import "github.com/antlr4-go/antlr/v4"

type BaseMyGoVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaseMyGoVisitor) VisitProgram(ctx *ProgramContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitPackageDecl(ctx *PackageDeclContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitBlockImport(ctx *BlockImportContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitSingleImport(ctx *SingleImportContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitImportSpec(ctx *ImportSpecContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitStatement(ctx *StatementContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitAnnotationDecl(ctx *AnnotationDeclContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitAnnotationTarget(ctx *AnnotationTargetContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitSpawnStmt(ctx *SpawnStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitSelectStmt(ctx *SelectStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitSelectReadBranch(ctx *SelectReadBranchContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitSelectWriteBranch(ctx *SelectWriteBranchContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitSelectOtherBranch(ctx *SelectOtherBranchContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitSelectRead(ctx *SelectReadContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitSelectWrite(ctx *SelectWriteContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitSelectOther(ctx *SelectOtherContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitDeferStmt(ctx *DeferStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitAssignmentStmt(ctx *AssignmentStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitModifier(ctx *ModifierContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitTypeParams(ctx *TypeParamsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitTypeParam(ctx *TypeParamContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitTypeArgs(ctx *TypeArgsContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitWhereClause(ctx *WhereClauseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitGenericConstraint(ctx *GenericConstraintContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitAnnotationUsage(ctx *AnnotationUsageContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitStructDecl(ctx *StructDeclContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitStructField(ctx *StructFieldContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitEnumDecl(ctx *EnumDeclContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitEnumVariant(ctx *EnumVariantContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitFnDecl(ctx *FnDeclContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitParamList(ctx *ParamListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitParam(ctx *ParamContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitPureTraitDecl(ctx *PureTraitDeclContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitBindTraitDecl(ctx *BindTraitDeclContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitTraitFnDecl(ctx *TraitFnDeclContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitBindTarget(ctx *BindTargetContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitTraitBodyItem(ctx *TraitBodyItemContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitSpecificBan(ctx *SpecificBanContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitRepeatBan(ctx *RepeatBanContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitReturnStmt(ctx *ReturnStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitBlock(ctx *BlockContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitIfStmt(ctx *IfStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitMatchStmt(ctx *MatchStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitValueMatchCase(ctx *ValueMatchCaseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitTypeMatchCase(ctx *TypeMatchCaseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitDefaultMatchCase(ctx *DefaultMatchCaseContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitWhileStmt(ctx *WhileStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitLoopStmt(ctx *LoopStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitRangeForStmt(ctx *RangeForStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitTraditionalForStmt(ctx *TraditionalForStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitIteratorForStmt(ctx *IteratorForStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitForInit(ctx *ForInitContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitBreakStmt(ctx *BreakStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitContinueStmt(ctx *ContinueStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitSingleLetDecl(ctx *SingleLetDeclContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitTupleLetDecl(ctx *TupleLetDeclContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitConstDecl(ctx *ConstDeclContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitTypeList(ctx *TypeListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitTypeType(ctx *TypeTypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitQualifiedName(ctx *QualifiedNameContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitExprStmt(ctx *ExprStmtContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitExprList(ctx *ExprListContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitStringExpr(ctx *StringExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitArrayIndexExpr(ctx *ArrayIndexExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitFloatExpr(ctx *FloatExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitNotIsExpr(ctx *NotIsExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitDerefExpr(ctx *DerefExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitLogicalAndExpr(ctx *LogicalAndExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitPostfixExpr(ctx *PostfixExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitIdentifierExpr(ctx *IdentifierExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitBinaryCompareExpr(ctx *BinaryCompareExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitArrayLiteralExpr(ctx *ArrayLiteralExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitIsExpr(ctx *IsExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitCastExpr(ctx *CastExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitCallExpr(ctx *CallExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitNotExpr(ctx *NotExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitThisExpr(ctx *ThisExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitTernaryExpr(ctx *TernaryExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitFuncCallExpr(ctx *FuncCallExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitNilExpr(ctx *NilExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitLambdaExpr(ctx *LambdaExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitStructLiteralExpr(ctx *StructLiteralExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitPanicUnwrapExpr(ctx *PanicUnwrapExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitTupleExpr(ctx *TupleExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitLogicalOrExpr(ctx *LogicalOrExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitMulDivExpr(ctx *MulDivExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitTryUnwrapExpr(ctx *TryUnwrapExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitAddrOfExpr(ctx *AddrOfExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitQuoteExpr(ctx *QuoteExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitInnerCallExpr(ctx *InnerCallExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitIntExpr(ctx *IntExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitParenExpr(ctx *ParenExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitMemberAccessExpr(ctx *MemberAccessExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitAddSubExpr(ctx *AddSubExprContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaseMyGoVisitor) VisitMethodCallExpr(ctx *MethodCallExprContext) interface{} {
	return v.VisitChildren(ctx)
}
