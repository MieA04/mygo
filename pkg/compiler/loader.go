package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/compiler/core"
	"github.com/miea04/mygo/pkg/compiler/loader"
	"github.com/miea04/mygo/pkg/compiler/semantic"
	"github.com/miea04/mygo/pkg/compiler/symbols"
)

type Package = core.Package
type SourceFile = core.SourceFile

type PackageLoader struct {
	LoadedPackages map[string]*Package // ImportPath -> Package
	RootPath       string              // Root directory of the compilation
}

func NewPackageLoader(rootPath string) *PackageLoader {
	return &PackageLoader{
		LoadedPackages: make(map[string]*Package),
		RootPath:       rootPath,
	}
}

// LoadPackage loads a package from a given path (relative or absolute).
// It recursively loads imported packages.
func (l *PackageLoader) LoadPackage(importPath string) (*Package, error) {
	// 1. Check cache
	if pkg, ok := l.LoadedPackages[importPath]; ok {
		return pkg, nil
	}

	// 2. Resolve directory path
	// For simplicity, we assume importPath is relative to RootPath or is a standard library path (not implemented yet)
	// or it is a relative path starting with "./" or "../" relative to the *importing file*?
	// To simplify, let's assume all imports are relative to the project root (RootPath).
	// Or relative to the current working directory if RootPath is not set.

	dirPath := filepath.Join(l.RootPath, importPath)
	info, err := os.Stat(dirPath)
	// Try loading as MyGo package first
	if err == nil && info.IsDir() {
		// 3. Scan files
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return nil, err
		}

		pkg := &Package{
			Name:       filepath.Base(dirPath), // Default package name is directory name
			ImportPath: importPath,
			DirPath:    dirPath,
			Scope:      symbols.NewScope(importPath, nil),
		}
		l.LoadedPackages[importPath] = pkg

		// 4. Parse files
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".mygo") {
				path := filepath.Join(dirPath, entry.Name())

				// Read file for SourceFile.Code
				content, err := os.ReadFile(path)
				if err != nil {
					return nil, err
				}

				// Create ANTLR stream
				input, err := antlr.NewFileStream(path)
				if err != nil {
					return nil, err
				}

				lexer := ast.NewMyGoLexer(input)
				stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
				parser := ast.NewMyGoParser(stream)

				// Error handling
				parser.RemoveErrorListeners()
				errorListener := &LoaderErrorListener{}
				parser.AddErrorListener(errorListener)

				tree := parser.Program()

				if len(errorListener.Errors) > 0 {
					return nil, fmt.Errorf("syntax errors in %s:\n%s", path, strings.Join(errorListener.Errors, "\n"))
				}

				file := &SourceFile{
					Path: path,
					Code: string(content),
					AST:  tree,
				}

				// Extract imports
				for _, child := range tree.GetChildren() {
					if pkgDecl, ok := child.(*ast.PackageDeclContext); ok {
						// Update package name from source (first one wins or check consistency)
						// For now, just overwrite
						pkg.Name = pkgDecl.ID().GetText()
					}
					if importStmt, ok := child.(ast.IImportStmtContext); ok {
						// Extract imports from block or single import
						if blockCtx, ok := importStmt.(*ast.BlockImportContext); ok {
							for _, spec := range blockCtx.AllImportSpec() {
								if specCtx, ok := spec.(*ast.ImportSpecContext); ok {
									rawStr := specCtx.STRING().GetText()
									importPath := strings.Trim(rawStr, "\"")
									file.Imports = append(file.Imports, core.ImportSpec{Path: importPath})
								}
							}
						} else if singleCtx, ok := importStmt.(*ast.SingleImportContext); ok {
							if specCtx, ok := singleCtx.ImportSpec().(*ast.ImportSpecContext); ok {
								rawStr := specCtx.STRING().GetText()
								importPath := strings.Trim(rawStr, "\"")
								file.Imports = append(file.Imports, core.ImportSpec{Path: importPath})
							}
						}
					}
				}

				pkg.Files = append(pkg.Files, file)
			}
		}

		// 5. Collect Declarations (Pass 1 for this package)
		collector := semantic.NewDeclarationCollector(pkg.Scope)
		for _, file := range pkg.Files {
			collector.SetCompilationUnit(pkg.Name, file.Path)
			for _, stmt := range file.AST.AllStatement() {
				stmt.Accept(collector)
			}
		}

		// 6. Recursively load imports
		for _, file := range pkg.Files {
			for _, spec := range file.Imports {
				imp := spec.Path
				importedPkg, err := l.LoadPackage(imp)
				if err != nil {
					return nil, fmt.Errorf("failed to load import '%s' in %s: %w", imp, file.Path, err)
				}

				// Define imported package symbol in the current package scope
				// This allows resolving types like "fmt.Stringer"
				// Note: Go/MyGo imports are file-level, but here we put them in package scope for simplicity.
				// This assumes no conflicting import names across files in the same package (or overwrites are fine).
				sym := &symbols.Symbol{
					MyGoName:      importedPkg.Name, // TODO: Support import aliases
					GoName:        importedPkg.Name,
					Kind:          symbols.KindPackage,
					Visibility:    symbols.VisibilityPrivate, // Imports are usually file-private
					ImportedScope: importedPkg.Scope,
				}
				pkg.Scope.DefineSymbol(sym)
			}
		}

		// 7. Collect Methods (Pass 1.5: BindTraitDecl)
		// This must run after imports are loaded because bind targets might be imported (future support)
		// or at least to ensure all local types are defined.
		methodCollector := semantic.NewMethodCollector(pkg.Scope)
		for _, file := range pkg.Files {
			methodCollector.SetCompilationUnit(pkg.Name, file.Path)
			for _, stmt := range file.AST.AllStatement() {
				stmt.Accept(methodCollector)
			}
		}

		return pkg, nil
	}

	// Fallback: Try loading as Go package
	// Prevent infinite recursion if Go loader calls back? No, Go loader uses packages.Load.
	// But we need to handle "not found" carefully.
	goPkg, err := loader.LoadGoPackage(importPath, l.RootPath)
	if err == nil {
		l.LoadedPackages[importPath] = goPkg
		return goPkg, nil
	}

	return nil, fmt.Errorf("package not found: %s (tried MyGo at %s, and Go package)", importPath, dirPath)
}

type LoaderErrorListener struct {
	*antlr.DefaultErrorListener
	Errors []string
}

func (l *LoaderErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	l.Errors = append(l.Errors, fmt.Sprintf("line %d:%d %s", line, column, msg))
}
