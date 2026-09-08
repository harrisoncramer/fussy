// Package analyzers assembles the set of analyzers a configuration asks for, which is what the
// golangci-lint plugin and the standalone binary both start from so the two cannot drift.
package analyzers

import (
	"go/ast"

	commentlength "github.com/harrisoncramer/fussy/analyzers/comment_length"
	commentprefix "github.com/harrisoncramer/fussy/analyzers/comment_prefix"
	commentreference "github.com/harrisoncramer/fussy/analyzers/comment_reference"
	contexttimeout "github.com/harrisoncramer/fussy/analyzers/context_timeout"
	forbidgetenv "github.com/harrisoncramer/fussy/analyzers/forbidgetenv"
	forbidnilnil "github.com/harrisoncramer/fussy/analyzers/forbidnilnil"
	paramsstruct "github.com/harrisoncramer/fussy/analyzers/params_struct"
	storeverb "github.com/harrisoncramer/fussy/analyzers/store_verb"
	switchdefault "github.com/harrisoncramer/fussy/analyzers/switch_default"
	tabletest "github.com/harrisoncramer/fussy/analyzers/table_test"
	testdouble "github.com/harrisoncramer/fussy/analyzers/test_double"
	"github.com/harrisoncramer/fussy/config"

	"golang.org/x/tools/go/analysis"
)

// BuildAll returns every analyzer the configuration leaves switched on.
func BuildAll(cfg config.Config) []*analysis.Analyzer {
	var built []*analysis.Analyzer

	if !cfg.CommentLength.Skip {
		built = append(built, commentlength.NewAnalyzer(cfg.CommentLength))
	}

	if !cfg.CommentPrefix.Skip {
		built = append(built, commentprefix.NewAnalyzer(cfg.CommentPrefix))
	}

	if !cfg.CommentReference.Skip {
		built = append(built, commentreference.NewAnalyzer(cfg.CommentReference))
	}

	if !cfg.ContextTimeout.Skip {
		built = append(built, contexttimeout.NewAnalyzer(cfg.ContextTimeout))
	}

	if !cfg.ForbidGetenv.Skip {
		built = append(built, forbidgetenv.NewAnalyzer(cfg.ForbidGetenv))
	}

	if !cfg.ForbidNilNil.Skip {
		built = append(built, forbidnilnil.NewAnalyzer(cfg.ForbidNilNil))
	}

	if !cfg.ParamsStruct.Skip {
		built = append(built, paramsstruct.NewAnalyzer(cfg.ParamsStruct))
	}

	if !cfg.StoreVerb.Skip {
		built = append(built, storeverb.NewAnalyzer(cfg.StoreVerb))
	}

	if !cfg.SwitchDefault.Skip {
		built = append(built, switchdefault.NewAnalyzer(cfg.SwitchDefault))
	}

	if !cfg.TableTest.Skip {
		built = append(built, tabletest.NewAnalyzer(cfg.TableTest))
	}

	if !cfg.TestDouble.Skip {
		built = append(built, testdouble.NewAnalyzer(cfg.TestDouble))
	}

	for _, a := range built {
		skipGenerated(a)
	}

	return built
}

func skipGenerated(a *analysis.Analyzer) {
	run := a.Run
	a.Run = func(pass *analysis.Pass) (any, error) {
		kept := pass.Files[:0:0]
		for _, file := range pass.Files {
			if !ast.IsGenerated(file) {
				kept = append(kept, file)
			}
		}
		pass.Files = kept

		return run(pass)
	}
}
