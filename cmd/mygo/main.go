package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/build"
	"github.com/miea04/mygo/pkg/compiler"
	"github.com/miea04/mygo/pkg/compiler/semantic"
	"github.com/miea04/mygo/pkg/compiler/symbols"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "build":
			if err := runBuild(os.Args[2:]); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			return
		case "transpile":
			if err := runTranspile(os.Args[2:], false); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	if err := runTranspile(os.Args[1:], true); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func runBuild(args []string) error {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outputFile := fs.String("o", "", "Output executable file path")
	rootDir := fs.String("root", ".", "Root directory for package resolution")
	keepWorkDir := fs.Bool("keep-work-dir", false, "Keep temporary workspace directory")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if len(fs.Args()) != 1 {
		return fmt.Errorf("usage: mygo build [-o output.exe] [-root dir] [-keep-work-dir] <source.mygo|directory>")
	}

	sourcePath := fs.Args()[0]

	// Resolve absolute paths
	absRoot, err := filepath.Abs(*rootDir)
	if err != nil {
		return err
	}
	absSource, err := filepath.Abs(sourcePath)
	if err != nil {
		return err
	}

	opts := &build.Options{
		SourcePath:  absSource,
		OutputPath:  *outputFile,
		KeepWorkDir: *keepWorkDir,
		RootDir:     absRoot,
	}

	builder := build.NewBuilder(opts)
	return builder.Build()
}

func runTranspile(args []string, allowDefaultSource bool) error {
	fs := flag.NewFlagSet("transpile", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outputFile := fs.String("o", "", "Output Go source file path")
	rootDir := fs.String("root", ".", "Root directory for package resolution")
	if err := fs.Parse(args); err != nil {
		return err
	}

	var pkg *compiler.Package
	var err error
	loader := compiler.NewPackageLoader(*rootDir)

	if len(fs.Args()) > 0 {
		sourcePath := fs.Args()[0]
		stat, statErr := os.Stat(sourcePath)
		if statErr != nil {
			return statErr
		}

		importPath := sourcePath
		if *rootDir != "." {
			absRoot, _ := filepath.Abs(*rootDir)
			absSource, _ := filepath.Abs(sourcePath)
			rel, err := filepath.Rel(absRoot, absSource)
			if err == nil {
				importPath = filepath.ToSlash(rel)
			}
		}

		if stat.IsDir() {
			pkg, err = loader.LoadPackage(importPath)
		} else {
			// Single file
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
				// Define imported package symbol in current scope
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
		if !allowDefaultSource {
			return fmt.Errorf("usage: mygo transpile [-o output.go] <source.mygo|directory>")
		}
		// Default demo code
		code := `
struct BinaryTree {}

fn test_iterators() {
    let arr = [1, 2, 3];
    let tree = BinaryTree{};

    for(k, v : arr) { print(k, v); }
    for(item : arr) { print(item); }
    for(idx : arr.index()) { print(idx); }

    for(k, v : tree) { print(k, v); }
    for(val : tree.item()) { print(val); }
}
`
		input := antlr.NewInputStream(code)
		lexer := ast.NewMyGoLexer(input)
		stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
		parser := ast.NewMyGoParser(stream)
		tree := parser.Program()

		pkg = &compiler.Package{
			Name:       "main",
			ImportPath: "main",
			DirPath:    ".",
			Scope:      symbols.NewScope("main", nil),
			Files: []*compiler.SourceFile{{
				Path: "default.mygo",
				Code: code,
				AST:  tree,
			}},
		}

		// Collect
		collector := semantic.NewDeclarationCollector(pkg.Scope)
		collector.SetCompilationUnit("main", "default.mygo")
		for _, stmt := range tree.AllStatement() {
			stmt.Accept(collector)
		}
	}

	if err != nil {
		return err
	}

	if pkg == nil {
		return fmt.Errorf("internal error: pkg is nil before compilation")
	}

	includePreamble := *outputFile != ""
	finalGoCode, err := compilePackage(pkg, loader, includePreamble)
	if err != nil {
		return err
	}

	if *outputFile == "" && allowDefaultSource && len(fs.Args()) == 0 {
		fmt.Println("🚀 【输入的 mygo 源码】:")
		fmt.Println(pkg.Files[0].Code)
		fmt.Println("⚙️  【AST 编译降级中...】")
	}

	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, []byte(finalGoCode), 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Transpile successful: %s\n", *outputFile)
		return nil
	}

	fmt.Println("\n✨ 【生成的原生 Go 源码】:")
	fmt.Println(finalGoCode)
	return nil
}

func compilePackage(pkg *compiler.Package, loader *compiler.PackageLoader, includePreamble bool) (string, error) {
	return compiler.TranspilePackage(pkg, loader, "main")
}
