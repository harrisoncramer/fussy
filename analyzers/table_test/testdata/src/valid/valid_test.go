package valid

import (
	"fmt"
	"testing"
)

func TestTheShapeEveryTableHas(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want int
	}{
		{name: "one", in: 1, want: 1},
		{name: "two", in: 2, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.in != tt.want {
				t.Fatal("mismatch")
			}
		})
	}
}

type testCase struct {
	name string
	in   int
}

// A table of a named struct type is the same shape written down once.
func TestNamedCaseType(t *testing.T) {
	tests := []testCase{{name: "one", in: 1}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.in
		})
	}
}

type paneKind int

const (
	paneLeft paneKind = iota
	paneRight
)

// A cross product over an enum and a slice of sizes is a different shape, and its subtest name
// is computed because no row carries one.
func TestCrossProduct(t *testing.T) {
	sizes := []int{10, 20}

	for _, kind := range []paneKind{paneLeft, paneRight} {
		for _, size := range sizes {
			t.Run(fmt.Sprintf("%d/%d", kind, size), func(t *testing.T) {})
		}
	}
}

// A slice of structs with no name field carries nothing for the subtest to be named from.
func TestFixtureSlice(t *testing.T) {
	fixtures := []struct {
		in   int
		want int
	}{{in: 1, want: 1}}

	for _, fixture := range fixtures {
		if fixture.in != fixture.want {
			t.Fatal("mismatch")
		}
	}
}

// A map keyed by the case name is ranged as a pair rather than as a table.
func TestMapOfCases(t *testing.T) {
	byName := map[string]int{"one": 1}

	for name, in := range byName {
		t.Run(name, func(t *testing.T) {
			_ = in
		})
	}
}

type runner struct{}

func (r runner) Run(in int) int {
	return in
}

// A Run on something other than a testing.T is not the subtest being named.
func TestRunOnAnotherType(t *testing.T) {
	tests := []struct {
		name string
		in   int
	}{{name: "one", in: 1}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := runner{}
			if r.Run(tt.in) != tt.in {
				t.Fatal("mismatch")
			}
		})
	}
}

// A case that splits into named sub-cases names those itself.
func TestNestedSubtests(t *testing.T) {
	tests := []struct {
		name string
		in   int
	}{{name: "one", in: 1}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("read", func(t *testing.T) {
				_ = tt.in
			})
			t.Run("write", func(t *testing.T) {
				_ = tt.in
			})
		})
	}
}
