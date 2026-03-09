package types_test

import (
	"strings"
	"testing"
)

func TestGenericConstraintPhase1_WhereStructConstraintTranspile(t *testing.T) {
	goCode := transpileSource(`
trait Flyable {
    fn fly();
}

where T: Flyable
struct Box<T> {
    item: T
}
`)

	if !strings.Contains(goCode, "type box[T flyable] struct") {
		t.Fatalf("expected where constraint to be emitted on struct generic params, got:\n%s", goCode)
	}
}

func TestGenericConstraintPhase1_WhereIntersectionConstraintTranspile(t *testing.T) {
	goCode := transpileSource(`
trait Flyable {
    fn fly();
}

trait Runnable {
    fn run();
}

where T: Flyable + Runnable
fn move<T>(x: T) {
    print(x);
}
`)

	if !strings.Contains(goCode, "func move[T interface{ flyable; runnable }](x T)") {
		t.Fatalf("expected where intersection constraint to transpile to Go interface intersection, got:\n%s", goCode)
	}
}

func TestGenericConstraintPhase1_WhereUnknownParamSemanticError(t *testing.T) {
	out := semanticOutput(`
trait Flyable {
    fn fly();
}

where X: Flyable
fn move<T>(x: T) {
    print(x);
}
`)

	if !strings.Contains(out, "E_WHERE_UNKNOWN_PARAM") {
		t.Fatalf("expected unknown where param error, got:\n%s", out)
	}
}

func TestGenericConstraintPhase1_WhereConflictSemanticError(t *testing.T) {
	out := semanticOutput(`
trait Flyable {
    fn fly();
}

trait Runnable {
    fn run();
}

where T: Runnable
fn move<T: Flyable>(x: T) {
    print(x);
}
`)

	if !strings.Contains(out, "E_WHERE_CONFLICT_CONSTRAINT") {
		t.Fatalf("expected where conflict constraint error, got:\n%s", out)
	}
}

func TestGenericConstraintPhase1_WhereFnDuplicateParamSemanticError(t *testing.T) {
	out := semanticOutput(`
where T: int, T: int
fn move<T>(x: T) {
    print(x);
}
`)

	if !strings.Contains(out, "E_WHERE_DUP_PARAM") {
		t.Fatalf("expected duplicate where param error in fn declaration, got:\n%s", out)
	}
}

func TestGenericConstraintPhase1_WhereStructDuplicateParamSemanticError(t *testing.T) {
	out := semanticOutput(`
where T: int, T: int
struct Box<T> {
    item: T
}
`)

	if !strings.Contains(out, "E_WHERE_DUP_PARAM") {
		t.Fatalf("expected duplicate where param error in struct declaration, got:\n%s", out)
	}
}
