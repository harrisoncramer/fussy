// Package paramsstruct holds a params struct to the name of the function that takes it, since a
// repository that also names them after the type built ends up with two rules and no way to
// tell which one a given struct is following.
package paramsstruct

import (
	"fmt"
	"go/ast"
	"go/types"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/harrisoncramer/fussy/config"

	"golang.org/x/tools/go/analysis"
)

// suffix is the ending that marks a struct as the parameters of one function.
const suffix = "Params"

// parameterName is the name every params struct is taken under, so a call site reads the same
// way wherever it is.
const parameterName = "p"

const (
	misnamedType = "%s takes %s, so name the struct %s%s after the function that takes it"
	misnamedArg  = "%s is taken as %q, so call it %q since every params struct is taken the same way"
)

// NewAnalyzer builds the params-struct analyzer, which reports a params struct named after
// anything but the function taking it, and a params parameter called anything but p.
func NewAnalyzer(cfg config.ParamsStructConfig) *analysis.Analyzer {
	excluded, compileErr := compileExcludes(cfg.Exclude)

	return &analysis.Analyzer{
		Name: "paramsstruct",
		Doc:  "Checks that a params struct is named after the function taking it and is taken as p",
		Run: func(pass *analysis.Pass) (any, error) {
			if compileErr != nil {
				return nil, compileErr
			}

			for _, file := range pass.Files {
				if isExcluded(excluded, pass.Fset.File(file.Pos()).Name()) {
					continue
				}

				for _, decl := range file.Decls {
					if fn, ok := decl.(*ast.FuncDecl); ok {
						check(pass, fn)
					}
				}
			}

			return nil, nil
		},
	}
}

func check(pass *analysis.Pass, fn *ast.FuncDecl) {
	if fn.Type.Params == nil {
		return
	}

	for _, field := range fn.Type.Params.List {
		object := paramsTypeOf(pass, field.Type)
		if object == nil {
			continue
		}

		if object.Pkg() == pass.Pkg && object.Name() != fn.Name.Name+suffix {
			pass.Report(analysis.Diagnostic{
				Pos:     field.Type.Pos(),
				End:     field.Type.End(),
				Message: fmt.Sprintf(misnamedType, fn.Name.Name, object.Name(), fn.Name.Name, suffix),
			})
		}

		for _, name := range field.Names {
			if name.Name != parameterName && name.Name != "_" {
				pass.Report(analysis.Diagnostic{
					Pos:     name.Pos(),
					End:     name.End(),
					Message: fmt.Sprintf(misnamedArg, object.Name(), name.Name, parameterName),
				})
			}
		}
	}
}

// paramsTypeOf returns the named struct behind a parameter when that struct is a params struct,
// looking through a pointer since a large one is sometimes taken by reference.
func paramsTypeOf(pass *analysis.Pass, expr ast.Expr) *types.TypeName {
	typ := pass.TypesInfo.TypeOf(expr)
	if typ == nil {
		return nil
	}

	if pointer, ok := typ.(*types.Pointer); ok {
		typ = pointer.Elem()
	}

	named, ok := typ.(*types.Named)
	if !ok || named.Obj().Pkg() == nil {
		return nil
	}

	if _, ok := named.Underlying().(*types.Struct); !ok {
		return nil
	}

	if !strings.HasSuffix(named.Obj().Name(), suffix) || named.Obj().Name() == suffix {
		return nil
	}

	return named.Obj()
}

func compileExcludes(patterns []string) ([]*regexp.Regexp, error) {
	var compiled []*regexp.Regexp
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("params_struct: exclude pattern %q: %w", pattern, err)
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
