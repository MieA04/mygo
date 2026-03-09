// Code generated from MyGo.g4 by ANTLR 4.13.1. DO NOT EDIT.

package ast // MyGo
import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by MyGoParser.
type MyGoVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by MyGoParser#program.
	VisitProgram(ctx *ProgramContext) interface{}

	// Visit a parse tree produced by MyGoParser#packageDecl.
	VisitPackageDecl(ctx *PackageDeclContext) interface{}

	// Visit a parse tree produced by MyGoParser#importStmt.
	VisitImportStmt(ctx *ImportStmtContext) interface{}

	// Visit a parse tree produced by MyGoParser#statement.
	VisitStatement(ctx *StatementContext) interface{}

	// Visit a parse tree produced by MyGoParser#spawnStmt.
	VisitSpawnStmt(ctx *SpawnStmtContext) interface{}

	// Visit a parse tree produced by MyGoParser#selectStmt.
	VisitSelectStmt(ctx *SelectStmtContext) interface{}

	// Visit a parse tree produced by MyGoParser#SelectReadBranch.
	VisitSelectReadBranch(ctx *SelectReadBranchContext) interface{}

	// Visit a parse tree produced by MyGoParser#SelectWriteBranch.
	VisitSelectWriteBranch(ctx *SelectWriteBranchContext) interface{}

	// Visit a parse tree produced by MyGoParser#SelectOtherBranch.
	VisitSelectOtherBranch(ctx *SelectOtherBranchContext) interface{}

	// Visit a parse tree produced by MyGoParser#selectRead.
	VisitSelectRead(ctx *SelectReadContext) interface{}

	// Visit a parse tree produced by MyGoParser#selectWrite.
	VisitSelectWrite(ctx *SelectWriteContext) interface{}

	// Visit a parse tree produced by MyGoParser#selectOther.
	VisitSelectOther(ctx *SelectOtherContext) interface{}

	// Visit a parse tree produced by MyGoParser#deferStmt.
	VisitDeferStmt(ctx *DeferStmtContext) interface{}

	// Visit a parse tree produced by MyGoParser#assignmentStmt.
	VisitAssignmentStmt(ctx *AssignmentStmtContext) interface{}

	// Visit a parse tree produced by MyGoParser#modifier.
	VisitModifier(ctx *ModifierContext) interface{}

	// Visit a parse tree produced by MyGoParser#typeParams.
	VisitTypeParams(ctx *TypeParamsContext) interface{}

	// Visit a parse tree produced by MyGoParser#typeParam.
	VisitTypeParam(ctx *TypeParamContext) interface{}

	// Visit a parse tree produced by MyGoParser#typeArgs.
	VisitTypeArgs(ctx *TypeArgsContext) interface{}

	// Visit a parse tree produced by MyGoParser#whereClause.
	VisitWhereClause(ctx *WhereClauseContext) interface{}

	// Visit a parse tree produced by MyGoParser#genericConstraint.
	VisitGenericConstraint(ctx *GenericConstraintContext) interface{}

	// Visit a parse tree produced by MyGoParser#structDecl.
	VisitStructDecl(ctx *StructDeclContext) interface{}

	// Visit a parse tree produced by MyGoParser#structField.
	VisitStructField(ctx *StructFieldContext) interface{}

	// Visit a parse tree produced by MyGoParser#enumDecl.
	VisitEnumDecl(ctx *EnumDeclContext) interface{}

	// Visit a parse tree produced by MyGoParser#enumVariant.
	VisitEnumVariant(ctx *EnumVariantContext) interface{}

	// Visit a parse tree produced by MyGoParser#fnDecl.
	VisitFnDecl(ctx *FnDeclContext) interface{}

	// Visit a parse tree produced by MyGoParser#paramList.
	VisitParamList(ctx *ParamListContext) interface{}

	// Visit a parse tree produced by MyGoParser#param.
	VisitParam(ctx *ParamContext) interface{}

	// Visit a parse tree produced by MyGoParser#PureTraitDecl.
	VisitPureTraitDecl(ctx *PureTraitDeclContext) interface{}

	// Visit a parse tree produced by MyGoParser#BindTraitDecl.
	VisitBindTraitDecl(ctx *BindTraitDeclContext) interface{}

	// Visit a parse tree produced by MyGoParser#traitFnDecl.
	VisitTraitFnDecl(ctx *TraitFnDeclContext) interface{}

	// Visit a parse tree produced by MyGoParser#bindTarget.
	VisitBindTarget(ctx *BindTargetContext) interface{}

	// Visit a parse tree produced by MyGoParser#traitBodyItem.
	VisitTraitBodyItem(ctx *TraitBodyItemContext) interface{}

	// Visit a parse tree produced by MyGoParser#SpecificBan.
	VisitSpecificBan(ctx *SpecificBanContext) interface{}

	// Visit a parse tree produced by MyGoParser#RepeatBan.
	VisitRepeatBan(ctx *RepeatBanContext) interface{}

	// Visit a parse tree produced by MyGoParser#returnStmt.
	VisitReturnStmt(ctx *ReturnStmtContext) interface{}

	// Visit a parse tree produced by MyGoParser#block.
	VisitBlock(ctx *BlockContext) interface{}

	// Visit a parse tree produced by MyGoParser#ifStmt.
	VisitIfStmt(ctx *IfStmtContext) interface{}

	// Visit a parse tree produced by MyGoParser#matchStmt.
	VisitMatchStmt(ctx *MatchStmtContext) interface{}

	// Visit a parse tree produced by MyGoParser#ValueMatchCase.
	VisitValueMatchCase(ctx *ValueMatchCaseContext) interface{}

	// Visit a parse tree produced by MyGoParser#TypeMatchCase.
	VisitTypeMatchCase(ctx *TypeMatchCaseContext) interface{}

	// Visit a parse tree produced by MyGoParser#DefaultMatchCase.
	VisitDefaultMatchCase(ctx *DefaultMatchCaseContext) interface{}

	// Visit a parse tree produced by MyGoParser#whileStmt.
	VisitWhileStmt(ctx *WhileStmtContext) interface{}

	// Visit a parse tree produced by MyGoParser#loopStmt.
	VisitLoopStmt(ctx *LoopStmtContext) interface{}

	// Visit a parse tree produced by MyGoParser#RangeForStmt.
	VisitRangeForStmt(ctx *RangeForStmtContext) interface{}

	// Visit a parse tree produced by MyGoParser#TraditionalForStmt.
	VisitTraditionalForStmt(ctx *TraditionalForStmtContext) interface{}

	// Visit a parse tree produced by MyGoParser#IteratorForStmt.
	VisitIteratorForStmt(ctx *IteratorForStmtContext) interface{}

	// Visit a parse tree produced by MyGoParser#forInit.
	VisitForInit(ctx *ForInitContext) interface{}

	// Visit a parse tree produced by MyGoParser#breakStmt.
	VisitBreakStmt(ctx *BreakStmtContext) interface{}

	// Visit a parse tree produced by MyGoParser#continueStmt.
	VisitContinueStmt(ctx *ContinueStmtContext) interface{}

	// Visit a parse tree produced by MyGoParser#SingleLetDecl.
	VisitSingleLetDecl(ctx *SingleLetDeclContext) interface{}

	// Visit a parse tree produced by MyGoParser#TupleLetDecl.
	VisitTupleLetDecl(ctx *TupleLetDeclContext) interface{}

	// Visit a parse tree produced by MyGoParser#ConstDecl.
	VisitConstDecl(ctx *ConstDeclContext) interface{}

	// Visit a parse tree produced by MyGoParser#typeList.
	VisitTypeList(ctx *TypeListContext) interface{}

	// Visit a parse tree produced by MyGoParser#typeType.
	VisitTypeType(ctx *TypeTypeContext) interface{}

	// Visit a parse tree produced by MyGoParser#qualifiedName.
	VisitQualifiedName(ctx *QualifiedNameContext) interface{}

	// Visit a parse tree produced by MyGoParser#exprStmt.
	VisitExprStmt(ctx *ExprStmtContext) interface{}

	// Visit a parse tree produced by MyGoParser#exprList.
	VisitExprList(ctx *ExprListContext) interface{}

	// Visit a parse tree produced by MyGoParser#StringExpr.
	VisitStringExpr(ctx *StringExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#ArrayIndexExpr.
	VisitArrayIndexExpr(ctx *ArrayIndexExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#FloatExpr.
	VisitFloatExpr(ctx *FloatExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#NotIsExpr.
	VisitNotIsExpr(ctx *NotIsExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#DerefExpr.
	VisitDerefExpr(ctx *DerefExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#LogicalAndExpr.
	VisitLogicalAndExpr(ctx *LogicalAndExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#PostfixExpr.
	VisitPostfixExpr(ctx *PostfixExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#IdentifierExpr.
	VisitIdentifierExpr(ctx *IdentifierExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#BinaryCompareExpr.
	VisitBinaryCompareExpr(ctx *BinaryCompareExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#ArrayLiteralExpr.
	VisitArrayLiteralExpr(ctx *ArrayLiteralExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#IsExpr.
	VisitIsExpr(ctx *IsExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#CastExpr.
	VisitCastExpr(ctx *CastExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#CallExpr.
	VisitCallExpr(ctx *CallExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#NotExpr.
	VisitNotExpr(ctx *NotExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#ThisExpr.
	VisitThisExpr(ctx *ThisExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#TernaryExpr.
	VisitTernaryExpr(ctx *TernaryExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#FuncCallExpr.
	VisitFuncCallExpr(ctx *FuncCallExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#NilExpr.
	VisitNilExpr(ctx *NilExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#LambdaExpr.
	VisitLambdaExpr(ctx *LambdaExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#StructLiteralExpr.
	VisitStructLiteralExpr(ctx *StructLiteralExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#PanicUnwrapExpr.
	VisitPanicUnwrapExpr(ctx *PanicUnwrapExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#TupleExpr.
	VisitTupleExpr(ctx *TupleExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#LogicalOrExpr.
	VisitLogicalOrExpr(ctx *LogicalOrExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#MulDivExpr.
	VisitMulDivExpr(ctx *MulDivExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#TryUnwrapExpr.
	VisitTryUnwrapExpr(ctx *TryUnwrapExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#AddrOfExpr.
	VisitAddrOfExpr(ctx *AddrOfExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#IntExpr.
	VisitIntExpr(ctx *IntExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#ParenExpr.
	VisitParenExpr(ctx *ParenExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#MemberAccessExpr.
	VisitMemberAccessExpr(ctx *MemberAccessExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#AddSubExpr.
	VisitAddSubExpr(ctx *AddSubExprContext) interface{}

	// Visit a parse tree produced by MyGoParser#MethodCallExpr.
	VisitMethodCallExpr(ctx *MethodCallExprContext) interface{}
}
