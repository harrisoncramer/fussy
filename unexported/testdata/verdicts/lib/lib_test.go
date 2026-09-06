package lib

import "testing"

func TestTestOnly(t *testing.T) {
	if TestOnly() == "" {
		t.Fatal("expected a value")
	}
}
