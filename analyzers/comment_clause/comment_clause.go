// Package commentclause reports the trailing justification clause a one-sentence rule invites.
//
// A package comment is exempt. It is the one comment in a file that exists to explain why, it is
// published on pkg.go.dev, and holding it to a bare statement would make it useless.
//
// Holding a comment to one sentence does not stop an author saying two things. It
// pushes them to glue the second onto the first with a comma and a "since", "because", "so that"
// or "rather than", which reads worse than the two sentences it replaced. The rule the clause is
// breaking is that a doc comment says what the thing is.
//
// Most of these clauses are worth nothing and should be deleted. The reasoning that is worth
// keeping belongs where a reader will find it without being sent looking: the package comment,
// which takes as many paragraphs as it needs, or one "Explainer:" paragraph beside the code it
// defends. A commit message is not that place, since a reader of the file has no way to know
// which of a thousand commits to go and read.
package commentclause

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/harrisoncramer/fussy/config"

	"golang.org/x/tools/go/analysis"
)

const explainerPrefix = "Explainer:"

// defaultClauses are the joins an author reaches for when a one-sentence rule leaves them a
// second thing to say.
var defaultClauses = []string{"since", "because", "so that", "rather than", "which is", "so as to"}

// NewAnalyzer builds the comment-clause analyzer, which reports a comment whose sentence carries
// a trailing justification.
func NewAnalyzer(cfg config.CommentClauseConfig) *analysis.Analyzer {
	excluded, compileErr := compileExcludes(cfg.Exclude)
	clause := compileClauses(cfg.Clauses)

	return &analysis.Analyzer{
		Name: "commentclause",
		Doc:  "Checks that comments do not tack a justification onto the end of a sentence",
		Run: func(pass *analysis.Pass) (any, error) {
			if compileErr != nil {
				return nil, compileErr
			}
			for _, file := range pass.Files {
				if ast.IsGenerated(file) {
					continue
				}

				if isExcluded(excluded, pass.Fset.File(file.Pos()).Name()) {
					continue
				}

				for _, group := range file.Comments {
					if group == file.Doc {
						continue
					}
					if message, ok := complaint(group.Text(), clause); ok {
						pass.Report(analysis.Diagnostic{Pos: group.Pos(), Message: message})
					}
				}
			}

			return nil, nil
		},
	}
}

func complaint(text string, clause *regexp.Regexp) (string, bool) {
	if strings.HasPrefix(strings.TrimSpace(text), explainerPrefix) {
		return "", false
	}

	found := clause.FindStringSubmatch(text)
	if found == nil {
		return "", false
	}

	return fmt.Sprintf(`comment tacks a %q clause onto its sentence: cut everything from the comma, or move the reasoning to the package comment or an "Explainer:" paragraph if it is worth keeping`, found[1]), true
}

// compileClauses builds the pattern for the joins configured, falling back to the ones every
// author reaches for.
func compileClauses(clauses []string) *regexp.Regexp {
	if len(clauses) == 0 {
		clauses = defaultClauses
	}

	quoted := make([]string, 0, len(clauses))
	for _, current := range clauses {
		quoted = append(quoted, regexp.QuoteMeta(current))
	}

	return regexp.MustCompile(`,\s+(` + strings.Join(quoted, "|") + `)\s`)
}

func compileExcludes(patterns []string) ([]*regexp.Regexp, error) {
	var compiled []*regexp.Regexp
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("comment_clause: exclude pattern %q: %w", pattern, err)
		}
		compiled = append(compiled, re)
	}

	return compiled, nil
}

func isExcluded(patterns []*regexp.Regexp, name string) bool {
	path := filepath.ToSlash(name)
	for _, pattern := range patterns {
		if pattern.MatchString(path) {
			return true
		}
	}

	return false
}
