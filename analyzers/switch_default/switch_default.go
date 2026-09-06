// Package switchdefault holds the default clause of a switch over an enum, or over a type, to
// failing rather than answering. A value added to the enum later takes that branch, and a branch
// that computes an answer turns that into a wrong result the caller believes, where a branch that
// returns an error or panics turns it into something a person sees.
package switchdefault

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/harrisoncramer/fussy/config"

	"golang.org/x/tools/go/analysis"
)

const (
	messageSilent = "default must return an error or panic, since a value added to the enum later takes this branch and nothing here says so"
	messageValue  = "default must return zero values beside its error, since a value returned here is a wrong answer for a value added to the enum later"
	messageNilErr = "default must return a non-nil error, since a nil error tells the caller a value added to the enum later was handled"
	messageOkFlag = "default must return false in its last result, since true tells the caller a value added to the enum later was handled"
	messageOpaque = "default must return a failure this rule can see, either a non-nil error or a call whose last result is an error"
	messageAssign = "default must not assign to state read after the switch; return an error or panic instead"
	messageEscape = "default must not leave the switch with break, continue, goto or fallthrough; return an error or panic instead"
)

// NewAnalyzer builds the switch-default analyzer, which holds the default clause of an enum or
// type switch to returning an error, panicking, or returning nothing but zero values.
func NewAnalyzer(cfg config.SwitchDefaultConfig) *analysis.Analyzer {
	excluded, compileErr := compileExcludes(cfg.Exclude)

	return &analysis.Analyzer{
		Name: "switchdefault",
		Doc:  "Checks that the default clause of an enum or type switch fails rather than answering",
		Run: func(pass *analysis.Pass) (any, error) {
			if compileErr != nil {
				return nil, compileErr
			}

			for _, file := range pass.Files {
				name := pass.Fset.Position(file.Pos()).Filename
				if strings.HasSuffix(name, "_test.go") || isExcluded(excluded, name) {
					continue
				}
				checkFile(pass, file, cfg.AllowSilentDefault)
			}

			return nil, nil
		},
	}
}

func checkFile(pass *analysis.Pass, file *ast.File, allowSilent bool) {
	ast.Inspect(file, func(node ast.Node) bool {
		switch fn := node.(type) {
		case *ast.FuncDecl:
			checkBody(pass, signatureOf(pass, fn.Name), fn.Body, allowSilent)
		case *ast.FuncLit:
			checkBody(pass, signatureOf(pass, fn), fn.Body, allowSilent)
		}

		return true
	})
}

// checkBody walks one function body, leaving the switches inside a nested literal to the walk
// that literal gets of its own, since a closure has a signature the returns answer to instead.
func checkBody(pass *analysis.Pass, signature *types.Signature, body *ast.BlockStmt, allowSilent bool) {
	if body == nil || signature == nil {
		return
	}

	ast.Inspect(body, func(node ast.Node) bool {
		switch stmt := node.(type) {
		case *ast.FuncLit:
			return false
		case *ast.SwitchStmt:
			if isEnumSwitch(pass, stmt.Tag) {
				checkClause(pass, signature, defaultClause(stmt.Body), allowSilent)
			}
		case *ast.TypeSwitchStmt:
			checkClause(pass, signature, defaultClause(stmt.Body), allowSilent)
		}

		return true
	})
}

func defaultClause(body *ast.BlockStmt) *ast.CaseClause {
	if body == nil {
		return nil
	}

	for _, stmt := range body.List {
		if clause, ok := stmt.(*ast.CaseClause); ok && clause.List == nil {
			return clause
		}
	}

	return nil
}

// checkClause reports every way the clause hands something other than a failure back to the code
// around it, which is the whole rule: what leaves the clause is what a caller can misread.
func checkClause(pass *analysis.Pass, signature *types.Signature, clause *ast.CaseClause, allowSilent bool) {
	if clause == nil {
		return
	}

	if len(clause.Body) == 0 {
		if !allowSilent {
			pass.Report(analysis.Diagnostic{Pos: clause.Pos(), End: clause.Colon + 1, Message: messageSilent})
		}

		return
	}

	found := scanClause(pass, clause, signature)

	for _, escape := range found.escapes {
		pass.Report(analysis.Diagnostic{Pos: escape.Pos(), End: escape.End(), Message: messageEscape})
	}

	for _, assign := range found.assigns {
		pass.Report(analysis.Diagnostic{Pos: assign.Pos(), End: assign.End(), Message: messageAssign})
	}

	for _, stmt := range found.returns {
		checkReturn(pass, signature, stmt, found.results)
	}

	if !allowSilent && found.quiet() && !terminates(clause.Body[len(clause.Body)-1]) {
		pass.Report(analysis.Diagnostic{Pos: clause.Pos(), End: clause.Colon + 1, Message: messageSilent})
	}
}

// checkReturn measures one return against the failure channel the signature offers, which is a
// trailing error, else a trailing boolean beside a value, else nothing but the zero values.
func checkReturn(pass *analysis.Pass, signature *types.Signature, stmt *ast.ReturnStmt, named map[*types.Var]ast.Expr) {
	results := signature.Results()
	values, ok := returnedValues(stmt, results, named)
	if !ok {
		checkOpaqueReturn(pass, results, stmt)
		return
	}

	failure := failureIndex(results)
	for i, value := range values {
		if value == nil || i == failure {
			continue
		}
		if !isZero(pass, value) {
			pass.Report(analysis.Diagnostic{Pos: value.Pos(), End: value.End(), Message: messageValue})
		}
	}

	if failure < 0 {
		return
	}

	value := values[failure]
	switch {
	case isError(results.At(failure).Type()):
		if value == nil || isNil(pass, value) {
			pass.Report(analysis.Diagnostic{Pos: stmt.Pos(), End: stmt.End(), Message: messageNilErr})
		}
	case value != nil && !isZero(pass, value):
		pass.Report(analysis.Diagnostic{Pos: value.Pos(), End: value.End(), Message: messageOkFlag})
	}
}

// checkOpaqueReturn covers a return handing every result to one call, which is worth trusting
// only where the last of those results is an error the caller has to look at.
func checkOpaqueReturn(pass *analysis.Pass, results *types.Tuple, stmt *ast.ReturnStmt) {
	if results.Len() > 0 && isError(results.At(results.Len()-1).Type()) {
		return
	}

	pass.Report(analysis.Diagnostic{Pos: stmt.Pos(), End: stmt.End(), Message: messageOpaque})
}

// returnedValues lines a return's expressions up with the results the signature declares, taking
// a bare return's values from the last assignment each named result was given in the clause.
func returnedValues(stmt *ast.ReturnStmt, results *types.Tuple, named map[*types.Var]ast.Expr) ([]ast.Expr, bool) {
	if len(stmt.Results) == results.Len() {
		return stmt.Results, true
	}

	if len(stmt.Results) != 0 {
		return nil, false
	}

	values := make([]ast.Expr, results.Len())
	for i := range results.Len() {
		values[i] = named[results.At(i)]
	}

	return values, true
}

// failureIndex names the result a caller reads to learn the switch did not know the value, and
// returns -1 for a signature offering none.
func failureIndex(results *types.Tuple) int {
	last := results.Len() - 1
	if last < 0 {
		return -1
	}

	if isError(results.At(last).Type()) {
		return last
	}

	if basic, ok := results.At(last).Type().Underlying().(*types.Basic); ok && basic.Kind() == types.Bool {
		return last
	}

	return -1
}

// clauseScan is what the clause hands to the code around it, gathered in one walk.
type clauseScan struct {
	returns []*ast.ReturnStmt
	escapes []ast.Stmt
	assigns []ast.Stmt
	results map[*types.Var]ast.Expr
}

func scanClause(pass *analysis.Pass, clause *ast.CaseClause, signature *types.Signature) *clauseScan {
	found := &clauseScan{results: map[*types.Var]ast.Expr{}}
	found.walkList(pass, clause.Body, clause, namedResults(signature), false)

	return found
}

func namedResults(signature *types.Signature) map[*types.Var]bool {
	named := map[*types.Var]bool{}
	results := signature.Results()
	for i := range results.Len() {
		if results.At(i).Name() != "" {
			named[results.At(i)] = true
		}
	}

	return named
}

// walk descends the statements of the clause, leaving the expressions alone, so a closure written
// inside it is answered by the walk that closure gets of its own.
func (s *clauseScan) walk(pass *analysis.Pass, stmt ast.Stmt, clause *ast.CaseClause, named map[*types.Var]bool, breakable bool) {
	switch node := stmt.(type) {
	case *ast.ReturnStmt:
		s.returns = append(s.returns, node)
	case *ast.BranchStmt:
		s.branch(node, breakable)
	case *ast.AssignStmt:
		s.assign(pass, node, clause, named)
	case *ast.IncDecStmt:
		s.target(pass, node, node.X, clause)
	case *ast.BlockStmt:
		s.walkList(pass, node.List, clause, named, breakable)
	case *ast.LabeledStmt:
		s.walk(pass, node.Stmt, clause, named, breakable)
	case *ast.IfStmt:
		s.walk(pass, node.Body, clause, named, breakable)
		if node.Else != nil {
			s.walk(pass, node.Else, clause, named, breakable)
		}
	case *ast.ForStmt:
		s.walk(pass, node.Body, clause, named, true)
	case *ast.RangeStmt:
		s.walk(pass, node.Body, clause, named, true)
	case *ast.SwitchStmt:
		s.walkList(pass, node.Body.List, clause, named, true)
	case *ast.TypeSwitchStmt:
		s.walkList(pass, node.Body.List, clause, named, true)
	case *ast.SelectStmt:
		s.walkList(pass, node.Body.List, clause, named, true)
	case *ast.CaseClause:
		s.walkList(pass, node.Body, clause, named, breakable)
	case *ast.CommClause:
		s.walkList(pass, node.Body, clause, named, breakable)
	}
}

// quiet reports whether the clause has been left to say something for itself, since a clause
// already reported for what it hands back does not need telling that it hands nothing back.
func (s *clauseScan) quiet() bool {
	return len(s.escapes) == 0 && len(s.assigns) == 0
}

func (s *clauseScan) walkList(pass *analysis.Pass, list []ast.Stmt, clause *ast.CaseClause, named map[*types.Var]bool, breakable bool) {
	for _, stmt := range list {
		s.walk(pass, stmt, clause, named, breakable)
	}
}

// branch keeps the break of a loop written inside the clause, since that one lands in the clause
// rather than past the switch.
func (s *clauseScan) branch(node *ast.BranchStmt, breakable bool) {
	if node.Label == nil && breakable {
		return
	}

	s.escapes = append(s.escapes, node)
}

func (s *clauseScan) assign(pass *analysis.Pass, node *ast.AssignStmt, clause *ast.CaseClause, named map[*types.Var]bool) {
	for i, target := range node.Lhs {
		if result, ok := s.namedResult(pass, target, named); ok {
			s.results[result] = valueAt(node, i)
			continue
		}
		s.target(pass, node, target, clause)
	}
}

// valueAt pairs an assignment's target with the expression given to it, and gives up on a call
// spread over several targets, where the clause is answering with something unreadable anyway.
func valueAt(node *ast.AssignStmt, i int) ast.Expr {
	if len(node.Rhs) != len(node.Lhs) {
		return nil
	}

	return node.Rhs[i]
}

func (s *clauseScan) namedResult(pass *analysis.Pass, target ast.Expr, named map[*types.Var]bool) (*types.Var, bool) {
	ident, ok := target.(*ast.Ident)
	if !ok {
		return nil, false
	}

	variable, ok := pass.TypesInfo.Uses[ident].(*types.Var)
	if !ok || !named[variable] {
		return nil, false
	}

	return variable, true
}

// target reports an assignment reaching outside the clause, which is the shape where a switch
// sets a variable the code after it reads.
func (s *clauseScan) target(pass *analysis.Pass, stmt ast.Stmt, target ast.Expr, clause *ast.CaseClause) {
	ident := rootIdent(target)
	if ident == nil || ident.Name == "_" {
		return
	}

	object := pass.TypesInfo.ObjectOf(ident)
	if object == nil || (clause.Pos() < object.Pos() && object.Pos() < clause.End()) {
		return
	}

	s.assigns = append(s.assigns, stmt)
}

// rootIdent digs out the variable an assignment target hangs off, so a field or an element set on
// something declared outside the clause is read as reaching outside it too.
func rootIdent(target ast.Expr) *ast.Ident {
	for {
		switch expr := target.(type) {
		case *ast.Ident:
			return expr
		case *ast.SelectorExpr:
			target = expr.X
		case *ast.IndexExpr:
			target = expr.X
		case *ast.StarExpr:
			target = expr.X
		case *ast.ParenExpr:
			target = expr.X
		default:
			return nil
		}
	}
}

// terminates reports whether the clause ends somewhere the code after the switch cannot be
// reached from, which is a return, a panic, or a call that ends the process.
func terminates(stmt ast.Stmt) bool {
	switch node := stmt.(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.ExprStmt:
		return isFatalCall(node.X)
	case *ast.LabeledStmt:
		return terminates(node.Stmt)
	case *ast.BlockStmt:
		return len(node.List) > 0 && terminates(node.List[len(node.List)-1])
	case *ast.IfStmt:
		return node.Else != nil && terminates(node.Body) && terminates(node.Else)
	case *ast.ForStmt:
		return node.Cond == nil
	}

	return false
}

func isFatalCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}

	switch fun := call.Fun.(type) {
	case *ast.Ident:
		return fun.Name == "panic"
	case *ast.SelectorExpr:
		return strings.HasPrefix(fun.Sel.Name, "Fatal") || fun.Sel.Name == "Exit"
	}

	return false
}

// isEnumSwitch holds the rule to a switch over a named type with constants of its own, which is
// what an enum is in go, rather than to a switch over a plain string or a comparison.
func isEnumSwitch(pass *analysis.Pass, tag ast.Expr) bool {
	if tag == nil {
		return false
	}

	named, ok := pass.TypesInfo.TypeOf(tag).(*types.Named)
	if !ok || named.Obj().Pkg() == nil {
		return false
	}

	if basic, ok := named.Underlying().(*types.Basic); !ok || basic.Info()&(types.IsInteger|types.IsString) == 0 {
		return false
	}

	return countConstants(named) > 1
}

func countConstants(named *types.Named) int {
	scope := named.Obj().Pkg().Scope()
	count := 0
	for _, name := range scope.Names() {
		declared, ok := scope.Lookup(name).(*types.Const)
		if ok && types.Identical(declared.Type(), named) {
			count++
		}
	}

	return count
}

// isZero reports whether an expression is the zero of its type, which folding the constant
// answers for a number, a string, a boolean and a constant of the enum's own type alike.
func isZero(pass *analysis.Pass, expr ast.Expr) bool {
	if isNil(pass, expr) {
		return true
	}

	if value := pass.TypesInfo.Types[expr].Value; value != nil {
		switch value.Kind() {
		case constant.Bool:
			return !constant.BoolVal(value)
		case constant.String:
			return constant.StringVal(value) == ""
		case constant.Int, constant.Float, constant.Complex:
			return constant.Sign(value) == 0
		case constant.Unknown:
			return false
		}
	}

	if lit, ok := expr.(*ast.CompositeLit); ok {
		return len(lit.Elts) == 0
	}

	return false
}

func isNil(pass *analysis.Pass, expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}

	_, ok = pass.TypesInfo.Uses[ident].(*types.Nil)

	return ok
}

func isError(typ types.Type) bool {
	named, ok := typ.(*types.Named)
	return ok && named.Obj().Pkg() == nil && named.Obj().Name() == "error"
}

func signatureOf(pass *analysis.Pass, node ast.Expr) *types.Signature {
	if ident, ok := node.(*ast.Ident); ok {
		if object := pass.TypesInfo.Defs[ident]; object != nil {
			signature, _ := object.Type().(*types.Signature)
			return signature
		}
	}

	signature, _ := pass.TypesInfo.TypeOf(node).(*types.Signature)

	return signature
}

func compileExcludes(patterns []string) ([]*regexp.Regexp, error) {
	var compiled []*regexp.Regexp
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("switch_default: exclude pattern %q: %w", pattern, err)
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
