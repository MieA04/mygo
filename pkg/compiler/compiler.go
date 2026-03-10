package compiler

import (
	"fmt"
	"strings"

	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/compiler/semantic"
	"github.com/miea04/mygo/pkg/compiler/symbols"
	"github.com/miea04/mygo/pkg/compiler/transpiler"
)

// TranspilePackage transpiles a single MyGo package to Go code.
func TranspilePackage(pkg *Package, loader *PackageLoader, moduleName string, emitHelpers bool) (string, error) {
	// Pass 1: Symbol Collection
	collector := semantic.NewDeclarationCollector(pkg.Scope)
	methodCollector := semantic.NewMethodCollector(pkg.Scope)

	for _, f := range pkg.Files {
		collector.SetCompilationUnit(pkg.Name, f.Path)
		for _, stmt := range f.AST.AllStatement() {
			stmt.Accept(collector)
		}
	}

	for _, f := range pkg.Files {
		methodCollector.SetCompilationUnit(pkg.Name, f.Path)
		for _, stmt := range f.AST.AllStatement() {
			stmt.Accept(methodCollector)
		}
	}

	// Pass 2: Semantic Analysis
	analyzer := semantic.NewSemanticAnalyzer(pkg.Scope)

	analyzer.PackageResolver = func(name string) *symbols.Scope {
		if loadedPkg, ok := loader.LoadedPackages[name]; ok {
			return loadedPkg.Scope
		}
		return nil
	}

	// 1. Semantic Analysis
	for _, f := range pkg.Files {
		analyzer.SetCompilationUnit(pkg.Name, f.Path)
		// Process imports
		for _, child := range f.AST.GetChildren() {
			if importStmt, ok := child.(*ast.ImportStmtContext); ok {
				analyzer.VisitImportStmt(importStmt)
			}
		}

		for _, stmt := range f.AST.AllStatement() {
			stmt.Accept(analyzer)
		}
	}

	if len(analyzer.GetErrors()) > 0 {
		return "", fmt.Errorf("semantic analysis failed:\n%s", strings.Join(analyzer.GetErrors(), "\n"))
	}

	// 2. Transpilation
	myTranspiler := transpiler.NewMyGoTranspiler(pkg.Scope)
	var body strings.Builder

	// Helpers
	if emitHelpers {
		body.WriteString(`
func _mygo_must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
`)
	}

	for _, f := range pkg.Files {
		myTranspiler.SetCurrentFile(f.Path)
		for _, stmt := range f.AST.AllStatement() {
			res := stmt.Accept(myTranspiler)
			if res != nil {
				body.WriteString(res.(string) + "\n")
			}
		}
	}

	if myTranspiler.NeedsTernaryHelper() && emitHelpers {
		body.WriteString(`
func _mygo_ternary[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}
`)
	}

	// 3. Assemble File
	var finalGoCode strings.Builder
	finalGoCode.WriteString(fmt.Sprintf("package %s\n\n", pkg.Name))

	// Collect imports
	imports := make(map[string]struct{})
	if strings.Contains(body.String(), "fmt.") {
		imports["fmt"] = struct{}{}
	}

	// Add user imports
	// We need to map MyGo imports to Go imports
	for _, f := range pkg.Files {
		for _, imp := range f.Imports {
			// If it's a loaded MyGo package (and thus local), prefix with moduleName.
			// Go packages (IsGoPackage=true) or not found packages are kept as is.
			if p, ok := loader.LoadedPackages[imp]; ok && !p.IsGoPackage {
				imports[fmt.Sprintf("%s/%s", moduleName, imp)] = struct{}{}
			} else {
				imports[imp] = struct{}{}
			}
		}
	}

	if len(imports) > 0 {
		finalGoCode.WriteString("import (\n")
		for imp := range imports {
			finalGoCode.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
		}
		finalGoCode.WriteString(")\n\n")
	}

	finalGoCode.WriteString(body.String())

	return finalGoCode.String(), nil
}
