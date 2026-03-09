package types_test

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/antlr4-go/antlr/v4"
	"github.com/miea04/mygo/pkg/ast"
	"github.com/miea04/mygo/pkg/compiler/semantic"
	"github.com/miea04/mygo/pkg/compiler/symbols"
	"github.com/miea04/mygo/pkg/compiler/transpiler"
	"github.com/miea04/mygo/pkg/compiler/types"
)

func analyzeScope(src string) *symbols.Scope {
	input := antlr.NewInputStream(src)
	lexer := ast.NewMyGoLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := ast.NewMyGoParser(stream)
	tree := parser.Program()
	global := symbols.NewScope("global", nil)
	analyzer := semantic.NewSemanticAnalyzer(global)
	for _, stmt := range tree.AllStatement() {
		stmt.Accept(analyzer)
	}
	return global
}

func transpileSource(src string) string {
	input := antlr.NewInputStream(src)
	lexer := ast.NewMyGoLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := ast.NewMyGoParser(stream)
	tree := parser.Program()
	global := symbols.NewScope("global", nil)
	analyzer := semantic.NewSemanticAnalyzer(global)
	for _, stmt := range tree.AllStatement() {
		stmt.Accept(analyzer)
	}
	tp := transpiler.NewMyGoTranspiler(global)
	var out strings.Builder
	for _, stmt := range tree.AllStatement() {
		out.WriteString(stmt.Accept(tp).(string))
		out.WriteString("\n")
	}
	return out.String()
}

func semanticOutput(src string) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	input := antlr.NewInputStream(src)
	lexer := ast.NewMyGoLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := ast.NewMyGoParser(stream)
	tree := parser.Program()
	global := symbols.NewScope("global", nil)
	analyzer := semantic.NewSemanticAnalyzer(global)
	for _, stmt := range tree.AllStatement() {
		stmt.Accept(analyzer)
	}

	_ = w.Close()
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestResolveBinaryOp_UsesTraitOverload(t *testing.T) {
	scope := analyzeScope(`
struct Vec2 {
    x: int
}

trait bind(v: Vec2) {
    fn add(rhs: Vec2): Vec2 {
        return v;
    }

    fn eq(rhs: Vec2): bool {
        return true;
    }
}
`)

	kind, _, method, negate, ok := types.ResolveBinaryOp("+", "Vec2", "Vec2", scope)
	if !ok {
		t.Fatalf("expected overload resolution success for Vec2 + Vec2")
	}
	if kind != "overload_method" {
		t.Fatalf("expected overload_method, got %q", kind)
	}
	if method != "add" {
		t.Fatalf("expected add method, got %q", method)
	}
	if negate {
		t.Fatalf("unexpected negate for + operator")
	}

	kind, resultType, method, negate, ok := types.ResolveBinaryOp("!=", "Vec2", "Vec2", scope)
	if !ok {
		t.Fatalf("expected overload resolution success for Vec2 != Vec2")
	}
	if kind != "overload_method" || method != "eq" || !negate {
		t.Fatalf("unexpected resolution for !=: kind=%q method=%q negate=%v", kind, method, negate)
	}
	if resultType != "bool" {
		t.Fatalf("expected bool result for eq overload, got %q", resultType)
	}
}

func TestResolveBinaryOp_BuiltinPrecedence(t *testing.T) {
	scope := analyzeScope(`
trait bind(i: int) {
    fn add(rhs: int): int {
        return i;
    }
}
`)

	kind, resultType, method, _, ok := types.ResolveBinaryOp("+", "int", "int", scope)
	if !ok {
		t.Fatalf("expected builtin resolution success for int + int")
	}
	if kind != "builtin_numeric" {
		t.Fatalf("expected builtin_numeric, got %q", kind)
	}
	if method != "" {
		t.Fatalf("expected no method for builtin numeric op, got %q", method)
	}
	if resultType != "int" {
		t.Fatalf("expected int result type, got %q", resultType)
	}
}

func TestTranspile_OperatorOverloadAndComparePromotion(t *testing.T) {
	goCode := transpileSource(`
struct Vec2 {
    x: int
}

trait bind(v: Vec2) {
    fn add(rhs: Vec2): Vec2 {
        return v;
    }
    fn eq(rhs: Vec2): bool {
        return true;
    }
}

fn main() {
    let a = Vec2{x: 1};
    let b = Vec2{x: 2};
    let c = a + b;
    let ok = a != b;
    let i: int = 1;
    let f: float64 = 1.5;
    let cmp = i == f;
    print(c, ok, cmp);
}
`)

	if !strings.Contains(goCode, "a.add(b)") {
		t.Fatalf("expected overloaded add call in transpiled code, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "!(a.eq(b))") {
		t.Fatalf("expected overloaded != lowering to negated eq call, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "float64(i) == f") {
		t.Fatalf("expected numeric compare implicit promotion, got:\n%s", goCode)
	}
}

func TestResolveBinaryOp_UsesTraitOverloadForMulDivAndOrder(t *testing.T) {
	scope := analyzeScope(`
struct Vec2 {
    x: int
}

trait bind(v: Vec2) {
    fn mul(rhs: Vec2): Vec2 {
        return v;
    }
    fn div(rhs: Vec2): Vec2 {
        return v;
    }
    fn lt(rhs: Vec2): bool {
        return true;
    }
    fn le(rhs: Vec2): bool {
        return true;
    }
    fn gt(rhs: Vec2): bool {
        return true;
    }
    fn ge(rhs: Vec2): bool {
        return true;
    }
}
`)

	cases := []struct {
		op           string
		wantMethod   string
		wantRetType  string
		expectNegate bool
	}{
		{op: "*", wantMethod: "mul", wantRetType: "Vec2"},
		{op: "/", wantMethod: "div", wantRetType: "Vec2"},
		{op: "<", wantMethod: "lt", wantRetType: "bool"},
		{op: "<=", wantMethod: "le", wantRetType: "bool"},
		{op: ">", wantMethod: "gt", wantRetType: "bool"},
		{op: ">=", wantMethod: "ge", wantRetType: "bool"},
	}

	for _, tc := range cases {
		kind, resultType, method, negate, ok := types.ResolveBinaryOp(tc.op, "Vec2", "Vec2", scope)
		if !ok {
			t.Fatalf("expected overload resolution success for Vec2 %s Vec2", tc.op)
		}
		if kind != "overload_method" {
			t.Fatalf("expected overload_method for %s, got %q", tc.op, kind)
		}
		if method != tc.wantMethod {
			t.Fatalf("expected method %q for %s, got %q", tc.wantMethod, tc.op, method)
		}
		if resultType != tc.wantRetType {
			t.Fatalf("expected result %q for %s, got %q", tc.wantRetType, tc.op, resultType)
		}
		if negate != tc.expectNegate {
			t.Fatalf("unexpected negate for %s: got %v", tc.op, negate)
		}
	}
}

func TestTranspile_OperatorOverloadMulDivAndOrder(t *testing.T) {
	goCode := transpileSource(`
struct Vec2 {
    x: int
}

trait bind(v: Vec2) {
    fn mul(rhs: Vec2): Vec2 {
        return v;
    }
    fn div(rhs: Vec2): Vec2 {
        return v;
    }
    fn lt(rhs: Vec2): bool {
        return true;
    }
    fn le(rhs: Vec2): bool {
        return true;
    }
    fn gt(rhs: Vec2): bool {
        return true;
    }
    fn ge(rhs: Vec2): bool {
        return true;
    }
}

fn main() {
    let a = Vec2{x: 1};
    let b = Vec2{x: 2};
    let p = a * b;
    let q = a / b;
    let l = a < b;
    let leq = a <= b;
    let g = a > b;
    let geq = a >= b;
    print(p, q, l, leq, g, geq);
}
`)

	expected := []string{
		"a.mul(b)",
		"a.div(b)",
		"a.lt(b)",
		"a.le(b)",
		"a.gt(b)",
		"a.ge(b)",
	}
	for _, fragment := range expected {
		if !strings.Contains(goCode, fragment) {
			t.Fatalf("expected overloaded operator lowering %q, got:\n%s", fragment, goCode)
		}
	}
}

func TestTranspile_IsAndNotIsUseAnyAssertion(t *testing.T) {
	goCode := transpileSource(`
struct Dog {
    name: string
}

fn main() {
    let d = Dog{name: "旺财"};
    let ok = d is Dog;
    let nok = d !is Dog;
    print(ok, nok);
}
`)

	if !strings.Contains(goCode, "any(d).(dog)") {
		t.Fatalf("expected is/!is to assert on any(expr), got:\n%s", goCode)
	}
}

func TestTranspile_IfIsSmartCastForIdentifier(t *testing.T) {
	goCode := transpileSource(`
trait Runnable {
    fn run();
}

struct Dog {}

trait bind(d: Dog) {
    fn run() {
        print("run");
    }
}

fn main() {
    let obj: any = Dog{};
    if obj is Runnable {
        obj.run();
    }
}
`)

	if !strings.Contains(goCode, "if _mygo_is_v, _mygo_is_ok := any(obj).(runnable); _mygo_is_ok") {
		t.Fatalf("expected if-is smart cast init clause, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "obj := _mygo_is_v") {
		t.Fatalf("expected identifier shadowing inside if branch, got:\n%s", goCode)
	}
}

func TestTranspile_MatchTypeCaseUsesTypeSwitch(t *testing.T) {
	goCode := transpileSource(`
trait Runnable {
    fn run();
}

struct Dog {}

trait bind(d: Dog) {
    fn run() {
        print("run");
    }
}

fn main() {
    let obj: any = Dog{};
    match obj {
        is Runnable => {
            obj.run();
        }
        other => {
            print("other");
        }
    }
}
`)

	if !strings.Contains(goCode, "switch _match_v := any(obj).(type)") {
		t.Fatalf("expected match type-case to use Go type switch, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "case runnable:") {
		t.Fatalf("expected type match case branch, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "obj := _match_v") {
		t.Fatalf("expected match type-case smart cast shadowing, got:\n%s", goCode)
	}
}

func TestTranspile_CastToTraitUsesAnyAssertion(t *testing.T) {
	goCode := transpileSource(`
trait Runnable {
    fn run();
}

struct Dog {}

trait bind(d: Dog) {
    fn run() {
        print("run");
    }
}

fn main() {
    let d = Dog{};
    let r = d to Runnable;
    r.run();
}
`)

	if !strings.Contains(goCode, "any(d).(runnable)") {
		t.Fatalf("expected cast-to-trait to use any assertion, got:\n%s", goCode)
	}
}

func TestTranspile_CastNumericStillUsesGoConversion(t *testing.T) {
	goCode := transpileSource(`
fn main() {
    let i: int = 7;
    let f = i to float64;
    print(f);
}
`)

	if !strings.Contains(goCode, "float64(i)") {
		t.Fatalf("expected numeric cast to keep Go conversion form, got:\n%s", goCode)
	}
}

func TestSemantic_IsUndefinedTypeReportsError(t *testing.T) {
	output := semanticOutput(`
struct Dog {}

fn main() {
    let d = Dog{};
    let ok = d is MissingType;
    print(ok);
}
`)

	if !strings.Contains(output, "E_IS_TYPE_UNDEFINED") {
		t.Fatalf("expected undefined type error code for is, got:\n%s", output)
	}
}

func TestSemantic_IsEnumReportsInvalidKind(t *testing.T) {
	output := semanticOutput(`
enum Status {
    Ok
}

fn main() {
    let s = Status.Ok;
    let ok = s is Status;
    print(ok);
}
`)

	if !strings.Contains(output, "E_IS_TYPE_INVALID_KIND") {
		t.Fatalf("expected invalid kind error code for enum is-check, got:\n%s", output)
	}
}

func TestSemantic_TraitInstantiationForbidden(t *testing.T) {
	output := semanticOutput(`
trait Runnable {
    fn run();
}

fn main() {
    let r = Runnable{};
    print(r);
}
`)

	if !strings.Contains(output, "E_TRAIT_INSTANTIATION_FORBIDDEN") {
		t.Fatalf("expected trait instantiation forbidden error, got:\n%s", output)
	}
}

func TestTranspile_IsTraitHierarchyByTypeName(t *testing.T) {
	goCode := transpileSource(`
trait TraitD {
    fn d(): int { return 1; }
}

trait TraitC {
    fn c(): int { return 1; }
}

trait TraitB {
    fn b(): int { return 1; }
}

trait TraitA {
    fn a(): int { return 1; }
}

trait bind(x: TraitA) combs(TraitB) {}
trait bind(x: TraitB) combs(TraitC) {}
trait bind(x: TraitC) combs(TraitD) {}

fn main() {
    let ok = TraitA is TraitD;
    let nok = TraitA !is TraitD;
    if TraitA is TraitD {
        print("ok");
    }
    print(ok, nok);
}
`)

	if !strings.Contains(goCode, "let ok = true") && !strings.Contains(goCode, "ok := true") {
		t.Fatalf("expected trait hierarchy is-check to fold to true, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "let nok = false") && !strings.Contains(goCode, "nok := false") {
		t.Fatalf("expected trait hierarchy !is-check to fold to false, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "if true") {
		t.Fatalf("expected if TraitA is TraitD to fold to if true, got:\n%s", goCode)
	}
}

func TestTranspile_IsTraitHierarchyByStructBinding(t *testing.T) {
	goCode := transpileSource(`
trait TraitD {
    fn d(): int { return 1; }
}

trait TraitC {
    fn c(): int { return 1; }
}

trait TraitB {
    fn b(): int { return 1; }
}

trait TraitA {
    fn a(): int { return 1; }
}

struct StructA {}

trait bind(x: TraitA) combs(TraitB) {}
trait bind(x: TraitB) combs(TraitC) {}
trait bind(x: TraitC) combs(TraitD) {}
trait bind(x: StructA) combs(TraitA) {}

fn main() {
    let ok = StructA is TraitD;
    let nok = StructA !is TraitD;
    print(ok, nok);
}
`)

	if !strings.Contains(goCode, "ok := true") {
		t.Fatalf("expected StructA is TraitD to fold to true, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "nok := false") {
		t.Fatalf("expected StructA !is TraitD to fold to false, got:\n%s", goCode)
	}
}

func TestTranspile_MatchTypeCaseTraitHierarchyByTypeName(t *testing.T) {
	goCode := transpileSource(`
trait TraitD {
    fn d(): int { return 1; }
}

trait TraitC {
    fn c(): int { return 1; }
}

trait TraitB {
    fn b(): int { return 1; }
}

trait TraitA {
    fn a(): int { return 1; }
}

trait bind(x: TraitA) combs(TraitB) {}
trait bind(x: TraitB) combs(TraitC) {}
trait bind(x: TraitC) combs(TraitD) {}

fn main() {
    match TraitA {
        is TraitD => {
            print("d");
        }
        other => {
            print("other");
        }
    }
}
`)

	if !strings.Contains(goCode, "switch true {") {
		t.Fatalf("expected compile-time folded type-name match switch, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "case true:") {
		t.Fatalf("expected TraitA is TraitD branch to fold to true, got:\n%s", goCode)
	}
}

func TestTranspile_MatchTypeCaseTraitHierarchyByStructBinding(t *testing.T) {
	goCode := transpileSource(`
trait TraitD {
    fn d(): int { return 1; }
}

trait TraitC {
    fn c(): int { return 1; }
}

trait TraitB {
    fn b(): int { return 1; }
}

trait TraitA {
    fn a(): int { return 1; }
}

struct StructA {}

trait bind(x: TraitA) combs(TraitB) {}
trait bind(x: TraitB) combs(TraitC) {}
trait bind(x: TraitC) combs(TraitD) {}
trait bind(x: StructA) combs(TraitA) {}

fn main() {
    match StructA {
        is TraitD => {
            print("d");
        }
        is TraitA => {
            print("a");
        }
        other => {
            print("other");
        }
    }
}
`)

	if !strings.Contains(goCode, "switch true {") {
		t.Fatalf("expected compile-time folded struct-name match switch, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "case true:") {
		t.Fatalf("expected StructA is TraitD/TraitA branch to contain true case, got:\n%s", goCode)
	}
}

func TestTranspile_MatchMixedValueAndTypeCaseUsesBoolSwitch(t *testing.T) {
	goCode := transpileSource(`
fn main() {
    let x: any = 1;
    match x {
        1 => {
            print("one");
        }
        is int => {
            print(x + 2);
        }
        other => {
            print("other");
        }
    }
}
`)

	if !strings.Contains(goCode, "switch true {") {
		t.Fatalf("expected mixed value/type match to use bool switch, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "case x == 1:") {
		t.Fatalf("expected value case to lower as equality condition, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "case func() bool { _, _ok := any(x).(int); return _ok }():") {
		t.Fatalf("expected type case to lower as runtime type condition, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "x := any(x).(int)") {
		t.Fatalf("expected mixed type case branch to re-bind narrowed identifier, got:\n%s", goCode)
	}
}
