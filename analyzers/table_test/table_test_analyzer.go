// Package tabletest holds a table-driven test to one shape, since four hundred tests written the
// same way are read at a glance and the one written differently is read twice.
package tabletest

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

const (
	sliceName   = "tests"
	bindingName = "tt"
	nameField   = "name"
)

const (
	misnamedSlice   = "a table of cases is declared as %q rather than %q"
	misnamedBinding = "the table binds each case as %q rather than %q"
	misnamedSubtest = "the subtest is named from something other than %s.%s, which is the field the table already carries"
)

// NewAnalyzer builds the table-test analyzer, which reports a table declared, bound or run under
// a name other than the one every other table in the tree uses.
func NewAnalyzer(cfg config.TableTestConfig) *analysis.Analyzer {
	excluded, compileErr := compileExcludes(cfg.Exclude)

	return &analysis.Analyzer{
		Name: "tabletest",
		Doc:  "Checks that a table-driven test declares tests, binds tt, and runs tt.name",
		Run: func(pass *analysis.Pass) (any, error) {
			if compileErr != nil {
				return nil, compileErr
			}

			for _, file := range pass.Files {
				path := pass.Fset.File(file.Pos()).Name()
				if !strings.HasSuffix(path, "_test.go") || isExcluded(excluded, path) {
					continue
				}

				check(pass, file)
			}

			return nil, nil
		},
	}
}

func check(pass *analysis.Pass, file *ast.File) {
	tables := map[types.Object]struct{}{}

	ast.Inspect(file, func(node ast.Node) bool {
		for _, ident := range declaredTable(pass, node) {
			object := pass.TypesInfo.Defs[ident]
			if object == nil {
				continue
			}

			tables[object] = struct{}{}
			if ident.Name != sliceName {
				pass.Report(analysis.Diagnostic{
					Pos:     ident.Pos(),
					End:     ident.End(),
					Message: fmt.Sprintf(misnamedSlice, ident.Name, sliceName),
				})
			}
		}

		return true
	})

	ast.Inspect(file, func(node ast.Node) bool {
		stmt, ok := node.(*ast.RangeStmt)
		if !ok || !rangesTable(pass, tables, stmt.X) {
			return true
		}

		binding, ok := stmt.Value.(*ast.Ident)
		if !ok {
			return true
		}

		if binding.Name != bindingName {
			pass.Report(analysis.Diagnostic{
				Pos:     binding.Pos(),
				End:     binding.End(),
				Message: fmt.Sprintf(misnamedBinding, binding.Name, bindingName),
			})
		}

		checkSubtestName(pass, stmt.Body, binding.Name)

		return true
	})
}

// checkSubtestName reports the subtest of a case named from something other than the case's own
// name field, and leaves a subtest nested inside that one to name itself however it likes.
func checkSubtestName(pass *analysis.Pass, body *ast.BlockStmt, binding string) {
	ast.Inspect(body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || !isSubtestCall(pass, call) || len(call.Args) == 0 {
			return true
		}

		if !isNameField(call.Args[0], binding) {
			pass.Report(analysis.Diagnostic{
				Pos:     call.Args[0].Pos(),
				End:     call.Args[0].End(),
				Message: fmt.Sprintf(misnamedSubtest, bindingName, nameField),
			})
		}

		return false
	})
}

// isSubtestCall holds the check to Run on a testing.T, so a table body calling Run on a server
// or a harness of its own is nothing to do with the subtest's name.
func isSubtestCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Run" {
		return false
	}

	typ := pass.TypesInfo.TypeOf(selector.X)

	return typ != nil && typ.String() == "*testing.T"
}

func isNameField(expr ast.Expr, binding string) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != nameField {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)

	return ok && ident.Name == binding
}

// declaredTable returns the identifiers a statement binds to a table of cases.
func declaredTable(pass *analysis.Pass, node ast.Node) []*ast.Ident {
	var found []*ast.Ident

	switch decl := node.(type) {
	case *ast.AssignStmt:
		for i, rhs := range decl.Rhs {
			if i >= len(decl.Lhs) || !isTableOfCases(pass.TypesInfo.TypeOf(rhs)) {
				continue
			}
			if ident, ok := decl.Lhs[i].(*ast.Ident); ok {
				found = append(found, ident)
			}
		}
	case *ast.ValueSpec:
		for i, name := range decl.Names {
			if i < len(decl.Values) && isTableOfCases(pass.TypesInfo.TypeOf(decl.Values[i])) {
				found = append(found, name)
			}
		}
	}

	return found
}

// isTableOfCases reports whether a type is a slice of structs holding a name field, the shape a
// table of cases has and a cross-product sweep over an enum does not.
func isTableOfCases(typ types.Type) bool {
	if typ == nil {
		return false
	}

	slice, ok := typ.Underlying().(*types.Slice)
	if !ok {
		return false
	}

	structure, ok := slice.Elem().Underlying().(*types.Struct)
	if !ok {
		return false
	}

	for i := range structure.NumFields() {
		field := structure.Field(i)
		if field.Name() != nameField {
			continue
		}

		basic, ok := field.Type().Underlying().(*types.Basic)

		return ok && basic.Kind() == types.String
	}

	return false
}

func rangesTable(pass *analysis.Pass, tables map[types.Object]struct{}, expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}

	_, found := tables[pass.TypesInfo.Uses[ident]]

	return found
}

func compileExcludes(patterns []string) ([]*regexp.Regexp, error) {
	var compiled []*regexp.Regexp
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("table_test: exclude pattern %q: %w", pattern, err)
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
