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

var docCmd = &cobra.Command{
	Use:   "doc [path]",
	Short: "Show documentation for package or symbol",
	RunE: func(cmd *cobra.Command, args []string) error {
		target := "."
		if len(args) > 0 {
			target = args[0]
		}

		// Check if target is a file or directory
		info, err := os.Stat(target)
		if err != nil {
			return err
		}

		var files []string
		if info.IsDir() {
			entries, _ := os.ReadDir(target)
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), ".mygo") {
					files = append(files, filepath.Join(target, e.Name()))
				}
			}
		} else {
			files = []string{target}
		}

		for _, file := range files {
			if err := extractDoc(file); err != nil {
				fmt.Printf("Error processing %s: %v\n", file, err)
			}
		}
		return nil
	},
}

func extractDoc(path string) error {
	input, err := antlr.NewFileStream(path)
	if err != nil {
		return err
	}

	lexer := ast.NewMyGoLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := ast.NewMyGoParser(stream)
	parser.BuildParseTrees = true

	// Add Error Listener
	errorListener := &DocErrorListener{}
	parser.RemoveErrorListeners()
	parser.AddErrorListener(errorListener)

	tree := parser.Program()

	if len(errorListener.errors) > 0 {
		return fmt.Errorf("syntax errors:\n%s", strings.Join(errorListener.errors, "\n"))
	}

	extractor := &DocExtractor{
		BaseMyGoVisitor: &ast.BaseMyGoVisitor{
			BaseParseTreeVisitor: &antlr.BaseParseTreeVisitor{},
		},
	}
	tree.Accept(extractor)
	return nil
}

type DocErrorListener struct {
	*antlr.DefaultErrorListener
	errors []string
}

func (l *DocErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	l.errors = append(l.errors, fmt.Sprintf("line %d:%d %s", line, column, msg))
}

type DocExtractor struct {
	*ast.BaseMyGoVisitor
}

func (d *DocExtractor) Visit(tree antlr.ParseTree) interface{} {
	return tree.Accept(d)
}

func (d *DocExtractor) VisitProgram(ctx *ast.ProgramContext) interface{} {
	for _, child := range ctx.GetChildren() {
		if t, ok := child.(antlr.ParseTree); ok {
			t.Accept(d)
		}
	}
	return nil
}

func (d *DocExtractor) VisitStatement(ctx *ast.StatementContext) interface{} {
	for _, child := range ctx.GetChildren() {
		if t, ok := child.(antlr.ParseTree); ok {
			t.Accept(d)
		}
	}
	return nil
}

func (d *DocExtractor) VisitFnDecl(ctx *ast.FnDeclContext) interface{} {
	// Check visibility
	isPub := false
	if mod := ctx.Modifier(); mod != nil {
		if mod.GetText() == "pub" {
			isPub = true
		}
	}

	if isPub {
		name := ctx.ID().GetText()
		sig := "fn " + name
		if ctx.ParamList() != nil {
			sig += "(" + ctx.ParamList().GetText() + ")"
		} else {
			sig += "()"
		}
		if ctx.TypeType() != nil {
			sig += ": " + ctx.TypeType().GetText()
		}

		fmt.Printf("fn %s\n    %s\n\n", name, sig)
	}
	return nil
}

func (d *DocExtractor) VisitStructDecl(ctx *ast.StructDeclContext) interface{} {
	// Check visibility
	isPub := false
	if mod := ctx.Modifier(); mod != nil {
		if mod.GetText() == "pub" {
			isPub = true
		}
	}

	if isPub {
		name := ctx.ID().GetText()
		fmt.Printf("type %s struct\n", name)

		fields := ctx.AllStructField()
		if len(fields) > 0 {
			fmt.Printf("    {\n")
			for _, field := range fields {
				fmt.Printf("        %s\n", field.GetText())
			}
			fmt.Printf("    }\n\n")
		} else {
			fmt.Printf("    {}\n\n")
		}
	}
	return nil
}

func (d *DocExtractor) VisitEnumDecl(ctx *ast.EnumDeclContext) interface{} {
	isPub := false
	if mod := ctx.Modifier(); mod != nil {
		if mod.GetText() == "pub" {
			isPub = true
		}
	}

	if isPub {
		name := ctx.ID().GetText()
		fmt.Printf("type %s enum\n", name)

		variants := ctx.AllEnumVariant()
		if len(variants) > 0 {
			fmt.Printf("    {\n")
			for _, v := range variants {
				fmt.Printf("        %s\n", v.GetText())
			}
			fmt.Printf("    }\n\n")
		} else {
			fmt.Printf("    {}\n\n")
		}
	}
	return nil
}

func (d *DocExtractor) VisitPureTraitDecl(ctx *ast.PureTraitDeclContext) interface{} {
	isPub := false
	if mod := ctx.Modifier(); mod != nil {
		if mod.GetText() == "pub" {
			isPub = true
		}
	}

	if isPub {
		name := ctx.ID().GetText()
		fmt.Printf("type %s trait\n", name)

		fns := ctx.AllTraitFnDecl()
		if len(fns) > 0 {
			fmt.Printf("    {\n")
			for _, fn := range fns {
				fmt.Printf("        %s\n", fn.GetText())
			}
			fmt.Printf("    }\n\n")
		} else {
			fmt.Printf("    {}\n\n")
		}
	}
	return nil
}

func init() {
	rootCmd.AddCommand(docCmd)
}
