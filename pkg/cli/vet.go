package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/compiler"
	"github.com/miea04/mygo/pkg/compiler/semantic"
	"github.com/miea04/mygo/pkg/compiler/symbols"
	"github.com/spf13/cobra"
)

var vetCmd = &cobra.Command{
	Use:   "vet [packages]",
	Short: "Run static analysis on MyGo packages",
	RunE: func(cmd *cobra.Command, args []string) error {
		targetDir, _ := os.Getwd()
		if len(args) > 0 {
			targetDir = args[0]
			// Resolve absolute path
			if !filepath.IsAbs(targetDir) {
				cwd, _ := os.Getwd()
				targetDir = filepath.Join(cwd, targetDir)
			}
		}

		loader := compiler.NewPackageLoader(targetDir)
		hasErrors := false

		err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
					return filepath.SkipDir
				}

				// Check for .mygo files
				hasMyGo := false
				entries, _ := os.ReadDir(path)
				for _, e := range entries {
					if strings.HasSuffix(e.Name(), ".mygo") {
						hasMyGo = true
						break
					}
				}

				if hasMyGo {
					rel, _ := filepath.Rel(targetDir, path)
					importPath := filepath.ToSlash(rel)
					if importPath == "." {
						importPath = ""
					}

					pkg, err := loader.LoadPackage(importPath)
					if err != nil {
						fmt.Printf("Error loading package %s: %v\n", importPath, err)
						hasErrors = true
						return nil
					}

					// Pass 1: Symbol Collection
					collector := semantic.NewDeclarationCollector(pkg.Scope)
					for _, f := range pkg.Files {
						collector.SetCompilationUnit(pkg.Name, f.Path)
						// Iterate over statements in AST
						for _, stmt := range f.AST.AllStatement() {
							stmt.Accept(collector)
						}
					}

					// Pass 2: Method Collection
					methodCollector := semantic.NewMethodCollector(pkg.Scope)
					for _, f := range pkg.Files {
						methodCollector.SetCompilationUnit(pkg.Name, f.Path)
						for _, stmt := range f.AST.AllStatement() {
							stmt.Accept(methodCollector)
						}
					}

					// Pass 3: Semantic Analysis
					analyzer := semantic.NewSemanticAnalyzer(pkg.Scope)
					analyzer.PackageResolver = func(name string) *symbols.Scope {
						if loadedPkg, ok := loader.LoadedPackages[name]; ok {
							return loadedPkg.Scope
						}
						return nil
					}

					for _, f := range pkg.Files {
						analyzer.SetCompilationUnit(pkg.Name, f.Path)
						// Process imports
						for _, child := range f.AST.GetChildren() {
							if importStmt, ok := child.(*ast.ImportStmtContext); ok {
								importStmt.Accept(analyzer)
							}
						}

						for _, stmt := range f.AST.AllStatement() {
							stmt.Accept(analyzer)
						}
					}

					if len(analyzer.GetErrors()) > 0 {
						hasErrors = true
						fmt.Printf("vet: %s\n", strings.Join(analyzer.GetErrors(), "\n"))
					}
				}
			}
			return nil
		})

		if err != nil {
			return err
		}

		if hasErrors {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(vetCmd)
}
