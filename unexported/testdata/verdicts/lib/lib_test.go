package lib

import "testing"

func TestTestOnly(t *testing.T) {
	if TestOnly() == "" {
		t.Fatal("expected a value")
	}
}

func TestChain(t *testing.T) {
	if TestChainHead() == "" {
		t.Fatal("expected a value")
	}
}
