package types

import (
	"testing"

	"github.com/miea04/mygo/pkg/compiler/symbols"
)

func TestResolveTypeWithScopeOptional(t *testing.T) {
	scope := symbols.NewScope("global", nil)
	cases := []struct {
		in   string
		want string
	}{
		{"int?", "Option<int>"},
		{"User?", "Option<User>"},
		{"Map<string,int>?", "Option<Map<string,int>>"},
		{"int", "int"},
	}
	for _, tc := range cases {
		got := ResolveTypeWithScope(tc.in, scope)
		if got != tc.want {
			t.Fatalf("ResolveTypeWithScope(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestResolveTypeWithScopeRejectsNestedOptional(t *testing.T) {
	scope := symbols.NewScope("global", nil)
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic for nested optional type")
		}
	}()
	_ = ResolveTypeWithScope("int??", scope)
}

func TestIsTypeAssignableOption(t *testing.T) {
	scope := symbols.NewScope("global", nil)
	if !IsTypeAssignable("Option<int>", "int", scope) {
		t.Fatalf("expected int assignable to Option<int>")
	}
	if !IsTypeAssignable("Option<int>", "nil", scope) {
		t.Fatalf("expected nil assignable to Option<int>")
	}
	if !IsTypeAssignable("Option<float64>", "int", scope) {
		t.Fatalf("expected int implicit promote assignable to Option<float64>")
	}
	if IsTypeAssignable("Option<int>", "string", scope) {
		t.Fatalf("did not expect string assignable to Option<int>")
	}
	if !IsTypeAssignable("Option<int>", "Option<int>", scope) {
		t.Fatalf("expected Option<int> assignable to Option<int>")
	}
}
