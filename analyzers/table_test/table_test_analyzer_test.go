package tabletest_test

import (
	"testing"

	tabletest "github.com/harrisoncramer/fussy/analyzers/table_test"
	"github.com/harrisoncramer/fussy/config"

	"golang.org/x/tools/go/analysis/analysistest"
)

var testdata = analysistest.TestData()

// TestInvalidTables pins that a table declared, bound or run under another name is found.
func TestInvalidTables(t *testing.T) {
	analyzer := tabletest.NewAnalyzer(config.TableTestConfig{})
	analysistest.Run(t, testdata, analyzer, "invalid")
}

// TestValidTables pins that a cross product over an enum, a map of cases, a slice with no name
// field, and a table outside a test file are all left alone.
func TestValidTables(t *testing.T) {
	analyzer := tabletest.NewAnalyzer(config.TableTestConfig{})
	analysistest.Run(t, testdata, analyzer, "valid")
}

// TestExcludedFilesAreLeftAlone pins that the rule can be pointed away from a path.
func TestExcludedFilesAreLeftAlone(t *testing.T) {
	analyzer := tabletest.NewAnalyzer(config.TableTestConfig{Exclude: []string{`src/excluded/`}})
	analysistest.Run(t, testdata, analyzer, "excluded")
}

// TestBadExcludePatternIsAnError pins that a pattern that will not compile fails the analyzer
// rather than silently matching nothing.
func TestBadExcludePatternIsAnError(t *testing.T) {
	analyzer := tabletest.NewAnalyzer(config.TableTestConfig{Exclude: []string{"("}})
	if _, err := analyzer.Run(nil); err == nil {
		t.Fatal("expected a bad exclude pattern to fail the analyzer")
	}
}
