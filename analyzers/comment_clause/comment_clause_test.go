package commentclause_test

import (
	"testing"

	commentclause "github.com/harrisoncramer/fussy/analyzers/comment_clause"
	"github.com/harrisoncramer/fussy/config"

	"golang.org/x/tools/go/analysis/analysistest"
)

var testdata = analysistest.TestData()

// TestInvalidCommentClause pins that a justification tacked on after a comma is found.
func TestInvalidCommentClause(t *testing.T) {
	analyzer := commentclause.NewAnalyzer(config.CommentClauseConfig{})
	analysistest.Run(t, testdata, analyzer, "invalid")
}

// TestValidCommentClause pins that a plain sentence, a clause with no comma, and an explainer are
// left alone.
func TestValidCommentClause(t *testing.T) {
	analyzer := commentclause.NewAnalyzer(config.CommentClauseConfig{})
	analysistest.Run(t, testdata, analyzer, "valid")
}

// TestExcludedFileIsNotReported pins that a file matching an exclude pattern is skipped whole.
func TestExcludedFileIsNotReported(t *testing.T) {
	analyzer := commentclause.NewAnalyzer(config.CommentClauseConfig{Exclude: []string{`excluded/excluded\.go$`}})
	analysistest.Run(t, testdata, analyzer, "excluded")
}

// TestBadExcludePatternIsAnError pins that an exclude pattern that will not compile fails the
// analyzer rather than silently matching nothing.
func TestBadExcludePatternIsAnError(t *testing.T) {
	analyzer := commentclause.NewAnalyzer(config.CommentClauseConfig{Exclude: []string{"("}})
	if _, err := analyzer.Run(nil); err == nil {
		t.Fatal("expected a bad exclude pattern to fail the analyzer")
	}
}
