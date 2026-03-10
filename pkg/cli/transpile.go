package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/compiler"
	"github.com/miea04/mygo/pkg/compiler/semantic"
	"github.com/miea04/mygo/pkg/compiler/symbols"
	"github.com/spf13/cobra"
)

var transpileCmd = &cobra.Command{
	Use:   "transpile [flags] <source.mygo|directory>",
	Short: "Transpile MyGo source code to Go",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFile, _ := cmd.Flags().GetString("output")
		rootDir, _ := cmd.Flags().GetString("root")

		loader := compiler.NewPackageLoader(rootDir)
		var pkg *compiler.Package
		var err error

		if len(args) > 0 {
			sourcePath := args[0]
			stat, statErr := os.Stat(sourcePath)
			if statErr != nil {
				return statErr
			}

			importPath := sourcePath
			if rootDir != "." {
				absRoot, _ := filepath.Abs(rootDir)
				absSource, _ := filepath.Abs(sourcePath)
				rel, err := filepath.Rel(absRoot, absSource)
				if err == nil {
					importPath = filepath.ToSlash(rel)
				}
			}

			if stat.IsDir() {
				pkg, err = loader.LoadPackage(importPath)
			} else {
				// Single file logic
				content, err := os.ReadFile(sourcePath)
				if err != nil {
					return err
				}
				input := antlr.NewInputStream(string(content))
				lexer := ast.NewMyGoLexer(input)
				stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
				parser := ast.NewMyGoParser(stream)
				tree := parser.Program()

				pkg = &compiler.Package{
					Name:       "main",
					ImportPath: "main",
					DirPath:    filepath.Dir(sourcePath),
					Scope:      symbols.NewScope("main", nil),
					Files: []*compiler.SourceFile{{
						Path: sourcePath,
						Code: string(content),
						AST:  tree,
					}},
				}

				// Parse imports
				for _, child := range tree.GetChildren() {
					if importStmt, ok := child.(*ast.ImportStmtContext); ok {
						rawStr := importStmt.STRING().GetText()
						importPath := strings.Trim(rawStr, "\"")
						pkg.Files[0].Imports = append(pkg.Files[0].Imports, importPath)
					}
				}

				// Load imports
				for _, imp := range pkg.Files[0].Imports {
					importedPkg, err := loader.LoadPackage(imp)
					if err != nil {
						return fmt.Errorf("failed to load import '%s': %w", imp, err)
					}
					// Define imported package symbol
					sym := &symbols.Symbol{
						MyGoName:      importedPkg.Name,
						GoName:        importedPkg.Name,
						Kind:          symbols.KindPackage,
						Visibility:    symbols.VisibilityPrivate,
						ImportedScope: importedPkg.Scope,
					}
					pkg.Scope.DefineSymbol(sym)
				}

				// Collect decls
				collector := semantic.NewDeclarationCollector(pkg.Scope)
				collector.SetCompilationUnit("main", sourcePath)
				for _, stmt := range tree.AllStatement() {
					stmt.Accept(collector)
				}

				// Collect methods
				methodCollector := semantic.NewMethodCollector(pkg.Scope)
				methodCollector.SetCompilationUnit("main", sourcePath)
				for _, stmt := range tree.AllStatement() {
					stmt.Accept(methodCollector)
				}
			}
		} else {
			return fmt.Errorf("usage: mygo transpile [-o output.go] <source.mygo|directory>")
		}

		if err != nil {
			return err
		}

		if pkg == nil {
			return fmt.Errorf("internal error: pkg is nil before compilation")
		}

		finalGoCode, err := compiler.TranspilePackage(pkg, loader, "main", true)
		if err != nil {
			return err
		}

		if outputFile != "" {
			if err := os.WriteFile(outputFile, []byte(finalGoCode), 0o644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}
			fmt.Printf("Transpile successful: %s\n", outputFile)
			return nil
		}

		fmt.Println("\n✨ 【生成的原生 Go 源码】:")
		fmt.Println(finalGoCode)
		return nil
	},
}

func init() {
	transpileCmd.Flags().StringP("output", "o", "", "Output Go source file path")
	transpileCmd.Flags().String("root", ".", "Root directory for package resolution")
	rootCmd.AddCommand(transpileCmd)
}
