package interpreter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/compiler/symbols"
)

type Value interface {
	String() string
	Type() string
}

type StringValue struct {
	Val string
}

func (v StringValue) String() string { return v.Val }
func (v StringValue) Type() string   { return "string" }

type IntValue struct {
	Val int
}

func (v IntValue) String() string { return strconv.Itoa(v.Val) }
func (v IntValue) Type() string   { return "int" }

type MetaValue struct {
	Props map[string]Value
}

func (v MetaValue) String() string { return "meta object" }
func (v MetaValue) Type() string   { return "meta" }

type ListValue struct {
	Val []Value
}

func (v ListValue) String() string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, item := range v.Val {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(item.String())
	}
	sb.WriteString("]")
	return sb.String()
}

func (v ListValue) Type() string { return "list" }

type Interpreter struct {
	*ast.BaseMyGoVisitor
	Scope *symbols.Scope
	Env   *Environment
}

func NewInterpreter(scope *symbols.Scope) *Interpreter {
	return &Interpreter{
		BaseMyGoVisitor: &ast.BaseMyGoVisitor{},
		Scope:           scope,
		Env:             NewEnvironment(nil),
	}
}

func (i *Interpreter) VisitBlock(ctx *ast.BlockContext) interface{} {
	// Push scope
	previousEnv := i.Env
	i.Env = NewEnvironment(previousEnv)
	defer func() { i.Env = previousEnv }()

	// Execute statements
	for _, stmt := range ctx.AllStatement() {
		// Explicitly cast to concrete type to bypass missing Accept method issue
		concreteStmt, ok := stmt.(*ast.StatementContext)
		if !ok {
			continue
		}
		res := i.VisitStatement(concreteStmt)
		// Check for return?
		if ret, isReturn := res.(ReturnValue); isReturn {
			return ret
		}
	}
	return nil
}

func (i *Interpreter) VisitIteratorForStmt(ctx *ast.IteratorForStmtContext) interface{} {
	// for id in expr { block }
	// or for id1, id2 in expr { block } (if we supported unpacking, but ListValue contains Values)

	iterable := ctx.Expr().Accept(i)
	listVal, ok := iterable.(ListValue)
	if !ok {
		return nil
	}

	ids := ctx.AllID()
	if len(ids) == 0 {
		return nil
	}
	loopVar := ids[0].GetText()

	// Iterate
	for _, item := range listVal.Val {
		// Create new scope for loop iteration? Or reuse?
		// Usually loops share scope or create new scope per iteration.
		// Let's create a new scope per iteration to be safe and clean.
		previousEnv := i.Env
		i.Env = NewEnvironment(previousEnv)

		i.Env.Set(loopVar, item)

		// Execute block
		// Explicitly cast to concrete BlockContext
		if block, ok := ctx.Block().(*ast.BlockContext); ok {
			res := i.VisitBlock(block)
			if res != nil {
				// Handle break/return/continue if we supported them
				// For now, if we get a ReturnValue, we should return it
				if _, isReturn := res.(ReturnValue); isReturn {
					i.Env = previousEnv // Restore before returning
					return res
				}
			}
		}

		i.Env = previousEnv
	}

	return nil
}

type ReturnValue struct {
	Val Value
}

func (i *Interpreter) VisitStatement(ctx *ast.StatementContext) interface{} {
	child := ctx.GetChild(0)
	if node, ok := child.(antlr.ParseTree); ok {
		return node.Accept(i)
	}
	return nil
}

func (i *Interpreter) VisitExprStmt(ctx *ast.ExprStmtContext) interface{} {
	if ctx.Expr() != nil {
		return ctx.Expr().Accept(i)
	}
	return nil
}

func (i *Interpreter) VisitIfStmt(ctx *ast.IfStmtContext) interface{} {
	cond := ctx.Expr(0).Accept(i)
	// Check if condition is true
	isTrue := false
	if b, ok := cond.(bool); ok {
		isTrue = b
	} else if _, ok := cond.(Value); ok {
		// Treat non-nil values as true (except empty string/0/false if we had BoolValue)
		// For now, simple check: non-nil is true?
		// Actually, let's implement BoolValue or treat specific values as false.
		// Since we don't have BoolValue yet, let's assume equality returns something we can check.
		// If VisitEqualityExpr returns bool, we use it.
		// But VisitEqualityExpr returns interface{}.
	} else if b, ok := cond.(bool); ok { // Redundant check, just to be safe
		isTrue = b
	}

	// Since we don't have BoolValue in Value interface yet, let's just use Go bool from visitors
	if isTrue {
		return ctx.Block(0).Accept(i)
	} else if ctx.Block(1) != nil {
		return ctx.Block(1).Accept(i)
	}
	return nil
}

func (i *Interpreter) VisitBinaryCompareExpr(ctx *ast.BinaryCompareExprContext) interface{} {
	left := ctx.Expr(0).Accept(i)
	right := ctx.Expr(1).Accept(i)
	op := ctx.GetOp().GetText()

	// Handle string equality
	lStr, lOk := left.(StringValue)
	rStr, rOk := right.(StringValue)
	if lOk && rOk {
		if op == "==" {
			return lStr.Val == rStr.Val
		} else if op == "!=" {
			return lStr.Val != rStr.Val
		}
	}

	// Handle int equality
	lInt, lOkInt := left.(IntValue)
	rInt, rOkInt := right.(IntValue)
	if lOkInt && rOkInt {
		if op == "==" {
			return lInt.Val == rInt.Val
		} else if op == "!=" {
			return lInt.Val != rInt.Val
		}
	}

	return false
}

func (i *Interpreter) VisitReturnStmt(ctx *ast.ReturnStmtContext) interface{} {
	if ctx.Expr() != nil {
		val := ctx.Expr().Accept(i)
		if v, ok := val.(Value); ok {
			return ReturnValue{Val: v}
		}
	}
	return ReturnValue{Val: nil}
}

func (i *Interpreter) VisitQuoteExpr(ctx *ast.QuoteExprContext) interface{} {
	if ctx.Block() == nil {
		return StringValue{Val: ""}
	}

	start := ctx.Block().GetStart()
	stop := ctx.Block().GetStop()
	input := start.GetInputStream()
	interval := antlr.NewInterval(start.GetStart(), stop.GetStop())
	blockText := input.GetTextFromInterval(interval)

	// Strip braces
	if strings.HasPrefix(blockText, "{") && strings.HasSuffix(blockText, "}") {
		blockText = blockText[1 : len(blockText)-1]
	}

	// Collect all visible variables from current and outer scopes
	vars := make(map[string]Value)
	currentEnv := i.Env
	for currentEnv != nil {
		for name, val := range currentEnv.Values {
			if _, exists := vars[name]; !exists {
				vars[name] = val
			}
		}
		currentEnv = currentEnv.Outer
	}

	// Naive regex replacement for Env variables
	res := blockText
	for name, val := range vars {
		if name == "meta" || val == nil {
			continue
		}
		// Skip MetaValue replacement to avoid breaking dot access like r.path
		// Users should extract primitive values to variables before using them in quote
		if val.Type() == "meta" || val.Type() == "list" {
			continue
		}
		// Regex replace \bname\b -> val.String()
		// Warning: This assumes variable names are valid regex word characters
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\b`)
		res = re.ReplaceAllString(res, val.String())
	}

	return StringValue{Val: res}
}

func (i *Interpreter) VisitIntExpr(ctx *ast.IntExprContext) interface{} {
	val, err := strconv.Atoi(ctx.GetText())
	if err != nil {
		return IntValue{Val: 0}
	}
	return IntValue{Val: val}
}

func (i *Interpreter) VisitArrayIndexExpr(ctx *ast.ArrayIndexExprContext) interface{} {
	// expr '[' expr ']'
	arr := ctx.Expr(0).Accept(i)
	idx := ctx.Expr(1).Accept(i)

	listVal, ok := arr.(ListValue)
	if !ok {
		return nil
	}

	var index int
	if intVal, ok := idx.(IntValue); ok {
		index = intVal.Val
	} else if strVal, ok := idx.(StringValue); ok {
		// Fallback for when index is a string literal but we expect int
		// (though parser should handle this as IntExpr usually)
		val, err := strconv.Atoi(strVal.Val)
		if err == nil {
			index = val
		} else {
			return nil
		}
	} else {
		return nil
	}

	if index >= 0 && index < len(listVal.Val) {
		return listVal.Val[index]
	}

	return nil
}

func (i *Interpreter) VisitStringExpr(ctx *ast.StringExprContext) interface{} {
	s := ctx.GetText()
	unquoted, err := strconv.Unquote(s)
	if err == nil {
		return StringValue{Val: unquoted}
	}
	// Fallback if unquote fails (shouldn't happen if lexer is correct)
	if len(s) >= 2 {
		return StringValue{Val: s[1 : len(s)-1]}
	}
	return StringValue{Val: ""}
}

func (i *Interpreter) VisitParenExpr(ctx *ast.ParenExprContext) interface{} {
	return ctx.Expr().Accept(i)
}

func (i *Interpreter) VisitAddSubExpr(ctx *ast.AddSubExprContext) interface{} {
	left := ctx.Expr(0).Accept(i)
	right := ctx.Expr(1).Accept(i)
	op := ctx.GetOp().GetText()

	if op == "+" {
		lStr, lOk := left.(StringValue)
		rStr, rOk := right.(StringValue)
		if lOk && rOk {
			return StringValue{Val: lStr.Val + rStr.Val}
		}
		// Allow string concatenation with other types
		if lOk {
			if rVal, ok := right.(Value); ok {
				return StringValue{Val: lStr.Val + rVal.String()}
			}
		}
		if rOk {
			if lVal, ok := left.(Value); ok {
				return StringValue{Val: lVal.String() + rStr.Val}
			}
		}
		// TODO: Int addition
	}
	return nil
}

func (i *Interpreter) VisitIdentifierExpr(ctx *ast.IdentifierExprContext) interface{} {
	name := ctx.GetText()
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		val := i.resolveValue(parts[0])
		for _, part := range parts[1:] {
			if val == nil {
				return nil
			}
			if meta, ok := val.(MetaValue); ok {
				if v, exists := meta.Props[part]; exists {
					val = v
				} else {
					return nil
				}
			} else {
				return nil
			}
		}
		return val
	}

	val := i.resolveValue(name)
	return val
}

func (i *Interpreter) resolveValue(name string) Value {
	if val, ok := i.Env.Get(name); ok {
		return val
	}
	return nil
}

func (i *Interpreter) VisitFuncCallExpr(ctx *ast.FuncCallExprContext) interface{} {
	funcName := ctx.QualifiedName().GetText()
	// fmt.Fprintf(os.Stderr, "DEBUG: VisitFuncCallExpr %s\n", funcName)

	if funcName == "find_all_annotated_with" {
		if ctx.ExprList() == nil || len(ctx.ExprList().AllExpr()) != 1 {
			// Error: expected 1 argument
			return nil
		}
		argExpr := ctx.ExprList().Expr(0)
		argVal := argExpr.Accept(i)

		annName := ""
		if s, ok := argVal.(StringValue); ok {
			annName = s.Val
		} else {
			// Error: expected string argument
			return nil
		}

		syms := i.Scope.CollectAnnotatedSymbols(annName)
		var list []Value
		for _, sym := range syms {
			meta := i.createSymbolMeta(sym)
			list = append(list, meta)
		}
		return ListValue{Val: list}
	} else if funcName == "println" {
		if ctx.ExprList() != nil {
			var args []interface{}
			for _, expr := range ctx.ExprList().AllExpr() {
				val := expr.Accept(i)
				if v, ok := val.(Value); ok {
					args = append(args, v.String())
				} else {
					args = append(args, val)
				}
			}
			fmt.Println(args...)
		} else {
			fmt.Println()
		}
		return nil
	}

	return nil
}

func (i *Interpreter) createSymbolMeta(sym *symbols.Symbol) MetaValue {
	props := make(map[string]Value)
	props["name"] = StringValue{Val: sym.MyGoName}
	props["go_name"] = StringValue{Val: sym.GoName}
	props["kind"] = StringValue{Val: string(sym.Kind)}
	props["pkg"] = StringValue{Val: sym.PackageName}

	var anns []Value
	for _, ann := range sym.Annotations {
		annProps := make(map[string]Value)
		annProps["name"] = StringValue{Val: ann.Name}
		var args []Value
		for _, arg := range ann.Args {
			args = append(args, StringValue{Val: arg})
		}
		annProps["args"] = ListValue{Val: args}
		anns = append(anns, MetaValue{Props: annProps})
	}
	props["annotations"] = ListValue{Val: anns}

	return MetaValue{Props: props}
}

func (i *Interpreter) VisitMemberAccessExpr(ctx *ast.MemberAccessExprContext) interface{} {
	obj := ctx.Expr().Accept(i)
	prop := ctx.ID().GetText()

	if meta, ok := obj.(MetaValue); ok {
		if val, exists := meta.Props[prop]; exists {
			return val
		}
	}
	return nil
}

func (i *Interpreter) VisitSingleLetDecl(ctx *ast.SingleLetDeclContext) interface{} {
	name := ctx.ID().GetText()
	var val Value
	if ctx.Expr() != nil {
		res := ctx.Expr().Accept(i)
		if v, ok := res.(Value); ok {
			val = v
		}
	}
	i.defineValue(name, val)
	return nil
}

func (i *Interpreter) defineValue(name string, val Value) {
	i.Env.Set(name, val)
}

func (i *Interpreter) VisitAssignmentStmt(ctx *ast.AssignmentStmtContext) interface{} {
	// LHS = RHS
	// AssignmentStmt: expr '=' expr ';'
	left := ctx.Expr(0)
	right := ctx.Expr(1)

	val := right.Accept(i)
	if v, ok := val.(Value); ok {
		// LHS must be ID
		if idCtx, ok := left.(*ast.IdentifierExprContext); ok {
			name := idCtx.GetText()
			i.Env.Update(name, v)
		}
	}
	return nil
}
