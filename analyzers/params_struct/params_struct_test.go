package paramsstruct_test

import (
	"testing"

	paramsstruct "github.com/harrisoncramer/fussy/analyzers/params_struct"
	"github.com/harrisoncramer/fussy/config"

	"golang.org/x/tools/go/analysis/analysistest"
)

var testdata = analysistest.TestData()

// TestInvalidParamsStructs pins that a struct named after the type built, and a parameter called
// anything but p, are both found.
func TestInvalidParamsStructs(t *testing.T) {
	analyzer := paramsstruct.NewAnalyzer(config.ParamsStructConfig{})
	analysistest.Run(t, testdata, analyzer, "invalid")
}

// TestValidParamsStructs pins that a struct named after its function, a pointer to one, a struct
// that is not a params struct, and one from another package are all left alone.
func TestValidParamsStructs(t *testing.T) {
	analyzer := paramsstruct.NewAnalyzer(config.ParamsStructConfig{})
	analysistest.Run(t, testdata, analyzer, "valid")
}

// TestExcludedFilesAreLeftAlone pins that the rule can be pointed away from a path.
func TestExcludedFilesAreLeftAlone(t *testing.T) {
	analyzer := paramsstruct.NewAnalyzer(config.ParamsStructConfig{Exclude: []string{`src/excluded/`}})
	analysistest.Run(t, testdata, analyzer, "excluded")
}

// TestBadExcludePatternIsAnError pins that a pattern that will not compile fails the analyzer
// rather than silently matching nothing.
func TestBadExcludePatternIsAnError(t *testing.T) {
	analyzer := paramsstruct.NewAnalyzer(config.ParamsStructConfig{Exclude: []string{"("}})
	if _, err := analyzer.Run(nil); err == nil {
		t.Fatal("expected a bad exclude pattern to fail the analyzer")
	}
}
