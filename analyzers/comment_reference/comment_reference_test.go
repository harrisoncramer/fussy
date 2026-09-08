package commentreference_test

import (
	"testing"

	commentreference "github.com/harrisoncramer/fussy/analyzers/comment_reference"
	"github.com/harrisoncramer/fussy/config"

	"golang.org/x/tools/go/analysis/analysistest"
)

var testdata = analysistest.TestData()

// TestInvalidCommentReference pins that a comment naming a test or a file is found.
func TestInvalidCommentReference(t *testing.T) {
	analyzer := commentreference.NewAnalyzer(config.CommentReferenceConfig{})
	analysistest.Run(t, testdata, analyzer, "invalid")
}

// TestValidCommentReference pins that a comment pointing nowhere, and an explainer that does, are
// left alone.
func TestValidCommentReference(t *testing.T) {
	analyzer := commentreference.NewAnalyzer(config.CommentReferenceConfig{})
	analysistest.Run(t, testdata, analyzer, "valid")
}

// TestExcludedFileIsNotReported pins that a file matching an exclude pattern is skipped whole.
func TestExcludedFileIsNotReported(t *testing.T) {
	analyzer := commentreference.NewAnalyzer(config.CommentReferenceConfig{Exclude: []string{`excluded/excluded\.go$`}})
	analysistest.Run(t, testdata, analyzer, "excluded")
}

// TestBadExcludePatternIsAnError pins that an exclude pattern that will not compile fails the
// analyzer rather than silently matching nothing.
func TestBadExcludePatternIsAnError(t *testing.T) {
	analyzer := commentreference.NewAnalyzer(config.CommentReferenceConfig{Exclude: []string{"("}})
	if _, err := analyzer.Run(nil); err == nil {
		t.Fatal("expected a bad exclude pattern to fail the analyzer")
	}
}
