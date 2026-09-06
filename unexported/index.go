package unexported

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

// declaration is one exported package-level identifier, kept beside where a person would go to
// change it.
type declaration struct {
	pkgPath string
	name    string
	kind    string
	reach   string
	file    string
	line    int
}

// siteSet is where an identifier is used from, with the internal flag covering only the uses no
// tracked declaration encloses, which are the ones nothing in this sweep can ever call dead, such
// as a call from an unexported function, an init, a generated file or a file behind a build tag,
// since a same-package use written inside another tracked declaration is filed as an edge instead
// so that a run of declarations only ever calling each other is not read as a package using them.
type siteSet struct {
	internal     bool
	ownTest      bool
	externalTest bool
	foreignTest  bool
	external     bool
}

// index gathers what is declared and what is used across every package of the sweep, keyed by
// package path and name so a declaration and a use of it match even though the two are separate
// types.Object values in separate variants of the same package.
type index struct {
	opts               Options
	fset               *token.FileSet
	pkgs               []*packages.Package
	decls              map[string]*declaration
	sites              map[string]*siteSet
	exported           map[string][]types.Object
	dirs               map[string]*packages.Package
	names              map[string]string
	edges              map[string]map[string]bool
	keptByExternalTest int
	keptByForeignTest  int
	skippedGenerated   int
}

// extent is where one tracked declaration begins and ends so a use written inside it can be
// attributed to it, with a method filed under its receiver type, since a type whose only mentions
// are in its own methods is as dead as one nothing mentions at all.
type extent struct {
	start token.Pos
	end   token.Pos
	keys  []string
}

func newIndex(opts Options, loaded *loaded) *index {
	names := map[string]string{}
	for _, p := range loaded.pkgs {
		names[p.PkgPath] = p.Name
	}

	return &index{
		opts:     opts,
		fset:     loaded.pkgs[0].Fset,
		pkgs:     loaded.pkgs,
		decls:    map[string]*declaration{},
		sites:    map[string]*siteSet{},
		exported: map[string][]types.Object{},
		dirs:     map[string]*packages.Package{},
		names:    names,
		edges:    map[string]map[string]bool{},
	}
}

func objectKey(pkgPath, name string) string {
	return pkgPath + "." + name
}

// readDeclarations records every exported package-level identifier of the packages the patterns
// matched, taking them from the variant compiled without the test files so a helper declared in
// a test is never mistaken for part of the package's surface.
func (idx *index) readDeclarations() {
	for _, p := range idx.pkgs {
		if !isDeclarationSource(p) {
			continue
		}

		idx.dirs[packageDir(p)] = p
		reach := reachOf(p)
		generated := generatedFiles(p)
		scope := p.Types.Scope()

		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			if !obj.Exported() {
				continue
			}

			kind, ok := kindOf(obj)
			if !ok {
				continue
			}

			idx.exported[p.PkgPath] = append(idx.exported[p.PkgPath], obj)

			position := idx.fset.Position(obj.Pos())
			if strings.HasSuffix(position.Filename, "_test.go") {
				continue
			}

			if generated[position.Filename] && !idx.opts.IncludeGenerated {
				idx.skippedGenerated++
				continue
			}

			key := objectKey(p.PkgPath, name)
			idx.decls[key] = &declaration{
				pkgPath: p.PkgPath,
				name:    name,
				kind:    kind,
				reach:   reach,
				file:    position.Filename,
				line:    position.Line,
			}
			idx.sites[key] = &siteSet{}
		}
	}
}

// isDeclarationSource picks the plain variant of a package, leaving out the variants compiled
// with test files and the main package the go tool synthesises to run them.
func isDeclarationSource(p *packages.Package) bool {
	if p.Types == nil || p.ID != p.PkgPath {
		return false
	}

	return !strings.HasSuffix(p.PkgPath, ".test") && !strings.HasSuffix(p.PkgPath, "_test")
}

// kindOf names the sort of declaration an object is, and refuses methods, which unexporting can
// break interface satisfaction with, and struct fields, which are not package-level at all.
func kindOf(obj types.Object) (string, bool) {
	if obj.Pkg() == nil || obj.Parent() != obj.Pkg().Scope() {
		return "", false
	}

	switch obj.(type) {
	case *types.Func:
		return KindFunc, true
	case *types.TypeName:
		return KindType, true
	case *types.Var:
		return KindVar, true
	case *types.Const:
		return KindConst, true
	}

	return "", false
}

// reachOf says whether anything outside this module could have imported the package, which is
// what separates a verdict the sweep can stand behind from one that only covers this repository.
func reachOf(p *packages.Package) string {
	if p.Name == "main" {
		return ReachModulePrivate
	}

	if slices.Contains(strings.Split(p.PkgPath, "/"), "internal") {
		return ReachModulePrivate
	}

	return ReachImportable
}

func generatedFiles(p *packages.Package) map[string]bool {
	generated := map[string]bool{}
	for _, file := range p.Syntax {
		if ast.IsGenerated(file) {
			generated[p.Fset.Position(file.Pos()).Filename] = true
		}
	}

	return generated
}

func packageDir(p *packages.Package) string {
	if len(p.GoFiles) > 0 {
		return filepath.Dir(p.GoFiles[0])
	}

	if len(p.IgnoredFiles) > 0 {
		return filepath.Dir(p.IgnoredFiles[0])
	}

	return ""
}

// readUses walks every variant of every package the patterns matched, since the variant compiled
// with a package's own test files is the only place an in-package test call is visible.
func (idx *index) readUses() {
	for _, p := range idx.pkgs {
		if p.TypesInfo == nil || p.Types == nil || strings.HasSuffix(p.PkgPath, ".test") {
			continue
		}

		extents := idx.extentsOf(p)
		for ident, obj := range p.TypesInfo.Uses {
			idx.recordUse(p.Types.Path(), obj, ident.Pos(), enclosing(extents, ident.Pos()))
		}
	}
}

// recordUse files one use under the package path and name of what it points at, which is the only
// key a declaration and a use of it reliably share, and files a same-package use written inside
// another tracked declaration as an edge between the two instead.
func (idx *index) recordUse(usingPath string, obj types.Object, pos token.Pos, from []string) {
	if obj == nil || obj.Pkg() == nil || obj.Parent() != obj.Pkg().Scope() {
		return
	}

	declPath := obj.Pkg().Path()
	key := objectKey(declPath, obj.Name())
	sites, ok := idx.sites[key]
	if !ok {
		return
	}

	inTestFile := strings.HasSuffix(idx.fset.Position(pos).Filename, "_test.go")
	if usingPath == declPath && !inTestFile && len(from) > 0 {
		for _, source := range from {
			idx.edge(source, key)
		}

		return
	}

	sites.add(usingPath, declPath, inTestFile)
}

func (idx *index) edge(from, to string) {
	if idx.edges[from] == nil {
		idx.edges[from] = map[string]bool{}
	}
	idx.edges[from][to] = true
}

// add files one use in the bucket its verdict turns on, which separates a package's own test
// files from an external test package because unexporting is safe for one and not the other.
func (s *siteSet) add(usingPath, declPath string, inTestFile bool) {
	switch {
	case usingPath == declPath && inTestFile:
		s.ownTest = true
	case usingPath == declPath:
		s.internal = true
	case usingPath == declPath+"_test":
		s.externalTest = true
	case inTestFile:
		s.foreignTest = true
	default:
		s.external = true
	}
}

// readUnloadedFiles reads the files no build configuration in this sweep compiled, so a caller
// behind a build tag counts as a caller rather than being silently absent.
func (idx *index) readUnloadedFiles(paths []string) {
	fset := token.NewFileSet()
	for _, path := range paths {
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			continue
		}
		idx.readUnloadedFile(path, file)
	}
}

// readUnloadedFile takes what one untypechecked file names, which is a guess made only in the
// direction of suppressing a finding rather than of inventing one.
func (idx *index) readUnloadedFile(path string, file *ast.File) {
	owner := idx.dirs[filepath.Dir(path)]
	imports, dots := idx.importsOf(file)
	scan := &textScan{
		idx:         idx,
		owner:       owner,
		imports:     imports,
		dots:        dots,
		packageName: file.Name.Name,
		beside:      owner != nil && owner.Name == file.Name.Name,
		inTest:      strings.HasSuffix(path, "_test.go"),
	}
	ast.Inspect(file, scan.visit)
}

// textScan reads one untypechecked file for the identifiers it names, since a build tag hiding a
// caller from the typechecker must not hide it from the sweep.
type textScan struct {
	idx         *index
	owner       *packages.Package
	imports     map[string]string
	dots        []string
	packageName string
	beside      bool
	inTest      bool
}

// visit takes a qualified name as a use of the package it names, and a bare one as a use of the
// package the file itself belongs to, which is only true where the file is another build of it.
func (t *textScan) visit(node ast.Node) bool {
	switch expr := node.(type) {
	case *ast.SelectorExpr:
		if ident, ok := expr.X.(*ast.Ident); ok {
			if imported, found := t.imports[ident.Name]; found {
				t.idx.recordTextUse(imported, expr.Sel.Name, t.qualifiedUsePath(imported), false)
				return false
			}
		}
		ast.Inspect(expr.X, t.visit)

		return false
	case *ast.Ident:
		if t.beside {
			t.idx.recordTextUse(t.owner.PkgPath, expr.Name, t.owner.PkgPath, t.inTest)
		}
		for _, dotted := range t.dots {
			t.idx.recordTextUse(dotted, expr.Name, "", false)
		}
	}

	return true
}

// qualifiedUsePath decides which package a qualified reference in an untypechecked file counts as
// coming from, which is the package's own black-box test only where the file really declares that
// test package, since a go:build ignore program sitting in the same directory is another package
// entirely and counting it as a test would say the wrong thing in the summary.
func (t *textScan) qualifiedUsePath(imported string) string {
	if t.owner != nil && t.owner.PkgPath == imported && t.packageName == t.owner.Name+"_test" {
		return imported + "_test"
	}

	return ""
}

func (idx *index) recordTextUse(declPath, name, usingPath string, inTestFile bool) {
	if sites, ok := idx.sites[objectKey(declPath, name)]; ok {
		sites.add(usingPath, declPath, inTestFile)
	}
}

// importsOf maps the name a file refers to each import by onto the path it stands for, taking the
// name from the loaded package where there is one so a package named unlike its directory still
// resolves, and gathers the dot imports separately since those put names in scope with nothing in
// front of them.
func (idx *index) importsOf(file *ast.File) (map[string]string, []string) {
	imports := map[string]string{}
	var dots []string
	for _, spec := range file.Imports {
		path := strings.Trim(spec.Path.Value, `"`)
		name := idx.names[path]
		if spec.Name != nil {
			name = spec.Name.Name
		}
		if name == "" {
			name = path[strings.LastIndex(path, "/")+1:]
		}
		if name == "." {
			dots = append(dots, path)
			continue
		}
		if name != "" && name != "_" {
			imports[name] = path
		}
	}

	return imports, dots
}

// decide turns each declaration's use sites into the one thing a person should do about it,
// reading liveness through the edges rather than off a single flag so a run of declarations that
// only ever mention each other is called dead rather than called used.
func (idx *index) decide() []Finding {
	// Explainer: the liveness and the exported-surface guard have to settle together rather
	// than run in sequence. The guard rescues a declaration by taking its verdict away, which
	// leaves it holding no verdict and no liveness, so on its own it is a dead end: the
	// declarations it names go on being condemned even though something the package still
	// shows a caller reaches them. Feeding each rescue back in as a liveness seed and drawing
	// the verdicts again is what closes that. Both halves only move one way, the guard removing
	// verdicts and the spread adding liveness, so the pair converges.
	idx.countKept()

	testAlive := idx.reachable(keptAliveByTests, nil)
	rescued := map[string]bool{}

	var verdicts map[string]Verdict
	var reasons map[string]string

	for {
		alive := idx.reachable(keptAlive, rescued)
		verdicts, reasons = idx.draw(alive, testAlive)
		idx.keepTheExportedSurfaceWhole(verdicts)

		if !idx.growRescued(rescued, verdicts, alive) {
			break
		}
	}

	kinds := map[string]bool{}
	for _, kind := range idx.opts.Kinds {
		kinds[kind] = true
	}

	var findings []Finding
	for key, verdict := range verdicts {
		decl := idx.decls[key]
		if !kinds[decl.kind] {
			continue
		}
		findings = append(findings, Finding{
			Verdict: verdict,
			Package: decl.pkgPath,
			Name:    decl.name,
			Kind:    decl.kind,
			Reach:   decl.reach,
			File:    relative(idx.opts.Dir, decl.file),
			Line:    decl.line,
			Reason:  reasons[key],
		})
	}

	return findings
}

// countKept tallies the exports left alone because a test needs them, once rather than once per
// pass, since what keeps a declaration is a fact about its use sites and never moves.
func (idx *index) countKept() {
	for key := range idx.decls {
		sites := idx.sites[key]
		switch {
		case sites.external:
		case sites.externalTest:
			idx.keptByExternalTest++
		case sites.foreignTest:
			idx.keptByForeignTest++
		}
	}
}

// draw reads one settled liveness into a verdict per declaration.
func (idx *index) draw(alive, testAlive map[string]bool) (map[string]Verdict, map[string]string) {
	verdicts := map[string]Verdict{}
	reasons := map[string]string{}

	for key := range idx.decls {
		sites := idx.sites[key]
		switch {
		case sites.external, sites.externalTest, sites.foreignTest:
		case alive[key]:
			verdicts[key] = VerdictUnexport
			reasons[key] = "only its own package uses it"
		case testAlive[key]:
			verdicts[key] = VerdictDelete
			reasons[key] = "only its own tests use it"
		default:
			verdicts[key] = VerdictDelete
			reasons[key] = "nothing uses it"
		}
	}

	return verdicts, reasons
}

// growRescued takes the declarations the guard just rescued as liveness seeds for the next pass,
// which are the ones left holding neither a verdict nor any liveness of their own.
func (idx *index) growRescued(rescued map[string]bool, verdicts map[string]Verdict, alive map[string]bool) bool {
	added := false
	for key := range idx.decls {
		if alive[key] || rescued[key] {
			continue
		}
		if _, condemned := verdicts[key]; condemned {
			continue
		}
		rescued[key] = true
		added = true
	}

	return added
}

// keepTheExportedSurfaceWhole drops the verdict on anything the rest of the package's exported
// surface still needs, since advice that breaks the surface is worse than no advice.
func (idx *index) keepTheExportedSurfaceWhole(verdicts map[string]Verdict) {
	// Explainer: the loop below is load-bearing rather than a convenience for cascades. A
	// constant asks whether its type is condemned by reading the verdicts as they stand, so
	// within one pass a constant visited before its type gets a stale answer. Running until
	// nothing more is dropped is what settles that, and collapsing this to a single pass would
	// quietly turn the members of a rescued enum back into findings.
	for pkgPath, objects := range idx.exported {
		for {
			reachable := reachableFrom(idx.retained(pkgPath, objects, verdicts))
			dropped := false
			for _, obj := range objects {
				key := objectKey(pkgPath, obj.Name())
				if _, found := verdicts[key]; !found {
					continue
				}
				if !isNamedByTheSurface(obj, reachable, verdicts, pkgPath) {
					continue
				}
				delete(verdicts, key)
				dropped = true
			}
			if !dropped {
				break
			}
		}
	}
}

// isNamedByTheSurface reports whether unexporting one declaration would break what the package
// still shows a caller, which is either a type an exported signature names or a constant of an
// exported type, since an enum a caller can switch over needs every member it has.
func isNamedByTheSurface(obj types.Object, reachable map[*types.TypeName]bool, verdicts map[string]Verdict, pkgPath string) bool {
	if name, ok := obj.(*types.TypeName); ok {
		return reachable[name]
	}

	if _, ok := obj.(*types.Const); !ok {
		return false
	}

	named, ok := types.Unalias(obj.Type()).(*types.Named)
	if !ok || named.Obj().Pkg() != obj.Pkg() || !named.Obj().Exported() {
		return false
	}

	_, condemned := verdicts[objectKey(pkgPath, named.Obj().Name())]

	return !condemned
}

// retained is the exported surface the sweep is leaving in place, which is everything the module
// still uses beside everything the sweep never looked at.
func (idx *index) retained(pkgPath string, objects []types.Object, verdicts map[string]Verdict) []types.Object {
	var kept []types.Object
	for _, obj := range objects {
		if _, found := verdicts[objectKey(pkgPath, obj.Name())]; !found {
			kept = append(kept, obj)
		}
	}

	return kept
}
