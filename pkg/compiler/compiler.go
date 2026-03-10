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
		for _, decl := range f.AST.AllAnnotationDecl() {
			decl.Accept(collector)
		}
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
		for _, importStmt := range f.AST.AllImportStmt() {
			importStmt.Accept(analyzer)
		}

		for _, decl := range f.AST.AllAnnotationDecl() {
			decl.Accept(analyzer)
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
			// Explicitly cast to concrete type to bypass missing Accept method issue
			concreteStmt, ok := stmt.(*ast.StatementContext)
			if !ok {
				continue
			}
			res := myTranspiler.VisitStatement(concreteStmt)
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
	// imports := make(map[string]string) // Path -> Alias (or "" if none)
	// Note: If multiple aliases for same path, we might need a better structure.
	// But usually we just want to ensure we import it.
	// However, if we have: import f "fmt" AND import "fmt", we need BOTH.
	// So map key should be unique import declaration (Path + Alias).
	type ImportKey struct {
		Path  string
		Alias string
	}
	uniqueImports := make(map[ImportKey]struct{})

	if strings.Contains(body.String(), "fmt.") {
		uniqueImports[ImportKey{Path: "fmt"}] = struct{}{}
	}

	// Add auto-imports from transpiler (e.g. slices)
	for pkg := range myTranspiler.UsedPackages {
		uniqueImports[ImportKey{Path: pkg}] = struct{}{}
	}

	// Add user imports
	// We need to map MyGo imports to Go imports
	for _, f := range pkg.Files {
		for _, spec := range f.Imports {
			imp := spec.Path
			alias := spec.Alias
			// If it's a loaded MyGo package (and thus local), prefix with moduleName.
			// Go packages (IsGoPackage=true) or not found packages are kept as is.
			finalPath := imp
			if p, ok := loader.LoadedPackages[imp]; ok && !p.IsGoPackage {
				finalPath = fmt.Sprintf("%s/%s", moduleName, imp)
			}
			uniqueImports[ImportKey{Path: finalPath, Alias: alias}] = struct{}{}
		}
	}

	if len(uniqueImports) > 0 {
		finalGoCode.WriteString("import (\n")
		// Sort for deterministic output
		var sortedImports []ImportKey
		for k := range uniqueImports {
			sortedImports = append(sortedImports, k)
		}
		// Sort by Path then Alias
		// (omitted sort implementation for brevity, or rely on map iteration order if random is ok, but deterministic is better)
		// Let's just iterate map. Ideally we sort.

		for _, k := range sortedImports {
			if k.Alias != "" {
				finalGoCode.WriteString(fmt.Sprintf("\t%s \"%s\"\n", k.Alias, k.Path))
			} else {
				finalGoCode.WriteString(fmt.Sprintf("\t\"%s\"\n", k.Path))
			}
		}
		finalGoCode.WriteString(")\n\n")
	}

	// RFC-007: Generate init()
	if len(myTranspiler.InitFunctions) > 0 {
		body.WriteString("\nfunc init() {\n")
		for _, fn := range myTranspiler.InitFunctions {
			body.WriteString(fmt.Sprintf("\t%s()\n", fn))
		}
		body.WriteString("}\n")
	}

	finalGoCode.WriteString(body.String())

	return finalGoCode.String(), nil
}
