// Package storeverb asks a storage layer to open every exported method with one of the verbs it
// has settled on, since a reader scanning a hundred methods for the one that writes wants the
// answer in the first word.
package storeverb

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

// defaultVerbs is the list a repository that has not written one of its own is held to.
var defaultVerbs = []string{
	"Abandon", "Archive", "Begin", "Claim", "Clear", "Count", "Create", "Delete", "Exists",
	"Finish", "Forget", "Get", "Lock", "List", "Mark", "Move", "Open", "Prune", "Record",
	"Release", "Rename", "Restore", "Search", "Send", "Set", "Unmark", "Update", "Upsert",
}

const message = "%s does not open with a verb, so name it after what it does: %s"

// NewAnalyzer builds the store-verb analyzer, which reports an exported method whose name does
// not open with a configured verb, and which reaches nothing until it is pointed at a path.
func NewAnalyzer(cfg config.StoreVerbConfig) *analysis.Analyzer {
	included, compileErr := compileIncludes(cfg.Include)
	verbs := cfg.Verbs
	if len(verbs) == 0 {
		verbs = defaultVerbs
	}

	return &analysis.Analyzer{
		Name: "storeverb",
		Doc:  "Checks that an exported method in the storage layer opens with one of its verbs",
		Run: func(pass *analysis.Pass) (any, error) {
			if compileErr != nil {
				return nil, compileErr
			}

			if len(cfg.Include) == 0 {
				return nil, nil
			}

			for _, file := range pass.Files {
				path := pass.Fset.File(file.Pos()).Name()
				if strings.HasSuffix(path, "_test.go") || !isIncluded(included, path) {
					continue
				}

				check(pass, file, verbs)
			}

			return nil, nil
		},
	}
}

func check(pass *analysis.Pass, file *ast.File, verbs []string) {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || !fn.Name.IsExported() || opensWithVerb(fn.Name.Name, verbs) {
			continue
		}

		pass.Report(analysis.Diagnostic{
			Pos:     fn.Name.Pos(),
			End:     fn.Name.End(),
			Message: fmt.Sprintf(message, fn.Name.Name, strings.Join(verbs, ", ")),
		})
	}
}

// opensWithVerb reports whether a name begins with a whole verb, so ListNames passes on List
// while Listener does not.
func opensWithVerb(name string, verbs []string) bool {
	for _, verb := range verbs {
		if !strings.HasPrefix(name, verb) {
			continue
		}

		rest := name[len(verb):]
		if rest == "" || unicode.IsUpper(rune(rest[0])) || unicode.IsDigit(rune(rest[0])) {
			return true
		}
	}

	return false
}

func compileIncludes(patterns []string) ([]*regexp.Regexp, error) {
	var compiled []*regexp.Regexp
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("store_verb: include pattern %q: %w", pattern, err)
		}
		compiled = append(compiled, re)
	}

	return compiled, nil
}

func isIncluded(patterns []*regexp.Regexp, name string) bool {
	path := filepath.ToSlash(name)
	for _, pattern := range patterns {
		if pattern.MatchString(path) {
			return true
		}
	}

	return false
}
