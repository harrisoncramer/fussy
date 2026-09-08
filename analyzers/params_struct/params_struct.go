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

			var taken []parameter
			for _, file := range pass.Files {
				path := pass.Fset.File(file.Pos()).Name()
				if strings.HasSuffix(path, "_test.go") || isExcluded(excluded, path) {
					continue
				}

				for _, decl := range file.Decls {
					if fn, ok := decl.(*ast.FuncDecl); ok {
						taken = append(taken, parametersOf(pass, fn)...)
					}
				}
			}

			report(pass, taken)

			return nil, nil
		},
	}
}

// parameter is one params struct a function takes, held until the whole package has been read.
type parameter struct {
	function *ast.FuncDecl
	object   *types.TypeName
	field    *ast.Field
}

func parametersOf(pass *analysis.Pass, fn *ast.FuncDecl) []parameter {
	if fn.Type.Params == nil {
		return nil
	}

	var taken []parameter
	for _, field := range fn.Type.Params.List {
		if object := paramsTypeOf(pass, field.Type); object != nil {
			taken = append(taken, parameter{function: fn, object: object, field: field})
		}
	}

	return taken
}

// report holds the type name only where one function takes the struct, leaving a wrapper and the
// function it forwards to sharing one.
func report(pass *analysis.Pass, taken []parameter) {
	functions := map[*types.TypeName]map[*ast.FuncDecl]bool{}
	for _, p := range taken {
		if functions[p.object] == nil {
			functions[p.object] = map[*ast.FuncDecl]bool{}
		}
		functions[p.object][p.function] = true
	}

	for _, p := range taken {
		name := p.function.Name.Name
		sole := len(functions[p.object]) == 1
		if sole && p.object.Pkg() == pass.Pkg && p.object.Name() != name+suffix {
			pass.Report(analysis.Diagnostic{
				Pos:     p.field.Type.Pos(),
				End:     p.field.Type.End(),
				Message: fmt.Sprintf(misnamedType, name, p.object.Name(), name, suffix),
			})
		}

		for _, ident := range p.field.Names {
			if ident.Name != parameterName && ident.Name != "_" {
				pass.Report(analysis.Diagnostic{
					Pos:     ident.Pos(),
					End:     ident.End(),
					Message: fmt.Sprintf(misnamedArg, p.object.Name(), ident.Name, parameterName),
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
