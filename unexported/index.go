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

// siteSet is where an identifier is used from, which is the whole of what the verdict turns on.
type siteSet struct {
	internal     bool
	ownTest      bool
	externalTest bool
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
	keptByExternalTest int
	skippedGenerated   int
}

func newIndex(opts Options, loaded *loaded) *index {
	return &index{
		opts:     opts,
		fset:     loaded.pkgs[0].Fset,
		pkgs:     loaded.pkgs,
		decls:    map[string]*declaration{},
		sites:    map[string]*siteSet{},
		exported: map[string][]types.Object{},
		dirs:     map[string]*packages.Package{},
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

		for ident, obj := range p.TypesInfo.Uses {
			idx.recordUse(p.Types.Path(), obj, ident.Pos())
		}
	}
}

// recordUse files one use under the package path and name of what it points at, which is the
// only key a declaration and a use of it reliably share.
func (idx *index) recordUse(usingPath string, obj types.Object, pos token.Pos) {
	if obj == nil || obj.Pkg() == nil || obj.Parent() != obj.Pkg().Scope() {
		return
	}

	declPath := obj.Pkg().Path()
	sites, ok := idx.sites[objectKey(declPath, obj.Name())]
	if !ok {
		return
	}

	inTestFile := strings.HasSuffix(idx.fset.Position(pos).Filename, "_test.go")
	sites.add(usingPath, declPath, inTestFile)
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
	scan := &textScan{
		idx:     idx,
		owner:   owner,
		imports: idx.importsOf(file),
		beside:  owner != nil && owner.Name == file.Name.Name,
		inTest:  strings.HasSuffix(path, "_test.go"),
	}
	ast.Inspect(file, scan.visit)
}

// textScan reads one untypechecked file for the identifiers it names, since a build tag hiding a
// caller from the typechecker must not hide it from the sweep.
type textScan struct {
	idx     *index
	owner   *packages.Package
	imports map[string]string
	beside  bool
	inTest  bool
}

// visit takes a qualified name as a use of the package it names, and a bare one as a use of the
// package the file itself belongs to, which is only true where the file is another build of it.
func (t *textScan) visit(node ast.Node) bool {
	switch expr := node.(type) {
	case *ast.SelectorExpr:
		if ident, ok := expr.X.(*ast.Ident); ok {
			if imported, found := t.imports[ident.Name]; found {
				t.idx.recordTextUse(imported, expr.Sel.Name, textUsePath(t.owner, imported), false)
				return false
			}
		}
		ast.Inspect(expr.X, t.visit)

		return false
	case *ast.Ident:
		if t.beside {
			t.idx.recordTextUse(t.owner.PkgPath, expr.Name, t.owner.PkgPath, t.inTest)
		}
	}

	return true
}

// textUsePath decides which package a qualified reference in an untypechecked file counts as
// coming from, which is the package's own directory for a build-tagged file sitting beside it
// and somewhere else entirely otherwise.
func textUsePath(owner *packages.Package, imported string) string {
	if owner != nil && owner.PkgPath == imported {
		return imported + "_test"
	}

	return ""
}

func (idx *index) recordTextUse(declPath, name, usingPath string, inTestFile bool) {
	if sites, ok := idx.sites[objectKey(declPath, name)]; ok {
		sites.add(usingPath, declPath, inTestFile)
	}
}

// importsOf maps the name a file refers to each import by onto the path it stands for, taking
// the name from the loaded package where there is one so a package named unlike its directory
// still resolves.
func (idx *index) importsOf(file *ast.File) map[string]string {
	imports := map[string]string{}
	for _, spec := range file.Imports {
		path := strings.Trim(spec.Path.Value, `"`)
		name := ""
		switch {
		case spec.Name != nil:
			name = spec.Name.Name
		case idx.packageName(path) != "":
			name = idx.packageName(path)
		default:
			name = path[strings.LastIndex(path, "/")+1:]
		}
		if name != "" && name != "_" && name != "." {
			imports[name] = path
		}
	}

	return imports
}

func (idx *index) packageName(path string) string {
	for _, p := range idx.pkgs {
		if p.PkgPath == path {
			return p.Name
		}
	}

	return ""
}

// decide turns each declaration's use sites into the one thing a person should do about it.
func (idx *index) decide() []Finding {
	verdicts := map[string]Verdict{}
	reasons := map[string]string{}

	for key := range idx.decls {
		sites := idx.sites[key]
		switch {
		case sites.external:
		case sites.externalTest:
			idx.keptByExternalTest++
		case sites.internal:
			verdicts[key] = VerdictUnexport
			reasons[key] = "only its own package uses it"
		case sites.ownTest:
			verdicts[key] = VerdictDelete
			reasons[key] = "only its own tests use it"
		default:
			verdicts[key] = VerdictDelete
			reasons[key] = "nothing uses it"
		}
	}

	idx.keepTheExportedSurfaceWhole(verdicts)

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

// keepTheExportedSurfaceWhole drops the verdict on anything the rest of the package's exported
// surface still needs, since advice that breaks the surface is worse than no advice.
func (idx *index) keepTheExportedSurfaceWhole(verdicts map[string]Verdict) {
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
