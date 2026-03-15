// Code generated from MyGo.g4 by ANTLR 4.13.1. DO NOT EDIT.

package ast // MyGo
import "github.com/antlr4-go/antlr/v4"

// MyGoListener is a complete listener for a parse tree produced by MyGoParser.
type MyGoListener interface {
	antlr.ParseTreeListener

	// EnterProgram is called when entering the program production.
	EnterProgram(c *ProgramContext)

	// EnterPackageDecl is called when entering the packageDecl production.
	EnterPackageDecl(c *PackageDeclContext)

	// EnterBlockImport is called when entering the BlockImport production.
	EnterBlockImport(c *BlockImportContext)

	// EnterSingleImport is called when entering the SingleImport production.
	EnterSingleImport(c *SingleImportContext)

	// EnterImportSpec is called when entering the importSpec production.
	EnterImportSpec(c *ImportSpecContext)

	// EnterStatement is called when entering the statement production.
	EnterStatement(c *StatementContext)

	// EnterAnnotationDecl is called when entering the annotationDecl production.
	EnterAnnotationDecl(c *AnnotationDeclContext)

	// EnterAnnotationTarget is called when entering the annotationTarget production.
	EnterAnnotationTarget(c *AnnotationTargetContext)

	// EnterSpawnStmt is called when entering the spawnStmt production.
	EnterSpawnStmt(c *SpawnStmtContext)

	// EnterSelectStmt is called when entering the selectStmt production.
	EnterSelectStmt(c *SelectStmtContext)

	// EnterSelectReadBranch is called when entering the SelectReadBranch production.
	EnterSelectReadBranch(c *SelectReadBranchContext)

	// EnterSelectWriteBranch is called when entering the SelectWriteBranch production.
	EnterSelectWriteBranch(c *SelectWriteBranchContext)

	// EnterSelectOtherBranch is called when entering the SelectOtherBranch production.
	EnterSelectOtherBranch(c *SelectOtherBranchContext)

	// EnterSelectRead is called when entering the selectRead production.
	EnterSelectRead(c *SelectReadContext)

	// EnterSelectWrite is called when entering the selectWrite production.
	EnterSelectWrite(c *SelectWriteContext)

	// EnterSelectOther is called when entering the selectOther production.
	EnterSelectOther(c *SelectOtherContext)

	// EnterDeferStmt is called when entering the deferStmt production.
	EnterDeferStmt(c *DeferStmtContext)

	// EnterAssignmentStmt is called when entering the assignmentStmt production.
	EnterAssignmentStmt(c *AssignmentStmtContext)

	// EnterModifier is called when entering the modifier production.
	EnterModifier(c *ModifierContext)

	// EnterTypeParams is called when entering the typeParams production.
	EnterTypeParams(c *TypeParamsContext)

	// EnterTypeParam is called when entering the typeParam production.
	EnterTypeParam(c *TypeParamContext)

	// EnterTypeArgs is called when entering the typeArgs production.
	EnterTypeArgs(c *TypeArgsContext)

	// EnterWhereClause is called when entering the whereClause production.
	EnterWhereClause(c *WhereClauseContext)

	// EnterGenericConstraint is called when entering the genericConstraint production.
	EnterGenericConstraint(c *GenericConstraintContext)

	// EnterAnnotationUsage is called when entering the annotationUsage production.
	EnterAnnotationUsage(c *AnnotationUsageContext)

	// EnterStructDecl is called when entering the structDecl production.
	EnterStructDecl(c *StructDeclContext)

	// EnterStructField is called when entering the structField production.
	EnterStructField(c *StructFieldContext)

	// EnterEnumDecl is called when entering the enumDecl production.
	EnterEnumDecl(c *EnumDeclContext)

	// EnterEnumVariant is called when entering the enumVariant production.
	EnterEnumVariant(c *EnumVariantContext)

	// EnterFnDecl is called when entering the fnDecl production.
	EnterFnDecl(c *FnDeclContext)

	// EnterParamList is called when entering the paramList production.
	EnterParamList(c *ParamListContext)

	// EnterParam is called when entering the param production.
	EnterParam(c *ParamContext)

	// EnterPureTraitDecl is called when entering the PureTraitDecl production.
	EnterPureTraitDecl(c *PureTraitDeclContext)

	// EnterBindTraitDecl is called when entering the BindTraitDecl production.
	EnterBindTraitDecl(c *BindTraitDeclContext)

	// EnterTraitFnDecl is called when entering the traitFnDecl production.
	EnterTraitFnDecl(c *TraitFnDeclContext)

	// EnterBindTarget is called when entering the bindTarget production.
	EnterBindTarget(c *BindTargetContext)

	// EnterTraitBodyItem is called when entering the traitBodyItem production.
	EnterTraitBodyItem(c *TraitBodyItemContext)

	// EnterBanDirective is called when entering the BanDirective production.
	EnterBanDirective(c *BanDirectiveContext)

	// EnterFlipBanDirective is called when entering the FlipBanDirective production.
	EnterFlipBanDirective(c *FlipBanDirectiveContext)

	// EnterFlipBanItem is called when entering the flipBanItem production.
	EnterFlipBanItem(c *FlipBanItemContext)

	// EnterReturnStmt is called when entering the returnStmt production.
	EnterReturnStmt(c *ReturnStmtContext)

	// EnterBlock is called when entering the block production.
	EnterBlock(c *BlockContext)

	// EnterIfStmt is called when entering the ifStmt production.
	EnterIfStmt(c *IfStmtContext)

	// EnterMatchStmt is called when entering the matchStmt production.
	EnterMatchStmt(c *MatchStmtContext)

	// EnterValueMatchCase is called when entering the ValueMatchCase production.
	EnterValueMatchCase(c *ValueMatchCaseContext)

	// EnterTypeMatchCase is called when entering the TypeMatchCase production.
	EnterTypeMatchCase(c *TypeMatchCaseContext)

	// EnterDefaultMatchCase is called when entering the DefaultMatchCase production.
	EnterDefaultMatchCase(c *DefaultMatchCaseContext)

	// EnterWhileStmt is called when entering the whileStmt production.
	EnterWhileStmt(c *WhileStmtContext)

	// EnterLoopStmt is called when entering the loopStmt production.
	EnterLoopStmt(c *LoopStmtContext)

	// EnterRangeForStmt is called when entering the RangeForStmt production.
	EnterRangeForStmt(c *RangeForStmtContext)

	// EnterTraditionalForStmt is called when entering the TraditionalForStmt production.
	EnterTraditionalForStmt(c *TraditionalForStmtContext)

	// EnterIteratorForStmt is called when entering the IteratorForStmt production.
	EnterIteratorForStmt(c *IteratorForStmtContext)

	// EnterForInit is called when entering the forInit production.
	EnterForInit(c *ForInitContext)

	// EnterBreakStmt is called when entering the breakStmt production.
	EnterBreakStmt(c *BreakStmtContext)

	// EnterContinueStmt is called when entering the continueStmt production.
	EnterContinueStmt(c *ContinueStmtContext)

	// EnterSingleLetDecl is called when entering the SingleLetDecl production.
	EnterSingleLetDecl(c *SingleLetDeclContext)

	// EnterTupleLetDecl is called when entering the TupleLetDecl production.
	EnterTupleLetDecl(c *TupleLetDeclContext)

	// EnterConstDecl is called when entering the ConstDecl production.
	EnterConstDecl(c *ConstDeclContext)

	// EnterTypeList is called when entering the typeList production.
	EnterTypeList(c *TypeListContext)

	// EnterTypeType is called when entering the typeType production.
	EnterTypeType(c *TypeTypeContext)

	// EnterQualifiedName is called when entering the qualifiedName production.
	EnterQualifiedName(c *QualifiedNameContext)

	// EnterExprStmt is called when entering the exprStmt production.
	EnterExprStmt(c *ExprStmtContext)

	// EnterExprList is called when entering the exprList production.
	EnterExprList(c *ExprListContext)

	// EnterStringExpr is called when entering the StringExpr production.
	EnterStringExpr(c *StringExprContext)

	// EnterArrayIndexExpr is called when entering the ArrayIndexExpr production.
	EnterArrayIndexExpr(c *ArrayIndexExprContext)

	// EnterFloatExpr is called when entering the FloatExpr production.
	EnterFloatExpr(c *FloatExprContext)

	// EnterNotIsExpr is called when entering the NotIsExpr production.
	EnterNotIsExpr(c *NotIsExprContext)

	// EnterDerefExpr is called when entering the DerefExpr production.
	EnterDerefExpr(c *DerefExprContext)

	// EnterTargetExpr is called when entering the TargetExpr production.
	EnterTargetExpr(c *TargetExprContext)

	// EnterLogicalAndExpr is called when entering the LogicalAndExpr production.
	EnterLogicalAndExpr(c *LogicalAndExprContext)

	// EnterPostfixExpr is called when entering the PostfixExpr production.
	EnterPostfixExpr(c *PostfixExprContext)

	// EnterIdentifierExpr is called when entering the IdentifierExpr production.
	EnterIdentifierExpr(c *IdentifierExprContext)

	// EnterBinaryCompareExpr is called when entering the BinaryCompareExpr production.
	EnterBinaryCompareExpr(c *BinaryCompareExprContext)

	// EnterArrayLiteralExpr is called when entering the ArrayLiteralExpr production.
	EnterArrayLiteralExpr(c *ArrayLiteralExprContext)

	// EnterIsExpr is called when entering the IsExpr production.
	EnterIsExpr(c *IsExprContext)

	// EnterCastExpr is called when entering the CastExpr production.
	EnterCastExpr(c *CastExprContext)

	// EnterCallExpr is called when entering the CallExpr production.
	EnterCallExpr(c *CallExprContext)

	// EnterNotExpr is called when entering the NotExpr production.
	EnterNotExpr(c *NotExprContext)

	// EnterThisExpr is called when entering the ThisExpr production.
	EnterThisExpr(c *ThisExprContext)

	// EnterTernaryExpr is called when entering the TernaryExpr production.
	EnterTernaryExpr(c *TernaryExprContext)

	// EnterFuncCallExpr is called when entering the FuncCallExpr production.
	EnterFuncCallExpr(c *FuncCallExprContext)

	// EnterNilExpr is called when entering the NilExpr production.
	EnterNilExpr(c *NilExprContext)

	// EnterLambdaExpr is called when entering the LambdaExpr production.
	EnterLambdaExpr(c *LambdaExprContext)

	// EnterStructLiteralExpr is called when entering the StructLiteralExpr production.
	EnterStructLiteralExpr(c *StructLiteralExprContext)

	// EnterPanicUnwrapExpr is called when entering the PanicUnwrapExpr production.
	EnterPanicUnwrapExpr(c *PanicUnwrapExprContext)

	// EnterTupleExpr is called when entering the TupleExpr production.
	EnterTupleExpr(c *TupleExprContext)

	// EnterPrefixExpr is called when entering the PrefixExpr production.
	EnterPrefixExpr(c *PrefixExprContext)

	// EnterLogicalOrExpr is called when entering the LogicalOrExpr production.
	EnterLogicalOrExpr(c *LogicalOrExprContext)

	// EnterMulDivExpr is called when entering the MulDivExpr production.
	EnterMulDivExpr(c *MulDivExprContext)

	// EnterTryUnwrapExpr is called when entering the TryUnwrapExpr production.
	EnterTryUnwrapExpr(c *TryUnwrapExprContext)

	// EnterAddrOfExpr is called when entering the AddrOfExpr production.
	EnterAddrOfExpr(c *AddrOfExprContext)

	// EnterQuoteExpr is called when entering the QuoteExpr production.
	EnterQuoteExpr(c *QuoteExprContext)

	// EnterInnerCallExpr is called when entering the InnerCallExpr production.
	EnterInnerCallExpr(c *InnerCallExprContext)

	// EnterIntExpr is called when entering the IntExpr production.
	EnterIntExpr(c *IntExprContext)

	// EnterParenExpr is called when entering the ParenExpr production.
	EnterParenExpr(c *ParenExprContext)

	// EnterMemberAccessExpr is called when entering the MemberAccessExpr production.
	EnterMemberAccessExpr(c *MemberAccessExprContext)

	// EnterAddSubExpr is called when entering the AddSubExpr production.
	EnterAddSubExpr(c *AddSubExprContext)

	// EnterMethodCallExpr is called when entering the MethodCallExpr production.
	EnterMethodCallExpr(c *MethodCallExprContext)

	// ExitProgram is called when exiting the program production.
	ExitProgram(c *ProgramContext)

	// ExitPackageDecl is called when exiting the packageDecl production.
	ExitPackageDecl(c *PackageDeclContext)

	// ExitBlockImport is called when exiting the BlockImport production.
	ExitBlockImport(c *BlockImportContext)

	// ExitSingleImport is called when exiting the SingleImport production.
	ExitSingleImport(c *SingleImportContext)

	// ExitImportSpec is called when exiting the importSpec production.
	ExitImportSpec(c *ImportSpecContext)

	// ExitStatement is called when exiting the statement production.
	ExitStatement(c *StatementContext)

	// ExitAnnotationDecl is called when exiting the annotationDecl production.
	ExitAnnotationDecl(c *AnnotationDeclContext)

	// ExitAnnotationTarget is called when exiting the annotationTarget production.
	ExitAnnotationTarget(c *AnnotationTargetContext)

	// ExitSpawnStmt is called when exiting the spawnStmt production.
	ExitSpawnStmt(c *SpawnStmtContext)

	// ExitSelectStmt is called when exiting the selectStmt production.
	ExitSelectStmt(c *SelectStmtContext)

	// ExitSelectReadBranch is called when exiting the SelectReadBranch production.
	ExitSelectReadBranch(c *SelectReadBranchContext)

	// ExitSelectWriteBranch is called when exiting the SelectWriteBranch production.
	ExitSelectWriteBranch(c *SelectWriteBranchContext)

	// ExitSelectOtherBranch is called when exiting the SelectOtherBranch production.
	ExitSelectOtherBranch(c *SelectOtherBranchContext)

	// ExitSelectRead is called when exiting the selectRead production.
	ExitSelectRead(c *SelectReadContext)

	// ExitSelectWrite is called when exiting the selectWrite production.
	ExitSelectWrite(c *SelectWriteContext)

	// ExitSelectOther is called when exiting the selectOther production.
	ExitSelectOther(c *SelectOtherContext)

	// ExitDeferStmt is called when exiting the deferStmt production.
	ExitDeferStmt(c *DeferStmtContext)

	// ExitAssignmentStmt is called when exiting the assignmentStmt production.
	ExitAssignmentStmt(c *AssignmentStmtContext)

	// ExitModifier is called when exiting the modifier production.
	ExitModifier(c *ModifierContext)

	// ExitTypeParams is called when exiting the typeParams production.
	ExitTypeParams(c *TypeParamsContext)

	// ExitTypeParam is called when exiting the typeParam production.
	ExitTypeParam(c *TypeParamContext)

	// ExitTypeArgs is called when exiting the typeArgs production.
	ExitTypeArgs(c *TypeArgsContext)

	// ExitWhereClause is called when exiting the whereClause production.
	ExitWhereClause(c *WhereClauseContext)

	// ExitGenericConstraint is called when exiting the genericConstraint production.
	ExitGenericConstraint(c *GenericConstraintContext)

	// ExitAnnotationUsage is called when exiting the annotationUsage production.
	ExitAnnotationUsage(c *AnnotationUsageContext)

	// ExitStructDecl is called when exiting the structDecl production.
	ExitStructDecl(c *StructDeclContext)

	// ExitStructField is called when exiting the structField production.
	ExitStructField(c *StructFieldContext)

	// ExitEnumDecl is called when exiting the enumDecl production.
	ExitEnumDecl(c *EnumDeclContext)

	// ExitEnumVariant is called when exiting the enumVariant production.
	ExitEnumVariant(c *EnumVariantContext)

	// ExitFnDecl is called when exiting the fnDecl production.
	ExitFnDecl(c *FnDeclContext)

	// ExitParamList is called when exiting the paramList production.
	ExitParamList(c *ParamListContext)

	// ExitParam is called when exiting the param production.
	ExitParam(c *ParamContext)

	// ExitPureTraitDecl is called when exiting the PureTraitDecl production.
	ExitPureTraitDecl(c *PureTraitDeclContext)

	// ExitBindTraitDecl is called when exiting the BindTraitDecl production.
	ExitBindTraitDecl(c *BindTraitDeclContext)

	// ExitTraitFnDecl is called when exiting the traitFnDecl production.
	ExitTraitFnDecl(c *TraitFnDeclContext)

	// ExitBindTarget is called when exiting the bindTarget production.
	ExitBindTarget(c *BindTargetContext)

	// ExitTraitBodyItem is called when exiting the traitBodyItem production.
	ExitTraitBodyItem(c *TraitBodyItemContext)

	// ExitBanDirective is called when exiting the BanDirective production.
	ExitBanDirective(c *BanDirectiveContext)

	// ExitFlipBanDirective is called when exiting the FlipBanDirective production.
	ExitFlipBanDirective(c *FlipBanDirectiveContext)

	// ExitFlipBanItem is called when exiting the flipBanItem production.
	ExitFlipBanItem(c *FlipBanItemContext)

	// ExitReturnStmt is called when exiting the returnStmt production.
	ExitReturnStmt(c *ReturnStmtContext)

	// ExitBlock is called when exiting the block production.
	ExitBlock(c *BlockContext)

	// ExitIfStmt is called when exiting the ifStmt production.
	ExitIfStmt(c *IfStmtContext)

	// ExitMatchStmt is called when exiting the matchStmt production.
	ExitMatchStmt(c *MatchStmtContext)

	// ExitValueMatchCase is called when exiting the ValueMatchCase production.
	ExitValueMatchCase(c *ValueMatchCaseContext)

	// ExitTypeMatchCase is called when exiting the TypeMatchCase production.
	ExitTypeMatchCase(c *TypeMatchCaseContext)

	// ExitDefaultMatchCase is called when exiting the DefaultMatchCase production.
	ExitDefaultMatchCase(c *DefaultMatchCaseContext)

	// ExitWhileStmt is called when exiting the whileStmt production.
	ExitWhileStmt(c *WhileStmtContext)

	// ExitLoopStmt is called when exiting the loopStmt production.
	ExitLoopStmt(c *LoopStmtContext)

	// ExitRangeForStmt is called when exiting the RangeForStmt production.
	ExitRangeForStmt(c *RangeForStmtContext)

	// ExitTraditionalForStmt is called when exiting the TraditionalForStmt production.
	ExitTraditionalForStmt(c *TraditionalForStmtContext)

	// ExitIteratorForStmt is called when exiting the IteratorForStmt production.
	ExitIteratorForStmt(c *IteratorForStmtContext)

	// ExitForInit is called when exiting the forInit production.
	ExitForInit(c *ForInitContext)

	// ExitBreakStmt is called when exiting the breakStmt production.
	ExitBreakStmt(c *BreakStmtContext)

	// ExitContinueStmt is called when exiting the continueStmt production.
	ExitContinueStmt(c *ContinueStmtContext)

	// ExitSingleLetDecl is called when exiting the SingleLetDecl production.
	ExitSingleLetDecl(c *SingleLetDeclContext)

	// ExitTupleLetDecl is called when exiting the TupleLetDecl production.
	ExitTupleLetDecl(c *TupleLetDeclContext)

	// ExitConstDecl is called when exiting the ConstDecl production.
	ExitConstDecl(c *ConstDeclContext)

	// ExitTypeList is called when exiting the typeList production.
	ExitTypeList(c *TypeListContext)

	// ExitTypeType is called when exiting the typeType production.
	ExitTypeType(c *TypeTypeContext)

	// ExitQualifiedName is called when exiting the qualifiedName production.
	ExitQualifiedName(c *QualifiedNameContext)

	// ExitExprStmt is called when exiting the exprStmt production.
	ExitExprStmt(c *ExprStmtContext)

	// ExitExprList is called when exiting the exprList production.
	ExitExprList(c *ExprListContext)

	// ExitStringExpr is called when exiting the StringExpr production.
	ExitStringExpr(c *StringExprContext)

	// ExitArrayIndexExpr is called when exiting the ArrayIndexExpr production.
	ExitArrayIndexExpr(c *ArrayIndexExprContext)

	// ExitFloatExpr is called when exiting the FloatExpr production.
	ExitFloatExpr(c *FloatExprContext)

	// ExitNotIsExpr is called when exiting the NotIsExpr production.
	ExitNotIsExpr(c *NotIsExprContext)

	// ExitDerefExpr is called when exiting the DerefExpr production.
	ExitDerefExpr(c *DerefExprContext)

	// ExitTargetExpr is called when exiting the TargetExpr production.
	ExitTargetExpr(c *TargetExprContext)

	// ExitLogicalAndExpr is called when exiting the LogicalAndExpr production.
	ExitLogicalAndExpr(c *LogicalAndExprContext)

	// ExitPostfixExpr is called when exiting the PostfixExpr production.
	ExitPostfixExpr(c *PostfixExprContext)

	// ExitIdentifierExpr is called when exiting the IdentifierExpr production.
	ExitIdentifierExpr(c *IdentifierExprContext)

	// ExitBinaryCompareExpr is called when exiting the BinaryCompareExpr production.
	ExitBinaryCompareExpr(c *BinaryCompareExprContext)

	// ExitArrayLiteralExpr is called when exiting the ArrayLiteralExpr production.
	ExitArrayLiteralExpr(c *ArrayLiteralExprContext)

	// ExitIsExpr is called when exiting the IsExpr production.
	ExitIsExpr(c *IsExprContext)

	// ExitCastExpr is called when exiting the CastExpr production.
	ExitCastExpr(c *CastExprContext)

	// ExitCallExpr is called when exiting the CallExpr production.
	ExitCallExpr(c *CallExprContext)

	// ExitNotExpr is called when exiting the NotExpr production.
	ExitNotExpr(c *NotExprContext)

	// ExitThisExpr is called when exiting the ThisExpr production.
	ExitThisExpr(c *ThisExprContext)

	// ExitTernaryExpr is called when exiting the TernaryExpr production.
	ExitTernaryExpr(c *TernaryExprContext)

	// ExitFuncCallExpr is called when exiting the FuncCallExpr production.
	ExitFuncCallExpr(c *FuncCallExprContext)

	// ExitNilExpr is called when exiting the NilExpr production.
	ExitNilExpr(c *NilExprContext)

	// ExitLambdaExpr is called when exiting the LambdaExpr production.
	ExitLambdaExpr(c *LambdaExprContext)

	// ExitStructLiteralExpr is called when exiting the StructLiteralExpr production.
	ExitStructLiteralExpr(c *StructLiteralExprContext)

	// ExitPanicUnwrapExpr is called when exiting the PanicUnwrapExpr production.
	ExitPanicUnwrapExpr(c *PanicUnwrapExprContext)

	// ExitTupleExpr is called when exiting the TupleExpr production.
	ExitTupleExpr(c *TupleExprContext)

	// ExitPrefixExpr is called when exiting the PrefixExpr production.
	ExitPrefixExpr(c *PrefixExprContext)

	// ExitLogicalOrExpr is called when exiting the LogicalOrExpr production.
	ExitLogicalOrExpr(c *LogicalOrExprContext)

	// ExitMulDivExpr is called when exiting the MulDivExpr production.
	ExitMulDivExpr(c *MulDivExprContext)

	// ExitTryUnwrapExpr is called when exiting the TryUnwrapExpr production.
	ExitTryUnwrapExpr(c *TryUnwrapExprContext)

	// ExitAddrOfExpr is called when exiting the AddrOfExpr production.
	ExitAddrOfExpr(c *AddrOfExprContext)

	// ExitQuoteExpr is called when exiting the QuoteExpr production.
	ExitQuoteExpr(c *QuoteExprContext)

	// ExitInnerCallExpr is called when exiting the InnerCallExpr production.
	ExitInnerCallExpr(c *InnerCallExprContext)

	// ExitIntExpr is called when exiting the IntExpr production.
	ExitIntExpr(c *IntExprContext)

	// ExitParenExpr is called when exiting the ParenExpr production.
	ExitParenExpr(c *ParenExprContext)

	// ExitMemberAccessExpr is called when exiting the MemberAccessExpr production.
	ExitMemberAccessExpr(c *MemberAccessExprContext)

	// ExitAddSubExpr is called when exiting the AddSubExpr production.
	ExitAddSubExpr(c *AddSubExprContext)

	// ExitMethodCallExpr is called when exiting the MethodCallExpr production.
	ExitMethodCallExpr(c *MethodCallExprContext)
}
