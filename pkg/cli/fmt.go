package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/miea04/mygo/pkg/ast"
	"github.com/spf13/cobra"
)

var fmtCmd = &cobra.Command{
	Use:   "fmt [files]",
	Short: "Format MyGo source code",
	RunE: func(cmd *cobra.Command, args []string) error {
		files := args
		if len(files) == 0 {
			// Find all .mygo files in current directory
			err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && strings.HasSuffix(info.Name(), ".mygo") {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		for _, file := range files {
			if err := formatFile(file); err != nil {
				fmt.Printf("Error formatting %s: %v\n", file, err)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(fmtCmd)
}

func formatFile(path string) error {
	input, err := antlr.NewFileStream(path)
	if err != nil {
		return err
	}

	lexer := ast.NewMyGoLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)
	parser := ast.NewMyGoParser(stream)
	parser.BuildParseTrees = true

	// Add Error Listener
	errorListener := &MyGoErrorListener{}
	parser.RemoveErrorListeners()
	parser.AddErrorListener(errorListener)

	tree := parser.Program()

	if len(errorListener.errors) > 0 {
		return fmt.Errorf("syntax errors:\n%s", strings.Join(errorListener.errors, "\n"))
	}

	printer := NewPrettyPrinter()
	formatted := tree.Accept(printer).(string)

	return os.WriteFile(path, []byte(formatted), 0644)
}

type MyGoErrorListener struct {
	*antlr.DefaultErrorListener
	errors []string
}

func (l *MyGoErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	l.errors = append(l.errors, fmt.Sprintf("line %d:%d %s", line, column, msg))
}

type PrettyPrinter struct {
	*ast.BaseMyGoVisitor
	indentLevel int
}

func NewPrettyPrinter() *PrettyPrinter {
	return &PrettyPrinter{
		BaseMyGoVisitor: &ast.BaseMyGoVisitor{
			BaseParseTreeVisitor: &antlr.BaseParseTreeVisitor{},
		},
	}
}

func (p *PrettyPrinter) indent() string {
	return strings.Repeat("    ", p.indentLevel)
}

func (p *PrettyPrinter) accept(node antlr.ParseTree) string {
	if node == nil {
		return ""
	}
	res := node.Accept(p)
	if str, ok := res.(string); ok {
		return str
	}

	// Fallback: get original text with whitespace
	if ctx, ok := node.(antlr.ParserRuleContext); ok {
		start := ctx.GetStart()
		stop := ctx.GetStop()
		if start == nil || stop == nil {
			return ""
		}
		stream := start.GetInputStream()
		if stream != nil {
			return stream.GetTextFromInterval(ctx.GetSourceInterval())
		}
	}
	if term, ok := node.(antlr.TerminalNode); ok {
		return term.GetText()
	}
	return node.GetText()
}

// VisitProgram visits the program rule
func (p *PrettyPrinter) VisitProgram(ctx *ast.ProgramContext) interface{} {
	var sb strings.Builder

	// Package declaration
	if ctx.PackageDecl() != nil {
		sb.WriteString(p.accept(ctx.PackageDecl()))
		sb.WriteString("\n\n")
	}

	// Imports
	for _, imp := range ctx.AllImportStmt() {
		sb.WriteString(p.accept(imp))
		sb.WriteString("\n")
	}
	if len(ctx.AllImportStmt()) > 0 {
		sb.WriteString("\n")
	}

	// Top level declarations
	for _, child := range ctx.GetChildren() {
		if _, ok := child.(*ast.PackageDeclContext); ok {
			continue
		}
		if _, ok := child.(*ast.ImportStmtContext); ok {
			continue
		}

		// Skip EOF
		if term, ok := child.(antlr.TerminalNode); ok {
			if term.GetSymbol().GetTokenType() == antlr.TokenEOF {
				continue
			}
		}

		if t, ok := child.(antlr.ParseTree); ok {
			res := p.accept(t)
			if res != "" {
				sb.WriteString(res)
				sb.WriteString("\n\n")
			}
		}
	}

	return strings.TrimSpace(sb.String()) + "\n"
}

func (p *PrettyPrinter) VisitStructDecl(ctx *ast.StructDeclContext) interface{} {
	var sb strings.Builder
	if ctx.Modifier() != nil {
		sb.WriteString(ctx.Modifier().GetText() + " ")
	}
	sb.WriteString("struct ")
	sb.WriteString(ctx.ID().GetText())
	sb.WriteString(" {\n")
	p.indentLevel++
	fields := ctx.AllStructField()
	for i, field := range fields {
		sb.WriteString(p.indent())
		sb.WriteString(field.ID().GetText())
		sb.WriteString(": ")
		sb.WriteString(field.TypeType().GetText())
		if i < len(fields)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	p.indentLevel--
	sb.WriteString(p.indent() + "}")
	return sb.String()
}

func (p *PrettyPrinter) VisitStructLiteralExpr(ctx *ast.StructLiteralExprContext) interface{} {
	var sb strings.Builder
	sb.WriteString(ctx.QualifiedName().GetText())
	if ctx.TypeArgs() != nil {
		sb.WriteString(ctx.TypeArgs().GetText())
	}
	sb.WriteString("{")

	ids := ctx.AllID()
	exprs := ctx.AllExpr()

	if len(ids) > 0 {
		sb.WriteString("\n")
		p.indentLevel++
		for i := 0; i < len(ids); i++ {
			sb.WriteString(p.indent())
			sb.WriteString(ids[i].GetText())
			sb.WriteString(": ")
			sb.WriteString(p.accept(exprs[i]))
			if i < len(ids)-1 {
				sb.WriteString(",")
			}
			sb.WriteString("\n")
		}
		p.indentLevel--
		sb.WriteString(p.indent())
	}
	sb.WriteString("}")
	return sb.String()
}

func (p *PrettyPrinter) VisitFuncCallExpr(ctx *ast.FuncCallExprContext) interface{} {
	var sb strings.Builder
	sb.WriteString(ctx.QualifiedName().GetText())
	if ctx.TypeArgs() != nil {
		sb.WriteString(ctx.TypeArgs().GetText())
	}
	sb.WriteString("(")
	if ctx.ExprList() != nil {
		sb.WriteString(p.accept(ctx.ExprList()))
	}
	sb.WriteString(")")
	return sb.String()
}

func (p *PrettyPrinter) VisitMethodCallExpr(ctx *ast.MethodCallExprContext) interface{} {
	var sb strings.Builder
	sb.WriteString(p.accept(ctx.Expr()))
	sb.WriteString(".")
	sb.WriteString(ctx.ID().GetText())
	if ctx.TypeArgs() != nil {
		sb.WriteString(ctx.TypeArgs().GetText())
	}
	sb.WriteString("(")
	if ctx.ExprList() != nil {
		sb.WriteString(p.accept(ctx.ExprList()))
	}
	sb.WriteString(")")
	return sb.String()
}

func (p *PrettyPrinter) VisitExprList(ctx *ast.ExprListContext) interface{} {
	var exprs []string
	for _, expr := range ctx.AllExpr() {
		exprs = append(exprs, p.accept(expr))
	}
	return strings.Join(exprs, ", ")
}

func (p *PrettyPrinter) VisitMemberAccessExpr(ctx *ast.MemberAccessExprContext) interface{} {
	return p.accept(ctx.Expr()) + "." + ctx.ID().GetText()
}

func (p *PrettyPrinter) VisitIntExpr(ctx *ast.IntExprContext) interface{} {
	return ctx.GetText()
}

func (p *PrettyPrinter) VisitAddSubExpr(ctx *ast.AddSubExprContext) interface{} {
	return p.accept(ctx.Expr(0)) + " " + ctx.GetOp().GetText() + " " + p.accept(ctx.Expr(1))
}

func (p *PrettyPrinter) VisitIdentifierExpr(ctx *ast.IdentifierExprContext) interface{} {
	return ctx.QualifiedName().GetText()
}

func (p *PrettyPrinter) VisitStringExpr(ctx *ast.StringExprContext) interface{} {
	return ctx.GetText()
}

func (p *PrettyPrinter) VisitFloatExpr(ctx *ast.FloatExprContext) interface{} {
	return ctx.GetText()
}

func (p *PrettyPrinter) VisitParenExpr(ctx *ast.ParenExprContext) interface{} {
	return "(" + p.accept(ctx.Expr()) + ")"
}

func (p *PrettyPrinter) VisitNotExpr(ctx *ast.NotExprContext) interface{} {
	return "!" + p.accept(ctx.Expr())
}

func (p *PrettyPrinter) VisitBinaryCompareExpr(ctx *ast.BinaryCompareExprContext) interface{} {
	return p.accept(ctx.Expr(0)) + " " + ctx.GetOp().GetText() + " " + p.accept(ctx.Expr(1))
}

func (p *PrettyPrinter) VisitLogicalAndExpr(ctx *ast.LogicalAndExprContext) interface{} {
	return p.accept(ctx.Expr(0)) + " && " + p.accept(ctx.Expr(1))
}

func (p *PrettyPrinter) VisitLogicalOrExpr(ctx *ast.LogicalOrExprContext) interface{} {
	return p.accept(ctx.Expr(0)) + " || " + p.accept(ctx.Expr(1))
}

func (p *PrettyPrinter) VisitAssignmentStmt(ctx *ast.AssignmentStmtContext) interface{} {
	return p.accept(ctx.Expr(0)) + " = " + p.accept(ctx.Expr(1)) + ";"
}

func (p *PrettyPrinter) VisitMulDivExpr(ctx *ast.MulDivExprContext) interface{} {
	return p.accept(ctx.Expr(0)) + " " + ctx.GetOp().GetText() + " " + p.accept(ctx.Expr(1))
}

func (p *PrettyPrinter) VisitSingleLetDecl(ctx *ast.SingleLetDeclContext) interface{} {
	var sb strings.Builder
	if ctx.Modifier() != nil {
		sb.WriteString(ctx.Modifier().GetText() + " ")
	}
	sb.WriteString("let ")
	sb.WriteString(ctx.ID().GetText())
	if ctx.TypeType() != nil {
		sb.WriteString(": " + ctx.TypeType().GetText())
	}
	if ctx.Expr() != nil {
		sb.WriteString(" = ")
		sb.WriteString(p.accept(ctx.Expr()))
	}
	sb.WriteString(";")
	return sb.String()
}

func (p *PrettyPrinter) VisitPackageDecl(ctx *ast.PackageDeclContext) interface{} {
	return "package " + ctx.ID().GetText()
}

func (p *PrettyPrinter) VisitBlockImport(ctx *ast.BlockImportContext) interface{} {
	var sb strings.Builder
	sb.WriteString("import {\n")
	p.indentLevel++
	for _, spec := range ctx.AllImportSpec() {
		sb.WriteString(p.indent())
		sb.WriteString(p.accept(spec))
		sb.WriteString(",\n")
	}
	p.indentLevel--
	sb.WriteString(p.indent() + "}")
	return sb.String()
}

func (p *PrettyPrinter) VisitSingleImport(ctx *ast.SingleImportContext) interface{} {
	return "import " + p.accept(ctx.ImportSpec())
}

func (p *PrettyPrinter) VisitImportSpec(ctx *ast.ImportSpecContext) interface{} {
	if ctx.ID() != nil {
		return ctx.STRING().GetText() + " as " + ctx.ID().GetText()
	}
	return ctx.STRING().GetText()
}

func (p *PrettyPrinter) VisitFnDecl(ctx *ast.FnDeclContext) interface{} {
	var sb strings.Builder
	if ctx.Modifier() != nil {
		sb.WriteString(ctx.Modifier().GetText() + " ")
	}
	sb.WriteString("fn ")
	sb.WriteString(ctx.ID().GetText())

	// Type params if any
	if ctx.TypeParams() != nil {
		sb.WriteString(ctx.TypeParams().GetText())
	}

	sb.WriteString("(")
	if ctx.ParamList() != nil {
		sb.WriteString(p.accept(ctx.ParamList()))
	}
	sb.WriteString(")")

	if ctx.TypeType() != nil {
		sb.WriteString(": " + ctx.TypeType().GetText())
	}

	sb.WriteString(" ")
	sb.WriteString(p.accept(ctx.Block()))
	return sb.String()
}

func (p *PrettyPrinter) VisitParamList(ctx *ast.ParamListContext) interface{} {
	var params []string
	for _, param := range ctx.AllParam() {
		params = append(params, p.accept(param))
	}
	return strings.Join(params, ", ")
}

func (p *PrettyPrinter) VisitParam(ctx *ast.ParamContext) interface{} {
	return ctx.ID().GetText() + ": " + ctx.TypeType().GetText()
}

func (p *PrettyPrinter) VisitBlock(ctx *ast.BlockContext) interface{} {
	var sb strings.Builder
	sb.WriteString("{\n")
	p.indentLevel++

	for _, stmt := range ctx.AllStatement() {
		sb.WriteString(p.indent())
		sb.WriteString(p.accept(stmt))
		sb.WriteString("\n")
	}

	p.indentLevel--
	sb.WriteString(p.indent() + "}")
	return sb.String()
}

func (p *PrettyPrinter) VisitStatement(ctx *ast.StatementContext) interface{} {
	if ctx.GetChildCount() > 0 {
		return p.accept(ctx.GetChild(0).(antlr.ParseTree))
	}
	return ""
}

func (p *PrettyPrinter) VisitReturnStmt(ctx *ast.ReturnStmtContext) interface{} {
	var sb strings.Builder
	sb.WriteString("return")
	if ctx.Expr() != nil {
		sb.WriteString(" " + p.accept(ctx.Expr()))
	}
	sb.WriteString(";")
	return sb.String()
}

func (p *PrettyPrinter) VisitVarDecl(ctx *ast.VarDeclContext) interface{} {
	// Simplified handling
	return ctx.GetText()
}

func (p *PrettyPrinter) VisitExprStmt(ctx *ast.ExprStmtContext) interface{} {
	return p.accept(ctx.Expr()) + ";"
}
