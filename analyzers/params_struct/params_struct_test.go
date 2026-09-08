package paramsstruct_test

import (
	"testing"

	paramsstruct "github.com/harrisoncramer/fussy/analyzers/params_struct"
	"github.com/harrisoncramer/fussy/config"

	"golang.org/x/tools/go/analysis/analysistest"
)

var testdata = analysistest.TestData()

// TestInvalidParamsStructs pins that a struct named after the type built, and a parameter called
// anything but params, are both found.
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

// TestTheParameterNameIsConfigurable pins that a repository can hold the parameter to a name of
// its own rather than to params.
func TestTheParameterNameIsConfigurable(t *testing.T) {
	analyzer := paramsstruct.NewAnalyzer(config.ParamsStructConfig{ParameterName: "p"})
	analysistest.Run(t, testdata, analyzer, "renamed")
}

// TestAllowedStructsKeepTheirOwnName pins that the allow list frees a struct from the function
// name without freeing it from the parameter name.
func TestAllowedStructsKeepTheirOwnName(t *testing.T) {
	analyzer := paramsstruct.NewAnalyzer(config.ParamsStructConfig{Allow: []string{"scriptEnvParams", "resumeParams"}})
	analysistest.Run(t, testdata, analyzer, "allowed")
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
