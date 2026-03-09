package types_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildAndRunTranspiledGo(t *testing.T, goCode string) string {
	t.Helper()

	workDir := t.TempDir()
	goFile := filepath.Join(workDir, "main.go")
	fullCode := "package main\n\n" + goCode + "\n"
	if err := os.WriteFile(goFile, []byte(fullCode), 0o644); err != nil {
		t.Fatalf("failed to write generated go file: %v", err)
	}

	exeFile := filepath.Join(workDir, "app.exe")
	buildCmd := exec.Command("go", "build", "-o", exeFile, goFile)
	buildOut, buildErr := buildCmd.CombinedOutput()
	if buildErr != nil {
		t.Fatalf("expected generated Go code to build, error: %v\nbuild output:\n%s\ngo code:\n%s", buildErr, string(buildOut), fullCode)
	}

	runCmd := exec.Command(exeFile)
	runOut, runErr := runCmd.CombinedOutput()
	if runErr != nil {
		t.Fatalf("expected generated executable to run, error: %v\nrun output:\n%s\ngo code:\n%s", runErr, string(runOut), fullCode)
	}

	return string(runOut)
}

func TestGenericConstraintE2E_WhereToGoBuildAndRun(t *testing.T) {
	cases := []struct {
		name           string
		mygoCode       string
		mustContainGo  []string
		mustContainRun string
	}{
		{
			name: "fn where 约束",
			mygoCode: `
where T: int
fn id<T>(x: T): T {
    return x;
}

fn main() {
    print(7);
}
`,
			mustContainGo:  []string{"func id[T int](x T) T"},
			mustContainRun: "7",
		},
		{
			name: "struct where 约束",
			mygoCode: `
where T: int
struct Box<T> {
    item: T
}

fn main() {
    let bx = Box<int>{ item: 3 };
    print(bx.item);
}
`,
			mustContainGo:  []string{"type box[T int] struct"},
			mustContainRun: "3",
		},
		{
			name: "trait where 约束",
			mygoCode: `
where T: int
trait Iterator<T> {
    fn next(): T;
}

fn main() {
    print(11);
}
`,
			mustContainGo:  []string{"type iterator[T int] interface"},
			mustContainRun: "11",
		},
		{
			name: "trait bind where 约束",
			mygoCode: `
enum Opt { None }

where T: int
trait bind<T>(o: Opt) {
    fn make(v: T): T {
        return v;
    }
}

fn main() {
    let o = opt_None{};
    let v = make<int>(o, 9);
    print(v);
}
`,
			mustContainGo:  []string{"func make[T int](o opt, v T) T"},
			mustContainRun: "9",
		},
		{
			name: "enum where 约束",
			mygoCode: `
where T: int
enum Option<T> { Some(T), None }

fn main() {
    print(13);
}
`,
			mustContainGo:  []string{"type option[T int] interface"},
			mustContainRun: "13",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			semOut := semanticOutput(tc.mygoCode)
			if strings.Contains(semOut, "E_WHERE_") {
				t.Fatalf("expected no where semantic error, got:\n%s", semOut)
			}

			goCode := transpileSource(tc.mygoCode)
			for _, fragment := range tc.mustContainGo {
				if !strings.Contains(goCode, fragment) {
					t.Fatalf("expected generated go code to contain %q, got:\n%s", fragment, goCode)
				}
			}

			runOut := buildAndRunTranspiledGo(t, goCode)
			if !strings.Contains(runOut, tc.mustContainRun) {
				t.Fatalf("expected run output to contain %q, got:\n%s", tc.mustContainRun, runOut)
			}
		})
	}
}
