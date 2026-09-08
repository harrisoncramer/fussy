// Package testdouble holds every test double to one word, since stub, mock and spy are used
// interchangeably for the same shape and a reader cannot tell the three apart.
package testdouble

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/harrisoncramer/fussy/config"

	"golang.org/x/tools/go/analysis"
)

// defaultForbidden is the set of words a double is named with when nothing has decided, and
// defaultPreferred is the one word they collapse into.
var defaultForbidden = []string{"stub", "mock", "spy"}

const defaultPreferred = "fake"

const message = "%s is a %s: name a test double %s%s, so one word covers all of them"

// NewAnalyzer builds the test-double analyzer, which reports a type in a test file named with a
// word the repository has not settled on.
func NewAnalyzer(cfg config.TestDoubleConfig) *analysis.Analyzer {
	excluded, compileErr := compileExcludes(cfg.Exclude)
	forbidden := cfg.Forbidden
	if len(forbidden) == 0 {
		forbidden = defaultForbidden
	}

	preferred := cfg.Preferred
	if preferred == "" {
		preferred = defaultPreferred
	}

	return &analysis.Analyzer{
		Name: "testdouble",
		Doc:  "Checks that a type in a test file is not named with a stub, mock or spy prefix",
		Run: func(pass *analysis.Pass) (any, error) {
			if compileErr != nil {
				return nil, compileErr
			}

			for _, file := range pass.Files {
				path := pass.Fset.File(file.Pos()).Name()
				if !strings.HasSuffix(path, "_test.go") || isExcluded(excluded, path) {
					continue
				}

				check(pass, file, forbidden, preferred)
			}

			return nil, nil
		},
	}
}

func check(pass *analysis.Pass, file *ast.File, forbidden []string, preferred string) {
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.TypeSpec)
		if !ok {
			return true
		}

		word, rest, found := prefixOf(spec.Name.Name, forbidden)
		if !found {
			return true
		}

		pass.Report(analysis.Diagnostic{
			Pos:     spec.Name.Pos(),
			End:     spec.Name.End(),
			Message: fmt.Sprintf(message, spec.Name.Name, word, matchCase(preferred, spec.Name.Name), rest),
		})

		return true
	})
}

// matchCase spells the preferred word the way the name it replaces opened, so the suggestion is
// a rename rather than a different identifier.
func matchCase(preferred, name string) string {
	if name == "" || !unicode.IsUpper(rune(name[0])) {
		return preferred
	}

	return string(unicode.ToUpper(rune(preferred[0]))) + preferred[1:]
}

// prefixOf returns the forbidden word a name opens with and what follows it, matching only a
// whole word so a type called mockingbird is left alone.
func prefixOf(name string, forbidden []string) (string, string, bool) {
	lowered := strings.ToLower(name)
	for _, word := range forbidden {
		if !strings.HasPrefix(lowered, strings.ToLower(word)) {
			continue
		}

		rest := name[len(word):]
		if rest == "" || unicode.IsUpper(rune(rest[0])) {
			return word, rest, true
		}
	}

	return "", "", false
}

func compileExcludes(patterns []string) ([]*regexp.Regexp, error) {
	var compiled []*regexp.Regexp
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("test_double: exclude pattern %q: %w", pattern, err)
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
