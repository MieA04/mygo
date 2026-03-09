package types

import "testing"

func TestCanImplicitPromote(t *testing.T) {
	cases := []struct {
		from string
		to   string
		want bool
	}{
		{from: "byte", to: "int", want: true},
		{from: "int", to: "float64", want: true},
		{from: "float64", to: "int", want: false},
		{from: "string", to: "int", want: false},
	}
	for _, tc := range cases {
		got := CanImplicitPromote(tc.from, tc.to)
		if got != tc.want {
			t.Fatalf("CanImplicitPromote(%q, %q) = %v, want %v", tc.from, tc.to, got, tc.want)
		}
	}
}

func TestCommonNumericType(t *testing.T) {
	tp, ok := CommonNumericType("int", "float64")
	if !ok {
		t.Fatalf("CommonNumericType should succeed for int and float64")
	}
	if tp != "float64" {
		t.Fatalf("CommonNumericType(int, float64) = %q, want %q", tp, "float64")
	}

	_, ok = CommonNumericType("string", "int")
	if ok {
		t.Fatalf("CommonNumericType should fail for string and int")
	}
}
