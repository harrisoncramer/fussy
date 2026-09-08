package excluded

import "testing"

func TestExcluded(t *testing.T) {
	cases := []struct {
		name string
	}{{name: "one"}}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {})
	}
}
