package valid

import "testing"

type harnessParams struct {
	Name string
}

// A test helper's params struct is named after the fixture it builds rather than one of the
// tests using it.
func newHarness(params harnessParams) string {
	return params.Name
}

func TestHarness(t *testing.T) {
	if newHarness(harnessParams{Name: "one"}) != "one" {
		t.Fatal("mismatch")
	}
}
