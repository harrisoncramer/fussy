package unexported

import (
	"go/ast"
	"go/token"
	"sort"

	"golang.org/x/tools/go/packages"
)

// extentsOf lists where each of a package's tracked declarations begins and ends, in the order a
// search can walk, so a use can be attributed to whichever declaration it was written inside.
func (idx *index) extentsOf(p *packages.Package) []extent {
	// Explainer: this has to be built per package variant, inside the loop that reads the uses,
	// and must not be hoisted into a cache keyed by package path. The variant compiled with a
	// package's test files re-parses its non-test files, which gives them a second range in the
	// FileSet and so a second set of positions. Extents and positions agree only while they come
	// from the same parse, so a cache shared between the two variants would quietly stop every
	// use in a package that has tests from finding the declaration it was written inside.

	var extents []extent
	for _, file := range p.Syntax {
		for _, decl := range file.Decls {
			extents = append(extents, idx.extentsOfDecl(p.PkgPath, decl)...)
		}
	}

	sort.Slice(extents, func(i, j int) bool { return extents[i].start < extents[j].start })

	return extents
}

// extentsOfDecl covers one top-level declaration, giving a var or const block one extent per spec
// rather than one for the whole block, since a block mixing an exported name with an unexported
// one would otherwise credit the unexported name's initialiser to the exported name and let a
// dead export take a live caller down with it.
func (idx *index) extentsOfDecl(pkgPath string, decl ast.Decl) []extent {
	if fn, ok := decl.(*ast.FuncDecl); ok {
		return idx.extentFor(pkgPath, fn.Pos(), fn.End(), declaredName(fn))
	}

	node, ok := decl.(*ast.GenDecl)
	if !ok {
		return nil
	}

	var extents []extent
	for _, spec := range node.Specs {
		extents = append(extents, idx.extentForSpec(pkgPath, spec)...)
	}

	return extents
}

// extentForSpec covers one spec of a var, const or type block, and covers nothing at all where
// any name it introduces is one the sweep is not following, since a use written there is a use
// the sweep can never prove dead.
func (idx *index) extentForSpec(pkgPath string, spec ast.Spec) []extent {
	switch node := spec.(type) {
	case *ast.TypeSpec:
		return idx.extentFor(pkgPath, node.Pos(), node.End(), node.Name.Name)
	case *ast.ValueSpec:
		names := make([]string, 0, len(node.Names))
		for _, name := range node.Names {
			names = append(names, name.Name)
		}

		return idx.extentFor(pkgPath, node.Pos(), node.End(), names...)
	}

	return nil
}

// extentFor records the span only where every name it declares is tracked, since one untracked
// name among them means the sweep cannot say who the uses inside really belong to.
func (idx *index) extentFor(pkgPath string, start, end token.Pos, names ...string) []extent {
	if len(names) == 0 {
		return nil
	}

	keys := make([]string, 0, len(names))
	for _, name := range names {
		key := objectKey(pkgPath, name)
		if _, found := idx.decls[key]; !found {
			return nil
		}
		keys = append(keys, key)
	}

	return []extent{{start: start, end: end, keys: keys}}
}

// declaredName is what a function declaration stands for, which is the receiver type for a
// method, since a type whose only mentions are in its own methods is as dead as one nothing
// mentions at all.
func declaredName(fn *ast.FuncDecl) string {
	if fn.Recv == nil {
		return fn.Name.Name
	}

	return receiverName(fn.Recv)
}

// receiverName digs the type name out of a method's receiver, past the pointer and the type
// parameters a generic receiver carries.
func receiverName(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) != 1 {
		return ""
	}

	expr := recv.List[0].Type
	for {
		switch node := expr.(type) {
		case *ast.StarExpr:
			expr = node.X
		case *ast.ParenExpr:
			expr = node.X
		case *ast.IndexExpr:
			expr = node.X
		case *ast.IndexListExpr:
			expr = node.X
		case *ast.Ident:
			return node.Name
		default:
			return ""
		}
	}
}

// enclosing names the tracked declarations a position was written inside, which is nothing at all
// for a position in an unexported function, an init, a test or a generated file.
func enclosing(extents []extent, pos token.Pos) []string {
	at := sort.Search(len(extents), func(i int) bool { return extents[i].start > pos })
	if at == 0 {
		return nil
	}

	found := extents[at-1]
	if pos >= found.end {
		return nil
	}

	return found.keys
}

// reachable spreads liveness out from the declarations something outside the sweep's own
// candidates keeps alive, which is what separates a declaration its package really uses from a
// run of dead declarations that only ever mention each other, and takes the declarations the
// exported-surface guard has rescued as seeds of their own alongside them.
func (idx *index) reachable(seeded func(*siteSet) bool, rescued map[string]bool) map[string]bool {
	alive := map[string]bool{}

	var queue []string
	add := func(key string) {
		if !alive[key] {
			alive[key] = true
			queue = append(queue, key)
		}
	}

	for key, sites := range idx.sites {
		if seeded(sites) {
			add(key)
		}
	}

	for key := range rescued {
		add(key)
	}

	for len(queue) > 0 {
		from := queue[len(queue)-1]
		queue = queue[:len(queue)-1]

		for to := range idx.edges[from] {
			if !alive[to] {
				alive[to] = true
				queue = append(queue, to)
			}
		}
	}

	return alive
}

// keptAlive is what seeds liveness: a use from another package, or one no tracked declaration
// encloses, since the sweep can never call either of those dead.
func keptAlive(sites *siteSet) bool {
	return sites.external || sites.externalTest || sites.foreignTest || sites.internal
}

// keptAliveByTests is what seeds the weaker liveness a package's own tests give, which is a
// verdict of its own rather than a reason to leave something exported.
func keptAliveByTests(sites *siteSet) bool {
	return sites.ownTest
}
