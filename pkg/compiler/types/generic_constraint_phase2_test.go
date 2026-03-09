package types_test

import (
	"strings"
	"testing"
)

func TestGenericConstraintPhase2_WhereTraitConstraintTranspile(t *testing.T) {
	goCode := transpileSource(`
where T: int
trait Iterator<T> {
    fn next(): T;
}
`)

	if !strings.Contains(goCode, "type iterator[T int] interface") {
		t.Fatalf("expected where trait constraint to be emitted in trait generics, got:\n%s", goCode)
	}
}

func TestGenericConstraintPhase2_WhereTraitUnknownParamSemanticError(t *testing.T) {
	out := semanticOutput(`
where X: int
trait Iterator<T> {
    fn next(): T;
}
`)

	if !strings.Contains(out, "E_WHERE_UNKNOWN_PARAM") {
		t.Fatalf("expected unknown where param error in trait declaration, got:\n%s", out)
	}
}

func TestGenericConstraintPhase2_WhereBindConstraintTranspile(t *testing.T) {
	goCode := transpileSource(`
enum Opt { None }

where T: int
trait bind<T>(o: Opt) {
    fn make(v: T): T {
        return v;
    }
}
`)

	if !strings.Contains(goCode, "func make[T int](o opt, v T) T") {
		t.Fatalf("expected where bind constraint to be emitted in generated bind method, got:\n%s", goCode)
	}
}

func TestGenericConstraintPhase2_WhereBindConflictSemanticError(t *testing.T) {
	out := semanticOutput(`
enum Opt { None }

where T: string
trait bind<T: int>(o: Opt) {
    fn make(v: T): T {
        return v;
    }
}
`)

	if !strings.Contains(out, "E_WHERE_CONFLICT_CONSTRAINT") {
		t.Fatalf("expected where conflict constraint error in bind declaration, got:\n%s", out)
	}
}

func TestGenericConstraintPhase2_WhereTraitDuplicateParamSemanticError(t *testing.T) {
	out := semanticOutput(`
where T: int, T: int
trait Iterator<T> {
    fn next(): T;
}
`)

	if !strings.Contains(out, "E_WHERE_DUP_PARAM") {
		t.Fatalf("expected duplicate where param error in trait declaration, got:\n%s", out)
	}
}

func TestGenericConstraintPhase2_WhereBindDuplicateParamSemanticError(t *testing.T) {
	out := semanticOutput(`
enum Opt { None }

where T: int, T: int
trait bind<T>(o: Opt) {
    fn make(v: T): T {
        return v;
    }
}
`)

	if !strings.Contains(out, "E_WHERE_DUP_PARAM") {
		t.Fatalf("expected duplicate where param error in bind declaration, got:\n%s", out)
	}
}
