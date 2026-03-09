package transpiler

import (
	"strings"

	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/compiler/symbols"
	"github.com/miea04/mygo/pkg/compiler/types"
)

// MyGoTranspiler converts MyGo AST to Go source code
type MyGoTranspiler struct {
	*ast.BaseMyGoVisitor
	Scope                *symbols.Scope
	CurrentScope         *symbols.Scope
	needsTernaryHelper   bool
	loopDepth            int
	currentMatchTarget   string
	currentMatchExpr     string
	currentMatchNode     ast.IExprContext
	currentMatchVar      string
	useTypeMatch         bool
	compileTimeTypeMatch bool
	mixedBoolMatch       bool
	isEnumMatch          bool
	enumGenericArgs      string // e.g., "[int, string]" for Result<int, string>
	currentImplType      string
	currentImplSymbol    *symbols.Symbol // Add this field
	currentBindVar       string
	currentBindTypeDef   string
	expectedType         string
}

func NewMyGoTranspiler(global *symbols.Scope) *MyGoTranspiler {
	return &MyGoTranspiler{
		BaseMyGoVisitor: &ast.BaseMyGoVisitor{},
		Scope:           global,
		CurrentScope:    global,
	}
}

func (v *MyGoTranspiler) resolveType(ctx ast.ITypeTypeContext) string {
	if ctx == nil {
		return ""
	}
	return v.resolveTypeStr(ctx.GetText())
}

func (v *MyGoTranspiler) resolveTypeStr(rawText string) string {
	t := types.ResolveTypeWithScope(rawText, v.CurrentScope)
	return v.toGoType(t)
}

func (v *MyGoTranspiler) toGoType(t string) string {
	if strings.HasPrefix(t, "chan<") && strings.HasSuffix(t, ">") {
		elem := t[5 : len(t)-1]
		return "chan " + v.toGoType(elem)
	}
	return t
}

func (v *MyGoTranspiler) pushScope(name string) {
	v.CurrentScope = symbols.NewScope(name, v.CurrentScope)
}

func (v *MyGoTranspiler) popScope() {
	if v.CurrentScope.Parent != nil {
		v.CurrentScope = v.CurrentScope.Parent
	}
}

func (v *MyGoTranspiler) NeedsTernaryHelper() bool {
	return v.needsTernaryHelper
}
