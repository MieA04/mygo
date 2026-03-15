package transpiler

import (
	"strings"
	"testing"

	"github.com/miea04/mygo/pkg/compiler/symbols"
)

func TestToGoTypeOption(t *testing.T) {
	scope := symbols.NewScope("global", nil)
	v := NewMyGoTranspiler(scope)
	got := v.toGoType("Option<int>")
	if got != "Option[int]" {
		t.Fatalf("toGoType Option<int> = %q, want %q", got, "Option[int]")
	}
}

func TestBoxOptionExpr(t *testing.T) {
	scope := symbols.NewScope("global", nil)
	v := NewMyGoTranspiler(scope)

	gotSome := v.boxOptionExpr("x", "int", "Option<int>")
	if gotSome != "__MYGO_OPTION__Some[int](x)" {
		t.Fatalf("box some = %q", gotSome)
	}

	gotNone := v.boxOptionExpr("nil", "nil", "Option<int>")
	if gotNone != "__MYGO_OPTION__None[int]()" {
		t.Fatalf("box none = %q", gotNone)
	}

	gotRaw := v.boxOptionExpr("x", "int", "int")
	if gotRaw != "x" {
		t.Fatalf("non-option target should keep expr, got %q", gotRaw)
	}

	if !v.needsOptionHelper {
		t.Fatalf("expected needsOptionHelper = true after option boxing")
	}
}

func TestBuildOptionTryUnwrap(t *testing.T) {
	scope := symbols.NewScope("global", nil)
	v := NewMyGoTranspiler(scope)
	got, ok := v.buildOptionTryUnwrap("opt", "Option<int>")
	if !ok {
		t.Fatalf("expected option try unwrap enabled")
	}
	if !strings.Contains(got, "__MYGO_TRY_UNWRAP_OPTION__opt__INNER__int__BLOCK__") {
		t.Fatalf("unexpected marker: %q", got)
	}
	if !strings.Contains(got, "return Option_None[int]{}") {
		t.Fatalf("unexpected block: %q", got)
	}
}

func TestBuildOptionPanicUnwrap(t *testing.T) {
	scope := symbols.NewScope("global", nil)
	v := NewMyGoTranspiler(scope)
	got, ok := v.buildOptionPanicUnwrap("opt", "Option<int>")
	if !ok {
		t.Fatalf("expected option panic unwrap enabled")
	}
	if !strings.Contains(got, "case Option_Some[int]:") {
		t.Fatalf("unexpected panic unwrap code: %q", got)
	}
	if !strings.Contains(got, "panic(\"panic unwrap on None\")") {
		t.Fatalf("unexpected panic unwrap code: %q", got)
	}
}

func TestParseTryUnwrapMarkerOption(t *testing.T) {
	scope := symbols.NewScope("global", nil)
	v := NewMyGoTranspiler(scope)
	code, ok := v.buildOptionTryUnwrap("opt", "Option<int>")
	if !ok {
		t.Fatalf("expected option marker")
	}
	kind, expr, inner, block, parsed := parseTryUnwrapMarker(code)
	if !parsed || kind != "option" {
		t.Fatalf("parse marker failed: kind=%s parsed=%v", kind, parsed)
	}
	if expr != "opt" || inner != "int" {
		t.Fatalf("marker payload invalid: expr=%q inner=%q", expr, inner)
	}
	if !strings.Contains(block, "Option_None[int]{}") {
		t.Fatalf("block invalid: %q", block)
	}
}

func TestSpecializedTaggedOptionalFieldGoType(t *testing.T) {
	scope := symbols.NewScope("global", nil)
	v := NewMyGoTranspiler(scope)

	got, ok := v.specializedTaggedOptionalFieldGoType("Option<int>", "\"json:\\\"id\\\"\"")
	if !ok || got != "*int" {
		t.Fatalf("specialized type = (%q, %v), want (*int, true)", got, ok)
	}

	if _, ok := v.specializedTaggedOptionalFieldGoType("Option<int>", "\"db:\\\"id\\\"\""); ok {
		t.Fatalf("db tag should not trigger specialization")
	}
}

func TestBoxTaggedOptionalFieldExpr(t *testing.T) {
	scope := symbols.NewScope("global", nil)
	v := NewMyGoTranspiler(scope)
	tag := "\"json:\\\"id\\\"\""

	gotValue := v.boxTaggedOptionalFieldExpr("x", "int", "Option<int>", tag)
	if !strings.Contains(gotValue, "func() *int") || !strings.Contains(gotValue, "__mygo_v := x") {
		t.Fatalf("value boxing invalid: %q", gotValue)
	}

	gotNil := v.boxTaggedOptionalFieldExpr("nil", "nil", "Option<int>", tag)
	if gotNil != "nil" {
		t.Fatalf("nil boxing invalid: %q", gotNil)
	}

	gotOption := v.boxTaggedOptionalFieldExpr("opt", "Option<int>", "Option<int>", tag)
	if !strings.Contains(gotOption, "case Option_Some[int]:") || !strings.Contains(gotOption, "default: return nil") {
		t.Fatalf("option boxing invalid: %q", gotOption)
	}
}

func TestResolveStructFieldTypeAndTag(t *testing.T) {
	scope := symbols.NewScope("global", nil)
	v := NewMyGoTranspiler(scope)

	sym := &symbols.Symbol{
		Kind: symbols.KindStruct,
		FieldMap: map[string]*symbols.FieldSymbol{
			"id": {Name: "id", Type: "Option<T>", Tag: "\"json:\\\"id\\\"\""},
		},
		GenericParams: []symbols.GenericParamMeta{{Name: "T"}},
	}

	fieldType, fieldTag, ok := v.resolveStructFieldTypeAndTag(sym, "[int]", "id")
	if !ok {
		t.Fatalf("expected resolve success")
	}
	if fieldType != "Option<int>" {
		t.Fatalf("field type = %q, want Option<int>", fieldType)
	}
	if fieldTag != "\"json:\\\"id\\\"\"" {
		t.Fatalf("field tag = %q", fieldTag)
	}
}

func BenchmarkBoxTaggedOptionalFieldExpr_JSONTag(b *testing.B) {
	scope := symbols.NewScope("global", nil)
	v := NewMyGoTranspiler(scope)
	tag := "\"json:\\\"id\\\"\""
	for i := 0; i < b.N; i++ {
		_ = v.boxTaggedOptionalFieldExpr("opt", "Option<int>", "Option<int>", tag)
	}
}

func BenchmarkBoxTaggedOptionalFieldExpr_NoJSONTag(b *testing.B) {
	scope := symbols.NewScope("global", nil)
	v := NewMyGoTranspiler(scope)
	tag := "\"db:\\\"id\\\"\""
	for i := 0; i < b.N; i++ {
		_ = v.boxTaggedOptionalFieldExpr("opt", "Option<int>", "Option<int>", tag)
	}
}
