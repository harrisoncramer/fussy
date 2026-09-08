// Package commentreference keeps a comment from pointing at somewhere else in the tree.
//
// A comment naming a test, a caller or another file costs twice. The name rots the
// moment either side is renamed, and nothing fails when it does. And a reader who has to open a
// second file to finish reading this line has been sent away by the very thing that was meant to
// save them the trip. What the code is goes here; where else it is used is what the compiler and
// a grep are for.
package commentreference

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

var (
	testName = regexp.MustCompile(`\bTest[A-Z]\w+`)
	fileName = regexp.MustCompile(`\b[\w/]+\.go\b`)
)

const (
	namesTest = "comment names the test %q: say what the code is and leave the test to say what it proves"
	namesFile = "comment names the file %q: say what the code is rather than where else to look"
)

// NewAnalyzer builds the comment-reference analyzer, which reports a comment pointing at a test
// or at another file by name.
func NewAnalyzer(cfg config.CommentReferenceConfig) *analysis.Analyzer {
	excluded, compileErr := compileExcludes(cfg.Exclude)

	return &analysis.Analyzer{
		Name: "commentreference",
		Doc:  "Checks that comments do not name a test or another file",
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

				documented := documentedNames(file)
				for _, group := range file.Comments {
					if group == file.Doc {
						continue
					}
					if message, ok := complaint(group.Text(), documented[group]); ok {
						pass.Report(analysis.Diagnostic{Pos: group.Pos(), Message: message})
					}
				}
			}

			return nil, nil
		},
	}
}

func complaint(text, own string) (string, bool) {
	if strings.HasPrefix(strings.TrimSpace(text), explainerPrefix) {
		return "", false
	}

	for _, found := range testName.FindAllString(text, -1) {
		if found != own {
			return fmt.Sprintf(namesTest, found), true
		}
	}

	for _, found := range fileName.FindAllStringIndex(text, -1) {
		name := text[found[0]:found[1]]
		if !strings.HasPrefix(text[found[1]:], ".dev") {
			return fmt.Sprintf(namesFile, name), true
		}
	}

	return "", false
}

// documentedNames maps a doc comment to the identifier it documents, which a comment is allowed
// to name however much it looks like a reference.
func documentedNames(file *ast.File) map[*ast.CommentGroup]string {
	names := map[*ast.CommentGroup]string{}

	ast.Inspect(file, func(node ast.Node) bool {
		switch decl := node.(type) {
		case *ast.FuncDecl:
			if decl.Doc != nil {
				names[decl.Doc] = decl.Name.Name
			}
		case *ast.TypeSpec:
			if decl.Doc != nil {
				names[decl.Doc] = decl.Name.Name
			}
		}

		return true
	})

	return names
}

func compileExcludes(patterns []string) ([]*regexp.Regexp, error) {
	var compiled []*regexp.Regexp
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("comment_reference: exclude pattern %q: %w", pattern, err)
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
