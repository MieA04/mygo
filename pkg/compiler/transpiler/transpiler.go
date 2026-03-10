package transpiler

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/compiler/interpreter"
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
	CurrentFile          string // Path to the current MyGo source file
	UsedPackages         map[string]bool

	// RFC-007 Annotations
	CurrentStructAnnotations map[string][]string
	CurrentProcessingStruct  string
	InitFunctions            []string
	CurrentOriginalFnName    string // For #inner_call resolution
}

func NewMyGoTranspiler(global *symbols.Scope) *MyGoTranspiler {
	return &MyGoTranspiler{
		BaseMyGoVisitor:          &ast.BaseMyGoVisitor{},
		Scope:                    global,
		CurrentScope:             global,
		UsedPackages:             make(map[string]bool),
		CurrentStructAnnotations: make(map[string][]string),
	}
}

func (v *MyGoTranspiler) AddImport(pkg string) {
	v.UsedPackages[pkg] = true
}

func (v *MyGoTranspiler) SetCurrentFile(path string) {
	v.CurrentFile = path
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
		return "chan " + v.resolveTypeStr(elem)
	}
	// Handle function type fn(...) -> func(...)
	if strings.HasPrefix(t, "fn") {
		rest := t[2:]
		// Replace return type colon if present: fn(): int -> func() int
		rest = strings.Replace(rest, ":", " ", 1)
		return "func" + rest
	}

	// Handle slice T[] -> []T
	if strings.HasSuffix(t, "[]") {
		elem := t[:len(t)-2]
		return "[]" + v.resolveTypeStr(elem)
	}
	// Handle array T[N] -> [N]T
	if strings.HasSuffix(t, "]") {
		lastOpen := strings.LastIndex(t, "[")
		if lastOpen != -1 {
			sizeStr := t[lastOpen+1 : len(t)-1]
			isNum := true
			if len(sizeStr) == 0 {
				isNum = false
			}
			for _, r := range sizeStr {
				if r < '0' || r > '9' {
					isNum = false
					break
				}
			}
			if isNum {
				elem := t[:lastOpen]
				return "[" + sizeStr + "]" + v.resolveTypeStr(elem)
			}
		}
	}
	if t == "float" {
		return "float64"
	}
	return t
}

func (v *MyGoTranspiler) pushScope(name string) {
	v.CurrentScope = symbols.NewScope(name, v.CurrentScope)
}

func (v *MyGoTranspiler) executeMacro(annName string, targetNode interface{}) string {
	sym := v.CurrentScope.Resolve(annName)
	if sym == nil || sym.Kind != symbols.KindAnnotation {
		return "" // Not a macro or not found
	}

	decl, ok := sym.ASTNode.(*ast.AnnotationDeclContext)
	if !ok {
		return ""
	}

	interp := interpreter.NewInterpreter(v.Scope)
	meta := interpreter.MetaValue{Props: make(map[string]interpreter.Value)}

	// Try to get name from targetNode
	var targetName string
	var targetBody string
	if ctx, ok := targetNode.(antlr.ParserRuleContext); ok {
		switch n := ctx.(type) {
		case *ast.FnDeclContext:
			targetName = n.ID().GetText()
			// Extract body source text (excluding braces)
			if n.Block() != nil {
				start := n.Block().GetStart()
				stop := n.Block().GetStop()
				input := start.GetInputStream()

				// Use character indices to skip '{' and '}'
				startIndex := start.GetStop() + 1
				stopIndex := stop.GetStart() - 1

				if stopIndex > startIndex {
					interval := antlr.NewInterval(startIndex, stopIndex)
					targetBody = input.GetTextFromInterval(interval)
					// Trim trailing semicolon to avoid double semicolon when used as 'body;' in macro
					targetBody = strings.TrimSpace(targetBody)
					if strings.HasSuffix(targetBody, ";") {
						targetBody = targetBody[:len(targetBody)-1]
					}
				}
			}
		case *ast.StructDeclContext:
			targetName = n.ID().GetText()
		}
	}

	meta.Props["name"] = interpreter.StringValue{Val: targetName}
	if targetBody != "" {
		meta.Props["body"] = interpreter.StringValue{Val: targetBody}
	}
	interp.Env.Set("meta", meta)

	res := decl.Block().Accept(interp)
	if ret, ok := res.(interpreter.ReturnValue); ok {
		if str, ok := ret.Val.(interpreter.StringValue); ok {
			return str.Val
		}
	}
	return ""
}

func (v *MyGoTranspiler) transpileMacroResult(code string) string {
	codeWithBraces := "{\n" + code + "\n}"
	input := antlr.NewInputStream(codeWithBraces)
	lexer := ast.NewMyGoLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := ast.NewMyGoParser(stream)

	// Suppress error output for cleaner logs, or keep for debugging
	// parser.RemoveErrorListeners()

	tree := parser.Block()

	res := v.VisitBlock(tree.(*ast.BlockContext))
	if s, ok := res.(string); ok {
		// Strip outer braces from result: "{\n ... \n}"
		// Find first { and last }
		first := strings.Index(s, "{")
		last := strings.LastIndex(s, "}")
		if first != -1 && last != -1 && last > first {
			return s[first+1 : last]
		}
		return s
	}
	return ""
}

func (v *MyGoTranspiler) VisitInnerCallExpr(ctx *ast.InnerCallExprContext) interface{} {
	if v.CurrentOriginalFnName == "" {
		return "panic(\"inner_call used outside of macro context\")"
	}

	args := ""
	if ctx.ExprList() != nil {
		var argStrs []string
		for _, expr := range ctx.ExprList().(*ast.ExprListContext).AllExpr() {
			argStrs = append(argStrs, expr.Accept(v).(string))
		}
		args = strings.Join(argStrs, ", ")
	}

	return fmt.Sprintf("%s(%s)", v.CurrentOriginalFnName, args)
}

func (v *MyGoTranspiler) popScope() {
	if v.CurrentScope.Parent != nil {
		v.CurrentScope = v.CurrentScope.Parent
	}
}

func (v *MyGoTranspiler) NeedsTernaryHelper() bool {
	return v.needsTernaryHelper
}
