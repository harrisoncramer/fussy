package storeverb_test

import (
	"testing"

	storeverb "github.com/harrisoncramer/fussy/analyzers/store_verb"
	"github.com/harrisoncramer/fussy/config"

	"golang.org/x/tools/go/analysis/analysistest"
)

var testdata = analysistest.TestData()

// TestStoreMethodsOpenWithAVerb pins that an adjective-first, noun-first or question-shaped
// method is found while a verb-first one, an unexported one and a plain function are not.
func TestStoreMethodsOpenWithAVerb(t *testing.T) {
	analyzer := storeverb.NewAnalyzer(config.StoreVerbConfig{Include: []string{`src/store/`}})
	analysistest.Run(t, testdata, analyzer, "store")
}

// TestFilesOutsideTheIncludeAreLeftAlone pins that the rule reaches only the paths it names.
func TestFilesOutsideTheIncludeAreLeftAlone(t *testing.T) {
	analyzer := storeverb.NewAnalyzer(config.StoreVerbConfig{Include: []string{`src/store/`}})
	analysistest.Run(t, testdata, analyzer, "elsewhere")
}

// TestEmptyIncludeIsOff pins that a rule with no path to reach checks nothing.
func TestEmptyIncludeIsOff(t *testing.T) {
	analyzer := storeverb.NewAnalyzer(config.StoreVerbConfig{})
	analysistest.Run(t, testdata, analyzer, "elsewhere")
}

// TestConfiguredVerbs pins that a repository's own list replaces the built-in one.
func TestConfiguredVerbs(t *testing.T) {
	analyzer := storeverb.NewAnalyzer(config.StoreVerbConfig{
		Include: []string{`src/verbs/`},
		Verbs:   []string{"Fetch"},
	})
	analysistest.Run(t, testdata, analyzer, "verbs")
}

// TestBadIncludePatternIsAnError pins that a pattern that will not compile fails the analyzer
// rather than silently matching nothing.
func TestBadIncludePatternIsAnError(t *testing.T) {
	analyzer := storeverb.NewAnalyzer(config.StoreVerbConfig{Include: []string{"("}})
	if _, err := analyzer.Run(nil); err == nil {
		t.Fatal("expected a bad include pattern to fail the analyzer")
	}
}

// TestConfiguredAllowList pins that a repository's own allow list replaces the built-in one, so
// a method the stdlib named is only exempt while it is named.
func TestConfiguredAllowList(t *testing.T) {
	analyzer := storeverb.NewAnalyzer(config.StoreVerbConfig{
		Include: []string{`src/allowed/`},
		Allow:   []string{"Rows"},
	})
	analysistest.Run(t, testdata, analyzer, "allowed")
}
