package testdouble_test

import (
	"testing"

	testdouble "github.com/harrisoncramer/fussy/analyzers/test_double"
	"github.com/harrisoncramer/fussy/config"

	"golang.org/x/tools/go/analysis/analysistest"
)

var testdata = analysistest.TestData()

// TestForbiddenPrefixes pins that a stub, a mock and a spy are all found.
func TestForbiddenPrefixes(t *testing.T) {
	analyzer := testdouble.NewAnalyzer(config.TestDoubleConfig{})
	analysistest.Run(t, testdata, analyzer, "invalid")
}

// TestValidDoubles pins that the settled word, a longer word opening with a forbidden one, and a
// type outside a test file are all left alone.
func TestValidDoubles(t *testing.T) {
	analyzer := testdouble.NewAnalyzer(config.TestDoubleConfig{})
	analysistest.Run(t, testdata, analyzer, "valid")
}

// TestConfiguredWords pins that a repository having picked mock can forbid fake instead.
func TestConfiguredWords(t *testing.T) {
	analyzer := testdouble.NewAnalyzer(config.TestDoubleConfig{
		Forbidden: []string{"fake", "stub", "spy"},
		Preferred: "mock",
	})
	analysistest.Run(t, testdata, analyzer, "configured")
}

// TestExcludedFilesAreLeftAlone pins that the rule can be pointed away from a path.
func TestExcludedFilesAreLeftAlone(t *testing.T) {
	analyzer := testdouble.NewAnalyzer(config.TestDoubleConfig{Exclude: []string{`src/excluded/`}})
	analysistest.Run(t, testdata, analyzer, "excluded")
}

// TestBadExcludePatternIsAnError pins that a pattern that will not compile fails the analyzer
// rather than silently matching nothing.
func TestBadExcludePatternIsAnError(t *testing.T) {
	analyzer := testdouble.NewAnalyzer(config.TestDoubleConfig{Exclude: []string{"("}})
	if _, err := analyzer.Run(nil); err == nil {
		t.Fatal("expected a bad exclude pattern to fail the analyzer")
	}
}
