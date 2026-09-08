// Package commentclause reports the trailing justification clause a one-sentence rule invites.
//
// A package comment is exempt. It is the one comment in a file that exists to explain why, it is
// published on pkg.go.dev, and holding it to a bare statement would make it useless.
//
// Holding a comment to one sentence does not stop an author saying two things. It
// pushes them to glue the second onto the first with a comma and a "since", "because", "so that"
// or "rather than", which reads worse than the two sentences it replaced. The rule the clause is
// breaking is that a comment says what the thing is; why it is that way belongs in the commit
// message, where it does not have to be maintained alongside the code.
package commentclause

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/harrisoncramer/fussy/config"

	"golang.org/x/tools/go/analysis"
)

const explainerPrefix = "Explainer:"

var clause = regexp.MustCompile(`,\s+(since|because|so that|rather than|which is|so as to)\s`)

// NewAnalyzer builds the comment-clause analyzer, which reports a comment whose sentence carries
// a trailing justification.
func NewAnalyzer(cfg config.CommentClauseConfig) *analysis.Analyzer {
	excluded, compileErr := compileExcludes(cfg.Exclude)

	return &analysis.Analyzer{
		Name: "commentclause",
		Doc:  "Checks that comments do not tack a justification onto the end of a sentence",
		Run: func(pass *analysis.Pass) (any, error) {
			if compileErr != nil {
				return nil, compileErr
			}
			for _, file := range pass.Files {
				if isExcluded(excluded, pass.Fset.File(file.Pos()).Name()) {
					continue
				}
				for _, group := range file.Comments {
					if group == file.Doc {
						continue
					}
					if message, ok := complaint(group.Text()); ok {
						pass.Report(analysis.Diagnostic{Pos: group.Pos(), Message: message})
					}
				}
			}

			return nil, nil
		},
	}
}

func complaint(text string) (string, bool) {
	if strings.HasPrefix(strings.TrimSpace(text), explainerPrefix) {
		return "", false
	}

	found := clause.FindStringSubmatch(text)
	if found == nil {
		return "", false
	}

	return fmt.Sprintf(`comment tacks a %q clause onto its sentence: cut everything from the comma and put the reasoning in the commit message`, found[1]), true
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
