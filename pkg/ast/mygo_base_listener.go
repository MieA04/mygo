// Code generated from MyGo.g4 by ANTLR 4.13.1. DO NOT EDIT.

package ast // MyGo
import "github.com/antlr4-go/antlr/v4"

// BaseMyGoListener is a complete listener for a parse tree produced by MyGoParser.
type BaseMyGoListener struct{}

var _ MyGoListener = &BaseMyGoListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseMyGoListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseMyGoListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseMyGoListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseMyGoListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterProgram is called when production program is entered.
func (s *BaseMyGoListener) EnterProgram(ctx *ProgramContext) {}

// ExitProgram is called when production program is exited.
func (s *BaseMyGoListener) ExitProgram(ctx *ProgramContext) {}

// EnterPackageDecl is called when production packageDecl is entered.
func (s *BaseMyGoListener) EnterPackageDecl(ctx *PackageDeclContext) {}

// ExitPackageDecl is called when production packageDecl is exited.
func (s *BaseMyGoListener) ExitPackageDecl(ctx *PackageDeclContext) {}

// EnterBlockImport is called when production BlockImport is entered.
func (s *BaseMyGoListener) EnterBlockImport(ctx *BlockImportContext) {}

// ExitBlockImport is called when production BlockImport is exited.
func (s *BaseMyGoListener) ExitBlockImport(ctx *BlockImportContext) {}

// EnterSingleImport is called when production SingleImport is entered.
func (s *BaseMyGoListener) EnterSingleImport(ctx *SingleImportContext) {}

// ExitSingleImport is called when production SingleImport is exited.
func (s *BaseMyGoListener) ExitSingleImport(ctx *SingleImportContext) {}

// EnterImportSpec is called when production importSpec is entered.
func (s *BaseMyGoListener) EnterImportSpec(ctx *ImportSpecContext) {}

// ExitImportSpec is called when production importSpec is exited.
func (s *BaseMyGoListener) ExitImportSpec(ctx *ImportSpecContext) {}

// EnterStatement is called when production statement is entered.
func (s *BaseMyGoListener) EnterStatement(ctx *StatementContext) {}

// ExitStatement is called when production statement is exited.
func (s *BaseMyGoListener) ExitStatement(ctx *StatementContext) {}

// EnterAnnotationDecl is called when production annotationDecl is entered.
func (s *BaseMyGoListener) EnterAnnotationDecl(ctx *AnnotationDeclContext) {}

// ExitAnnotationDecl is called when production annotationDecl is exited.
func (s *BaseMyGoListener) ExitAnnotationDecl(ctx *AnnotationDeclContext) {}

// EnterAnnotationTarget is called when production annotationTarget is entered.
func (s *BaseMyGoListener) EnterAnnotationTarget(ctx *AnnotationTargetContext) {}

// ExitAnnotationTarget is called when production annotationTarget is exited.
func (s *BaseMyGoListener) ExitAnnotationTarget(ctx *AnnotationTargetContext) {}

// EnterSpawnStmt is called when production spawnStmt is entered.
func (s *BaseMyGoListener) EnterSpawnStmt(ctx *SpawnStmtContext) {}

// ExitSpawnStmt is called when production spawnStmt is exited.
func (s *BaseMyGoListener) ExitSpawnStmt(ctx *SpawnStmtContext) {}

// EnterSelectStmt is called when production selectStmt is entered.
func (s *BaseMyGoListener) EnterSelectStmt(ctx *SelectStmtContext) {}

// ExitSelectStmt is called when production selectStmt is exited.
func (s *BaseMyGoListener) ExitSelectStmt(ctx *SelectStmtContext) {}

// EnterSelectReadBranch is called when production SelectReadBranch is entered.
func (s *BaseMyGoListener) EnterSelectReadBranch(ctx *SelectReadBranchContext) {}

// ExitSelectReadBranch is called when production SelectReadBranch is exited.
func (s *BaseMyGoListener) ExitSelectReadBranch(ctx *SelectReadBranchContext) {}

// EnterSelectWriteBranch is called when production SelectWriteBranch is entered.
func (s *BaseMyGoListener) EnterSelectWriteBranch(ctx *SelectWriteBranchContext) {}

// ExitSelectWriteBranch is called when production SelectWriteBranch is exited.
func (s *BaseMyGoListener) ExitSelectWriteBranch(ctx *SelectWriteBranchContext) {}

// EnterSelectOtherBranch is called when production SelectOtherBranch is entered.
func (s *BaseMyGoListener) EnterSelectOtherBranch(ctx *SelectOtherBranchContext) {}

// ExitSelectOtherBranch is called when production SelectOtherBranch is exited.
func (s *BaseMyGoListener) ExitSelectOtherBranch(ctx *SelectOtherBranchContext) {}

// EnterSelectRead is called when production selectRead is entered.
func (s *BaseMyGoListener) EnterSelectRead(ctx *SelectReadContext) {}

// ExitSelectRead is called when production selectRead is exited.
func (s *BaseMyGoListener) ExitSelectRead(ctx *SelectReadContext) {}

// EnterSelectWrite is called when production selectWrite is entered.
func (s *BaseMyGoListener) EnterSelectWrite(ctx *SelectWriteContext) {}

// ExitSelectWrite is called when production selectWrite is exited.
func (s *BaseMyGoListener) ExitSelectWrite(ctx *SelectWriteContext) {}

// EnterSelectOther is called when production selectOther is entered.
func (s *BaseMyGoListener) EnterSelectOther(ctx *SelectOtherContext) {}

// ExitSelectOther is called when production selectOther is exited.
func (s *BaseMyGoListener) ExitSelectOther(ctx *SelectOtherContext) {}

// EnterDeferStmt is called when production deferStmt is entered.
func (s *BaseMyGoListener) EnterDeferStmt(ctx *DeferStmtContext) {}

// ExitDeferStmt is called when production deferStmt is exited.
func (s *BaseMyGoListener) ExitDeferStmt(ctx *DeferStmtContext) {}

// EnterAssignmentStmt is called when production assignmentStmt is entered.
func (s *BaseMyGoListener) EnterAssignmentStmt(ctx *AssignmentStmtContext) {}

// ExitAssignmentStmt is called when production assignmentStmt is exited.
func (s *BaseMyGoListener) ExitAssignmentStmt(ctx *AssignmentStmtContext) {}

// EnterModifier is called when production modifier is entered.
func (s *BaseMyGoListener) EnterModifier(ctx *ModifierContext) {}

// ExitModifier is called when production modifier is exited.
func (s *BaseMyGoListener) ExitModifier(ctx *ModifierContext) {}

// EnterTypeParams is called when production typeParams is entered.
func (s *BaseMyGoListener) EnterTypeParams(ctx *TypeParamsContext) {}

// ExitTypeParams is called when production typeParams is exited.
func (s *BaseMyGoListener) ExitTypeParams(ctx *TypeParamsContext) {}

// EnterTypeParam is called when production typeParam is entered.
func (s *BaseMyGoListener) EnterTypeParam(ctx *TypeParamContext) {}

// ExitTypeParam is called when production typeParam is exited.
func (s *BaseMyGoListener) ExitTypeParam(ctx *TypeParamContext) {}

// EnterTypeArgs is called when production typeArgs is entered.
func (s *BaseMyGoListener) EnterTypeArgs(ctx *TypeArgsContext) {}

// ExitTypeArgs is called when production typeArgs is exited.
func (s *BaseMyGoListener) ExitTypeArgs(ctx *TypeArgsContext) {}

// EnterWhereClause is called when production whereClause is entered.
func (s *BaseMyGoListener) EnterWhereClause(ctx *WhereClauseContext) {}

// ExitWhereClause is called when production whereClause is exited.
func (s *BaseMyGoListener) ExitWhereClause(ctx *WhereClauseContext) {}

// EnterGenericConstraint is called when production genericConstraint is entered.
func (s *BaseMyGoListener) EnterGenericConstraint(ctx *GenericConstraintContext) {}

// ExitGenericConstraint is called when production genericConstraint is exited.
func (s *BaseMyGoListener) ExitGenericConstraint(ctx *GenericConstraintContext) {}

// EnterAnnotationUsage is called when production annotationUsage is entered.
func (s *BaseMyGoListener) EnterAnnotationUsage(ctx *AnnotationUsageContext) {}

// ExitAnnotationUsage is called when production annotationUsage is exited.
func (s *BaseMyGoListener) ExitAnnotationUsage(ctx *AnnotationUsageContext) {}

// EnterStructDecl is called when production structDecl is entered.
func (s *BaseMyGoListener) EnterStructDecl(ctx *StructDeclContext) {}

// ExitStructDecl is called when production structDecl is exited.
func (s *BaseMyGoListener) ExitStructDecl(ctx *StructDeclContext) {}

// EnterStructField is called when production structField is entered.
func (s *BaseMyGoListener) EnterStructField(ctx *StructFieldContext) {}

// ExitStructField is called when production structField is exited.
func (s *BaseMyGoListener) ExitStructField(ctx *StructFieldContext) {}

// EnterEnumDecl is called when production enumDecl is entered.
func (s *BaseMyGoListener) EnterEnumDecl(ctx *EnumDeclContext) {}

// ExitEnumDecl is called when production enumDecl is exited.
func (s *BaseMyGoListener) ExitEnumDecl(ctx *EnumDeclContext) {}

// EnterEnumVariant is called when production enumVariant is entered.
func (s *BaseMyGoListener) EnterEnumVariant(ctx *EnumVariantContext) {}

// ExitEnumVariant is called when production enumVariant is exited.
func (s *BaseMyGoListener) ExitEnumVariant(ctx *EnumVariantContext) {}

// EnterFnDecl is called when production fnDecl is entered.
func (s *BaseMyGoListener) EnterFnDecl(ctx *FnDeclContext) {}

// ExitFnDecl is called when production fnDecl is exited.
func (s *BaseMyGoListener) ExitFnDecl(ctx *FnDeclContext) {}

// EnterParamList is called when production paramList is entered.
func (s *BaseMyGoListener) EnterParamList(ctx *ParamListContext) {}

// ExitParamList is called when production paramList is exited.
func (s *BaseMyGoListener) ExitParamList(ctx *ParamListContext) {}

// EnterParam is called when production param is entered.
func (s *BaseMyGoListener) EnterParam(ctx *ParamContext) {}

// ExitParam is called when production param is exited.
func (s *BaseMyGoListener) ExitParam(ctx *ParamContext) {}

// EnterPureTraitDecl is called when production PureTraitDecl is entered.
func (s *BaseMyGoListener) EnterPureTraitDecl(ctx *PureTraitDeclContext) {}

// ExitPureTraitDecl is called when production PureTraitDecl is exited.
func (s *BaseMyGoListener) ExitPureTraitDecl(ctx *PureTraitDeclContext) {}

// EnterBindTraitDecl is called when production BindTraitDecl is entered.
func (s *BaseMyGoListener) EnterBindTraitDecl(ctx *BindTraitDeclContext) {}

// ExitBindTraitDecl is called when production BindTraitDecl is exited.
func (s *BaseMyGoListener) ExitBindTraitDecl(ctx *BindTraitDeclContext) {}

// EnterTraitFnDecl is called when production traitFnDecl is entered.
func (s *BaseMyGoListener) EnterTraitFnDecl(ctx *TraitFnDeclContext) {}

// ExitTraitFnDecl is called when production traitFnDecl is exited.
func (s *BaseMyGoListener) ExitTraitFnDecl(ctx *TraitFnDeclContext) {}

// EnterBindTarget is called when production bindTarget is entered.
func (s *BaseMyGoListener) EnterBindTarget(ctx *BindTargetContext) {}

// ExitBindTarget is called when production bindTarget is exited.
func (s *BaseMyGoListener) ExitBindTarget(ctx *BindTargetContext) {}

// EnterTraitBodyItem is called when production traitBodyItem is entered.
func (s *BaseMyGoListener) EnterTraitBodyItem(ctx *TraitBodyItemContext) {}

// ExitTraitBodyItem is called when production traitBodyItem is exited.
func (s *BaseMyGoListener) ExitTraitBodyItem(ctx *TraitBodyItemContext) {}

// EnterSpecificBan is called when production SpecificBan is entered.
func (s *BaseMyGoListener) EnterSpecificBan(ctx *SpecificBanContext) {}

// ExitSpecificBan is called when production SpecificBan is exited.
func (s *BaseMyGoListener) ExitSpecificBan(ctx *SpecificBanContext) {}

// EnterRepeatBan is called when production RepeatBan is entered.
func (s *BaseMyGoListener) EnterRepeatBan(ctx *RepeatBanContext) {}

// ExitRepeatBan is called when production RepeatBan is exited.
func (s *BaseMyGoListener) ExitRepeatBan(ctx *RepeatBanContext) {}

// EnterReturnStmt is called when production returnStmt is entered.
func (s *BaseMyGoListener) EnterReturnStmt(ctx *ReturnStmtContext) {}

// ExitReturnStmt is called when production returnStmt is exited.
func (s *BaseMyGoListener) ExitReturnStmt(ctx *ReturnStmtContext) {}

// EnterBlock is called when production block is entered.
func (s *BaseMyGoListener) EnterBlock(ctx *BlockContext) {}

// ExitBlock is called when production block is exited.
func (s *BaseMyGoListener) ExitBlock(ctx *BlockContext) {}

// EnterIfStmt is called when production ifStmt is entered.
func (s *BaseMyGoListener) EnterIfStmt(ctx *IfStmtContext) {}

// ExitIfStmt is called when production ifStmt is exited.
func (s *BaseMyGoListener) ExitIfStmt(ctx *IfStmtContext) {}

// EnterMatchStmt is called when production matchStmt is entered.
func (s *BaseMyGoListener) EnterMatchStmt(ctx *MatchStmtContext) {}

// ExitMatchStmt is called when production matchStmt is exited.
func (s *BaseMyGoListener) ExitMatchStmt(ctx *MatchStmtContext) {}

// EnterValueMatchCase is called when production ValueMatchCase is entered.
func (s *BaseMyGoListener) EnterValueMatchCase(ctx *ValueMatchCaseContext) {}

// ExitValueMatchCase is called when production ValueMatchCase is exited.
func (s *BaseMyGoListener) ExitValueMatchCase(ctx *ValueMatchCaseContext) {}

// EnterTypeMatchCase is called when production TypeMatchCase is entered.
func (s *BaseMyGoListener) EnterTypeMatchCase(ctx *TypeMatchCaseContext) {}

// ExitTypeMatchCase is called when production TypeMatchCase is exited.
func (s *BaseMyGoListener) ExitTypeMatchCase(ctx *TypeMatchCaseContext) {}

// EnterDefaultMatchCase is called when production DefaultMatchCase is entered.
func (s *BaseMyGoListener) EnterDefaultMatchCase(ctx *DefaultMatchCaseContext) {}

// ExitDefaultMatchCase is called when production DefaultMatchCase is exited.
func (s *BaseMyGoListener) ExitDefaultMatchCase(ctx *DefaultMatchCaseContext) {}

// EnterWhileStmt is called when production whileStmt is entered.
func (s *BaseMyGoListener) EnterWhileStmt(ctx *WhileStmtContext) {}

// ExitWhileStmt is called when production whileStmt is exited.
func (s *BaseMyGoListener) ExitWhileStmt(ctx *WhileStmtContext) {}

// EnterLoopStmt is called when production loopStmt is entered.
func (s *BaseMyGoListener) EnterLoopStmt(ctx *LoopStmtContext) {}

// ExitLoopStmt is called when production loopStmt is exited.
func (s *BaseMyGoListener) ExitLoopStmt(ctx *LoopStmtContext) {}

// EnterRangeForStmt is called when production RangeForStmt is entered.
func (s *BaseMyGoListener) EnterRangeForStmt(ctx *RangeForStmtContext) {}

// ExitRangeForStmt is called when production RangeForStmt is exited.
func (s *BaseMyGoListener) ExitRangeForStmt(ctx *RangeForStmtContext) {}

// EnterTraditionalForStmt is called when production TraditionalForStmt is entered.
func (s *BaseMyGoListener) EnterTraditionalForStmt(ctx *TraditionalForStmtContext) {}

// ExitTraditionalForStmt is called when production TraditionalForStmt is exited.
func (s *BaseMyGoListener) ExitTraditionalForStmt(ctx *TraditionalForStmtContext) {}

// EnterIteratorForStmt is called when production IteratorForStmt is entered.
func (s *BaseMyGoListener) EnterIteratorForStmt(ctx *IteratorForStmtContext) {}

// ExitIteratorForStmt is called when production IteratorForStmt is exited.
func (s *BaseMyGoListener) ExitIteratorForStmt(ctx *IteratorForStmtContext) {}

// EnterForInit is called when production forInit is entered.
func (s *BaseMyGoListener) EnterForInit(ctx *ForInitContext) {}

// ExitForInit is called when production forInit is exited.
func (s *BaseMyGoListener) ExitForInit(ctx *ForInitContext) {}

// EnterBreakStmt is called when production breakStmt is entered.
func (s *BaseMyGoListener) EnterBreakStmt(ctx *BreakStmtContext) {}

// ExitBreakStmt is called when production breakStmt is exited.
func (s *BaseMyGoListener) ExitBreakStmt(ctx *BreakStmtContext) {}

// EnterContinueStmt is called when production continueStmt is entered.
func (s *BaseMyGoListener) EnterContinueStmt(ctx *ContinueStmtContext) {}

// ExitContinueStmt is called when production continueStmt is exited.
func (s *BaseMyGoListener) ExitContinueStmt(ctx *ContinueStmtContext) {}

// EnterSingleLetDecl is called when production SingleLetDecl is entered.
func (s *BaseMyGoListener) EnterSingleLetDecl(ctx *SingleLetDeclContext) {}

// ExitSingleLetDecl is called when production SingleLetDecl is exited.
func (s *BaseMyGoListener) ExitSingleLetDecl(ctx *SingleLetDeclContext) {}

// EnterTupleLetDecl is called when production TupleLetDecl is entered.
func (s *BaseMyGoListener) EnterTupleLetDecl(ctx *TupleLetDeclContext) {}

// ExitTupleLetDecl is called when production TupleLetDecl is exited.
func (s *BaseMyGoListener) ExitTupleLetDecl(ctx *TupleLetDeclContext) {}

// EnterConstDecl is called when production ConstDecl is entered.
func (s *BaseMyGoListener) EnterConstDecl(ctx *ConstDeclContext) {}

// ExitConstDecl is called when production ConstDecl is exited.
func (s *BaseMyGoListener) ExitConstDecl(ctx *ConstDeclContext) {}

// EnterTypeList is called when production typeList is entered.
func (s *BaseMyGoListener) EnterTypeList(ctx *TypeListContext) {}

// ExitTypeList is called when production typeList is exited.
func (s *BaseMyGoListener) ExitTypeList(ctx *TypeListContext) {}

// EnterTypeType is called when production typeType is entered.
func (s *BaseMyGoListener) EnterTypeType(ctx *TypeTypeContext) {}

// ExitTypeType is called when production typeType is exited.
func (s *BaseMyGoListener) ExitTypeType(ctx *TypeTypeContext) {}

// EnterQualifiedName is called when production qualifiedName is entered.
func (s *BaseMyGoListener) EnterQualifiedName(ctx *QualifiedNameContext) {}

// ExitQualifiedName is called when production qualifiedName is exited.
func (s *BaseMyGoListener) ExitQualifiedName(ctx *QualifiedNameContext) {}

// EnterExprStmt is called when production exprStmt is entered.
func (s *BaseMyGoListener) EnterExprStmt(ctx *ExprStmtContext) {}

// ExitExprStmt is called when production exprStmt is exited.
func (s *BaseMyGoListener) ExitExprStmt(ctx *ExprStmtContext) {}

// EnterExprList is called when production exprList is entered.
func (s *BaseMyGoListener) EnterExprList(ctx *ExprListContext) {}

// ExitExprList is called when production exprList is exited.
func (s *BaseMyGoListener) ExitExprList(ctx *ExprListContext) {}

// EnterStringExpr is called when production StringExpr is entered.
func (s *BaseMyGoListener) EnterStringExpr(ctx *StringExprContext) {}

// ExitStringExpr is called when production StringExpr is exited.
func (s *BaseMyGoListener) ExitStringExpr(ctx *StringExprContext) {}

// EnterArrayIndexExpr is called when production ArrayIndexExpr is entered.
func (s *BaseMyGoListener) EnterArrayIndexExpr(ctx *ArrayIndexExprContext) {}

// ExitArrayIndexExpr is called when production ArrayIndexExpr is exited.
func (s *BaseMyGoListener) ExitArrayIndexExpr(ctx *ArrayIndexExprContext) {}

// EnterFloatExpr is called when production FloatExpr is entered.
func (s *BaseMyGoListener) EnterFloatExpr(ctx *FloatExprContext) {}

// ExitFloatExpr is called when production FloatExpr is exited.
func (s *BaseMyGoListener) ExitFloatExpr(ctx *FloatExprContext) {}

// EnterNotIsExpr is called when production NotIsExpr is entered.
func (s *BaseMyGoListener) EnterNotIsExpr(ctx *NotIsExprContext) {}

// ExitNotIsExpr is called when production NotIsExpr is exited.
func (s *BaseMyGoListener) ExitNotIsExpr(ctx *NotIsExprContext) {}

// EnterDerefExpr is called when production DerefExpr is entered.
func (s *BaseMyGoListener) EnterDerefExpr(ctx *DerefExprContext) {}

// ExitDerefExpr is called when production DerefExpr is exited.
func (s *BaseMyGoListener) ExitDerefExpr(ctx *DerefExprContext) {}

// EnterLogicalAndExpr is called when production LogicalAndExpr is entered.
func (s *BaseMyGoListener) EnterLogicalAndExpr(ctx *LogicalAndExprContext) {}

// ExitLogicalAndExpr is called when production LogicalAndExpr is exited.
func (s *BaseMyGoListener) ExitLogicalAndExpr(ctx *LogicalAndExprContext) {}

// EnterPostfixExpr is called when production PostfixExpr is entered.
func (s *BaseMyGoListener) EnterPostfixExpr(ctx *PostfixExprContext) {}

// ExitPostfixExpr is called when production PostfixExpr is exited.
func (s *BaseMyGoListener) ExitPostfixExpr(ctx *PostfixExprContext) {}

// EnterIdentifierExpr is called when production IdentifierExpr is entered.
func (s *BaseMyGoListener) EnterIdentifierExpr(ctx *IdentifierExprContext) {}

// ExitIdentifierExpr is called when production IdentifierExpr is exited.
func (s *BaseMyGoListener) ExitIdentifierExpr(ctx *IdentifierExprContext) {}

// EnterBinaryCompareExpr is called when production BinaryCompareExpr is entered.
func (s *BaseMyGoListener) EnterBinaryCompareExpr(ctx *BinaryCompareExprContext) {}

// ExitBinaryCompareExpr is called when production BinaryCompareExpr is exited.
func (s *BaseMyGoListener) ExitBinaryCompareExpr(ctx *BinaryCompareExprContext) {}

// EnterArrayLiteralExpr is called when production ArrayLiteralExpr is entered.
func (s *BaseMyGoListener) EnterArrayLiteralExpr(ctx *ArrayLiteralExprContext) {}

// ExitArrayLiteralExpr is called when production ArrayLiteralExpr is exited.
func (s *BaseMyGoListener) ExitArrayLiteralExpr(ctx *ArrayLiteralExprContext) {}

// EnterIsExpr is called when production IsExpr is entered.
func (s *BaseMyGoListener) EnterIsExpr(ctx *IsExprContext) {}

// ExitIsExpr is called when production IsExpr is exited.
func (s *BaseMyGoListener) ExitIsExpr(ctx *IsExprContext) {}

// EnterCastExpr is called when production CastExpr is entered.
func (s *BaseMyGoListener) EnterCastExpr(ctx *CastExprContext) {}

// ExitCastExpr is called when production CastExpr is exited.
func (s *BaseMyGoListener) ExitCastExpr(ctx *CastExprContext) {}

// EnterCallExpr is called when production CallExpr is entered.
func (s *BaseMyGoListener) EnterCallExpr(ctx *CallExprContext) {}

// ExitCallExpr is called when production CallExpr is exited.
func (s *BaseMyGoListener) ExitCallExpr(ctx *CallExprContext) {}

// EnterNotExpr is called when production NotExpr is entered.
func (s *BaseMyGoListener) EnterNotExpr(ctx *NotExprContext) {}

// ExitNotExpr is called when production NotExpr is exited.
func (s *BaseMyGoListener) ExitNotExpr(ctx *NotExprContext) {}

// EnterThisExpr is called when production ThisExpr is entered.
func (s *BaseMyGoListener) EnterThisExpr(ctx *ThisExprContext) {}

// ExitThisExpr is called when production ThisExpr is exited.
func (s *BaseMyGoListener) ExitThisExpr(ctx *ThisExprContext) {}

// EnterTernaryExpr is called when production TernaryExpr is entered.
func (s *BaseMyGoListener) EnterTernaryExpr(ctx *TernaryExprContext) {}

// ExitTernaryExpr is called when production TernaryExpr is exited.
func (s *BaseMyGoListener) ExitTernaryExpr(ctx *TernaryExprContext) {}

// EnterFuncCallExpr is called when production FuncCallExpr is entered.
func (s *BaseMyGoListener) EnterFuncCallExpr(ctx *FuncCallExprContext) {}

// ExitFuncCallExpr is called when production FuncCallExpr is exited.
func (s *BaseMyGoListener) ExitFuncCallExpr(ctx *FuncCallExprContext) {}

// EnterNilExpr is called when production NilExpr is entered.
func (s *BaseMyGoListener) EnterNilExpr(ctx *NilExprContext) {}

// ExitNilExpr is called when production NilExpr is exited.
func (s *BaseMyGoListener) ExitNilExpr(ctx *NilExprContext) {}

// EnterLambdaExpr is called when production LambdaExpr is entered.
func (s *BaseMyGoListener) EnterLambdaExpr(ctx *LambdaExprContext) {}

// ExitLambdaExpr is called when production LambdaExpr is exited.
func (s *BaseMyGoListener) ExitLambdaExpr(ctx *LambdaExprContext) {}

// EnterStructLiteralExpr is called when production StructLiteralExpr is entered.
func (s *BaseMyGoListener) EnterStructLiteralExpr(ctx *StructLiteralExprContext) {}

// ExitStructLiteralExpr is called when production StructLiteralExpr is exited.
func (s *BaseMyGoListener) ExitStructLiteralExpr(ctx *StructLiteralExprContext) {}

// EnterPanicUnwrapExpr is called when production PanicUnwrapExpr is entered.
func (s *BaseMyGoListener) EnterPanicUnwrapExpr(ctx *PanicUnwrapExprContext) {}

// ExitPanicUnwrapExpr is called when production PanicUnwrapExpr is exited.
func (s *BaseMyGoListener) ExitPanicUnwrapExpr(ctx *PanicUnwrapExprContext) {}

// EnterTupleExpr is called when production TupleExpr is entered.
func (s *BaseMyGoListener) EnterTupleExpr(ctx *TupleExprContext) {}

// ExitTupleExpr is called when production TupleExpr is exited.
func (s *BaseMyGoListener) ExitTupleExpr(ctx *TupleExprContext) {}

// EnterLogicalOrExpr is called when production LogicalOrExpr is entered.
func (s *BaseMyGoListener) EnterLogicalOrExpr(ctx *LogicalOrExprContext) {}

// ExitLogicalOrExpr is called when production LogicalOrExpr is exited.
func (s *BaseMyGoListener) ExitLogicalOrExpr(ctx *LogicalOrExprContext) {}

// EnterMulDivExpr is called when production MulDivExpr is entered.
func (s *BaseMyGoListener) EnterMulDivExpr(ctx *MulDivExprContext) {}

// ExitMulDivExpr is called when production MulDivExpr is exited.
func (s *BaseMyGoListener) ExitMulDivExpr(ctx *MulDivExprContext) {}

// EnterTryUnwrapExpr is called when production TryUnwrapExpr is entered.
func (s *BaseMyGoListener) EnterTryUnwrapExpr(ctx *TryUnwrapExprContext) {}

// ExitTryUnwrapExpr is called when production TryUnwrapExpr is exited.
func (s *BaseMyGoListener) ExitTryUnwrapExpr(ctx *TryUnwrapExprContext) {}

// EnterAddrOfExpr is called when production AddrOfExpr is entered.
func (s *BaseMyGoListener) EnterAddrOfExpr(ctx *AddrOfExprContext) {}

// ExitAddrOfExpr is called when production AddrOfExpr is exited.
func (s *BaseMyGoListener) ExitAddrOfExpr(ctx *AddrOfExprContext) {}

// EnterQuoteExpr is called when production QuoteExpr is entered.
func (s *BaseMyGoListener) EnterQuoteExpr(ctx *QuoteExprContext) {}

// ExitQuoteExpr is called when production QuoteExpr is exited.
func (s *BaseMyGoListener) ExitQuoteExpr(ctx *QuoteExprContext) {}

// EnterInnerCallExpr is called when production InnerCallExpr is entered.
func (s *BaseMyGoListener) EnterInnerCallExpr(ctx *InnerCallExprContext) {}

// ExitInnerCallExpr is called when production InnerCallExpr is exited.
func (s *BaseMyGoListener) ExitInnerCallExpr(ctx *InnerCallExprContext) {}

// EnterIntExpr is called when production IntExpr is entered.
func (s *BaseMyGoListener) EnterIntExpr(ctx *IntExprContext) {}

// ExitIntExpr is called when production IntExpr is exited.
func (s *BaseMyGoListener) ExitIntExpr(ctx *IntExprContext) {}

// EnterParenExpr is called when production ParenExpr is entered.
func (s *BaseMyGoListener) EnterParenExpr(ctx *ParenExprContext) {}

// ExitParenExpr is called when production ParenExpr is exited.
func (s *BaseMyGoListener) ExitParenExpr(ctx *ParenExprContext) {}

// EnterMemberAccessExpr is called when production MemberAccessExpr is entered.
func (s *BaseMyGoListener) EnterMemberAccessExpr(ctx *MemberAccessExprContext) {}

// ExitMemberAccessExpr is called when production MemberAccessExpr is exited.
func (s *BaseMyGoListener) ExitMemberAccessExpr(ctx *MemberAccessExprContext) {}

// EnterAddSubExpr is called when production AddSubExpr is entered.
func (s *BaseMyGoListener) EnterAddSubExpr(ctx *AddSubExprContext) {}

// ExitAddSubExpr is called when production AddSubExpr is exited.
func (s *BaseMyGoListener) ExitAddSubExpr(ctx *AddSubExprContext) {}

// EnterMethodCallExpr is called when production MethodCallExpr is entered.
func (s *BaseMyGoListener) EnterMethodCallExpr(ctx *MethodCallExprContext) {}

// ExitMethodCallExpr is called when production MethodCallExpr is exited.
func (s *BaseMyGoListener) ExitMethodCallExpr(ctx *MethodCallExprContext) {}
