package main

import (
	"testing"

	"fussytest/verdicts/lib"
)

func TestCrossPackage(t *testing.T) {
	if lib.CrossTest() == "" {
		t.Fatal("expected a value")
	}
}
