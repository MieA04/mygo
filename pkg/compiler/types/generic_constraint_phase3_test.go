package types_test

import (
	"strings"
	"testing"
)

func TestGenericConstraintPhase3_WhereEnumConstraintTranspile(t *testing.T) {
	goCode := transpileSource(`
where T: int
enum Option<T> { Some(T), None }
`)

	if !strings.Contains(goCode, "type option[T int] interface") {
		t.Fatalf("expected where enum constraint to be emitted in enum generics, got:\n%s", goCode)
	}
}

func TestGenericConstraintPhase3_WhereEnumUnknownParamSemanticError(t *testing.T) {
	out := semanticOutput(`
where X: int
enum Option<T> { Some(T), None }
`)

	if !strings.Contains(out, "E_WHERE_UNKNOWN_PARAM") {
		t.Fatalf("expected unknown where param error in enum declaration, got:\n%s", out)
	}
}

func TestGenericConstraintPhase3_WhereEnumConflictSemanticError(t *testing.T) {
	out := semanticOutput(`
where T: string
enum Option<T: int> { Some(T), None }
`)

	if !strings.Contains(out, "E_WHERE_CONFLICT_CONSTRAINT") {
		t.Fatalf("expected where conflict constraint error in enum declaration, got:\n%s", out)
	}
}

func TestGenericConstraintPhase3_WhereEnumDuplicateParamSemanticError(t *testing.T) {
	out := semanticOutput(`
where T: int, T: int
enum Option<T> { Some(T), None }
`)

	if !strings.Contains(out, "E_WHERE_DUP_PARAM") {
		t.Fatalf("expected duplicate where param error in enum declaration, got:\n%s", out)
	}
}
