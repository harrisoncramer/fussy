package undriven

import "testing"

type hookChoice struct {
	name  string
	label string
}

func availableHooks() []hookChoice {
	return []hookChoice{{name: "pre-launch"}}
}

// A slice of the same shape as a table, built and asserted against rather than ranged over, is
// the test's data and not its cases.
func TestHooksAreListed(t *testing.T) {
	available := availableHooks()
	if len(available) != 1 {
		t.Fatalf("got %d hooks", len(available))
	}

	out := make([]hookChoice, 0, 1)
	out = append(out, hookChoice{name: "post-launch"})
	if out[0].name != "post-launch" {
		t.Fatal("wrong hook")
	}

	got := availableHooks()
	if got[0].name != "pre-launch" {
		t.Fatal("wrong first hook")
	}
}

// A table nothing ranges over is left alone until something drives it.
func TestDeclaredButHandedToAHelper(t *testing.T) {
	cases := []hookChoice{{name: "one"}}
	runAll(t, cases)
}

func runAll(t *testing.T, cases []hookChoice) {
	t.Helper()
	if len(cases) == 0 {
		t.Fatal("no cases")
	}
}
