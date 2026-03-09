package types_test

import (
	"strings"
	"testing"
)

func TestPointerAchieve_TranspileAddrDerefAndNil(t *testing.T) {
	goCode := transpileSource(`
struct User {
    name: string
}

fn main() {
    let u = User{name: "alice"};
    let p: *User = &u;
    let u2 = *p;
    if p != nil {
        print(u2.name);
    }
}
`)

	if !strings.Contains(goCode, "var p *user = &u") {
		t.Fatalf("expected pointer declaration with addr-of, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "u2 := *p") {
		t.Fatalf("expected deref expression transpilation, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "p != nil") {
		t.Fatalf("expected nil comparison transpilation, got:\n%s", goCode)
	}
}

func TestPointerAchieve_AnyToPointerCastUsesAssertion(t *testing.T) {
	goCode := transpileSource(`
struct User {
    name: string
}

fn main() {
    let u = User{name: "alice"};
    let x: any = u;
    let p = x to *User;
    print(p.name);
}
`)

	if !strings.Contains(goCode, "any(x).(*user)") {
		t.Fatalf("expected cast from any to pointer to use assertion, got:\n%s", goCode)
	}
}

func TestPointerAchieve_SemanticRejectNonAddressableAddrOf(t *testing.T) {
	output := semanticOutput(`
fn main() {
    let x = &(1 + 2);
    print(x);
}
`)

	if !strings.Contains(output, "E_PTR_ADDR_NON_ADDRESSABLE") {
		t.Fatalf("expected non-addressable addr-of error, got:\n%s", output)
	}
}

func TestPointerAchieve_SemanticRejectDerefNonPointer(t *testing.T) {
	output := semanticOutput(`
fn main() {
    let i: int = 1;
    let x = *i;
    print(x);
}
`)

	if !strings.Contains(output, "E_PTR_DEREF_NON_POINTER") {
		t.Fatalf("expected non-pointer deref error, got:\n%s", output)
	}
}

func TestPointerAchieve_SemanticRejectNilAssignToInt(t *testing.T) {
	output := semanticOutput(`
fn main() {
    let i: int = nil;
    print(i);
}
`)

	if !strings.Contains(output, "E_PTR_NIL_ASSIGN_INVALID") {
		t.Fatalf("expected nil assign invalid error, got:\n%s", output)
	}
}

func TestPointerAchieve_IfIsPointerSmartCast(t *testing.T) {
	goCode := transpileSource(`
struct User {
    name: string
}

fn main() {
    let u = User{name: "alice"};
    let x: any = &u;
    if x is *User {
        print((*x).name);
    }
}
`)

	if !strings.Contains(goCode, "if _mygo_is_v, _mygo_is_ok := any(x).(*user); _mygo_is_ok") {
		t.Fatalf("expected pointer is-check to use any assertion, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "x := _mygo_is_v") {
		t.Fatalf("expected pointer smart cast shadowing in if branch, got:\n%s", goCode)
	}
}

func TestPointerAchieve_MatchIsPointerUsesTypeSwitch(t *testing.T) {
	goCode := transpileSource(`
struct User {
    name: string
}

fn main() {
    let u = User{name: "alice"};
    let x: any = &u;
    match x {
        is *User => {
            print((*x).name);
        }
        other => {
            print("other");
        }
    }
}
`)

	if !strings.Contains(goCode, "switch _match_v := any(x).(type)") {
		t.Fatalf("expected pointer type-case to use type switch, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "case *user:") {
		t.Fatalf("expected pointer type case branch, got:\n%s", goCode)
	}
	if !strings.Contains(goCode, "x := _match_v") {
		t.Fatalf("expected pointer smart cast shadowing in match branch, got:\n%s", goCode)
	}
}

func TestPointerAchieve_SemanticAllowsPointerTypeCheck(t *testing.T) {
	output := semanticOutput(`
struct User {
    name: string
}

fn main() {
    let u = User{name: "alice"};
    let x: any = &u;
    let ok = x is *User;
    print(ok);
}
`)

	if strings.Contains(output, "E_IS_TYPE_UNDEFINED") || strings.Contains(output, "E_IS_TYPE_INVALID_KIND") {
		t.Fatalf("expected pointer type check to be accepted, got:\n%s", output)
	}
}

func TestPointerAchieve_SemanticRejectDirectNilDeref(t *testing.T) {
	output := semanticOutput(`
fn main() {
    let x = *nil;
    print(x);
}
`)

	if !strings.Contains(output, "E_PTR_DEREF_NIL") {
		t.Fatalf("expected direct nil deref error, got:\n%s", output)
	}
}

func TestPointerAchieve_SemanticWarnPossibleNilDeref(t *testing.T) {
	output := semanticOutput(`
struct User {
    name: string
}

fn main() {
    let u = User{name: "alice"};
    let p: *User = &u;
    let x = *p;
    print(x.name);
}
`)

	if !strings.Contains(output, "W_PTR_DEREF_NIL_POSSIBLE") {
		t.Fatalf("expected possible nil deref warning, got:\n%s", output)
	}
}

func TestPointerAchieve_SemanticSuppressWarnInsideNonNilGuard(t *testing.T) {
	output := semanticOutput(`
struct User {
    name: string
}

fn main() {
    let u = User{name: "alice"};
    let p: *User = &u;
    if p != nil {
        let x = *p;
        print(x.name);
    }
}
`)

	if strings.Contains(output, "W_PTR_DEREF_NIL_POSSIBLE") {
		t.Fatalf("expected nil-deref warning to be suppressed inside non-nil guard, got:\n%s", output)
	}
}

func TestPointerAchieve_SemanticSuppressWarnInsideElseOfNilGuard(t *testing.T) {
	output := semanticOutput(`
struct User {
    name: string
}

fn main() {
    let u = User{name: "alice"};
    let p: *User = &u;
    if p == nil {
        print("nil");
    } else {
        let x = *p;
        print(x.name);
    }
}
`)

	if strings.Contains(output, "W_PTR_DEREF_NIL_POSSIBLE") {
		t.Fatalf("expected nil-deref warning to be suppressed inside else of nil-guard, got:\n%s", output)
	}
}

func TestPointerAchieve_SemanticErrorInsideThenOfNilGuard(t *testing.T) {
	output := semanticOutput(`
struct User {
    name: string
}

fn main() {
    let u = User{name: "alice"};
    let p: *User = &u;
    if p == nil {
        let x = *p;
        print(x.name);
    } else {
        print("ok");
    }
}
`)

	if !strings.Contains(output, "E_PTR_DEREF_GUARDED_NIL") {
		t.Fatalf("expected nil-deref error in then branch of nil-guard, got:\n%s", output)
	}
}

func TestPointerAchieve_SemanticSuppressWarnInsideElseIfAfterNilGuard(t *testing.T) {
	output := semanticOutput(`
struct User {
    name: string
}

fn main() {
    let u = User{name: "alice"};
    let p: *User = &u;
    let ok = true;
    if p == nil {
        print("nil");
    } else if ok {
        let x = *p;
        print(x.name);
    }
}
`)

	if strings.Contains(output, "W_PTR_DEREF_NIL_POSSIBLE") {
		t.Fatalf("expected nil-deref warning to be suppressed in else-if branch after nil-guard, got:\n%s", output)
	}
}

func TestPointerAchieve_SemanticSuppressWarnInsideElseAfterMultiNilGuards(t *testing.T) {
	output := semanticOutput(`
struct User {
    name: string
}

fn main() {
    let u1 = User{name: "a"};
    let u2 = User{name: "b"};
    let p: *User = &u1;
    let q: *User = &u2;
    if p == nil {
        print("p nil");
    } else if q == nil {
        print("q nil");
    } else {
        let x = *p;
        let y = *q;
        print(x.name, y.name);
    }
}
`)

	if strings.Contains(output, "W_PTR_DEREF_NIL_POSSIBLE") {
		t.Fatalf("expected nil-deref warning to be suppressed in else after multi nil-guards, got:\n%s", output)
	}
}

func TestPointerAchieve_SemanticErrorInsideElseAfterNonNilGuard(t *testing.T) {
	output := semanticOutput(`
struct User {
    name: string
}

fn main() {
    let u = User{name: "alice"};
    let p: *User = &u;
    if p != nil {
        print("ok");
    } else {
        let x = *p;
        print(x.name);
    }
}
`)

	if !strings.Contains(output, "E_PTR_DEREF_GUARDED_NIL") {
		t.Fatalf("expected nil-deref error in else branch after non-nil guard, got:\n%s", output)
	}
}

func TestPointerAchieve_SemanticErrorInsideElseIfAfterNonNilGuard(t *testing.T) {
	output := semanticOutput(`
struct User {
    name: string
}

fn main() {
    let u = User{name: "alice"};
    let p: *User = &u;
    let ok = true;
    if p != nil {
        print("ok");
    } else if ok {
        let x = *p;
        print(x.name);
    }
}
`)

	if !strings.Contains(output, "E_PTR_DEREF_GUARDED_NIL") {
		t.Fatalf("expected nil-deref error in else-if branch after non-nil guard, got:\n%s", output)
	}
}

func TestPointerAchieve_SemanticSuppressWarnInsideAndGuard(t *testing.T) {
	output := semanticOutput(`
struct User {
    name: string
}

fn main() {
    let u = User{name: "alice"};
    let p: *User = &u;
    let ok = true;
    if p != nil && ok {
        let x = *p;
        print(x.name);
    }
}
`)

	if strings.Contains(output, "W_PTR_DEREF_NIL_POSSIBLE") {
		t.Fatalf("expected nil-deref warning to be suppressed in and-guard then branch, got:\n%s", output)
	}
}

func TestPointerAchieve_SemanticSuppressWarnInsideElseAfterOrGuard(t *testing.T) {
	output := semanticOutput(`
struct User {
    name: string
}

fn main() {
    let u = User{name: "alice"};
    let p: *User = &u;
    let ok = false;
    if p == nil || ok {
        print("skip");
    } else {
        let x = *p;
        print(x.name);
    }
}
`)

	if strings.Contains(output, "W_PTR_DEREF_NIL_POSSIBLE") {
		t.Fatalf("expected nil-deref warning to be suppressed in else branch after or-guard, got:\n%s", output)
	}
}

func TestPointerAchieve_SemanticKeepWarnInsideOrGuardThen(t *testing.T) {
	output := semanticOutput(`
struct User {
    name: string
}

fn main() {
    let u = User{name: "alice"};
    let p: *User = &u;
    let ok = true;
    if p != nil || ok {
        let x = *p;
        print(x.name);
    }
}
`)

	if !strings.Contains(output, "W_PTR_DEREF_NIL_POSSIBLE") {
		t.Fatalf("expected nil-deref warning in or-guard then branch, got:\n%s", output)
	}
}

func TestPointerAchieve_SemanticErrorInsideNotGuardThen(t *testing.T) {
	output := semanticOutput(`
struct User {
    name: string
}

fn main() {
    let u = User{name: "alice"};
    let p: *User = &u;
    if !(p != nil) {
        let x = *p;
        print(x.name);
    }
}
`)

	if !strings.Contains(output, "E_PTR_DEREF_GUARDED_NIL") {
		t.Fatalf("expected nil-deref error in not-guard then branch, got:\n%s", output)
	}
}
