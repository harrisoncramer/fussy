package switchdefault_test

import (
	"testing"

	switchdefault "github.com/harrisoncramer/fussy/analyzers/switch_default"
	"github.com/harrisoncramer/fussy/config"

	"golang.org/x/tools/go/analysis/analysistest"
)

var testdata = analysistest.TestData()

// TestFailingDefaultsAreLeftAlone pins the shapes a default is allowed to take, which are an
// error, a panic, a comma-ok false, and the zero values of a signature offering neither.
func TestFailingDefaultsAreLeftAlone(t *testing.T) {
	analyzer := switchdefault.NewAnalyzer(config.SwitchDefaultConfig{})
	analysistest.Run(t, testdata, analyzer, "valid")
}

// TestAnsweringDefaultsAreReported pins that a default handing a value, a claim of success, or
// an assignment back to the code around it is found.
func TestAnsweringDefaultsAreReported(t *testing.T) {
	analyzer := switchdefault.NewAnalyzer(config.SwitchDefaultConfig{})
	analysistest.Run(t, testdata, analyzer, "invalid")
}

// TestAllowSilentDefaultKeepsTheValueChecks pins that letting a default say nothing is not a way
// to let it answer, since the returns are still measured.
func TestAllowSilentDefaultKeepsTheValueChecks(t *testing.T) {
	analyzer := switchdefault.NewAnalyzer(config.SwitchDefaultConfig{AllowSilentDefault: true})
	analysistest.Run(t, testdata, analyzer, "silent")
}

// TestExcludedPathsAreLeftAlone pins that the rule reaches past the paths a repository excludes
// from it, since a generated switch is not one an author can change.
func TestExcludedPathsAreLeftAlone(t *testing.T) {
	analyzer := switchdefault.NewAnalyzer(config.SwitchDefaultConfig{Exclude: []string{`src/excluded/`}})
	analysistest.Run(t, testdata, analyzer, "excluded")
}

// TestBadExcludePatternIsAnError pins that an exclude pattern that will not compile fails the
// analyzer rather than silently matching nothing.
func TestBadExcludePatternIsAnError(t *testing.T) {
	analyzer := switchdefault.NewAnalyzer(config.SwitchDefaultConfig{Exclude: []string{"("}})
	if _, err := analyzer.Run(nil); err == nil {
		t.Fatal("expected a bad exclude pattern to fail the analyzer")
	}
}
