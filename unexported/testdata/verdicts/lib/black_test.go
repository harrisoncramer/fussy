package lib_test

import (
	"testing"

	"fussytest/verdicts/lib"
)

func TestBlackBoxOnly(t *testing.T) {
	if lib.BlackBoxOnly() == "" {
		t.Fatal("expected a value")
	}
}
