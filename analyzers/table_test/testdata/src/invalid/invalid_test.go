package invalid

import "testing"

func TestSliceIsNamedCases(t *testing.T) {
	cases := []struct { // want `a table of cases is declared as "cases" rather than "tests"`
		name string
		in   int
	}{{name: "one", in: 1}}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			_ = tt.in
		})
	}
}

func TestBindingIsNamedTest(t *testing.T) {
	tests := []struct {
		name string
		in   int
	}{{name: "one", in: 1}}

	for _, test := range tests { // want `the table binds each case as "test" rather than "tt"`
		t.Run(test.name, func(t *testing.T) {
			_ = test.in
		})
	}
}

func TestSubtestNameIsBuilt(t *testing.T) {
	tests := []struct {
		name string
		in   int
	}{{name: "one", in: 1}}

	for _, tt := range tests {
		t.Run(tt.name+"/extra", func(t *testing.T) { // want `the subtest is named from something other than tt.name, which is the field the table already carries`
			_ = tt.in
		})
	}
}

var table = []struct { // want `a table of cases is declared as "table" rather than "tests"`
	name string
}{{name: "one"}}

func TestPackageLevelTable(t *testing.T) {
	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {})
	}
}
